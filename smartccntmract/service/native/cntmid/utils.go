/*
 * Copyright (C) 2018 The cntmology Authors
 * This file is part of The cntmology library.
 *
 * The cntmology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
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
	cmn "github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/core/genesis"
	"github.com/cntmio/cntmology/core/states"
	"github.com/cntmio/cntmology/core/store/common"
	"github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/smartccntmract/service/native"
)

const flag_exist = 0x01

func checkIDExistence(srvc *native.NativeService, encID []byte) bool {
	val, err := srvc.CloneCache.Get(common.ST_STORAGE, encID)
	if err == nil {
		t, ok := val.(*states.StorageItem)
		if ok {
			if len(t.Value) > 0 && t.Value[0] == flag_exist {
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
	enc = append(genesis.OntIDCcntmractAddress[:], enc...)
	return enc, nil
}

func decodeID(data []byte) ([]byte, error) {
	if len(data) == 0 || len(data) != int(data[0])+1 {
		return nil, errors.New("decode cntm ID error: invalid data length")
	}
	return data[1:], nil
}

func setRecovery(srvc *native.NativeService, encID, recovery []byte) error {
	key := append(encID, FIELD_RECOVERY)
	val := &states.StorageItem{Value: recovery}
	srvc.CloneCache.Add(common.ST_STORAGE, key, val)
	return nil
}

func getRecovery(srvc *native.NativeService, encID []byte) ([]byte, error) {
	key := append(encID, FIELD_RECOVERY)
	item, err := getStorageItem(srvc, key)
	if err != nil {
		return nil, errors.New("get recovery error: " + err.Error())
	}
	return item.Value, nil
}

func getStorageItem(srvc *native.NativeService, key []byte) (*states.StorageItem, error) {
	val, err := srvc.CloneCache.Get(common.ST_STORAGE, key)
	if err != nil {
		return nil, err
	}
	t, ok := val.(*states.StorageItem)
	if !ok {
		return nil, errors.New("invalid value type")
	}
	return t, nil
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
	addr, err := cmn.AddressParseFromBytes(key)
	if srvc.CcntmextRef.CheckWitness(addr) {
		return nil
	}

	return errors.New("check witness failed, " + hex.EncodeToString(key))
}
