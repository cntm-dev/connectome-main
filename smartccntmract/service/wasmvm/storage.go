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
	"errors"
	"math"

	"github.com/cntmio/cntmology/core/states"
	"github.com/cntmio/wagon/exec"
)

func StorageRead(proc *exec.Process, keyPtr uint32, klen uint32, val uint32, vlen uint32, offset uint32) uint32 {
	self := proc.HostData().(*Runtime)
	self.checkGas(STORAGE_GET_GAS)
	keybytes, err := ReadWasmMemory(proc, keyPtr, klen)
	if err != nil {
		panic(err)
	}

	key := serializeStorageKey(self.Service.CcntmextRef.CurrentCcntmext().CcntmractAddress, keybytes)

	raw, err := self.Service.CacheDB.Get(key)
	if err != nil {
		panic(err)
	}

	if raw == nil {
		return math.MaxUint32
	}

	item, err := states.GetValueFromRawStorageItem(raw)
	if err != nil {
		panic(err)
	}

	length := vlen
	itemlen := uint32(len(item))
	if itemlen < vlen {
		length = itemlen
	}

	if uint32(len(item)) < offset {
		panic(errors.New("offset is invalid"))
	}
	_, err = proc.WriteAt(item[offset:offset+length], int64(val))

	if err != nil {
		panic(err)
	}
	return uint32(len(item))
}

func StorageWrite(proc *exec.Process, keyPtr uint32, keyLen uint32, valPtr uint32, valLen uint32) {
	self := proc.HostData().(*Runtime)
	keybytes, err := ReadWasmMemory(proc, keyPtr, keyLen)
	if err != nil {
		panic(err)
	}

	valbytes, err := ReadWasmMemory(proc, valPtr, valLen)
	if err != nil {
		panic(err)
	}

	cost := uint64((len(keybytes)+len(valbytes)-1)/1024+1) * STORAGE_PUT_GAS
	self.checkGas(cost)

	key := serializeStorageKey(self.Service.CcntmextRef.CurrentCcntmext().CcntmractAddress, keybytes)

	self.Service.CacheDB.Put(key, states.GenRawStorageItem(valbytes))
}

func StorageDelete(proc *exec.Process, keyPtr uint32, keyLen uint32) {
	self := proc.HostData().(*Runtime)
	self.checkGas(STORAGE_DELETE_GAS)
	keybytes, err := ReadWasmMemory(proc, keyPtr, keyLen)
	if err != nil {
		panic(err)
	}
	key := serializeStorageKey(self.Service.CcntmextRef.CurrentCcntmext().CcntmractAddress, keybytes)

	self.Service.CacheDB.Delete(key)
}
