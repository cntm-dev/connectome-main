package neovm

import (
	. "github.com/Ontology/vm/neovm/errors"
)

func opNop(e *ExecutionEngine) (VMState, error) {
	return NONE, nil
}

func opJmp(e *ExecutionEngine) (VMState, error) {
	offset := int(e.ccntmext.OpReader.ReadInt16())

	offset = e.ccntmext.GetInstructionPointer() + offset - 3

	if offset > len(e.ccntmext.Code) {
		return FAULT, ErrFault
	}
	var fValue = true

	if e.opCode == JMPIF {
		fValue = PopBoolean(e)
	} else if e.opCode == JMPIFNOT {
		fValue = !PopBoolean(e)
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
