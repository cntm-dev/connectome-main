package neovm

import (
	vm "github.com/cntmio/cntmology/vm/neovm"
	"github.com/cntmio/cntmology/errors"
)

// get current execute transaction
func GetCodeCcntmainer(service *NeoVmService, engine *vm.ExecutionEngine) error {
	vm.PushData(engine, service.Tx)
	return nil
}

// get current ccntmract address
func GetExecutingAddress(service *NeoVmService, engine *vm.ExecutionEngine) error {
	ccntmext := service.CcntmextRef.CurrentCcntmext(); if ccntmext == nil {
		return errors.NewErr("Current ccntmext invalid")
	}
	vm.PushData(engine, ccntmext.CcntmractAddress[:])
	return nil
}

// get previous call ccntmract address
func GetCallingAddress(service *NeoVmService, engine *vm.ExecutionEngine) error {
	ccntmext := service.CcntmextRef.CallingCcntmext(); if ccntmext == nil {
		return errors.NewErr("Calling ccntmext invalid")
	}
	vm.PushData(engine, ccntmext.CcntmractAddress[:])
	return nil
}

// get entry call ccntmract address
func GetEntryAddress(service *NeoVmService, engine *vm.ExecutionEngine) error {
	ccntmext := service.CcntmextRef.EntryCcntmext(); if ccntmext == nil {
		return errors.NewErr("Entry ccntmext invalid")
	}
	vm.PushData(engine, ccntmext.CcntmractAddress[:])
	return nil
}

