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
	"fmt"

	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/log"
	"github.com/cntmio/cntmology/smartccntmract/service/native"
	"github.com/cntmio/cntmology/smartccntmract/service/native/utils"
)

func GetPublicKeyByID(srvc *native.NativeService) ([]byte, error) {
	args := common.NewZeroCopySource(srvc.Input)
	// arg0: ID
	arg0, err := utils.DecodeVarBytes(args)
	if err != nil {
		return nil, errors.New("get public key failed: argument 0 error")
	}
	// arg1: key ID
	arg1, err := utils.DecodeUint32(args)
	if err != nil {
		return nil, errors.New("get public key failed: argument 1 error")
	}

	key, err := encodeID(arg0)
	if err != nil {
		return nil, fmt.Errorf("get public key failed: %s", err)
	}

	pk, err := getPk(srvc, key, arg1)
	if err != nil {
		return nil, fmt.Errorf("get public key failed: %s", err)
	} else if pk == nil {
		return nil, errors.New("get public key failed: not found")
	} else if pk.revoked {
		return nil, errors.New("get public key failed: revoked")
	}

	return pk.key, nil
}

func GetDDO(srvc *native.NativeService) ([]byte, error) {
	log.Debug("GetDDO")
	var0, err := GetPublicKeys(srvc)
	if err != nil {
		return nil, fmt.Errorf("get DDO error: %s", err)
	} else if var0 == nil {
		log.Debug("DDO: null")
		return nil, nil
	}
	sink := common.NewZeroCopySink(nil)
	sink.WriteVarBytes(var0)

	var1, err := GetAttributes(srvc)
	sink.WriteVarBytes(var1)

	source := common.NewZeroCopySource(srvc.Input)
	did, _, irregular, eof := source.NextVarBytes()
	if irregular || eof {
		return nil, fmt.Errorf("read did failed")
	}
	key, _ := encodeID(did)
	var2, err := getRecovery(srvc, key)
	sink.WriteVarBytes(var2)
	res := sink.Bytes()
	log.Debug("DDO:", hex.EncodeToString(res))
	return res, nil
}

func GetPublicKeys(srvc *native.NativeService) ([]byte, error) {
	log.Debug("GetPublicKeys")
	args := common.NewZeroCopySource(srvc.Input)
	did, err := utils.DecodeVarBytes(args)
	if err != nil {
		return nil, fmt.Errorf("get public keys error: invalid argument, %s", err)
	}
	if len(did) == 0 {
		return nil, errors.New("get public keys error: invalid ID")
	}
	key, err := encodeID(did)
	if err != nil {
		return nil, fmt.Errorf("get public keys error: %s", err)
	}
	key = append(key, FIELD_PK)
	list, err := getAllPk(srvc, key)
	if err != nil {
		return nil, fmt.Errorf("get public keys error: %s", err)
	} else if list == nil {
		return nil, nil
	}

	sink := common.NewZeroCopySink(nil)
	for i, v := range list {
		if v.revoked {
			ccntminue
		}
		sink.WriteUint32(uint32(i + 1))
		sink.WriteVarBytes(v.key)
	}

	return sink.Bytes(), nil
}

func GetAttributes(srvc *native.NativeService) ([]byte, error) {
	log.Debug("GetAttributes")
	source := common.NewZeroCopySource(srvc.Input)
	did, err := utils.DecodeVarBytes(source)
	if err != nil {
		return nil, fmt.Errorf("get public keys error: invalid argument, %s", err)
	}
	if len(did) == 0 {
		return nil, errors.New("get attributes error: invalid ID")
	}
	key, err := encodeID(did)
	if err != nil {
		return nil, fmt.Errorf("get public keys error: %s", err)
	}
	res, err := getAllAttr(srvc, key)
	if err != nil {
		return nil, fmt.Errorf("get attributes error: %s", err)
	}

	return res, nil
}

func GetKeyState(srvc *native.NativeService) ([]byte, error) {
	log.Debug("GetKeyState")
	source := common.NewZeroCopySource(srvc.Input)
	// arg0: ID
	arg0, _, irregular, eof := source.NextVarBytes()
	if irregular || eof {
		return nil, fmt.Errorf("get key state failed: argument 0 error")
	}
	// arg1: public key ID
	arg1, err := utils.DecodeVarUint(source)
	if err != nil {
		return nil, fmt.Errorf("get key state failed: argument 1 error, %s", err)
	}

	key, err := encodeID(arg0)
	if err != nil {
		return nil, fmt.Errorf("get key state failed: %s", err)
	}

	owner, err := getPk(srvc, key, uint32(arg1))
	if err != nil {
		return nil, fmt.Errorf("get key state failed: %s", err)
	} else if owner == nil {
		log.Debug("key state: not exist")
		return []byte("not exist"), nil
	}

	log.Debug("key state: ", owner.revoked)
	if owner.revoked {
		return []byte("revoked"), nil
	} else {
		return []byte("in use"), nil
	}
}
