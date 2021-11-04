package neovm

import (
	"github.com/Ontology/common"
	"github.com/Ontology/vm/neovm/types"
	"github.com/Ontology/vm/neovm/utils"
	"io"
	vmtypes "github.com/Ontology/vm/types"
)

type ExecutionCcntmext struct {
	Code               []byte
	OpReader           *utils.VmReader
	PushOnly           bool
	BreakPoints        []uint
	InstructionPointer int
	CodeHash           common.Uint160
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

func (ec *ExecutionCcntmext) GetCodeHash() (common.Uint160, error) {
	if ec.CodeHash.CompareTo(common.Uint160{}) == 0 {
		code := &vmtypes.VmCode{
			Code: ec.Code,
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

func (ec *ExecutionCcntmext) GetStackItem() types.StackItemInterface {
	return nil
}

func (ec *ExecutionCcntmext) GetExecutionCcntmext() *ExecutionCcntmext {
	return ec
}
