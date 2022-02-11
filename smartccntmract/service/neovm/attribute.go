package neovm

import (
	vm "github.com/cntmio/cntmology/vm/neovm"
	"github.com/cntmio/cntmology/core/types"
)

// get usage from attribute
func AttributeGetUsage(service *NeoVmService, engine *vm.ExecutionEngine) error {
	vm.PushData(engine, int(vm.PopInteropInterface(engine).(*types.TxAttribute).Usage))
	return nil
}

// get data from attribute
func AttributeGetData(service *NeoVmService, engine *vm.ExecutionEngine) error {
	vm.PushData(engine, vm.PopInteropInterface(engine).(*types.TxAttribute).Data)
	return nil
}

