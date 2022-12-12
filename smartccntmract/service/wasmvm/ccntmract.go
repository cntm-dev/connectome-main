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
	"github.com/go-interpreter/wagon/exec"
	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/core/payload"
	"github.com/cntmio/cntmology/errors"
)

func (self *Runtime) CcntmractCreate(proc *exec.Process,
	codePtr uint32,
	codeLen uint32,
	needStorage uint32,
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

	if uint32(proc.MemAllocated()) < codeLen+nameLen+verLen+authorLen+emailLen+descLen {
		panic(errors.NewErr("ccntmract create len is greater than memory size"))
	}

	code := make([]byte, codeLen)
	_, err := proc.ReadAt(code, int64(codePtr))
	if err != nil {
		panic(err)
	}

	cost := CcntmRACT_CREATE_GAS + uint64(uint64(codeLen)/PER_UNIT_CODE_LEN)*UINT_DEPLOY_CODE_LEN_GAS
	self.checkGas(cost)

	name := make([]byte, nameLen)
	_, err = proc.ReadAt(name, int64(namePtr))
	if err != nil {
		panic(err)
	}

	version := make([]byte, verLen)
	_, err = proc.ReadAt(version, int64(verPtr))
	if err != nil {
		panic(err)
	}

	author := make([]byte, authorLen)
	_, err = proc.ReadAt(author, int64(authorPtr))
	if err != nil {
		panic(err)
	}

	email := make([]byte, emailLen)
	_, err = proc.ReadAt(email, int64(emailPtr))
	if err != nil {
		panic(err)
	}

	desc := make([]byte, descLen)
	_, err = proc.ReadAt(desc, int64(descPtr))
	if err != nil {
		panic(err)
	}

	dep, err := payload.CreateDeployCode(code, needStorage, name, version, author, email, desc)
	if err != nil {
		panic(err)
	}

	ccntmractAddr := dep.Address()
	if self.isCcntmractExist(ccntmractAddr) {
		panic(errors.NewErr("ccntmract has been deployed"))
	}

	err = self.Service.CacheDB.PutCcntmract(dep)
	if err != nil {
		panic(err)
	}

	length, err := proc.WriteAt(ccntmractAddr[:], int64(newAddressPtr))
	return uint32(length)

}

func (self *Runtime) CcntmractMigrate(proc *exec.Process,
	codePtr uint32,
	codeLen uint32,
	needStorage uint32,
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

	if uint32(proc.MemAllocated()) < codeLen+nameLen+verLen+authorLen+emailLen+descLen {
		panic(errors.NewErr("ccntmract migrate len is greater than memory size"))
	}

	code := make([]byte, codeLen)
	_, err := proc.ReadAt(code, int64(codePtr))
	if err != nil {
		panic(err)
	}

	cost := CcntmRACT_CREATE_GAS + uint64(uint64(codeLen)/PER_UNIT_CODE_LEN)*UINT_DEPLOY_CODE_LEN_GAS
	self.checkGas(cost)

	name := make([]byte, nameLen)
	_, err = proc.ReadAt(name, int64(namePtr))
	if err != nil {
		panic(err)
	}

	version := make([]byte, verLen)
	_, err = proc.ReadAt(version, int64(verPtr))
	if err != nil {
		panic(err)
	}

	author := make([]byte, authorLen)
	_, err = proc.ReadAt(author, int64(authorPtr))
	if err != nil {
		panic(err)
	}

	email := make([]byte, emailLen)
	_, err = proc.ReadAt(email, int64(emailPtr))
	if err != nil {
		panic(err)
	}

	desc := make([]byte, descLen)
	_, err = proc.ReadAt(desc, int64(descPtr))
	if err != nil {
		panic(err)
	}

	dep, err := payload.CreateDeployCode(code, needStorage, name, version, author, email, desc)
	if err != nil {
		panic(err)
	}

	ccntmractAddr := dep.Address()
	if self.isCcntmractExist(ccntmractAddr) {
		panic(errors.NewErr("ccntmract has been deployed"))
	}
	oldAddress := self.Service.CcntmextRef.CurrentCcntmext().CcntmractAddress

	self.Service.CacheDB.PutCcntmract(dep)
	self.Service.CacheDB.DeleteCcntmract(oldAddress)

	iter := self.Service.CacheDB.NewIterator(oldAddress[:])
	for has := iter.First(); has; has = iter.Next() {
		key := iter.Key()
		val := iter.Value()

		newkey := serializeStorageKey(ccntmractAddr, key)

		self.Service.CacheDB.Put(newkey, val)
		self.Service.CacheDB.Delete(key)
	}

	iter.Release()
	if err := iter.Error(); err != nil {
		panic(err)
	}

	length, err := proc.WriteAt(ccntmractAddr[:], int64(newAddressPtr))
	if err != nil {
		panic(err)
	}

	return uint32(length)
}

func (self *Runtime) CcntmractDelete(proc *exec.Process) {
	ccntmractAddress := self.Service.CcntmextRef.CurrentCcntmext().CcntmractAddress
	iter := self.Service.CacheDB.NewIterator(ccntmractAddress[:])

	for has := iter.First(); has; has = iter.Next() {
		self.Service.CacheDB.Delete(iter.Key())
	}
	iter.Release()
	if err := iter.Error(); err != nil {
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
