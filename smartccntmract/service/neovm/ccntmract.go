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

package neovm

import (
	"fmt"

	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/core/payload"
	scommon "github.com/cntmio/cntmology/core/store/common"
	"github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/errors"
	vm "github.com/cntmio/cntmology/vm/neovm"
)

// CcntmractCreate create a new smart ccntmract on blockchain, and put it to vm stack
func CcntmractCreate(service *NeoVmService, engine *vm.ExecutionEngine) error {
	ccntmract, err := isCcntmractParamValid(engine)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[CcntmractCreate] ccntmract parameters invalid!")
	}
	ccntmractAddress := types.AddressFromVmCode(ccntmract.Code)
	state, err := service.CloneCache.GetOrAdd(scommon.ST_CcntmRACT, ccntmractAddress[:], ccntmract)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[CcntmractCreate] GetOrAdd error!")
	}
	vm.PushData(engine, state)
	return nil
}

// CcntmractMigrate migrate old smart ccntmract to a new ccntmract, and destroy old ccntmract
func CcntmractMigrate(service *NeoVmService, engine *vm.ExecutionEngine) error {
	ccntmract, err := isCcntmractParamValid(engine)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[CcntmractMigrate] ccntmract parameters invalid!")
	}
	ccntmractAddress := types.AddressFromVmCode(ccntmract.Code)

	if err := isCcntmractExist(service, ccntmractAddress); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[CcntmractMigrate] ccntmract invalid!")
	}
	ccntmext := service.CcntmextRef.CurrentCcntmext()

	service.CloneCache.Add(scommon.ST_CcntmRACT, ccntmractAddress[:], ccntmract)
	items, err := storeMigration(service, ccntmext.CcntmractAddress, ccntmractAddress)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[CcntmractMigrate] ccntmract store migration error!")
	}
	service.CloneCache.Delete(scommon.ST_CcntmRACT, ccntmext.CcntmractAddress[:])
	for _, v := range items {
		service.CloneCache.Delete(scommon.ST_STORAGE, []byte(v.Key))
	}
	vm.PushData(engine, ccntmract)
	return nil
}

// CcntmractDestory destroy a ccntmract
func CcntmractDestory(service *NeoVmService, engine *vm.ExecutionEngine) error {
	ccntmext := service.CcntmextRef.CurrentCcntmext()
	if ccntmext == nil {
		return errors.NewErr("[CcntmractDestory] current ccntmract ccntmext invalid!")
	}
	item, err := service.CloneCache.Store.TryGet(scommon.ST_CcntmRACT, ccntmext.CcntmractAddress[:])

	if err != nil || item == nil {
		return errors.NewErr("[CcntmractDestory] get current ccntmract fail!")
	}

	service.CloneCache.Delete(scommon.ST_CcntmRACT, ccntmext.CcntmractAddress[:])
	stateValues, err := service.CloneCache.Store.Find(scommon.ST_CcntmRACT, ccntmext.CcntmractAddress[:])
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[CcntmractDestory] find error!")
	}
	for _, v := range stateValues {
		service.CloneCache.Delete(scommon.ST_STORAGE, []byte(v.Key))
	}
	return nil
}

// CcntmractGetStorageCcntmext put ccntmract storage ccntmext to vm stack
func CcntmractGetStorageCcntmext(service *NeoVmService, engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 1 {
		return errors.NewErr("[GetStorageCcntmext] Too few input parameter!")
	}
	opInterface, err := vm.PopInteropInterface(engine)
	if err != nil {
		return err
	}
	if opInterface == nil {
		return errors.NewErr("[GetStorageCcntmext] Pop data nil!")
	}
	ccntmractState, ok := opInterface.(*payload.DeployCode)
	if !ok {
		return errors.NewErr("[GetStorageCcntmext] Pop data not ccntmract!")
	}
	address := types.AddressFromVmCode(ccntmractState.Code)
	item, err := service.CloneCache.Store.TryGet(scommon.ST_CcntmRACT, address[:])
	if err != nil || item == nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[GetStorageCcntmext] Get StorageCcntmext nil")
	}
	if address != service.CcntmextRef.CurrentCcntmext().CcntmractAddress {
		return errors.NewErr("[GetStorageCcntmext] CodeHash not equal!")
	}
	vm.PushData(engine, NewStorageCcntmext(address))
	return nil
}

// CcntmractGetCode put ccntmract to vm stack
func CcntmractGetCode(service *NeoVmService, engine *vm.ExecutionEngine) error {
	i, err := vm.PopInteropInterface(engine)
	if err != nil {
		return err
	}
	vm.PushData(engine, i.(*payload.DeployCode).Code)
	return nil
}

func isCcntmractParamValid(engine *vm.ExecutionEngine) (*payload.DeployCode, error) {
	if vm.EvaluationStackCount(engine) < 7 {
		return nil, errors.NewErr("[Ccntmract] Too few input parameters")
	}
	code, err := vm.PopByteArray(engine)
	if err != nil {
		return nil, err
	}
	if len(code) > 1024*1024 {
		return nil, errors.NewErr("[Ccntmract] Code too lcntm!")
	}
	needStorage, err := vm.PopBoolean(engine)
	if err != nil {
		return nil, err
	}
	name, err := vm.PopByteArray(engine)
	if err != nil {
		return nil, err
	}
	if len(name) > 252 {
		return nil, errors.NewErr("[Ccntmract] Name too lcntm!")
	}
	version, err := vm.PopByteArray(engine)
	if err != nil {
		return nil, err
	}
	if len(version) > 252 {
		return nil, errors.NewErr("[Ccntmract] Version too lcntm!")
	}
	author, err := vm.PopByteArray(engine)
	if err != nil {
		return nil, err
	}
	if len(author) > 252 {
		return nil, errors.NewErr("[Ccntmract] Author too lcntm!")
	}
	email, err := vm.PopByteArray(engine)
	if err != nil {
		return nil, err
	}
	if len(email) > 252 {
		return nil, errors.NewErr("[Ccntmract] Email too lcntm!")
	}
	desc, err := vm.PopByteArray(engine)
	if err != nil {
		return nil, err
	}
	if len(desc) > 65536 {
		return nil, errors.NewErr("[Ccntmract] Desc too lcntm!")
	}
	ccntmract := &payload.DeployCode{
		Code:        code,
		NeedStorage: needStorage,
		Name:        string(name),
		Version:     string(version),
		Author:      string(author),
		Email:       string(email),
		Description: string(desc),
	}
	return ccntmract, nil
}

func isCcntmractExist(service *NeoVmService, ccntmractAddress common.Address) error {
	item, err := service.CloneCache.Get(scommon.ST_CcntmRACT, ccntmractAddress[:])

	if err != nil || item != nil {
		return fmt.Errorf("[Ccntmract] Get ccntmract %x error or ccntmract exist!", ccntmractAddress)
	}
	return nil
}

func storeMigration(service *NeoVmService, oldAddr common.Address, newAddr common.Address) ([]*scommon.StateItem, error) {
	stateValues, err := service.CloneCache.Store.Find(scommon.ST_STORAGE, oldAddr[:])
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[Ccntmract] Find error!")
	}
	for _, v := range stateValues {
		service.CloneCache.Add(scommon.ST_STORAGE, getStorageKey(newAddr, []byte(v.Key)[20:]), v.Value)
	}
	return stateValues, nil
}
