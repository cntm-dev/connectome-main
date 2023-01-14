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
	"github.com/cntmio/cntmology/smartccntmract/service/native"
	"github.com/cntmio/cntmology/smartccntmract/states"
	vm "github.com/cntmio/cntmology/vm/neovm"
)

func NativeInvoke(service *NeoVmService, engine *vm.Executor) error {
	version, err := engine.EvalStack.PopAsInt64()
	if err != nil {
		return err
	}
	address, err := engine.EvalStack.PopAsBytes()
	if err != nil {
		return err
	}
	addr, err := common.AddressParseFromBytes(address)
	if err != nil {
		return fmt.Errorf("invoke native ccntmract:%s, address invalid", address)
	}
	method, err := engine.EvalStack.PopAsBytes()
	if err != nil {
		return err
	}
	if len(method) > METHOD_LENGTH_LIMIT {
		return fmt.Errorf("invoke native ccntmract:%s method:%s too lcntm, over max length 1024 limit", address, method)
	}
	args, err := engine.EvalStack.Pop()
	if err != nil {
		return err
	}
	sink := new(common.ZeroCopySink)
	if err := args.BuildParamToNative(sink); err != nil {
		return err
	}

	ccntmract := &states.Ccntmract{
		Version: byte(version),
		Address: addr,
		Method:  string(method),
		Args:    buf.Bytes(),
	}

	bf := new(bytes.Buffer)
	if err := ccntmract.Serialize(bf); err != nil {
		return err
	}

	native := &native.NativeService{
		CloneCache: service.CloneCache,
		Code:       bf.Bytes(),
		Tx:         service.Tx,
		Height:     service.Height,
		Time:       service.Time,
		CcntmextRef: service.CcntmextRef,
		ServiceMap: make(map[string]native.Handler),
	}

	result, err := native.Invoke()
	if err != nil {
		return err
	}
	return engine.EvalStack.PushBytes(result)
}
