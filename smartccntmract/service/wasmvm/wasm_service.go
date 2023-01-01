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
	"github.com/hashicorp/golang-lru"
	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/core/store"
	"github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/errors"
	"github.com/cntmio/cntmology/smartccntmract/ccntmext"
	"github.com/cntmio/cntmology/smartccntmract/event"
	"github.com/cntmio/cntmology/smartccntmract/states"
	"github.com/cntmio/cntmology/smartccntmract/storage"
)

type WasmVmService struct {
	Store         store.LedgerStore
	CacheDB       *storage.CacheDB
	CcntmextRef    ccntmext.CcntmextRef
	Notifications []*event.NotifyEventInfo
	Code          []byte
	Tx            *types.Transaction
	Time          uint32
	Height        uint32
	BlockHash     common.Uint256
	PreExec       bool
	GasPrice      uint64
	GasLimit      *uint64
	ExecStep      *uint64
	GasFactor     uint64
	IsTerminate   bool
	vm            *exec.VM
}

var (
	ERR_CHECK_STACK_SIZE  = errors.NewErr("[WasmVmService] vm over max stack size!")
	ERR_EXECUTE_CODE      = errors.NewErr("[WasmVmService] vm execute code invalid!")
	ERR_GAS_INSUFFICIENT  = errors.NewErr("[WasmVmService] gas insufficient")
	VM_EXEC_STEP_EXCEED   = errors.NewErr("[WasmVmService] vm execute step exceed!")
	CcntmRACT_NOT_EXIST    = errors.NewErr("[WasmVmService] Get ccntmract code from db fail")
	DEPLOYCODE_TYPE_ERROR = errors.NewErr("[WasmVmService] DeployCode type error!")
	VM_EXEC_FAULT         = errors.NewErr("[WasmVmService] vm execute state fault!")
	VM_INIT_FAULT         = errors.NewErr("[WasmVmService] vm init state fault!")

	CODE_CACHE_SIZE      = 100
	CcntmRACT_METHOD_NAME = "invoke"

	//max memory size of wasm vm
	WASM_MEM_LIMITATION  uint64 = 10 * 1024 * 1024
	VM_STEP_LIMIT               = 40000000
	WASM_CALLSTACK_LIMIT        = 1024

	CodeCache *lru.ARCCache
)

func init() {
	CodeCache, _ = lru.NewARC(CODE_CACHE_SIZE)
	//if err != nil{
	//	log.Info("NewARC block error %s", err)
	//}
}

func (this *WasmVmService) Invoke() (interface{}, error) {
	if len(this.Code) == 0 {
		return nil, ERR_EXECUTE_CODE
	}

	ccntmract := &states.WasmCcntmractParam{}
	sink := common.NewZeroCopySource(this.Code)
	err := ccntmract.Deserialization(sink)
	if err != nil {
		return nil, err
	}

	code, err := this.CacheDB.GetCcntmract(ccntmract.Address)
	if err != nil {
		return nil, err
	}

	if code == nil {
		return nil, errors.NewErr("wasm ccntmract does not exist")
	}

	wasmCode, err := code.GetWasmCode()
	if err != nil {
		return nil, errors.NewErr("not a wasm ccntmract")
	}
	this.CcntmextRef.PushCcntmext(&ccntmext.Ccntmext{CcntmractAddress: ccntmract.Address, Code: wasmCode})
	host := &Runtime{Service: this, Input: ccntmract.Args}

	var compiled *exec.CompiledModule
	if CodeCache != nil {
		cached, ok := CodeCache.Get(ccntmract.Address.ToHexString())
		if ok {
			compiled = cached.(*exec.CompiledModule)
		}
	}

	if compiled == nil {
		compiled, err = ReadWasmModule(wasmCode, false)
		if err != nil {
			return nil, err
		}
		CodeCache.Add(ccntmract.Address.ToHexString(), compiled)
	}

	vm, err := exec.NewVMWithCompiled(compiled, WASM_MEM_LIMITATION)
	if err != nil {
		return nil, VM_INIT_FAULT
	}

	vm.HostData = host

	vm.AvaliableGas = &exec.Gas{GasLimit: this.GasLimit, LocalGasCounter: 0, GasPrice: this.GasPrice, GasFactor: this.GasFactor, ExecStep: this.ExecStep}
	vm.CallStackDepth = uint32(WASM_CALLSTACK_LIMIT)
	vm.RecoverPanic = true

	entryName := CcntmRACT_METHOD_NAME

	entry, ok := compiled.RawModule.Export.Entries[entryName]

	if ok == false {
		return nil, errors.NewErr("[Call]Method:" + entryName + " does not exist!")
	}

	//get entry index
	index := int64(entry.Index)

	//get function index
	fidx := compiled.RawModule.Function.Types[int(index)]

	//get  function type
	ftype := compiled.RawModule.Types.Entries[int(fidx)]

	//no returns of the entry function
	if len(ftype.ReturnTypes) > 0 {
		return nil, errors.NewErr("[Call]ExecCode error! Invoke function sig error")
	}

	//no args for passed in, all args in runtime input buffer
	this.vm = vm

	_, err = vm.ExecCode(index)

	if err != nil {
		return nil, errors.NewErr("[Call]ExecCode error!" + err.Error())
	}

	//pop the current ccntmext
	this.CcntmextRef.PopCcntmext()

	return host.Output, nil
}
