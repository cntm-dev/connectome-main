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
	vm "github.com/cntmio/cntmology/vm/neovm"
	"github.com/cntmio/cntmology/errors"
)

// GetCodeCcntmainer push current transaction to vm stack
func GetCodeCcntmainer(service *NeoVmService, engine *vm.ExecutionEngine) error {
	vm.PushData(engine, service.Tx)
	return nil
}

// GetExecutingAddress push current ccntmext to vm stack
func GetExecutingAddress(service *NeoVmService, engine *vm.ExecutionEngine) error {
	ccntmext := service.CcntmextRef.CurrentCcntmext(); if ccntmext == nil {
		return errors.NewErr("Current ccntmext invalid")
	}
	vm.PushData(engine, ccntmext.CcntmractAddress[:])
	return nil
}

// GetExecutingAddress push previous ccntmext to vm stack
func GetCallingAddress(service *NeoVmService, engine *vm.ExecutionEngine) error {
	ccntmext := service.CcntmextRef.CallingCcntmext(); if ccntmext == nil {
		return errors.NewErr("Calling ccntmext invalid")
	}
	vm.PushData(engine, ccntmext.CcntmractAddress[:])
	return nil
}

// GetExecutingAddress push entry call ccntmext to vm stack
func GetEntryAddress(service *NeoVmService, engine *vm.ExecutionEngine) error {
	ccntmext := service.CcntmextRef.EntryCcntmext(); if ccntmext == nil {
		return errors.NewErr("Entry ccntmext invalid")
	}
	vm.PushData(engine, ccntmext.CcntmractAddress[:])
	return nil
}

