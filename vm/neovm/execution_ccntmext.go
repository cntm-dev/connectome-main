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
	"io"

	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/vm/neovm/types"
	"github.com/cntmio/cntmology/vm/neovm/utils"
	vmtypes "github.com/cntmio/cntmology/vm/types"
)

type ExecutionCcntmext struct {
	Code               []byte
	OpReader           *utils.VmReader
	PushOnly           bool
	BreakPoints        []uint
	InstructionPointer int
	CodeHash           common.Address
	engine             *ExecutionEngine
}

func NewExecutionCcntmext(engine *ExecutionEngine, code []byte, pushOnly bool, breakPoints []uint) *ExecutionCcntmext {
	var executionCcntmext ExecutionCcntmext
	executionCcntmext.Code = code
	executionCcntmext.OpReader = utils.NewVmReader(code)
	executionCcntmext.PushOnly = pushOnly
	executionCcntmext.BreakPoints = breakPoints
	executionCcntmext.InstructionPointer = 0
	executionCcntmext.engine = engine
	return &executionCcntmext
}

func (ec *ExecutionCcntmext) GetInstructionPointer() int {
	return ec.OpReader.Position()
}

func (ec *ExecutionCcntmext) SetInstructionPointer(offset int64) {
	ec.OpReader.Seek(offset, io.SeekStart)
}

func (ec *ExecutionCcntmext) GetCodeHash() (common.Address, error) {
	empty := common.Address{}
	if ec.CodeHash == empty {
		code := &vmtypes.VmCode{
			Code:   ec.Code,
			VmType: vmtypes.NEOVM,
		}
		ec.CodeHash = code.AddressFromVmCode()
	}
	return ec.CodeHash, nil
}

func (ec *ExecutionCcntmext) NextInstruction() OpCode {
	return OpCode(ec.Code[ec.OpReader.Position()])
}

func (ec *ExecutionCcntmext) Clone() *ExecutionCcntmext {
	executionCcntmext := NewExecutionCcntmext(ec.engine, ec.Code, ec.PushOnly, ec.BreakPoints)
	executionCcntmext.InstructionPointer = ec.InstructionPointer
	executionCcntmext.SetInstructionPointer(int64(ec.GetInstructionPointer()))
	return executionCcntmext
}

func (ec *ExecutionCcntmext) GetStackItem() types.StackItems {
	return nil
}

func (ec *ExecutionCcntmext) GetExecutionCcntmext() *ExecutionCcntmext {
	return ec
}
