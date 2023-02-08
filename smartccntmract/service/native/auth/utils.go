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

package auth

import (
	"fmt"
	"sort"

	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/smartccntmract/event"
	"github.com/cntmio/cntmology/smartccntmract/service/native"
	"github.com/cntmio/cntmology/smartccntmract/service/native/utils"
)

var (
	PreAdmin          = []byte{0x01}
	PreRoleFunc       = []byte{0x02}
	PreRoleToken      = []byte{0x03}
	PreDelegateStatus = []byte{0x04}
)

//type(this.ccntmractAddr.Admin) = []byte
func concatCcntmractAdminKey(native *native.NativeService, ccntmractAddr common.Address) []byte {
	this := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	adminKey := append(this[:], ccntmractAddr[:]...)
	adminKey = append(adminKey, PreAdmin...)

	return adminKey
}

func getCcntmractAdmin(native *native.NativeService, ccntmractAddr common.Address) ([]byte, error) {
	key := concatCcntmractAdminKey(native, ccntmractAddr)
	item, err := utils.GetStorageItem(native.CacheDB, key)
	if err != nil {
		return nil, err
	}
	if item == nil { //is not set
		return nil, nil
	}
	return item.Value, nil
}

func putCcntmractAdmin(native *native.NativeService, ccntmractAddr common.Address, adminOntID []byte) error {
	key := concatCcntmractAdminKey(native, ccntmractAddr)
	utils.PutBytes(native, key, adminOntID)
	return nil
}

//type(this.ccntmractAddr.RoleFunc.role) = roleFuncs
func concatRoleFuncKey(native *native.NativeService, ccntmractAddr common.Address, role []byte) []byte {
	this := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	roleFuncKey := append(this[:], ccntmractAddr[:]...)
	roleFuncKey = append(roleFuncKey, PreRoleFunc...)
	roleFuncKey = append(roleFuncKey, role...)

	return roleFuncKey
}

func getRoleFunc(native *native.NativeService, ccntmractAddr common.Address, role []byte) (*roleFuncs, error) {
	key := concatRoleFuncKey(native, ccntmractAddr, role)
	item, err := utils.GetStorageItem(native.CacheDB, key)
	if err != nil {
		return nil, err
	}
	if item == nil { //is not set
		return nil, nil
	}
	source := common.NewZeroCopySource(item.Value)
	rF := new(roleFuncs)
	err = rF.Deserialization(source)
	if err != nil {
		return nil, fmt.Errorf("deserialize roleFuncs object failed. data: %x", item.Value)
	}
	return rF, nil
}

func putRoleFunc(native *native.NativeService, ccntmractAddr common.Address, role []byte, funcs *roleFuncs) error {
	key := concatRoleFuncKey(native, ccntmractAddr, role)
	utils.PutBytes(native, key, common.SerializeToBytes(funcs))
	return nil
}

//type(this.ccntmractAddr.RoleP.cntmID) = roleTokens
func concatOntIDTokenKey(native *native.NativeService, ccntmractAddr common.Address, cntmID []byte) []byte {
	this := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	tokenKey := append(this[:], ccntmractAddr[:]...)
	tokenKey = append(tokenKey, PreRoleToken...)
	tokenKey = append(tokenKey, cntmID...)

	return tokenKey
}

func getOntIDToken(native *native.NativeService, ccntmractAddr common.Address, cntmID []byte) (*roleTokens, error) {
	key := concatOntIDTokenKey(native, ccntmractAddr, cntmID)
	item, err := utils.GetStorageItem(native.CacheDB, key)
	if err != nil {
		return nil, err
	}
	if item == nil { //is not set
		return nil, nil
	}
	source := common.NewZeroCopySource(item.Value)
	rT := new(roleTokens)
	err = rT.Deserialization(source)
	if err != nil {
		return nil, fmt.Errorf("deserialize roleTokens object failed. data: %x", item.Value)
	}
	return rT, nil
}

func putOntIDToken(native *native.NativeService, ccntmractAddr common.Address, cntmID []byte, tokens *roleTokens) error {
	key := concatOntIDTokenKey(native, ccntmractAddr, cntmID)
	utils.PutBytes(native, key, common.SerializeToBytes(tokens))
	return nil
}

//type(this.ccntmractAddr.DelegateStatus.cntmID)
func concatDelegateStatusKey(native *native.NativeService, ccntmractAddr common.Address, cntmID []byte) []byte {
	this := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	key := append(this[:], ccntmractAddr[:]...)
	key = append(key, PreDelegateStatus...)
	key = append(key, cntmID...)

	return key
}

func getDelegateStatus(native *native.NativeService, ccntmractAddr common.Address, cntmID []byte) (*Status, error) {
	key := concatDelegateStatusKey(native, ccntmractAddr, cntmID)
	item, err := utils.GetStorageItem(native.CacheDB, key)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, nil
	}
	status := new(Status)
	source := common.NewZeroCopySource(item.Value)
	err = status.Deserialization(source)
	if err != nil {
		return nil, fmt.Errorf("deserialize Status object failed. data: %x", item.Value)
	}
	return status, nil
}

func putDelegateStatus(native *native.NativeService, ccntmractAddr common.Address, cntmID []byte, status *Status) error {
	key := concatDelegateStatusKey(native, ccntmractAddr, cntmID)
	utils.PutBytes(native, key, common.SerializeToBytes(status))
	return nil
}

//remove duplicates in the slice of string and sorts the slice in increasing order.
func StringsDedupAndSort(s []string) []string {
	smap := make(map[string]int)
	for i, str := range s {
		if str == "" {
			ccntminue
		}
		smap[str] = i
	}
	ret := make([]string, len(smap))
	i := 0
	for str := range smap {
		ret[i] = str
		i++
	}
	sort.Strings(ret)
	return ret
}

func pushEvent(native *native.NativeService, s interface{}) {
	event := new(event.NotifyEventInfo)
	event.CcntmractAddress = native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	event.States = s
	native.Notifications = append(native.Notifications, event)
}

func serializeAddress(sink *common.ZeroCopySink, addr common.Address) {
	sink.WriteVarBytes(addr[:])
}
