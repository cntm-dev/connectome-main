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
	"github.com/cntmio/cntmology/errors"
	vm "github.com/cntmio/cntmology/vm/neovm"
)

// CcntmractCreate create a new smart ccntmract on blockchain, and put it to vm stack
func CcntmractCreate(service *NeoVmService, engine *vm.Executor) error {
	ccntmract, err := isCcntmractParamValid(engine)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[CcntmractCreate] ccntmract parameters invalid!")
	}
	ccntmractAddress := ccntmract.Address()
	dep, err := service.CacheDB.GetCcntmract(ccntmractAddress)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[CcntmractCreate] GetOrAdd error!")
	}
	if dep == nil {
		err := service.CacheDB.PutCcntmract(ccntmract)
		if err != nil {
			return err
		}
		dep = ccntmract
	}
	return engine.EvalStack.PushAsInteropValue(dep)
}

// CcntmractMigrate migrate old smart ccntmract to a new ccntmract, and destroy old ccntmract
func CcntmractMigrate(service *NeoVmService, engine *vm.Executor) error {
	ccntmract, err := isCcntmractParamValid(engine)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[CcntmractMigrate] ccntmract parameters invalid!")
	}
	newAddr := ccntmract.Address()

	if err := isCcntmractExist(service, newAddr); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[CcntmractMigrate] ccntmract invalid!")
	}
	ccntmext := service.CcntmextRef.CurrentCcntmext()
	oldAddr := ccntmext.CcntmractAddress

	err = service.CacheDB.PutCcntmract(ccntmract)
	if err != nil {
		return err
	}
	service.CacheDB.DeleteCcntmract(oldAddr)

	iter := service.CacheDB.NewIterator(oldAddr[:])
	for has := iter.First(); has; has = iter.Next() {
		key := iter.Key()
		val := iter.Value()

		newKey := genStorageKey(newAddr, key[20:])
		service.CacheDB.Put(newKey, val)
		service.CacheDB.Delete(key)
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		return err
	}
	return engine.EvalStack.PushAsInteropValue(ccntmract)
}

// CcntmractDestory destroy a ccntmract
func CcntmractDestory(service *NeoVmService, engine *vm.Executor) error {
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
func CcntmractGetStorageCcntmext(service *NeoVmService, engine *vm.Executor) error {
	opInterface, err := engine.EvalStack.PopAsInteropValue()
	if err != nil {
		return err
	}
	if opInterface.Data == nil {
		return errors.NewErr("[GetStorageCcntmext] Pop data nil!")
	}
	ccntmractState, ok := opInterface.Data.(*payload.DeployCode)
	if !ok {
		return errors.NewErr("[GetStorageCcntmext] Pop data not ccntmract!")
	}
	address := ccntmractState.Address()
	item, err := service.CacheDB.GetCcntmract(address)
	if err != nil || item == nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[GetStorageCcntmext] Get StorageCcntmext nil")
	}
	if address != service.CcntmextRef.CurrentCcntmext().CcntmractAddress {
		return errors.NewErr("[GetStorageCcntmext] CodeHash not equal!")
	}
	return engine.EvalStack.PushAsInteropValue(NewStorageCcntmext(address))
}

// CcntmractGetCode put ccntmract to vm stack
func CcntmractGetCode(service *NeoVmService, engine *vm.Executor) error {
	i, err := engine.EvalStack.PopAsInteropValue()
	if err != nil {
		return err
	}
	if d, ok := i.Data.(*payload.DeployCode); ok {
		return engine.EvalStack.PushBytes(d.Code)
	}
	return fmt.Errorf("[CcntmractGetCode] Type error ")
}

func isCcntmractParamValid(engine *vm.Executor) (*payload.DeployCode, error) {
	if engine.EvalStack.Count() < 7 {
		return nil, errors.NewErr("[Ccntmract] Too few input parameters")
	}
	code, err := engine.EvalStack.PopAsBytes()
	if err != nil {
		return nil, err
	}

	vmType, err := engine.EvalStack.PopAsInt64()
	if err != nil {
		return nil, err
	}
	name, err := engine.EvalStack.PopAsBytes()
	if err != nil {
		return nil, err
	}

	version, err := engine.EvalStack.PopAsBytes()
	if err != nil {
		return nil, err
	}

	author, err := engine.EvalStack.PopAsBytes()
	if err != nil {
		return nil, err
	}

	email, err := engine.EvalStack.PopAsBytes()
	if err != nil {
		return nil, err
	}

	desc, err := engine.EvalStack.PopAsBytes()
	if err != nil {
		return nil, err
	}

	ccntmract, err := payload.CreateDeployCode(code, uint32(vmType), name, version, author, email, desc)

	if err != nil {
		return nil, err
	}

	return ccntmract, nil
}

func isCcntmractExist(service *NeoVmService, ccntmractAddress common.Address) error {
	item, err := service.CacheDB.GetCcntmract(ccntmractAddress)

	if err != nil || item != nil {
		return fmt.Errorf("[Ccntmract] Get ccntmract %x error or ccntmract exist", ccntmractAddress)
	}
	return nil
}
