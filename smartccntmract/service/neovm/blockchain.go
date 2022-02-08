package neovm

import (
	vm "github.com/cntmio/cntmology/vm/neovm"
	"github.com/cntmio/cntmology/errors"
	"github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/common"
	vmtypes "github.com/cntmio/cntmology/vm/neovm/types"
)

func BlockChainGetHeight(service *NeoVmService, engine *vm.ExecutionEngine) error {
	vm.PushData(engine, service.Store.GetCurrentBlockHeight())
	return nil
}

func BlockChainGetHeader(service *NeoVmService, engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 1 {
		return errors.NewErr("[BlockChainGetHeader] Too few input parameters ")
	}
	data := vm.PopByteArray(engine)
	var (
		header *types.Header
		err error
	)
	l := len(data)
	if l <= 5 {
		b := vmtypes.ConvertBytesToBigInteger(data)
		height := uint32(b.Int64())
		hash := service.Store.GetBlockHash(height)
		header, err = service.Store.GetHeaderByHash(hash)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[BlockChainGetHeader] GetHeader error!.")
		}
	} else if l == 32 {
		hash, _ := common.Uint256ParseFromBytes(data)
		header, err = service.Store.GetHeaderByHash(hash)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[BlockChainGetHeader] GetHeader error!.")
		}
	} else {
		return errors.NewErr("[BlockChainGetHeader] data invalid.")
	}
	vm.PushData(engine, header)
	return nil
}

func BlockChainGetBlock(service *NeoVmService, engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 1 {
		return errors.NewErr("[BlockChainGetBlock] Too few input parameters ")
	}
	data := vm.PopByteArray(engine)
	var block *types.Block
	l := len(data)
	if l <= 5 {
		b := vmtypes.ConvertBytesToBigInteger(data)
		height := uint32(b.Int64())
		var err error
		block, err = service.Store.GetBlockByHeight(height)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[BlockChainGetBlock] GetBlock error!.")
		}
	} else if l == 32 {
		hash, err := common.Uint256ParseFromBytes(data)
		if err != nil {
			return err
		}
		block, err = service.Store.GetBlockByHash(hash)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[BlockChainGetBlock] GetBlock error!.")
		}
	} else {
		return errors.NewErr("[BlockChainGetBlock] data invalid.")
	}
	vm.PushData(engine, block)
	return nil
}

func BlockChainGetTransaction(service *NeoVmService, engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 1 {
		return errors.NewErr("[BlockChainGetTransaction] Too few input parameters ")
	}
	d := vm.PopByteArray(engine)
	hash, err := common.Uint256ParseFromBytes(d); if err != nil {
		return err
	}
	t, _, err := service.Store.GetTransaction(hash); if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[BlockChainGetTransaction] GetTransaction error!")
	}
	vm.PushData(engine, t)
	return nil
}

func BlockChainGetCcntmract(service *NeoVmService, engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 1 {
		return errors.NewErr("[GetCcntmract] Too few input parameters ")
	}
	address, err := common.AddressParseFromBytes(vm.PopByteArray(engine)); if err != nil {
		return err
	}
	item, err := service.Store.GetCcntmractState(address); if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[GetCcntmract] GetAsset error!")
	}
	vm.PushData(engine, item)
	return nil
}

