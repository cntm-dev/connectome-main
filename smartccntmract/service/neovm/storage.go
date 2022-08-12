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
	"bytes"

	"fmt"
	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/core/states"
	scommon "github.com/cntmio/cntmology/core/store/common"
	"github.com/cntmio/cntmology/errors"
	vm "github.com/cntmio/cntmology/vm/neovm"
)

// StoragePut put smart ccntmract storage item to cache
func StoragePut(service *NeoVmService, engine *vm.ExecutionEngine) error {
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

	key := vm.PopByteArray(engine)
	if len(key) > 1024 {
		return errors.NewErr("[StoragePut] Storage key to lcntm")
	}

	value := vm.PopByteArray(engine)
	service.CloneCache.Add(scommon.ST_STORAGE, getStorageKey(ccntmext.Address, key), &states.StorageItem{Value: value})
	return nil
}

// StorageDelete delete smart ccntmract storage item from cache
func StorageDelete(service *NeoVmService, engine *vm.ExecutionEngine) error {
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

	service.CloneCache.Delete(scommon.ST_STORAGE, getStorageKey(ccntmext.Address, vm.PopByteArray(engine)))

	return nil
}

// StorageGet push smart ccntmract storage item from cache to vm stack
func StorageGet(service *NeoVmService, engine *vm.ExecutionEngine) error {
	ccntmext, err := getCcntmext(engine)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[StorageGet] get pop ccntmext error!")
	}

	item, err := service.CloneCache.Get(scommon.ST_STORAGE, getStorageKey(ccntmext.Address, vm.PopByteArray(engine)))
	if err != nil {
		return err
	}

	if item == nil {
		vm.PushData(engine, []byte{})
	} else {
		vm.PushData(engine, item.(*states.StorageItem).Value)
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
	item, err := service.CloneCache.Get(scommon.ST_CcntmRACT, ccntmext.Address[:])
	if err != nil || item == nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[CheckStorageCcntmext] get ccntmext fail!")
	}
	return nil
}

func getCcntmext(engine *vm.ExecutionEngine) (*StorageCcntmext, error) {
	if vm.EvaluationStackCount(engine) < 2 {
		return nil, errors.NewErr("[Ccntmext] Too few input parameters ")
	}
	opInterface := vm.PopInteropInterface(engine)
	if opInterface == nil {
		return nil, errors.NewErr("[Ccntmext] Get storageCcntmext nil")
	}
	ccntmext, ok := opInterface.(*StorageCcntmext)
	if !ok {
		return nil, errors.NewErr("[Ccntmext] Get storageCcntmext invalid")
	}
	return ccntmext, nil
}

func getStorageKey(address common.Address, key []byte) []byte {
	buf := bytes.NewBuffer(nil)
	buf.Write(address[:])
	buf.Write(key)
	return buf.Bytes()
}
