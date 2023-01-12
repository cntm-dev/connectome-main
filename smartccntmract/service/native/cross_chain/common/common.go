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
package common

import (
	"fmt"
	"math/big"

	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/log"
	ctypes "github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/errors"
	"github.com/cntmio/cntmology/smartccntmract/service/native"
	"github.com/cntmio/cntmology/smartccntmract/service/neovm"
	ntypes "github.com/cntmio/cntmology/vm/neovm/types"
)

func CrossChainNeoVMCall(this *native.NativeService, address common.Address, method string, args []byte,
	fromCcntmractAddress []byte, fromChainID uint64) (interface{}, error) {
	dep, err := this.CacheDB.GetCcntmract(address)
	if err != nil {
		return nil, errors.NewErr("[NeoVMCall] Get ccntmract ccntmext error!")
	}
	log.Debugf("[NeoVMCall] native invoke neovm ccntmract address:%s", address.ToHexString())
	if dep == nil {
		return nil, errors.NewErr("[NeoVMCall] native invoke neovm ccntmract is nil")
	}
	m, err := ntypes.VmValueFromBytes([]byte(method))
	if err != nil {
		return nil, err
	}
	array := ntypes.NewArrayValue()
	a, err := ntypes.VmValueFromBytes(args)
	if err != nil {
		return nil, err
	}
	if err := array.Append(a); err != nil {
		return nil, err
	}
	fca, err := ntypes.VmValueFromBytes(fromCcntmractAddress)
	if err != nil {
		return nil, err
	}
	if err := array.Append(fca); err != nil {
		return nil, err
	}
	fci, err := ntypes.VmValueFromBigInt(new(big.Int).SetUint64(fromChainID))
	if err != nil {
		return nil, err
	}
	if err := array.Append(fci); err != nil {
		return nil, err
	}
	if !this.CcntmextRef.CheckUseGas(neovm.NATIVE_INVOKE_GAS) {
		return nil, fmt.Errorf("[CrossChainNeoVMCall], check use gaslimit insufficient！")
	}
	engine, err := this.CcntmextRef.NewExecuteEngine(dep.GetRawCode(), ctypes.InvokeNeo)
	if err != nil {
		return nil, err
	}
	evalStack := engine.(*neovm.NeoVmService).Engine.EvalStack
	if err := evalStack.Push(ntypes.VmValueFromArrayVal(array)); err != nil {
		return nil, err
	}
	if err := evalStack.Push(m); err != nil {
		return nil, err
	}
	return engine.Invoke()
}
