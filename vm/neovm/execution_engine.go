// Copyright 2017 The Ontology Authors
// This file is part of the Ontology library.
//
// The Ontology library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Ontology library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// alcntm with the Ontology library. If not, see <http://www.gnu.org/licenses/>.

package neovm

import (
	"github.com/Ontology/vm/neovm/interfaces"
	"io"
	_ "math/big"
	_ "sort"
	. "github.com/Ontology/vm/neovm/errors"
	"github.com/Ontology/common"
	"github.com/Ontology/common/log"
)

func NewExecutionEngine(ccntmainer interfaces.ICodeCcntmainer, crypto interfaces.ICrypto, table interfaces.ICodeTable, service IInteropService) *ExecutionEngine {
	var engine ExecutionEngine

	engine.crypto = crypto
	engine.table = table

	engine.codeCcntmainer = ccntmainer
	engine.invocationStack = NewRandAccessStack()
	engine.opCount = 0

	engine.evaluationStack = NewRandAccessStack()
	engine.altStack = NewRandAccessStack()
	engine.state = BREAK

	engine.ccntmext = nil
	engine.opCode = 0

	engine.service = NewInteropService()

	if service != nil {
		engine.service.MergeMap(service.GetServiceMap())
	}
	return &engine
}

type ExecutionEngine struct {
	crypto          interfaces.ICrypto
	table           interfaces.ICodeTable
	service         *InteropService

	codeCcntmainer   interfaces.ICodeCcntmainer
	invocationStack *RandomAccessStack
	opCount         int

	evaluationStack *RandomAccessStack
	altStack        *RandomAccessStack
	state           VMState

	ccntmext         *ExecutionCcntmext

	//current opcode
	opCode          OpCode
	gas             int64
}

func (e *ExecutionEngine) Create(caller common.Uint160, code []byte) ([]byte, error) {
	return code, nil
}

