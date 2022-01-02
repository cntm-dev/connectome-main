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
	. "github.com/Ontology/vm/neovm/errors"
	"github.com/Ontology/common/log"
	"fmt"
)

func opNop(e *ExecutionEngine) (VMState, error) {
	return NONE, nil
}

func opJmp(e *ExecutionEngine) (VMState, error) {
	offset := int(e.ccntmext.OpReader.ReadInt16())

	offset = e.ccntmext.GetInstructionPointer() + offset - 3

	if offset < 0 || offset > len(e.ccntmext.Code) {
		log.Error(fmt.Sprintf("[opJmp] offset:%v > e.ccntmex.Code len:%v error", offset, len(e.ccntmext.Code)))
		return FAULT, ERR_FAULT
	}
	var fValue = true

	if e.opCode > JMP {
		if EvaluationStackCount(e) < 1 {
			log.Error(fmt.Sprintf("[opJmp] stack count:%v > 1 error", EvaluationStackCount(e)))
			return FAULT, ERR_UNDER_STACK_LEN
		}
		fValue = PopBoolean(e)
		if e.opCode == JMPIFNOT {
			fValue = !fValue
		}
	}

	if fValue {
		e.ccntmext.SetInstructionPointer(int64(offset))
	}
	return NONE, nil
}

func opCall(e *ExecutionEngine) (VMState, error) {

	e.invocationStack.Push(e.ccntmext.Clone())
	e.ccntmext.SetInstructionPointer(int64(e.ccntmext.GetInstructionPointer() + 2))
	e.opCode = JMP
	ccntmext, err := e.CurrentCcntmext()
	if err != nil {
		return FAULT, err
	}
	e.ccntmext = ccntmext
	return opJmp(e)
}

func opRet(e *ExecutionEngine) (VMState, error) {
	e.invocationStack.Pop()
	return NONE, nil
}

func opAppCall(e *ExecutionEngine) (VMState, error) {
	codeHash := e.ccntmext.OpReader.ReadBytes(20)
	if len(codeHash) == 0 {
		codeHash = PopByteArray(e)
	}

	code, err := e.table.GetCode(codeHash)
	if code == nil {
		return FAULT, err
	}

	if e.opCode == TAILCALL {
		e.invocationStack.Pop()
	}
	e.LoadCode(code, false)
	return NONE, nil
}

func opSysCall(e *ExecutionEngine) (VMState, error) {
	s := e.ccntmext.OpReader.ReadVarString()

	success, err := e.service.Invoke(s, e)
	if success {
		return NONE, nil
	} else {
		return FAULT, err
	}
}
