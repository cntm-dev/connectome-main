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
	"reflect"

	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/core/payload"
	"github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/core/utils"
	"github.com/cntmio/cntmology/vm/crossvm_codec"
	vm "github.com/cntmio/cntmology/vm/neovm"
)

//neovm ccntmract call wasmvm ccntmract
func WASMInvoke(service *NeoVmService, engine *vm.Executor) error {
	address, err := engine.EvalStack.PopAsBytes()
	if err != nil {
		return err
	}

	ccntmractAddress, err := common.AddressParseFromBytes(address)
	if err != nil {
		return fmt.Errorf("invoke wasm ccntmract:%s, address invalid", address)
	}

	dp, err := service.CacheDB.GetCcntmract(ccntmractAddress)
	if err != nil {
		return err
	}
	if dp == nil {
		return fmt.Errorf("wasm ccntmract does not exist")
	}

	if dp.VmType() != payload.WASMVM_TYPE {
		return fmt.Errorf("not a wasm ccntmract")
	}

	parambytes, err := engine.EvalStack.PopAsBytes()
	if err != nil {
		return err
	}
	list, err := crossvm_codec.DeserializeCallParam(parambytes)
	if err != nil {
		return err
	}

	params, ok := list.([]interface{})
	if !ok {
		return fmt.Errorf("wasm invoke error: wrcntm param type:%s", reflect.TypeOf(list).String())
	}

	inputs, err := utils.BuildWasmVMInvokeCode(ccntmractAddress, params)
	if err != nil {
		return err
	}

	newservice, err := service.CcntmextRef.NewExecuteEngine(inputs, types.InvokeWasm)
	if err != nil {
		return err
	}

	tmpRes, err := newservice.Invoke()
	if err != nil {
		return err
	}

	return engine.EvalStack.PushBytes(tmpRes.([]byte))
}
