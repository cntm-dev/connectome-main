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
	"github.com/cntmio/cntmology/core/states"
	"github.com/cntmio/cntmology/errors"
	vm "github.com/cntmio/cntmology/vm/neovm"
)

// StoragePut put smart ccntmract storage item to cache
func StoragePut(service *NeoVmService, engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 3 {
		return errors.NewErr("[Ccntmext] Too few input parameters ")
	}
	ccntmext, err := getCcntmext(engine)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[StoragePut] get pop ccntmext error!")
	}
	if ccntmext.IsReadOnly {
		return fmt.Errorf("%s", "[StoragePut] storage read only!")
	}
	if err := checkStorageCcntmext(service, ccntmext); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[StoragePut] check ccntmext error!")
	}

	key, err := vm.PopByteArray(engine)
	if err != nil {
		return err
	}
	if len(key) > 1024 {
		return errors.NewErr("[StoragePut] Storage key to lcntm")
	}

	value, err := vm.PopByteArray(engine)
	if err != nil {
		return err
	}

	service.CacheDB.Put(genStorageKey(ccntmext.Address, key), states.GenRawStorageItem(value))
	return nil
}

// StorageDelete delete smart ccntmract storage item from cache
func StorageDelete(service *NeoVmService, engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 2 {
		return errors.NewErr("[Ccntmext] Too few input parameters ")
	}
	ccntmext, err := getCcntmext(engine)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[StorageDelete] get pop ccntmext error!")
	}
	if ccntmext.IsReadOnly {
		return fmt.Errorf("%s", "[StorageDelete] storage read only!")
	}
	if err := checkStorageCcntmext(service, ccntmext); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[StorageDelete] check ccntmext error!")
	}
	ba, err := vm.PopByteArray(engine)
	if err != nil {
		return err
	}
	service.CacheDB.Delete(genStorageKey(ccntmext.Address, ba))

	return nil
}

// StorageGet push smart ccntmract storage item from cache to vm stack
func StorageGet(service *NeoVmService, engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 2 {
		return errors.NewErr("[Ccntmext] Too few input parameters ")
	}
	ccntmext, err := getCcntmext(engine)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[StorageGet] get pop ccntmext error!")
	}
	ba, err := vm.PopByteArray(engine)
	if err != nil {
		return err
	}

	raw, err := service.CacheDB.Get(genStorageKey(ccntmext.Address, ba))
	if err != nil {
		return err
	}

	if len(raw) == 0 {
		vm.PushData(engine, []byte{})
	} else {
		value, err := states.GetValueFromRawStorageItem(raw)
		if err != nil {
			return err
		}
		vm.PushData(engine, value)
	}
	return nil
}

// StorageGetCcntmext push smart ccntmract storage ccntmext to vm stack
func StorageGetCcntmext(service *NeoVmService, engine *vm.ExecutionEngine) error {
	vm.PushData(engine, NewStorageCcntmext(service.CcntmextRef.CurrentCcntmext().CcntmractAddress))
	return nil
}

func StorageGetReadOnlyCcntmext(service *NeoVmService, engine *vm.ExecutionEngine) error {
	ccntmext := NewStorageCcntmext(service.CcntmextRef.CurrentCcntmext().CcntmractAddress)
	ccntmext.IsReadOnly = true
	vm.PushData(engine, ccntmext)
	return nil
}

func checkStorageCcntmext(service *NeoVmService, ccntmext *StorageCcntmext) error {
	item, err := service.CacheDB.GetCcntmract(ccntmext.Address)
	if err != nil || item == nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[CheckStorageCcntmext] get ccntmext fail!")
	}
	return nil
}

func getCcntmext(engine *vm.ExecutionEngine) (*StorageCcntmext, error) {
	opInterface, err := vm.PopInteropInterface(engine)
	if err != nil {
		return nil, err
	}
	if opInterface == nil {
		return nil, errors.NewErr("[Ccntmext] Get storageCcntmext nil")
	}
	ccntmext, ok := opInterface.(*StorageCcntmext)
	if !ok {
		return nil, errors.NewErr("[Ccntmext] Get storageCcntmext invalid")
	}
	return ccntmext, nil
}

func genStorageKey(address common.Address, key []byte) []byte {
	res := make([]byte, 0, len(address[:])+len(key))
	res = append(res, address[:]...)
	res = append(res, key...)
	return res
}
