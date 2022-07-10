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
	"bytes"
	"fmt"

	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/serialization"
	"github.com/cntmio/cntmology/smartccntmract/event"
	"github.com/cntmio/cntmology/smartccntmract/service/native"
	"github.com/cntmio/cntmology/smartccntmract/service/native/utils"
)

var (
	PreAdmin          = []byte{0x01}
	PreRoleFunc       = []byte{0x02}
	PreRoleToken      = []byte{0x03}
	PreDelegateStatus = []byte{0x04}
	//PreDelegateList   = []byte{0x04}
)

//type(this.ccntmractAddr.Admin) = []byte
func concatCcntmractAdminKey(native *native.NativeService, ccntmractAddr []byte) ([]byte, error) {
	this := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	adminKey, err := packKeys(this, [][]byte{ccntmractAddr, PreAdmin})

	return adminKey, err
}

func getCcntmractAdmin(native *native.NativeService, ccntmractAddr []byte) ([]byte, error) {
	key, err := concatCcntmractAdminKey(native, ccntmractAddr)
	if err != nil {
		return nil, err
	}
	item, err := utils.GetStorageItem(native, key)
	if err != nil {
		return nil, err
	}
	if item == nil { //is not set
		return nil, nil
	}
	return item.Value, nil
}

func putCcntmractAdmin(native *native.NativeService, ccntmractAddr, adminOntID []byte) error {
	key, err := concatCcntmractAdminKey(native, ccntmractAddr)
	if err != nil {
		return err
	}
	utils.PutBytes(native, key, adminOntID)
	return nil
}

//type(this.ccntmractAddr.RoleFunc.role) = roleFuncs
func concatRoleFuncKey(native *native.NativeService, ccntmractAddr, role []byte) ([]byte, error) {
	this := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	roleFuncKey, err := packKeys(this, [][]byte{ccntmractAddr, PreRoleFunc, role})

	return roleFuncKey, err
}

func getRoleFunc(native *native.NativeService, ccntmractAddr, role []byte) (*roleFuncs, error) {
	key, err := concatRoleFuncKey(native, ccntmractAddr, role)
	if err != nil {
		return nil, err
	}
	item, err := utils.GetStorageItem(native, key)
	if err != nil {
		return nil, err
	}
	if item == nil { //is not set
		return nil, nil
	}
	rd := bytes.NewReader(item.Value)
	rF := new(roleFuncs)
	err = rF.Deserialize(rd)
	if err != nil {
		return nil, fmt.Errorf("deserialize roleFuncs object failed. data: %x", item.Value)
	}
	return rF, nil
}

func putRoleFunc(native *native.NativeService, ccntmractAddr, role []byte, funcs *roleFuncs) error {
	key, _ := concatRoleFuncKey(native, ccntmractAddr, role)
	bf := new(bytes.Buffer)
	err := funcs.Serialize(bf)
	if err != nil {
		return fmt.Errorf("serialize roleFuncs failed, caused by %v", err)
	}
	utils.PutBytes(native, key, bf.Bytes())
	return nil
}

//type(this.ccntmractAddr.RoleP.cntmID) = roleTokens
func concatOntIDTokenKey(native *native.NativeService, ccntmractAddr, cntmID []byte) ([]byte, error) {
	this := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	tokenKey, err := packKeys(this, [][]byte{ccntmractAddr, PreRoleToken, cntmID})

	return tokenKey, err
}

func getOntIDToken(native *native.NativeService, ccntmractAddr, cntmID []byte) (*roleTokens, error) {
	key, err := concatOntIDTokenKey(native, ccntmractAddr, cntmID)
	if err != nil {
		return nil, err
	}
	item, err := utils.GetStorageItem(native, key)
	if err != nil {
		return nil, err
	}
	if item == nil { //is not set
		return nil, nil
	}
	rd := bytes.NewReader(item.Value)
	rT := new(roleTokens)
	err = rT.Deserialize(rd)
	if err != nil {
		return nil, fmt.Errorf("deserialize roleTokens object failed. data: %x", item.Value)
	}
	return rT, nil
}

func putOntIDToken(native *native.NativeService, ccntmractAddr, cntmID []byte, tokens *roleTokens) error {
	key, _ := concatOntIDTokenKey(native, ccntmractAddr, cntmID)
	bf := new(bytes.Buffer)
	err := tokens.Serialize(bf)
	if err != nil {
		return fmt.Errorf("serialize roleFuncs failed, caused by %v", err)
	}
	utils.PutBytes(native, key, bf.Bytes())
	return nil
}

//type(this.ccntmractAddr.DelegateStatus.cntmID)
func concatDelegateStatusKey(native *native.NativeService, ccntmractAddr, cntmID []byte) ([]byte, error) {
	this := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	key, err := packKeys(this, [][]byte{ccntmractAddr, PreDelegateStatus, cntmID})

	return key, err
}

func getDelegateStatus(native *native.NativeService, ccntmractAddr, cntmID []byte) (*Status, error) {
	key, err := concatDelegateStatusKey(native, ccntmractAddr, cntmID)
	if err != nil {
		return nil, err
	}
	item, err := utils.GetStorageItem(native, key)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, nil
	}
	status := new(Status)
	rd := bytes.NewReader(item.Value)
	err = status.Deserialize(rd)
	if err != nil {
		return nil, fmt.Errorf("deserialize Status object failed. data: %x", item.Value)
	}
	return status, nil
}

func putDelegateStatus(native *native.NativeService, ccntmractAddr, cntmID []byte, status *Status) error {
	key, _ := concatDelegateStatusKey(native, ccntmractAddr, cntmID)
	bf := new(bytes.Buffer)
	err := status.Serialize(bf)
	if err != nil {
		return fmt.Errorf("serialize Status failed, caused by %v", err)
	}
	utils.PutBytes(native, key, bf.Bytes())
	return nil
}

/*
 * pack data to be used as a key in the kv storage
 * key := field || ser_items[1] || ... || ser_items[n]
 */
func packKeys(field common.Address, items [][]byte) ([]byte, error) {
	w := new(bytes.Buffer)
	for _, item := range items {
		err := serialization.WriteVarBytes(w, item)
		if err != nil {
			return nil, fmt.Errorf("packKeys failed when serialize %x", item)
		}
	}
	key := append(field[:], w.Bytes()...)
	return key, nil
}

/*
 * pack data to be used as a key in the kv storage
 * key := field || ser_data
 */
func packKey(field common.Address, data []byte) ([]byte, error) {
	return packKeys(field, [][]byte{data})
}

//remote duplicates in the slice of string
func stringSliceUniq(s []string) []string {
	smap := make(map[string]int)
	for i, str := range s {
		if str == "" {
			ccntminue
		}
		smap[str] = i
	}
	ret := make([]string, len(smap))
	i := 0
	for str, _ := range smap {
		ret[i] = str
		i++
	}
	return ret
}

func pushEvent(native *native.NativeService, s interface{}) {
	event := new(event.NotifyEventInfo)
	event.CcntmractAddress = native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	event.States = s
	native.Notifications = append(native.Notifications, event)
}

func invokeEvent(native *native.NativeService, fn string, ret bool) {
	pushEvent(native, []interface{}{fn, ret})
}
