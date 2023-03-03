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
package cntmvm

import (
	"github.com/conntectome/cntm/core/types"
	"github.com/conntectome/cntm/errors"
	vm "github.com/conntectome/cntm/vm/cntmvm"
	vmtypes "github.com/conntectome/cntm/vm/cntmvm/types"
)

// BlockGetTransactionCount put block's transactions count to vm stack
func BlockGetTransactionCount(service *CntmVmService, engine *vm.Executor) error {
	i, err := engine.EvalStack.PopAsInteropValue()
	if err != nil {
		return err
	}
	if block, ok := i.Data.(*types.Block); ok {
		val := vmtypes.VmValueFromInt64(int64(len(block.Transactions)))
		return engine.EvalStack.Push(val)
	}
	return errors.NewErr("[BlockGetTransactionCount] Wrong type ")
}

// BlockGetTransactions put block's transactions to vm stack
func BlockGetTransactions(service *CntmVmService, engine *vm.Executor) error {
	i, err := engine.EvalStack.PopAsInteropValue()
	if err != nil {
		return err
	}
	if block, ok := i.Data.(*types.Block); ok {
		transactions := block.Transactions
		transactionList := make([]vmtypes.VmValue, 0)

		for _, v := range transactions {
			transactionList = append(transactionList, vmtypes.VmValueFromInteropValue(vmtypes.NewInteropValue(v)))
		}

		return engine.EvalStack.PushAsArray(transactionList)
	}
	return errors.NewErr("[BlockGetTransactions] Wrong type ")
}

// BlockGetTransaction put block's transaction to vm stack
func BlockGetTransaction(service *CntmVmService, engine *vm.Executor) error {
	i, err := engine.EvalStack.PopAsInteropValue()
	if err != nil {
		return err
	}
	index, err := engine.EvalStack.PopAsInt64()
	if err != nil {
		return err
	}
	if block, ok := i.Data.(*types.Block); ok {
		if index < 0 || int(index) >= len(block.Transactions) {
			return errors.NewErr("[BlockGetTransaction] index out of bounds")
		}
		return engine.EvalStack.PushAsInteropValue(block.Transactions[index])

	}
	return errors.NewErr("[BlockGetTransaction] Wrong type ")
}
