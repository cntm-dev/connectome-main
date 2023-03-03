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
	"fmt"

	"github.com/conntectome/cntm/core/types"
	vm "github.com/conntectome/cntm/vm/cntmvm"
	vmtypes "github.com/conntectome/cntm/vm/cntmvm/types"
)

// GetExecutingAddress push transaction's hash to vm stack
func TransactionGetHash(service *CntmVmService, engine *vm.Executor) error {
	txn, err := engine.EvalStack.PopAsInteropValue()
	if err != nil {
		return fmt.Errorf("[TransactionGetHash] PopAsInteropValue error:%s", err)
	}
	if tx, ok := txn.Data.(*types.Transaction); ok {
		txHash := tx.Hash()
		return engine.EvalStack.PushBytes(txHash.ToArray())
	}
	return fmt.Errorf("[TransactionGetHash] Type error")
}

// TransactionGetType push transaction's type to vm stack
func TransactionGetType(service *CntmVmService, engine *vm.Executor) error {
	txn, err := engine.EvalStack.PopAsInteropValue()
	if err != nil {
		return fmt.Errorf("[TransactionGetType] PopAsInteropValue error:%s", err)
	}
	if tx, ok := txn.Data.(*types.Transaction); ok {
		return engine.EvalStack.PushInt64(int64(tx.TxType))
	}
	return fmt.Errorf("[TransactionGetType] Type error")
}

// TransactionGetAttributes push transaction's attributes to vm stack
func TransactionGetAttributes(service *CntmVmService, engine *vm.Executor) error {
	_, err := engine.EvalStack.PopAsInteropValue()
	if err != nil {
		return fmt.Errorf("[TransactionGetAttributes] PopAsInteropValue error: %s", err)
	}
	attributList := make([]vmtypes.VmValue, 0)
	return engine.EvalStack.PushAsArray(attributList)
}
