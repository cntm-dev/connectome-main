package vm

import "DNA/vm/utils"

type ExecutionCcntmext struct {
	Script             []byte
	OpReader           *utils.VmReader
	PushOnly           bool
	BreakPoints        []uint
	InstructionPointer int
}

func NewExecutionCcntmext(script []byte, pushOnly bool, breakPoints []uint) *ExecutionCcntmext {
	var executionCcntmext ExecutionCcntmext
	executionCcntmext.Script = script
	executionCcntmext.OpReader = utils.NewVmReader(script)
	executionCcntmext.PushOnly = pushOnly
	executionCcntmext.BreakPoints = breakPoints
	executionCcntmext.InstructionPointer = executionCcntmext.OpReader.Position()
	return &executionCcntmext
}

func (ec *ExecutionCcntmext) NextInstruction() OpCode {
	return OpCode(ec.Script[ec.OpReader.Position()])
}

func (ec *ExecutionCcntmext) Clone() *ExecutionCcntmext {
	return NewExecutionCcntmext(ec.Script, ec.PushOnly, ec.BreakPoints)
}
