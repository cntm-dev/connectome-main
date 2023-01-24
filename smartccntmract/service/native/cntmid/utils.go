/*
 * Copyright (C) 2018 The cntmology Authors
 * This file is part of The cntmology library.
 *
 * The cntmology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or * (at your option) any later version.
 *
 * The cntmology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * alcntm with The cntmology.  If not, see <http://www.gnu.org/licenses/>.
 */
package cntmid

import (
	"encoding/hex"
	"errors"

	"github.com/cntmio/cntmology-crypto/keypair"
	"github.com/cntmio/cntmology/core/states"
	"github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/smartccntmract/service/native"
	"github.com/cntmio/cntmology/smartccntmract/service/native/utils"
)

const flag_exist = 0x01

func checkIDExistence(srvc *native.NativeService, encID []byte) bool {
	val, err := srvc.CacheDB.Get(encID)
	if err == nil {
		val, err := states.GetValueFromRawStorageItem(val)
		if err == nil {
			if len(val) > 0 && val[0] == flag_exist {
				return true
			}
		}
	}
	return false
}

const (
	FIELD_PK byte = 1 + iota
	FIELD_ATTR
	FIELD_RECOVERY
)

func encodeID(id []byte) ([]byte, error) {
	length := len(id)
	if length == 0 || length > 255 {
		return nil, errors.New("encode cntm ID error: invalid ID length")
	}
	enc := []byte{byte(length)}
	enc = append(enc, id...)
	return enc, nil
}

func decodeID(data []byte) ([]byte, error) {
	if len(data) == 0 || len(data) != int(data[0])+1 {
		return nil, errors.New("decode cntm ID error: invalid data length")
	}
	return data[1:], nil
}

func setRecovery(srvc *native.NativeService, encID []byte, recovery com.Address) error {
	key := append(encID, FIELD_RECOVERY)
	val := states.StorageItem{Value: recovery[:]}
	srvc.CacheDB.Put(key, val.ToArray())
	return nil
}

func getRecovery(srvc *native.NativeService, encID []byte) ([]byte, error) {
	key := append(encID, FIELD_RECOVERY)
	item, err := utils.GetStorageItem(srvc, key)
	if err != nil {
		return nil, errors.New("get recovery error: " + err.Error())
	} else if item == nil {
		return nil, nil
	}
	return item.Value, nil
}

func checkWitness(srvc *native.NativeService, key []byte) error {
	// try as if key is a public key
	pk, err := keypair.DeserializePublicKey(key)
	if err == nil {
		addr := types.AddressFromPubKey(pk)
		if srvc.CcntmextRef.CheckWitness(addr) {
			return nil
		}
	}

	// try as if key is an address
	addr, err := common.AddressParseFromBytes(key)
	if err == nil && srvc.CcntmextRef.CheckWitness(addr) {
		return nil
	}

	return errors.New("check witness failed, " + hex.EncodeToString(key))
}

func checkWitnessByIndex(srvc *native.NativeService, encId []byte, index uint32) error {
	pk, err := getPk(srvc, encId, index)
	if err != nil {
		return err
	} else if pk.revoked {
		return errors.New("revoked key")
	}

	//verify access
	if !pk.isAuthentication {
		return fmt.Errorf("pk do not have access")
	}

	return checkWitness(srvc, pk.key)
}

func checkWitnessWithoutAuth(srvc *native.NativeService, encId []byte, index uint32) error {
	pk, err := getPk(srvc, encId, index)
	if err != nil {
		return err
	} else if pk.revoked {
		return errors.New("revoked key")
	}

	return checkWitness(srvc, pk.key)
}

func deleteID(srvc *native.NativeService, encId []byte) error {
	key := append(encId, FIELD_PK)
	srvc.CacheDB.Delete(key)

	key = append(encId, FIELD_CcntmROLLER)
	srvc.CacheDB.Delete(key)

	key = append(encId, FIELD_RECOVERY)
	srvc.CacheDB.Delete(key)
	if srvc.Height >= config.GetNewOntIdHeight() {
		key = append(encId, FIELD_SERVICE)
		srvc.CacheDB.Delete(key)

		key = append(encId, FIELD_CREATED)
		srvc.CacheDB.Delete(key)

		key = append(encId, FIELD_UPDATED)
		srvc.CacheDB.Delete(key)

		key = append(encId, FIELD_PROOF)
		srvc.CacheDB.Delete(key)

		key = append(encId, FIELD_CcntmEXT)
		srvc.CacheDB.Delete(key)
	}
	err := deleteAllAttr(srvc, encId)
	if err != nil {
		return err
	}

	//set flag to revoke
	utils.PutBytes(srvc, encId, []byte{flag_revoke})
	return nil
}

func updateTimeAndClearProof(srvc *native.NativeService, encId []byte) {
	key := append(encId, FIELD_UPDATED)
	updateTime(srvc, key)
}

func createTimeAndClearProof(srvc *native.NativeService, encId []byte) {
	key := append(encId, FIELD_CREATED)
	updateTime(srvc, key)
}
