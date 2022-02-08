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

package wasm

import (
	"bytes"
	"fmt"

	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/core/states"
	"github.com/cntmio/cntmology/core/store"
	scommon "github.com/cntmio/cntmology/core/store/common"
	"github.com/cntmio/cntmology/errors"
	"github.com/cntmio/cntmology/smartccntmract/storage"
	"github.com/cntmio/cntmology/vm/wasmvm/exec"
	"github.com/cntmio/cntmology/vm/wasmvm/memory"
	"github.com/cntmio/cntmology/vm/wasmvm/util"
	"github.com/cntmio/cntmology/vm/wasmvm/wasm"
	"github.com/cntmio/cntmology/common/log"
	"github.com/cntmio/cntmology/vm/types"
)

type LogLevel byte
const(
	Debug LogLevel = iota
	Info
	Error
)

type WasmStateMachine struct {
	*WasmStateReader
	ldgerStore store.LedgerStore
	CloneCache *storage.CloneCache
	time       uint32
}

func NewWasmStateMachine(ldgerStore store.LedgerStore, dbCache scommon.StateStore, time uint32) *WasmStateMachine {

	var stateMachine WasmStateMachine
	stateMachine.ldgerStore = ldgerStore
	stateMachine.CloneCache = storage.NewCloneCache(dbCache)
	stateMachine.WasmStateReader = NewWasmStateReader(ldgerStore)
	stateMachine.time = time

	stateMachine.Register("PutStorage", stateMachine.putstore)
	stateMachine.Register("GetStorage", stateMachine.getstore)
	stateMachine.Register("DeleteStorage", stateMachine.deletestore)
	stateMachine.Register("CallCcntmract", stateMachine.callCcntmract)
	stateMachine.Register("CcntmractLogDebug", stateMachine.ccntmractLogDebug)
	stateMachine.Register("CcntmractLogInfo", stateMachine.ccntmractLogInfo)
	stateMachine.Register("CcntmractLogError", stateMachine.ccntmractLogError)

	return &stateMachine
}

//======================store apis here============================================
func (s *WasmStateMachine) putstore(engine *exec.ExecutionEngine) (bool, error) {

	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()

	if len(params) != 2 {
		return false, errors.NewErr("[putstore] parameter count error")
	}

	key, err := vm.GetPointerMemory(params[0])
	if err != nil {
		return false, err
	}
	if len(key) > 1024 {
		return false, errors.NewErr("[putstore] Get Storage key to lcntm")
	}

	value, err := vm.GetPointerMemory(params[1])
	if err != nil {
		return false, err
	}
	k, err := serializeStorageKey(vm.CodeHash, key)
	if err != nil {
		return false, err
	}

	s.CloneCache.Add(scommon.ST_STORAGE, k, &states.StorageItem{Value: value})

	vm.RestoreCtx()

	return true, nil
}

func (s *WasmStateMachine) getstore(engine *exec.ExecutionEngine) (bool, error) {

	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()

	if len(params) != 1 {
		return false, errors.NewErr("[getstore] parameter count error ")
	}

	key, err := vm.GetPointerMemory(params[0])
	if err != nil {
		return false, err
	}
	k, err := serializeStorageKey(vm.CodeHash, key)
	if err != nil {
		return false, err
	}
	item, err := s.CloneCache.Get(scommon.ST_STORAGE, k)
	if err != nil {
		return false, err
	}

	if item == nil {
		vm.RestoreCtx()
		if envCall.GetReturns() {
			vm.PushResult(uint64(memory.VM_NIL_POINTER))
		}
		return true, nil
	}

	idx, err := vm.SetPointerMemory(item.(*states.StorageItem).Value)
	if err != nil {
		return false, err
	}

	vm.RestoreCtx()
	if envCall.GetReturns() {
		vm.PushResult(uint64(idx))
	}
	return true, nil
}

func (s *WasmStateMachine) deletestore(engine *exec.ExecutionEngine) (bool, error) {

	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()

	if len(params) != 1 {
		return false, errors.NewErr("[deletestore] parameter count error")
	}

	key, err := vm.GetPointerMemory(params[0])
	if err != nil {
		return false, err
	}

	k, err := serializeStorageKey(vm.CodeHash, key)
	if err != nil {
		return false, err
	}

	s.CloneCache.Delete(scommon.ST_STORAGE, k)
	vm.RestoreCtx()

	return true, nil
}

func (s *WasmStateMachine) GetCcntmractCodeFromAddress(address common.Address) ([]byte, error) {

	dcode, err := s.ldgerStore.GetCcntmractState(address)
	if err != nil {
		return nil, err
	}

	return dcode.Code.Code, nil

}

