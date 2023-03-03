/*
 * Copyright (C) 2018 The cntm Authors
 * This file is part of The cntm library.
 *
 * The cntm is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The cntm is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The cntm.  If not, see <http://www.gnu.org/licenses/>.
 */

package cntmvm

import (
	"fmt"

	"github.com/conntectome/cntm/core/types"
	"github.com/conntectome/cntm/errors"
	vm "github.com/conntectome/cntm/vm/cntmvm"
)

// HeaderGetHash put header's hash to vm stack
func HeaderGetHash(service *CntmVmService, engine *vm.Executor) error {
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
		return errors.NewErr("[HeaderGetHash] Wrong type!")
	}
	h := data.Hash()
	return engine.EvalStack.PushBytes(h.ToArray())
}

// HeaderGetVersion put header's version to vm stack
func HeaderGetVersion(service *CntmVmService, engine *vm.Executor) error {
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
		return errors.NewErr("[HeaderGetVersion] Wrong type!")
	}
	return engine.EvalStack.PushInt64(int64(data.Version))
}

// HeaderGetPrevHash put header's prevblockhash to vm stack
func HeaderGetPrevHash(service *CntmVmService, engine *vm.Executor) error {
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
		return errors.NewErr("[HeaderGetPrevHash] Wrong type!")
	}
	return engine.EvalStack.PushBytes(data.PrevBlockHash.ToArray())
}

// HeaderGetMerkleRoot put header's merkleroot to vm stack
func HeaderGetMerkleRoot(service *CntmVmService, engine *vm.Executor) error {
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
		return errors.NewErr("[HeaderGetMerkleRoot] Wrong type!")
	}
	return engine.EvalStack.PushBytes(data.TransactionsRoot.ToArray())
}

// HeaderGetIndex put header's height to vm stack
func HeaderGetIndex(service *CntmVmService, engine *vm.Executor) error {
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
		return fmt.Errorf("[HeaderGetIndex] Wrong type")
	}
	return engine.EvalStack.PushUint32(data.Height)
}

// HeaderGetTimestamp put header's timestamp to vm stack
func HeaderGetTimestamp(service *CntmVmService, engine *vm.Executor) error {
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
		return errors.NewErr("[HeaderGetTimestamp] Wrong type")
	}
	return engine.EvalStack.PushUint32(data.Timestamp)
}

// HeaderGetConsensusData put header's consensus data to vm stack
func HeaderGetConsensusData(service *CntmVmService, engine *vm.Executor) error {
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
		return errors.NewErr("[HeaderGetConsensusData] Wrong type")
	}
	return engine.EvalStack.PushUint64(data.ConsensusData)
}

// HeaderGetNextConsensus put header's consensus to vm stack
func HeaderGetNextConsensus(service *CntmVmService, engine *vm.Executor) error {
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
		return errors.NewErr("[HeaderGetNextConsensus] Wrong type")
	}
	return engine.EvalStack.PushBytes(data.NextBookkeeper[:])
}
