package neovm

import (
	vm "github.com/cntmio/cntmology/vm/neovm"
)

func GetCodeCcntmainer(service *NeoVmService, engine *vm.ExecutionEngine) error {
	vm.PushData(engine, service.Tx)
	return nil
}

func GetExecutingAddress(service *NeoVmService, engine *vm.ExecutionEngine) error {
	vm.PushData(engine, service.CcntmextRef.CurrentCcntmext().CcntmractAddress[:])
	return nil
}

func GetCallingAddress(service *NeoVmService, engine *vm.ExecutionEngine) error {
	vm.PushData(engine, service.CcntmextRef.CallingCcntmext().CcntmractAddress[:])
	return nil
}

func GetEntryAddress(service *NeoVmService, engine *vm.ExecutionEngine) error {
	vm.PushData(engine, service.CcntmextRef.EntryCcntmext().CcntmractAddress[:])
	return nil
}