//call other ccntmract
func (s *WasmStateMachine) callCcntmract(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 3 {
		return false, errors.NewErr("parameter count error while call readMessage")
	}
	ccntmractAddressIdx := params[0]
	addr, err := vm.GetPointerMemory(ccntmractAddressIdx)
	if err != nil {
		return false, errors.NewErr("get Ccntmract address failed")
	}
	//the ccntmract codes
	ccntmractBytes, err := s.getCcntmractFromAddr(addr)
	if err != nil {
		return false, err
	}
	vmcode := types.VmCode{VmType:types.WASMVM,Code:ccntmractBytes}
	codeHash := vmcode.AddressFromVmCode()
	bf := bytes.NewBuffer(ccntmractBytes)

	module, err := wasm.ReadModule(bf, emptyImporter)
	if err != nil {
		return false, errors.NewErr("load Module failed")
	}

	methodName, err := vm.GetPointerMemory(params[1])
	if err != nil {
		return false, errors.NewErr("[callCcntmract]get Ccntmract methodName failed")
	}

	arg, err := vm.GetPointerMemory(params[2])
	if err != nil {
		return false, errors.NewErr("[callCcntmract]get Ccntmract arg failed")
	}
	res, err := vm.CallCcntmract(vm.CodeHash, codeHash, module, methodName, arg)
	if err != nil {
		return false, errors.NewErr("[callCcntmract]CallProductCcntmract failed")
	}
	vm.RestoreCtx()
	if envCall.GetReturns() {
		vm.PushResult(uint64(res))
	}
	return true, nil
}

func (s *WasmStateMachine) ccntmractLogDebug(engine *exec.ExecutionEngine) (bool, error) {
	 _ ,err := ccntmractLog(Debug,engine)
	 if err!= nil{
	 	return false,err
	 }

	engine.GetVM().RestoreCtx()
	return true, nil
}

func (s *WasmStateMachine) ccntmractLogInfo(engine *exec.ExecutionEngine) (bool, error) {
	_ ,err := ccntmractLog(Info,engine)
	if err!= nil{
		return false,err
	}

	engine.GetVM().RestoreCtx()
	return true, nil
}

func (s *WasmStateMachine) ccntmractLogError(engine *exec.ExecutionEngine) (bool, error) {
	_ ,err := ccntmractLog(Error,engine)
	if err!= nil{
		return false,err
	}

	engine.GetVM().RestoreCtx()
	return true, nil
}



func ccntmractLog(lv LogLevel,engine *exec.ExecutionEngine ) (bool, error){
	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 1 {
		return false, errors.NewErr("parameter count error while call ccntmractLcntm")
	}

	Idx := params[0]
	addr, err := vm.GetPointerMemory(Idx)
	if err != nil {
		return false, errors.NewErr("get Ccntmract address failed")
	}

	msg := fmt.Sprintf("[WASM Ccntmract] Address:%s message:%s",vm.CodeHash.ToHexString(),util.TrimBuffToString(addr))

	switch lv {
	case Debug:
		log.Debug(msg)
	case Info:
		log.Info(msg)
	case Error:
		log.Error(msg)
	}
	return true ,nil

}

func serializeStorageKey(codeHash common.Address, key []byte) ([]byte, error) {
	bf := new(bytes.Buffer)
	storageKey := &states.StorageKey{CodeHash: codeHash, Key: key}
	if _, err := storageKey.Serialize(bf); err != nil {
		return []byte{}, errors.NewErr("[serializeStorageKey] StorageKey serialize error!")
	}
	return bf.Bytes(), nil
}

func (s *WasmStateMachine) getCcntmractFromAddr(addr []byte) ([]byte, error) {

	//just for test
	/*	ccntmract := util.TrimBuffToString(addr)
		code, err := ioutil.ReadFile(fmt.Sprintf("./testdata2/%s.wasm", ccntmract))
		if err != nil {
			return nil, err
		}

		return code, nil*/
	//Fixme get the ccntmract code from ledger
	addrbytes, err := common.HexToBytes(util.TrimBuffToString(addr))
	if err != nil {
		return nil, errors.NewErr("get ccntmract address error")
	}
	ccntmactaddress, err := common.AddressParseFromBytes(addrbytes)
	if err != nil {
		return nil, errors.NewErr("get ccntmract address error")
	}
	dpcode, err := s.GetCcntmractCodeFromAddress(ccntmactaddress)
	if err != nil {
		return nil, errors.NewErr("get ccntmract  error")
	}
	return dpcode, nil
}

func emptyImporter(name string) (*wasm.Module, error) {
	return nil, nil
}
