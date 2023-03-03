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

// AttributeGetUsage put attribute's usage to vm stack
func AttributeGetUsage(service *CntmVmService, engine *vm.Executor) error {
	i, err := engine.EvalStack.PopAsInteropValue()
	if err != nil {
		return err
	}
	if a, ok := i.Data.(*types.TxAttribute); ok {
		return engine.EvalStack.PushInt64(int64(a.Usage))
	}
	return errors.NewErr("[AttributeGetUsage] Wrong type!")
}

// AttributeGetData put attribute's data to vm stack
func AttributeGetData(service *CntmVmService, engine *vm.Executor) error {
	i, err := engine.EvalStack.PopAsInteropValue()
	if err != nil {
		return err
	}
	if a, ok := i.Data.(*types.TxAttribute); ok {
		val, err := vmtypes.VmValueFromBytes(a.Data)
		if err != nil {
			return err
		}
		return engine.EvalStack.Push(val)
	}
	return errors.NewErr("[AttributeGetData] Wrong type!")
}
