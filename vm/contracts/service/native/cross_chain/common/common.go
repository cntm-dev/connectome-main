/*
 * Copyright (C) 2018 The cntm Authors
 * This file is part of The cntm library.
 *
 * The cntm is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The cntm is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The cntm.  If not, see <http://www.gnu.org/licenses/>.
 */
package common

import (
	"fmt"
	"math/big"

	"github.com/conntectome/cntm/common"
	"github.com/conntectome/cntm/common/log"
	ctypes "github.com/conntectome/cntm/core/types"
	"github.com/conntectome/cntm/errors"
	"github.com/conntectome/cntm/smartcontract/service/native"
	"github.com/conntectome/cntm/smartcontract/service/cntmvm"
	ntypes "github.com/conntectome/cntm/vm/cntmvm/types"
)

func CrossChainCntmVMCall(this *native.NativeService, address common.Address, method string, args []byte,
	fromContractAddress []byte, fromChainID uint64) (interface{}, error) {
	dep, err := this.CacheDB.GetContract(address)
	if err != nil {
		return nil, errors.NewErr("[CntmVMCall] Get contract context error!")
	}
	log.Debugf("[CntmVMCall] native invoke cntmvm contract address:%s", address.ToHexString())
	if dep == nil {
		return nil, errors.NewErr("[CntmVMCall] native invoke cntmvm contract is nil")
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
	fca, err := ntypes.VmValueFromBytes(fromContractAddress)
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
	if !this.ContextRef.CheckUseGas(cntmvm.NATIVE_INVOKE_GAS) {
		return nil, fmt.Errorf("[CrossChainCntmVMCall], check use gaslimit insufficient！")
	}
	engine, err := this.ContextRef.NewExecuteEngine(dep.GetRawCode(), ctypes.InvokeCntm)
	if err != nil {
		return nil, err
	}
	evalStack := engine.(*cntmvm.CntmVmService).Engine.EvalStack
	if err := evalStack.Push(ntypes.VmValueFromArrayVal(array)); err != nil {
		return nil, err
	}
	if err := evalStack.Push(m); err != nil {
		return nil, err
	}
	return engine.Invoke()
}