func (e *ExecutionEngine) Call(caller common.Uint160, code, input []byte) ([]byte, error) {
	e.LoadCode(code, false)
	e.LoadCode(input, false)
	err := e.Execute()
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (e *ExecutionEngine) GetCodeCcntmainer() interfaces.ICodeCcntmainer {
	return e.codeCcntmainer
}

func (e *ExecutionEngine) GetState() VMState {
	return e.state
}

func (e *ExecutionEngine) GetEvaluationStack() *RandomAccessStack {
	return e.evaluationStack
}

func (e *ExecutionEngine) GetEvaluationStackCount() int {
	return e.evaluationStack.Count()
}

func (e *ExecutionEngine) GetExecuteResult() bool {
	if e.evaluationStack.Count() < 1 {
		return false
	}
	return e.evaluationStack.Pop().GetStackItem().GetBoolean()
}

func (e *ExecutionEngine) CurrentCcntmext() (*ExecutionCcntmext, error) {
	if e.invocationStack.Count() < 1 {
		log.Error("[CurrentCcntmext], Get current ccntmext fail!")
		return nil, ErrOverStackLen
	}
	ccntmext := e.invocationStack.Peek(0).GetExecutionCcntmext()
	if ccntmext == nil {
		return nil, ErrCurrentCcntmextNil
	}
	return ccntmext, nil
}

func (e *ExecutionEngine) CallingCcntmext() (*ExecutionCcntmext, error) {
	if e.invocationStack.Count() < 2 {
		log.Error("[CallingCcntmext], Get calling ccntmext fail!")
		return nil, ErrOverStackLen
	}
	ccntmext := e.invocationStack.Peek(1).GetExecutionCcntmext()
	if ccntmext == nil {
		return nil, ErrCallingCcntmextNil
	}
	return ccntmext, nil
}

func (e *ExecutionEngine) EntryCcntmext() (*ExecutionCcntmext, error) {
	if e.invocationStack.Count() < 3 {
		log.Error("[EntryCcntmext], Get entry ccntmext fail!")
		return nil, ErrOverStackLen
	}
	ccntmext := e.invocationStack.Peek(e.invocationStack.Count() - 1).GetExecutionCcntmext()
	if ccntmext == nil {
		return nil, ErrEntryCcntmextNil
	}
	return ccntmext, nil
}

func (e *ExecutionEngine) LoadCode(script []byte, pushOnly bool) {
	e.invocationStack.Push(NewExecutionCcntmext(e, script, pushOnly, nil))
}

func (e *ExecutionEngine) Execute() error {
	e.state = e.state & (^BREAK)
	for {
		if e.state == FAULT || e.state == HALT || e.state == BREAK {
			break
		}
		err := e.StepInto()
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *ExecutionEngine) StepInto() error {
	if e.invocationStack.Count() == 0 {
		e.state = HALT
		return nil
	}
	ccntmext, err := e.CurrentCcntmext()
	if err != nil {
		return err
	}
	var opCode OpCode

	if ccntmext.GetInstructionPointer() >= len(ccntmext.Code) {
		opCode = RET
	} else {
		o, err := ccntmext.OpReader.ReadByte()
		if err == io.EOF {
			e.state = FAULT
			return err
		}
		opCode = OpCode(o)
	}
	e.opCode = opCode
	e.ccntmext = ccntmext
	if !e.checkStackSize() {
		return ErrOverLimitStack
	}
	state, err := e.ExecuteOp()

	if state == HALT || state == FAULT {
		e.state = state
		return err
	}
	for _, v := range ccntmext.BreakPoints {
		if v == uint(ccntmext.InstructionPointer) {
			e.state = HALT
			return nil
		}
	}
	return nil
}

func (e *ExecutionEngine) ExecuteOp() (VMState, error) {
	if e.opCode > PUSH16 && e.opCode != RET && e.ccntmext.PushOnly {
		return FAULT, ErrBadValue
	}

	if e.opCode >= PUSHBYTES1 && e.opCode <= PUSHBYTES75 {
		PushData(e, e.ccntmext.OpReader.ReadBytes(int(e.opCode)))
		return NONE, nil
	}

	opExec := OpExecList[e.opCode]
	if opExec.Exec == nil {
		return FAULT, ErrNotSupportOpCode
	}

	if opExec.Validator != nil {
		if err := opExec.Validator(e); err != nil {
			return FAULT, err
		}
	}
	return opExec.Exec(e)
}

func (e *ExecutionEngine) StepOut() {
	e.state = e.state & (^BREAK)
	c := e.invocationStack.Count()
	for {
		if e.state == FAULT || e.state == HALT || e.state == BREAK || e.invocationStack.Count() >= c {
			break
		}
		e.StepInto()
	}
}

func (e *ExecutionEngine) StepOver() {
	if e.state == FAULT || e.state == HALT {
		return
	}
	e.state = e.state & (^BREAK)
	c := e.invocationStack.Count()
	for {
		if e.state == FAULT || e.state == HALT || e.state == BREAK || e.invocationStack.Count() > c {
			break
		}
		e.StepInto()
	}
}

func (e *ExecutionEngine) AddBreakPoint(position uint) {
	e.ccntmext.BreakPoints = append(e.ccntmext.BreakPoints, position)
}

func (e *ExecutionEngine) RemoveBreakPoint(position uint) bool {
	if e.invocationStack.Count() == 0 {
		return false
	}
	bs := make([]uint, 0)
	breakPoints := e.ccntmext.BreakPoints
	for _, v := range breakPoints {
		if v != position {
			bs = append(bs, v)
		}
	}
	e.ccntmext.BreakPoints = bs
	return true
}

func (e *ExecutionEngine) checkStackSize() bool {
	size := 0
	if e.opCode < PUSH16 {
		size = 1
	} else {
		switch e.opCode {
		case DEPTH, DUP, OVER, TUCK:
			size = 1
		case UNPACK:
			item := Peek(e)
			if item == nil {
				return false
			}
			size = len(item.GetStackItem().GetArray())
		}
	}
	size += e.evaluationStack.Count() + e.altStack.Count()
	if uint32(size) > StackLimit {
		return false
	}
	return true
}
