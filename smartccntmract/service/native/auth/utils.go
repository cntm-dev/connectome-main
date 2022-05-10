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
	"github.com/cntmio/cntmology/core/states"
	"github.com/cntmio/cntmology/core/store/common"
	"github.com/cntmio/cntmology/smartccntmract/service/native"
)

var (
	RoleF        = []byte{0x01}
	RoleP        = []byte{0x02}
	FuncPerson   = []byte{0x03}
	DelegateList = []byte{0x04}
	Admin        = []byte{0x05}
)

//this.ccntmractAddr.Admin
func GetCcntmractAdminKey(native *native.NativeService, ccntmractAddr []byte) ([]byte, error) {
	this := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	adminKey, err := PackKeys(this[:], [][]byte{ccntmractAddr, Admin})

	return adminKey, err
}

//this.ccntmractAddr.RoleF.role
func GetRoleFKey(native *native.NativeService, ccntmractAddr, role []byte) ([]byte, error) {
	this := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	roleFKey, err := PackKeys(this[:], [][]byte{ccntmractAddr, RoleF, role})

	return roleFKey, err
}

//this.ccntmractAddr.RoleP.role
func GetRolePKey(native *native.NativeService, ccntmractAddr, role []byte) ([]byte, error) {
	this := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	rolePKey, err := PackKeys(this[:], [][]byte{ccntmractAddr, RoleP, role})

	return rolePKey, err
}

//this.ccntmractAddr.FuncOntID.func.cntmID
func GetFuncOntIDKey(native *native.NativeService, ccntmractAddr, fn, cntmID []byte) ([]byte, error) {
	this := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	funcOntIDKey, err := PackKeys(this[:], [][]byte{ccntmractAddr, FuncPerson, fn, cntmID})

	return funcOntIDKey, err
}

//this.ccntmractAddr.DelegateList.role.cntmID
func GetDelegateListKey(native *native.NativeService, ccntmractAddr, role, cntmID []byte) ([]byte, error) {
	this := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	delegateListKey, err := PackKeys(this[:], [][]byte{ccntmractAddr, DelegateList, role, cntmID})

	return delegateListKey, err
}

func PutBytes(native *native.NativeService, key []byte, value []byte) {
	native.CloneCache.Add(common.ST_STORAGE, key, &states.StorageItem{Value: value})
}

func writeAuthToken(native *native.NativeService, ccntmractAddr, fn, cntmID, auth []byte) error {
	key, err := GetFuncOntIDKey(native, ccntmractAddr, fn, cntmID)
	if err != nil {
		return err
	}
	PutBytes(native, key, auth)
	return nil
}
