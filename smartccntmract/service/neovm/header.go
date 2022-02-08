package neovm

import (
	vm "github.com/cntmio/cntmology/vm/neovm"
	"github.com/cntmio/cntmology/errors"
	"github.com/cntmio/cntmology/core/types"
)

func HeaderGetHash(service *NeoVmService, engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 1 {
		return errors.NewErr("[HeaderGetHash] Too few input parameters ")
	}
	d := vm.PopInteropInterface(engine); if d == nil {
		return errors.NewErr("[HeaderGetHash] Pop blockdata nil!")
	}
	var data *types.Header
	if b, ok := d.(*types.Block); ok {
		data = b.Header
	} else if h, ok := d.(*types.Header); ok {
		data = h
	} else {
		return errors.NewErr("[HeaderGetHash] Wrcntm type!")
	}
	h := data.Hash()
	vm.PushData(engine, h.ToArray())
	return nil
}

func HeaderGetVersion(service *NeoVmService, engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 1 {
		return errors.NewErr("[HeaderGetVersion] Too few input parameters ")
	}
	d := vm.PopInteropInterface(engine); if d == nil {
		return errors.NewErr("[HeaderGetVersion] Pop blockdata nil!")
	}
	var data *types.Header
	if b, ok := d.(*types.Block); ok {
		data = b.Header
	} else if h, ok := d.(*types.Header); ok {
		data = h
	} else {
		return errors.NewErr("[HeaderGetVersion] Wrcntm type!")
	}
	vm.PushData(engine, data.Version)
	return nil
}

func HeaderGetPrevHash(service *NeoVmService, engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 1 {
		return errors.NewErr("[HeaderGetPrevHash] Too few input parameters ")
	}
	d := vm.PopInteropInterface(engine); if d == nil {
		return errors.NewErr("[HeaderGetPrevHash] Pop blockdata nil!")
	}
	var data *types.Header
	if b, ok := d.(*types.Block); ok {
		data = b.Header
	} else if h, ok := d.(*types.Header); ok {
		data = h
	} else {
		return errors.NewErr("[HeaderGetPrevHash] Wrcntm type!")
	}
	vm.PushData(engine, data.PrevBlockHash.ToArray())
	return nil
}

func HeaderGetMerkleRoot(service *NeoVmService, engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 1 {
		return errors.NewErr("[HeaderGetMerkleRoot] Too few input parameters ")
	}
	d := vm.PopInteropInterface(engine); if d == nil {
		return errors.NewErr("[HeaderGetMerkleRoot] Pop blockdata nil!")
	}
	var data *types.Header
	if b, ok := d.(*types.Block); ok {
		data = b.Header
	} else if h, ok := d.(*types.Header); ok {
		data = h
	} else {
		return errors.NewErr("[HeaderGetMerkleRoot] Wrcntm type!")
	}
	vm.PushData(engine, data.TransactionsRoot.ToArray())
	return nil
}

func HeaderGetIndex(service *NeoVmService, engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 1 {
		return errors.NewErr("[HeaderGetIndex] Too few input parameters ")
	}
	d := vm.PopInteropInterface(engine); if d == nil {
		return errors.NewErr("[HeaderGetIndex] Pop blockdata nil!")
	}
	var data *types.Header
	if b, ok := d.(*types.Block); ok {
		data = b.Header
	} else if h, ok := d.(*types.Header); ok {
		data = h
	} else {
		return errors.NewErr("[HeaderGetIndex] Wrcntm type!")
	}
	vm.PushData(engine, data.Height)
	return nil
}

func HeaderGetTimestamp(service *NeoVmService, engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 1 {
		return errors.NewErr("[HeaderGetTimestamp] Too few input parameters ")
	}
	d := vm.PopInteropInterface(engine); if d == nil {
		return errors.NewErr("[HeaderGetTimestamp] Pop blockdata nil!")
	}
	var data *types.Header
	if b, ok := d.(*types.Block); ok {
		data = b.Header
	} else if h, ok := d.(*types.Header); ok {
		data = h
	} else {
		return errors.NewErr("[HeaderGetTimestamp] Wrcntm type!")
	}
	vm.PushData(engine, data.Timestamp)
	return nil
}

func HeaderGetConsensusData(service *NeoVmService, engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 1 {
		return errors.NewErr("[HeaderGetConsensusData] Too few input parameters ")
	}
	d := vm.PopInteropInterface(engine); if d == nil {
		return errors.NewErr("[HeaderGetConsensusData] Pop blockdata nil!")
	}
	var data *types.Header
	if b, ok := d.(*types.Block); ok {
		data = b.Header
	} else if h, ok := d.(*types.Header); ok {
		data = h
	} else {
		return errors.NewErr("[HeaderGetConsensusData] Wrcntm type!")
	}
	vm.PushData(engine, data.ConsensusData)
	return nil
}

func HeaderGetNextConsensus(service *NeoVmService, engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 1 {
		return errors.NewErr("[HeaderGetNextConsensus] Too few input parameters ")
	}
	d := vm.PopInteropInterface(engine); if d == nil {
		return errors.NewErr("[HeaderGetNextConsensus] Pop blockdata nil!")
	}
	var data *types.Header
	if b, ok := d.(*types.Block); ok {
		data = b.Header
	} else if h, ok := d.(*types.Header); ok {
		data = h
	} else {
		return errors.NewErr("[HeaderGetNextConsensus] Wrcntm type!")
	}
	vm.PushData(engine, data.NextBookkeeper[:])
	return nil
}








