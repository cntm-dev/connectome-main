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
package util

import (
	"bytes"
	"fmt"
	"math/big"
	"reflect"

	"github.com/cntmio/cntmology/common"
	cstate "github.com/cntmio/cntmology/smartccntmract/states"
	"github.com/cntmio/cntmology/vm/crossvm_codec"
	"github.com/cntmio/cntmology/vm/neovm"
)

//create paramters for neovm ccntmract
func CreateNeoInvokeParam(ccntmractAddress common.Address, input []byte) ([]byte, error) {

	list, err := crossvm_codec.DeserializeInput(input)
	if err != nil {
		return nil, err
	}

	if list == nil {
		return nil, nil
	}

	builder := neovm.NewParamsBuilder(new(bytes.Buffer))
	err = BuildNeoVMParam(builder, list)
	if err != nil {
		return nil, err
	}
	args := append(builder.ToArray(), 0x67)
	args = append(args, ccntmractAddress[:]...)
	return args, nil
}

//buildNeoVMParamInter build neovm invoke param code
func BuildNeoVMParam(builder *neovm.ParamsBuilder, smartCcntmractParams []interface{}) error {
	//VM load params in reverse order
	for i := len(smartCcntmractParams) - 1; i >= 0; i-- {
		switch v := smartCcntmractParams[i].(type) {
		case bool:
			builder.EmitPushBool(v)
		case byte:
			builder.EmitPushInteger(big.NewInt(int64(v)))
		case int:
			builder.EmitPushInteger(big.NewInt(int64(v)))
		case uint:
			builder.EmitPushInteger(big.NewInt(int64(v)))
		case int32:
			builder.EmitPushInteger(big.NewInt(int64(v)))
		case uint32:
			builder.EmitPushInteger(big.NewInt(int64(v)))
		case int64:
			builder.EmitPushInteger(big.NewInt(int64(v)))
		case common.Fixed64:
			builder.EmitPushInteger(big.NewInt(int64(v.GetData())))
		case uint64:
			val := big.NewInt(0)
			builder.EmitPushInteger(val.SetUint64(uint64(v)))
		case string:
			builder.EmitPushByteArray([]byte(v))
		case *big.Int:
			builder.EmitPushInteger(v)
		case []byte:
			builder.EmitPushByteArray(v)
		case common.Address:
			builder.EmitPushByteArray(v[:])
		case common.Uint256:
			builder.EmitPushByteArray(v.ToArray())
		case []interface{}:
			err := BuildNeoVMParam(builder, v)
			if err != nil {
				return err
			}
			builder.EmitPushInteger(big.NewInt(int64(len(v))))
			builder.Emit(neovm.PACK)
		default:
			object := reflect.ValueOf(v)
			kind := object.Kind().String()
			if kind == "ptr" {
				object = object.Elem()
				kind = object.Kind().String()
			}
			switch kind {
			case "slice":
				ps := make([]interface{}, 0)
				for i := 0; i < object.Len(); i++ {
					ps = append(ps, object.Index(i).Interface())
				}
				err := BuildNeoVMParam(builder, []interface{}{ps})
				if err != nil {
					return err
				}
			case "struct":
				builder.EmitPushInteger(big.NewInt(0))
				builder.Emit(neovm.NEWSTRUCT)
				builder.Emit(neovm.TOALTSTACK)
				for i := 0; i < object.NumField(); i++ {
					field := object.Field(i)
					builder.Emit(neovm.DUPFROMALTSTACK)
					err := BuildNeoVMParam(builder, []interface{}{field.Interface()})
					if err != nil {
						return err
					}
					builder.Emit(neovm.APPEND)
				}
				builder.Emit(neovm.FROMALTSTACK)
			default:
				return fmt.Errorf("unsupported param:%s", v)
			}
		}
	}
	return nil
}

//build param bytes for wasm ccntmract
func BuildWasmVMInvokeCode(ccntmractAddress common.Address, params []interface{}) ([]byte, error) {
	ccntmract := &cstate.WasmCcntmractParam{}
	ccntmract.Address = ccntmractAddress
	//bf := bytes.NewBuffer(nil)
	bf := common.NewZeroCopySink(nil)
	argbytes, err := buildWasmCcntmractParam(params, bf)
	if err != nil {
		return nil, fmt.Errorf("build wasm ccntmract param failed:%s", err)
	}
	ccntmract.Args = argbytes
	sink := common.NewZeroCopySink(nil)
	ccntmract.Serialization(sink)
	return sink.Bytes(), nil

}

//build param bytes for wasm ccntmract
func buildWasmCcntmractParam(params []interface{}, bf *common.ZeroCopySink) ([]byte, error) {
	for _, param := range params {
		switch param.(type) {
		case string:
			bf.WriteString(param.(string))
		case int:
			bf.WriteInt32(param.(int32))
		case int64:
			bf.WriteInt64(param.(int64))
		case uint16:
			bf.WriteUint16(param.(uint16))
		case uint32:
			bf.WriteUint32(param.(uint32))
		case uint64:
			bf.WriteUint64(param.(uint64))
		case []byte:
			bf.WriteVarBytes(param.([]byte))
		case common.Uint256:
			bf.WriteHash(param.(common.Uint256))
		case common.Address:
			bf.WriteAddress(param.(common.Address))
		case byte:
			bf.WriteByte(param.(byte))
		case []interface{}:
			_, err := buildWasmCcntmractParam(param.([]interface{}), bf)
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("not a supported type :%v\n", param)
		}
	}
	return bf.Bytes(), nil

}
