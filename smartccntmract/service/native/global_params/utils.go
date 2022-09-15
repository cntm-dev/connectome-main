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
	"github.com/cntmio/cntmology/common/config"
	cstates "github.com/cntmio/cntmology/core/states"
	"github.com/cntmio/cntmology/smartccntmract/event"
	"github.com/cntmio/cntmology/smartccntmract/service/native"
	"github.com/cntmio/cntmology/smartccntmract/service/native/utils"
)

const (
	PARAM    = "param"
	TRANSFER = "transfer"
	ADMIN    = "admin"
	OPERATOR = "operator"
)

func getRoleStorageItem(role common.Address) *cstates.StorageItem {
	bf := new(bytes.Buffer)
	utils.WriteAddress(bf, role)
	return &cstates.StorageItem{Value: bf.Bytes()}
}

func getParamStorageItem(params Params) *cstates.StorageItem {
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

func getStorageParam(native *native.NativeService, key []byte) (Params, error) {
	item, err := utils.GetStorageItem(native, key)
	params := Params{}
	if err != nil || item == nil {
		return params, err
	}
	bf := bytes.NewBuffer(item.Value)
	err = params.Deserialize(bf)
	return params, err
}

func GetStorageRole(native *native.NativeService, key []byte) (common.Address, error) {
	item, err := utils.GetStorageItem(native, key)
	var role common.Address
	if err != nil || item == nil {
		return role, err
	}
	bf := bytes.NewBuffer(item.Value)
	role, err = utils.ReadAddress(bf)
	return role, err
}

func NotifyRoleChange(native *native.NativeService, ccntmract common.Address, functionName string,
	newAddr common.Address) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			CcntmractAddress: ccntmract,
			States:          []interface{}{functionName, newAddr.ToBase58()},
		})
}

func NotifyTransferAdmin(native *native.NativeService, ccntmract common.Address, functionName string,
	originAdmin, newAdmin common.Address) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			CcntmractAddress: ccntmract,
			States:          []interface{}{functionName, originAdmin.ToBase58(), newAdmin.ToBase58()},
		})
}

func NotifyParamChange(native *native.NativeService, ccntmract common.Address, functionName string, params Params) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	paramsString := ""
	for _, param := range params {
		paramsString += param.Key + "," + param.Value + ";"
	}
	paramsString = paramsString[:len(paramsString)-1]
	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			CcntmractAddress: ccntmract,
			States:          []interface{}{functionName, paramsString},
		})
}
