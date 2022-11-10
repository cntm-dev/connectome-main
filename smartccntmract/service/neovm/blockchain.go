/*
 * Copyright (C) 2018 The cntmology Authors
 * This file is part of The cntmology library.
 *
 * The cntmology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The cntmology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * alcntm with The cntmology.  If not, see <http://www.gnu.org/licenses/>.
 */

package neovm

import (
	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/errors"
	vm "github.com/cntmio/cntmology/vm/neovm"
)

// BlockChainGetHeight put blockchain's height to vm stack
func BlockChainGetHeight(service *NeoVmService, engine *vm.ExecutionEngine) error {
	vm.PushData(engine, service.Store.GetCurrentBlockHeight())
	return nil
}

// BlockChainGetHeader put blockchain's header to vm stack
func BlockChainGetHeader(service *NeoVmService, engine *vm.ExecutionEngine) error {
	var (
		header *types.Header
		err    error
	)
	data, err := vm.PopByteArray(engine)
	if err != nil {
		return err
	}

	l := len(data)
	if l <= 5 {
		b := common.BigIntFromNeoBytes(data)
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

// BlockChainGetBlock put blockchain's block to vm stack
func BlockChainGetBlock(service *NeoVmService, engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 1 {
		return errors.NewErr("[BlockChainGetBlock] Too few input parameters ")
	}
	data, err := vm.PopByteArray(engine)
	if err != nil {
		return err
	}

	var block *types.Block
	l := len(data)
	if l <= 5 {
		b := common.BigIntFromNeoBytes(data)
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

// BlockChainGetTransaction put blockchain's transaction to vm stack
func BlockChainGetTransaction(service *NeoVmService, engine *vm.ExecutionEngine) error {
	d, err := vm.PopByteArray(engine)
	if err != nil {
		return err
	}
	hash, err := common.Uint256ParseFromBytes(d)
	if err != nil {
		return err
	}
	t, _, err := service.Store.GetTransaction(hash)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[BlockChainGetTransaction] GetTransaction error!")
	}
	vm.PushData(engine, t)
	return nil
}

// BlockChainGetCcntmract put blockchain's ccntmract to vm stack
func BlockChainGetCcntmract(service *NeoVmService, engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 1 {
		return errors.NewErr("[GetCcntmract] Too few input parameters ")
	}
	b, err := vm.PopByteArray(engine)
	if err != nil {
		return err
	}
	address, err := common.AddressParseFromBytes(b)
	if err != nil {
		return err
	}
	item, err := service.Store.GetCcntmractState(address)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[GetCcntmract] GetAsset error!")
	}
	vm.PushData(engine, item)
	return nil
}

// BlockChainGetTransactionHeight put transaction in block height to vm stack
func BlockChainGetTransactionHeight(service *NeoVmService, engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 1 {
		return errors.NewErr("[BlockChainGetTransactionHeight] Too few input parameters ")
	}
	d, err := vm.PopByteArray(engine)
	if err != nil {
		return err
	}
	hash, err := common.Uint256ParseFromBytes(d)
	if err != nil {
		return err
	}
	_, h, err := service.Store.GetTransaction(hash)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[BlockChainGetTransaction] GetTransaction error!")
	}
	vm.PushData(engine, h)
	return nil
}
