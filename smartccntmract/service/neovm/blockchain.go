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
	"math/big"

	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/errors"
	vm "github.com/cntmio/cntmology/vm/neovm"
)

// BlockChainGetHeight put blockchain's height to vm stack
func BlockChainGetHeight(service *NeoVmService, engine *vm.Executor) error {
	return engine.EvalStack.PushUint32(service.Height - 1)
}

func BlockChainGetHeightNew(service *NeoVmService, engine *vm.Executor) error {
	return engine.EvalStack.PushUint32(service.Height)
}

func BlockChainGetHeaderNew(service *NeoVmService, engine *vm.Executor) error {
	data, err := engine.EvalStack.PopAsBytes()
	if err != nil {
		return err
	}
	b := common.BigIntFromNeoBytes(data)
	if b.Cmp(big.NewInt(int64(service.Height))) != 0 {
		return errors.NewErr("can only get current block header")
	}

	header := &HeaderValue{Height: service.Height, Timestamp: service.Time, Hash: service.BlockHash}
	return engine.EvalStack.PushAsInteropValue(header)
}

// BlockChainGetCcntmract put blockchain's ccntmract to vm stack
func BlockChainGetCcntmract(service *NeoVmService, engine *vm.Executor) error {
	b, err := engine.EvalStack.PopAsBytes()
	if err != nil {
		return err
	}
	address, err := common.AddressParseFromBytes(b)
	if err != nil {
		return err
	}
	item, err := service.Store.GetCcntmractState(address)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[BlockChainGetCcntmract] GetCcntmract error!")
	}
	err = engine.EvalStack.PushAsInteropValue(item)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[BlockChainGetCcntmract] PushAsInteropValue error!")
	}
	return nil
}
