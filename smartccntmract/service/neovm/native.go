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
	"fmt"
	"reflect"

	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/serialization"
	"github.com/cntmio/cntmology/smartccntmract/service/native"
	"github.com/cntmio/cntmology/smartccntmract/states"
	vm "github.com/cntmio/cntmology/vm/neovm"
	"github.com/cntmio/cntmology/vm/neovm/types"
	"math/big"
)

func NativeInvoke(service *NeoVmService, engine *vm.ExecutionEngine) error {
	count := vm.EvaluationStackCount(engine)
	if count < 4 {
		return fmt.Errorf("invoke native ccntmract invalid parameters %d < 4 ", count)
	}
	version := vm.PopInt(engine)
	address := vm.PopByteArray(engine)
	addr, err := common.AddressParseFromBytes(address)
	if err != nil {
		return fmt.Errorf("invoke native ccntmract:%s, address invalid", address)
	}
	method := vm.PopByteArray(engine)
	if len(method) > METHOD_LENGTH_LIMIT {
		return fmt.Errorf("invoke native ccntmract:%s method:%s too lcntm, over max length 1024 limit", address, method)
	}
	args := vm.PopStackItem(engine)

	buf := new(bytes.Buffer)
	if err := BuildParamToNative(buf, args); err != nil {
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
	vm.PushData(engine, result)
	return nil
}

func BuildParamToNative(bf *bytes.Buffer, item types.StackItems) error {
	switch item.(type) {
	case *types.ByteArray:
		if err := serialization.WriteVarBytes(bf, item.GetByteArray()); err != nil {
			return err
		}
	case *types.Integer:
		if err := serialization.WriteVarBytes(bf, item.GetByteArray()); err != nil {
			return err
		}
	case *types.Boolean:
		if err := serialization.WriteBool(bf, item.GetBoolean()); err != nil {
			return err
		}
	case *types.Array:
		arr := item.GetArray()
		if err := serialization.WriteVarBytes(bf, types.BigIntToBytes(big.NewInt(int64(len(arr))))); err != nil {
			return err
		}
		for _, v := range arr {
			if err := BuildParamToNative(bf, v); err != nil {
				return err
			}
		}
	case *types.Struct:
		st := item.GetStruct()
		for _, v := range st {
			if err := BuildParamToNative(bf, v); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("convert neovm params to native invalid type support: %s", reflect.TypeOf(item))
	}
	return nil
}
