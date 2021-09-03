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
		return FAULT, ErrFault
	}
	var fValue = true

	if e.opCode > JMP {
		if EvaluationStackCount(e) < 1 {
			log.Error(fmt.Sprintf("[opJmp] stack count:%v > 1 error", EvaluationStackCount(e)))
			return FAULT, ErrUnderStackLen
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
	e.ccntmext = e.CurrentCcntmext()
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
