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
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"reflect"

	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/serialization"
	"github.com/cntmio/cntmology/smartccntmract/service/native"
	"github.com/cntmio/cntmology/smartccntmract/states"
	vm "github.com/cntmio/cntmology/vm/neovm"
	"github.com/cntmio/cntmology/vm/neovm/types"
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

func BuildParamToNative(bf *bytes.Buffer, item types.StackItems) error {
	if CircularRefAndDepthDetection(item) {
		return errors.New("invoke native circular reference!")
	}
	return buildParamToNative(bf, item)
}

func buildParamToNative(bf *bytes.Buffer, item types.StackItems) error {
	switch item.(type) {
	case *types.ByteArray:
		a, _ := item.GetByteArray()
		if err := serialization.WriteVarBytes(bf, a); err != nil {
			return err
		}
	case *types.Integer:
		i, _ := item.GetByteArray()
		if err := serialization.WriteVarBytes(bf, i); err != nil {
			return err
		}
	case *types.Boolean:
		b, _ := item.GetBoolean()
		if err := serialization.WriteBool(bf, b); err != nil {
			return err
		}
	case *types.Array:
		arr, _ := item.GetArray()
		if err := serialization.WriteVarBytes(bf, common.BigIntToNeoBytes(big.NewInt(int64(len(arr))))); err != nil {
			return err
		}
		for _, v := range arr {
			if err := buildParamToNative(bf, v); err != nil {
				return err
			}
		}
	case *types.Struct:
		st, _ := item.GetStruct()
		for _, v := range st {
			if err := buildParamToNative(bf, v); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("convert neovm params to native invalid type support: %s", reflect.TypeOf(item))
	}
	return nil
}
