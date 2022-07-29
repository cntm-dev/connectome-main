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

package global_params

import (
	"bytes"

	"github.com/cntmio/cntmology/common"
	cstates "github.com/cntmio/cntmology/core/states"
	"github.com/cntmio/cntmology/smartccntmract/service/native"
	"github.com/cntmio/cntmology/smartccntmract/service/native/utils"
)

const (
	PARAM    = "param"
	TRANSFER = "transfer"
	ADMIN    = "admin"
	OPERATOR = "operator"
)

func getRoleStorageItem(role *Role) *cstates.StorageItem {
	bf := new(bytes.Buffer)
	role.Serialize(bf)
	return &cstates.StorageItem{Value: bf.Bytes()}
}

func getParamStorageItem(params *Params) *cstates.StorageItem {
	bf := new(bytes.Buffer)
	params.Serialize(bf)
	return &cstates.StorageItem{Value: bf.Bytes()}
}

func generateParamKey(ccntmract common.Address, valueType paramType) []byte {
	key := append(ccntmract[:], PARAM...)
	key = append(key[:], byte(valueType))
	return key
}

func generateAdminKey(ccntmract common.Address, isTransferAdmin bool) []byte {
	if isTransferAdmin {
		return append(ccntmract[:], TRANSFER...)
	} else {
		return append(ccntmract[:], ADMIN...)
	}
}

func GenerateOperatorKey(ccntmract common.Address) []byte {
	return append(ccntmract[:], OPERATOR...)
}

func getStorageParam(native *native.NativeService, key []byte) (*Params, error) {
	item, err := utils.GetStorageItem(native, key)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, nil
	}
	params := new(Params)
	bf := bytes.NewBuffer(item.Value)
	params.Deserialize(bf)
	return params, nil
}

func GetStorageRole(native *native.NativeService, key []byte) (*Role, error) {
	item, err := utils.GetStorageItem(native, key)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, nil
	}
	role := new(Role)
	bf := bytes.NewBuffer(item.Value)
	role.Deserialize(bf)
	return role, nil
}
