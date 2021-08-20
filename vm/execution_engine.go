package vm

import (
	"github.com/Ontology/vm/interfaces"
	"github.com/Ontology/vm/utils"
	"io"
	_ "math/big"
	_ "sort"
)

const MAXSTEPS int = 1200

func NewExecutionEngine(ccntmainer interfaces.IScriptCcntmainer, crypto interfaces.ICrypto, maxSteps int, table interfaces.IScriptTable, service *InteropService) *ExecutionEngine {
	var engine ExecutionEngine

	engine.crypto = crypto
	engine.table = table

	engine.scriptCcntmainer = ccntmainer
	engine.invocationStack = utils.NewRandAccessStack()
	engine.opCount = 0

	engine.evaluationStack = utils.NewRandAccessStack()
	engine.altStack = utils.NewRandAccessStack()
	engine.state = BREAK

	engine.ccntmext = nil
	engine.opCode = 0

	engine.maxSteps = maxSteps

	if service != nil {
		engine.service = service
	}

	engine.service = NewInteropService()

	return &engine
}

type ExecutionEngine struct {
	crypto  interfaces.ICrypto
	table   interfaces.IScriptTable
	service *InteropService

	scriptCcntmainer interfaces.IScriptCcntmainer
	invocationStack *utils.RandomAccessStack
	opCount         int

	maxSteps int

	evaluationStack *utils.RandomAccessStack
	altStack        *utils.RandomAccessStack
	state           VMState

	ccntmext *ExecutionCcntmext

	//current opcode
	opCode OpCode
}

func (e *ExecutionEngine) GetState() VMState {
	return e.state
}

func (e *ExecutionEngine) GetEvaluationStack() *utils.RandomAccessStack {
	return e.evaluationStack
}

func (e *ExecutionEngine) GetExecuteResult() bool {
	return AssertStackItem(e.evaluationStack.Pop()).GetBoolean()
}

func (e *ExecutionEngine) ExecutingScript() []byte {
	ccntmext := AssertExecutionCcntmext(e.invocationStack.Peek(0))
	if ccntmext != nil {
		return ccntmext.Script
	}
	return nil
}

func (e *ExecutionEngine) CallingScript() []byte {
	if e.invocationStack.Count() > 1 {
		ccntmext := AssertExecutionCcntmext(e.invocationStack.Peek(1))
		if ccntmext != nil {
			return ccntmext.Script
		}
		return nil
	}
	return nil
}

func (e *ExecutionEngine) EntryScript() []byte {
	ccntmext := AssertExecutionCcntmext(e.invocationStack.Peek(e.invocationStack.Count() - 1))
	if ccntmext != nil {
		return ccntmext.Script
	}
	return nil
}

func (e *ExecutionEngine) LoadScript(script []byte, pushOnly bool) {
	e.invocationStack.Push(NewExecutionCcntmext(script, pushOnly, nil))
}

func (e *ExecutionEngine) Execute() {
	e.state = e.state & (^BREAK)
	for {
		if e.state == FAULT || e.state == HALT || e.state == BREAK {
			break
		}
		e.StepInto()
	}
}

func (e *ExecutionEngine) StepInto() {
	if e.invocationStack.Count() == 0 {
		e.state = VMState(e.state | HALT)
	}
	if e.state&HALT == HALT || e.state&FAULT == FAULT {
		return
	}
	ccntmext := AssertExecutionCcntmext(e.invocationStack.Pop())
	if ccntmext.InstructionPointer >= len(ccntmext.Script) {
		e.opCode = RET
	}
	for {
		opCode, err := ccntmext.OpReader.ReadByte()
		if err == io.EOF && opCode == 0 {
			return
		}
		e.opCount++
		state, err := e.ExecuteOp(OpCode(opCode), ccntmext)
		if state == VMState(HALT) {
			e.state = VMState(e.state | HALT)
			return
		}
	}
}

func (e *ExecutionEngine) ExecuteOp(opCode OpCode, ccntmext *ExecutionCcntmext) (VMState, error) {
	if opCode > PUSH16 && opCode != RET && ccntmext.PushOnly {
		return FAULT, nil
	}
	if opCode > PUSH16 && e.opCount > e.maxSteps {
		return FAULT, nil
	}
	if opCode >= PUSHBYTES1 && opCode <= PUSHBYTES75 {
		err := pushData(e, ccntmext.OpReader.ReadBytes(int(opCode)))
		if err != nil {
			return FAULT, err
		}
		return NONE, nil
	}
	e.opCode = opCode
	e.ccntmext = ccntmext
	opExec := OpExecList[opCode]
	if opExec.Exec == nil {
		return FAULT, nil
	}
	state, err := opExec.Exec(e)
	if err != nil {
		return state, err
	}
	return NONE, nil
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
	//b := e.ccntmext.BreakPoints
	//b = append(b, position)
}

func (e *ExecutionEngine) RemoveBreakPoint(position uint) bool {
	//if e.invocationStack.Count() == 0 { return false }
	//b := e.ccntmext.BreakPoints
	return true
}
