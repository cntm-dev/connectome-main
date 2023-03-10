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

	"github.com/cntmio/cntmology/vm/neovm/utils"
)

type ExecutionCcntmext struct {
	Code               []byte
	OpReader           *utils.VmReader
	InstructionPointer int
	vmFlags            VmFeatureFlag
}

func NewExecutionCcntmext(code []byte, flag VmFeatureFlag) *ExecutionCcntmext {
	var ccntmext ExecutionCcntmext
	ccntmext.Code = code
	ccntmext.OpReader = utils.NewVmReader(code)
	ccntmext.OpReader.AllowEOF = flag.AllowReaderEOF
	ccntmext.vmFlags = flag

	ccntmext.InstructionPointer = 0
	return &ccntmext
}

func (ec *ExecutionCcntmext) GetInstructionPointer() int {
	return ec.OpReader.Position()
}

func (ec *ExecutionCcntmext) SetInstructionPointer(offset int64) error {
	_, err := ec.OpReader.Seek(offset, io.SeekStart)
	return err
}

func (ec *ExecutionCcntmext) NextInstruction() OpCode {
	return OpCode(ec.Code[ec.OpReader.Position()])
}

func (self *ExecutionCcntmext) ReadOpCode() (val OpCode, eof bool) {
	code, err := self.OpReader.ReadByte()
	if err != nil {
		eof = true
		return
	}
	val = OpCode(code)
	return val, false
}

func (ec *ExecutionCcntmext) Clone() *ExecutionCcntmext {
	ccntmext := NewExecutionCcntmext(ec.Code, ec.vmFlags)
	ccntmext.InstructionPointer = ec.InstructionPointer
	_ = ccntmext.SetInstructionPointer(int64(ec.GetInstructionPointer()))
	return ccntmext
}
