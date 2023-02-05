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

package wasmvm

import (
	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/config"
	"github.com/cntmio/cntmology/core/payload"
	"github.com/cntmio/cntmology/errors"
	"github.com/cntmio/wagon/exec"
)

func migrateCcntmractStorage(service *WasmVmService, newAddress common.Address) error {
	oldAddress := service.CcntmextRef.CurrentCcntmext().CcntmractAddress
	service.CacheDB.DeleteCcntmract(oldAddress)

	iter := service.CacheDB.NewIterator(oldAddress[:])
	for has := iter.First(); has; has = iter.Next() {
		key := iter.Key()
		val := iter.Value()

		newkey := serializeStorageKey(newAddress, key[20:])

		service.CacheDB.Put(newkey, val)
		service.CacheDB.Delete(key)
	}

	iter.Release()
	if err := iter.Error(); err != nil {
		return err
	}

	return nil
}

func deleteCcntmractStorage(service *WasmVmService) error {
	ccntmractAddress := service.CcntmextRef.CurrentCcntmext().CcntmractAddress
	iter := service.CacheDB.NewIterator(ccntmractAddress[:])

	for has := iter.First(); has; has = iter.Next() {
		service.CacheDB.Delete(iter.Key())
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		return err
	}

	service.CacheDB.DeleteCcntmract(ccntmractAddress)
	return nil
}

func CcntmractCreate(proc *exec.Process,
	codePtr uint32,
	codeLen uint32,
	vmType uint32,
	namePtr uint32,
	nameLen uint32,
	verPtr uint32,
	verLen uint32,
	authorPtr uint32,
	authorLen uint32,
	emailPtr uint32,
	emailLen uint32,
	descPtr uint32,
	descLen uint32,
	newAddressPtr uint32) uint32 {
	self := proc.HostData().(*Runtime)
	code, err := ReadWasmMemory(proc, codePtr, codeLen)
	if err != nil {
		panic(err)
	}

	cost := CcntmRACT_CREATE_GAS + uint64(uint64(codeLen)/PER_UNIT_CODE_LEN)*UINT_DEPLOY_CODE_LEN_GAS
	self.checkGas(cost)

	name, err := ReadWasmMemory(proc, namePtr, nameLen)
	if err != nil {
		panic(err)
	}

	version, err := ReadWasmMemory(proc, verPtr, verLen)
	if err != nil {
		panic(err)
	}

	author, err := ReadWasmMemory(proc, authorPtr, authorLen)
	if err != nil {
		panic(err)
	}

	email, err := ReadWasmMemory(proc, emailPtr, emailLen)
	if err != nil {
		panic(err)
	}

	desc, err := ReadWasmMemory(proc, descPtr, descLen)
	if err != nil {
		panic(err)
	}

	dep, err := payload.CreateDeployCode(code, vmType, name, version, author, email, desc)
	if err != nil {
		panic(err)
	}

	wasmCode, err := dep.GetWasmCode()
	if err != nil {
		panic(err)
	}
	_, err = ReadWasmModule(wasmCode, config.DefConfig.Common.WasmVerifyMethod)
	if err != nil {
		panic(err)
	}

	ccntmractAddr := dep.Address()
	if self.isCcntmractExist(ccntmractAddr) {
		panic(errors.NewErr("ccntmract has been deployed"))
	}

	self.Service.CacheDB.PutCcntmract(dep)

	length, err := proc.WriteAt(ccntmractAddr[:], int64(newAddressPtr))
	if err != nil {
		panic(err)
	}
	return uint32(length)

}

func CcntmractMigrate(proc *exec.Process,
	codePtr uint32,
	codeLen uint32,
	vmType uint32,
	namePtr uint32,
	nameLen uint32,
	verPtr uint32,
	verLen uint32,
	authorPtr uint32,
	authorLen uint32,
	emailPtr uint32,
	emailLen uint32,
	descPtr uint32,
	descLen uint32,
	newAddressPtr uint32) uint32 {

	self := proc.HostData().(*Runtime)

	code, err := ReadWasmMemory(proc, codePtr, codeLen)
	if err != nil {
		panic(err)
	}

	cost := CcntmRACT_CREATE_GAS + uint64(uint64(codeLen)/PER_UNIT_CODE_LEN)*UINT_DEPLOY_CODE_LEN_GAS
	self.checkGas(cost)

	name, err := ReadWasmMemory(proc, namePtr, nameLen)
	if err != nil {
		panic(err)
	}

	version, err := ReadWasmMemory(proc, verPtr, verLen)
	if err != nil {
		panic(err)
	}

	author, err := ReadWasmMemory(proc, authorPtr, authorLen)
	if err != nil {
		panic(err)
	}

	email, err := ReadWasmMemory(proc, emailPtr, emailLen)
	if err != nil {
		panic(err)
	}

	desc, err := ReadWasmMemory(proc, descPtr, descLen)
	if err != nil {
		panic(err)
	}

	dep, err := payload.CreateDeployCode(code, vmType, name, version, author, email, desc)
	if err != nil {
		panic(err)
	}

	wasmCode, err := dep.GetWasmCode()
	if err != nil {
		panic(err)
	}
	_, err = ReadWasmModule(wasmCode, config.DefConfig.Common.WasmVerifyMethod)
	if err != nil {
		panic(err)
	}

	ccntmractAddr := dep.Address()
	if self.isCcntmractExist(ccntmractAddr) {
		panic(errors.NewErr("ccntmract has been deployed"))
	}
	self.Service.CacheDB.PutCcntmract(dep)

	err = migrateCcntmractStorage(self.Service, ccntmractAddr)
	if err != nil {
		panic(err)
	}

	length, err := proc.WriteAt(ccntmractAddr[:], int64(newAddressPtr))
	if err != nil {
		panic(err)
	}

	return uint32(length)
}

func CcntmractDestroy(proc *exec.Process) {
	self := proc.HostData().(*Runtime)
	err := deleteCcntmractStorage(self.Service)
	if err != nil {
		panic(err)
	}
	//the ccntmract has been deleted ,quit the ccntmract operation
	proc.Terminate()
}

func (self *Runtime) isCcntmractExist(ccntmractAddress common.Address) bool {
	item, err := self.Service.CacheDB.GetCcntmract(ccntmractAddress)
	if err != nil {
		panic(err)
	}
	return item != nil
}
