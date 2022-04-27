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
	"github.com/cntmio/cntmology/vm/wasmvm/exec"
	"github.com/cntmio/cntmology/errors"
	"github.com/cntmio/cntmology/core/types"
	"bytes"
)

func (this * WasmVmService)attributeGetUsage(engine *exec.ExecutionEngine)(bool,error){
	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 1 {
		return false, errors.NewErr("[transactionGetHash] parameter count error")
	}

	attributebytes ,err:= vm.GetPointerMemory(params[0])
	if err != nil{
		return false,nil
	}

	attr := types.TxAttribute{}
	err = attr.Deserialize(bytes.NewBuffer(attributebytes))
	if err != nil{
		return false,nil
	}
	vm.RestoreCtx()
	vm.PushResult(uint64(attr.Usage))
	return true,nil
}
func (this * WasmVmService)attributeGetData(engine *exec.ExecutionEngine)(bool,error){
	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 1 {
		return false, errors.NewErr("[transactionGetHash] parameter count error")
	}

	attributebytes ,err:= vm.GetPointerMemory(params[0])
	if err != nil{
		return false,nil
	}

	attr := types.TxAttribute{}
	err = attr.Deserialize(bytes.NewBuffer(attributebytes))
	if err != nil{
		return false,nil
	}

	idx,err := vm.SetPointerMemory(attr.Data)
	if err != nil{
		return false,nil
	}

	vm.RestoreCtx()
	vm.PushResult(uint64(idx))
	return true,nil
}