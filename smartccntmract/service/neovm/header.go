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
	"fmt"

	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/errors"
	vm "github.com/cntmio/cntmology/vm/neovm"
)

type HeaderValue struct {
	Height    uint32
	Hash      common.Uint256
	Timestamp uint32
}

func (self *HeaderValue) ToArray() []byte {
	return []byte(fmt.Sprintf("interop value: header %d", self.Height))
}

// HeaderGetHash put header's hash to vm stack
func HeaderGetHash(service *NeoVmService, engine *vm.Executor) error {
	d, err := engine.EvalStack.PopAsInteropValue()
	if err != nil {
		return err
	}
	var data common.Uint256
	switch val := d.Data.(type) {
	case *types.Block:
		data = val.Header.Hash()
	case *types.Header:
		data = val.Hash()
	case *HeaderValue:
		data = val.Hash
	default:
		return errors.NewErr("[HeaderGetHash] Wrcntm type!")
	}
	return engine.EvalStack.PushBytes(data.ToArray())
}

// HeaderGetVersion put header's version to vm stack
func HeaderGetVersion(service *NeoVmService, engine *vm.Executor) error {
	d, err := engine.EvalStack.PopAsInteropValue()
	if err != nil {
		return err
	}
	var data *types.Header
	if b, ok := d.Data.(*types.Block); ok {
		data = b.Header
	} else if h, ok := d.Data.(*types.Header); ok {
		data = h
	} else {
		return errors.NewErr("[HeaderGetVersion] Wrcntm type!")
	}
	return engine.EvalStack.PushInt64(int64(data.Version))
}

// HeaderGetPrevHash put header's prevblockhash to vm stack
func HeaderGetPrevHash(service *NeoVmService, engine *vm.Executor) error {
	d, err := engine.EvalStack.PopAsInteropValue()
	if err != nil {
		return err
	}
	var data *types.Header
	if b, ok := d.Data.(*types.Block); ok {
		data = b.Header
	} else if h, ok := d.Data.(*types.Header); ok {
		data = h
	} else {
		return errors.NewErr("[HeaderGetPrevHash] Wrcntm type!")
	}
	return engine.EvalStack.PushBytes(data.PrevBlockHash.ToArray())
}

// HeaderGetMerkleRoot put header's merkleroot to vm stack
func HeaderGetMerkleRoot(service *NeoVmService, engine *vm.Executor) error {
	d, err := engine.EvalStack.PopAsInteropValue()
	if err != nil {
		return err
	}
	var data *types.Header
	if b, ok := d.Data.(*types.Block); ok {
		data = b.Header
	} else if h, ok := d.Data.(*types.Header); ok {
		data = h
	} else {
		return errors.NewErr("[HeaderGetMerkleRoot] Wrcntm type!")
	}
	return engine.EvalStack.PushBytes(data.TransactionsRoot.ToArray())
}

// HeaderGetIndex put header's height to vm stack
func HeaderGetIndex(service *NeoVmService, engine *vm.Executor) error {
	d, err := engine.EvalStack.PopAsInteropValue()
	if err != nil {
		return err
	}
	var data uint32
	switch val := d.Data.(type) {
	case *types.Block:
		data = val.Header.Height
	case *types.Header:
		data = val.Height
	case *HeaderValue:
		data = val.Height
	default:
		return fmt.Errorf("[HeaderGetIndex] Wrcntm type")
	}

	return engine.EvalStack.PushUint32(data)
}

// HeaderGetTimestamp put header's timestamp to vm stack
func HeaderGetTimestamp(service *NeoVmService, engine *vm.Executor) error {
	d, err := engine.EvalStack.PopAsInteropValue()
	if err != nil {
		return err
	}
	var data uint32
	switch val := d.Data.(type) {
	case *types.Block:
		data = val.Header.Timestamp
	case *types.Header:
		data = val.Timestamp
	case *HeaderValue:
		data = val.Timestamp
	default:
		return fmt.Errorf("[HeaderGetIndex] Wrcntm type")
	}

	return engine.EvalStack.PushUint32(data)
}

// HeaderGetConsensusData put header's consensus data to vm stack
func HeaderGetConsensusData(service *NeoVmService, engine *vm.Executor) error {
	d, err := engine.EvalStack.PopAsInteropValue()
	if err != nil {
		return err
	}
	var data *types.Header
	if b, ok := d.Data.(*types.Block); ok {
		data = b.Header
	} else if h, ok := d.Data.(*types.Header); ok {
		data = h
	} else {
		return errors.NewErr("[HeaderGetConsensusData] Wrcntm type")
	}
	return engine.EvalStack.PushUint64(data.ConsensusData)
}

// HeaderGetNextConsensus put header's consensus to vm stack
func HeaderGetNextConsensus(service *NeoVmService, engine *vm.Executor) error {
	d, err := engine.EvalStack.PopAsInteropValue()
	if err != nil {
		return err
	}
	var data *types.Header
	if b, ok := d.Data.(*types.Block); ok {
		data = b.Header
	} else if h, ok := d.Data.(*types.Header); ok {
		data = h
	} else {
		return errors.NewErr("[HeaderGetNextConsensus] Wrcntm type")
	}
	return engine.EvalStack.PushBytes(data.NextBookkeeper[:])
}
