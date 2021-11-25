// Copyright 2017 The Ontology Authors
// This file is part of the Ontology library.
//
// The Ontology library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Ontology library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// alcntm with the Ontology library. If not, see <http://www.gnu.org/licenses/>.

package smartccntmract

import (
	"github.com/Ontology/common"
	"github.com/Ontology/core/ccntmract"
	sig "github.com/Ontology/core/signature"
	"github.com/Ontology/smartccntmract/service/neovm"
	"github.com/Ontology/vm/types"
	"github.com/Ontology/vm/neovm"
	"github.com/Ontology/vm/neovm/interfaces"
	"math/big"
	storecomm"github.com/Ontology/core/store/common"
	"github.com/Ontology/errors"
	"github.com/Ontology/common/log"
	scommon "github.com/Ontology/smartccntmract/common"
	"reflect"
	"github.com/Ontology/vm/wasmvm/memory"
	"encoding/binary"
	"github.com/Ontology/vm/wasmvm/exec"
	"github.com/Ontology/vm/wasmvm/util"
	"github.com/Ontology/smartccntmract/service/wasm"
)

type SmartCcntmract struct {
	Engine         Engine
	Code           []byte
	Input          []byte
	ParameterTypes []ccntmract.CcntmractParameterType
	Caller         common.Address
	CodeHash       common.Address
	VMType         types.VmType
	ReturnType     ccntmract.CcntmractParameterType
}

type Ccntmext struct {
	VmType         types.VmType
	Caller         common.Address
	StateMachine   *service.StateMachine
	WasmStateMachine *wasm.WasmStateMachine //add for wasm state machine
	DBCache        storecomm.IStateStore
	Code           []byte
	Input          []byte
	CodeHash       common.Address
	Time           *big.Int
	BlockNumber    *big.Int
	CacheCodeTable interfaces.ICodeTable
	SignableData   sig.SignableData
	Gas            common.Fixed64
	ReturnType     ccntmract.CcntmractParameterType
	ParameterTypes []ccntmract.CcntmractParameterType
}

type Engine interface {
	Create(caller common.Address, code []byte) ([]byte, error)
	Call(caller common.Address, code, input []byte) ([]byte, error)
}

func NewSmartCcntmract(ccntmext *Ccntmext) (*SmartCcntmract, error) {
	var e Engine
	switch ccntmext.VmType {
	case types.NEOVM:
		e = neovm.NewExecutionEngine(
			ccntmext.SignableData,
			new(neovm.ECDsaCrypto),
			ccntmext.CacheCodeTable,
			ccntmext.StateMachine,
		)
		//add wasmvm case
	case types.WASMVM:
		e = exec.NewExecutionEngine(ccntmext.SignableData,
			new(neovm.ECDsaCrypto),
			ccntmext.CacheCodeTable,
			ccntmext.WasmStateMachine)
	default:
		return nil, errors.NewErr("[NewSmartCcntmract] Invalid vm type!")
	}
	return &SmartCcntmract{
		Engine:         e,
		Code:           ccntmext.Code,
		CodeHash:       ccntmext.CodeHash,
		Input:          ccntmext.Input,
		Caller:         ccntmext.Caller,
		VMType:         ccntmext.VmType,
		ReturnType:     ccntmext.ReturnType,
		ParameterTypes: ccntmext.ParameterTypes,
	}, nil
}

func (sc *SmartCcntmract) DeployCcntmract() ([]byte, error) {
	return sc.Engine.Create(sc.Caller, sc.Code)
}

func (sc *SmartCcntmract) InvokeCcntmract() (interface{}, error) {
	res, err := sc.Engine.Call(sc.Caller, sc.Code, sc.Input)
	if err != nil {
		return nil, err
	}
	switch sc.VMType {
	case types.NEOVM:
		return sc.InvokeResult()
	case types.WASMVM:
		//todo add trasmform types
		//todo current we have multi-interface per wasm smart ccntmract
		mem := sc.Engine.(*exec.ExecutionEngine).GetMemory()
		switch sc.ReturnType {
		case ccntmract.Boolean:
			if len(res) > 0 && int(res[0]) == 0 {
				return false, nil
			} else {
				return true, nil
			}
		case ccntmract.Integer:
			if len(res) == 4 { //int32 case
				return int32(binary.LittleEndian.Uint32(res)), nil
			}
			if len(res) == 8 { //int64 case
				return int64(binary.LittleEndian.Uint64(res)), nil
			}

		case ccntmract.ByteArray:
			var idx int
			if len(res) == 4 {
				idx = int(binary.LittleEndian.Uint32(res))
			}else if len(res) == 8 {
				idx = int(binary.LittleEndian.Uint64(res))
			}
			bytes,err := mem.GetPointerMemory(uint64(idx))
			if err != nil{
				return nil,err
			}
			return bytes,nil

		case ccntmract.String:
			var idx int
			if len(res) == 4 {
				idx = int(binary.LittleEndian.Uint32(res))
			}else if len(res) == 8 {
				idx = int(binary.LittleEndian.Uint64(res))
			}
			bytes,err := mem.GetPointerMemory(uint64(idx))
			if err != nil{
				return nil,err
			}
			return string(bytes),nil
		case ccntmract.Hash160,ccntmract.Hash256,ccntmract.PublicKey:
			var idx int
			if len(res) == 4 {
				idx = int(binary.LittleEndian.Uint32(res))
			}else if len(res) == 8 {
				idx = int(binary.LittleEndian.Uint64(res))
			}
			bytes,err := mem.GetPointerMemory(uint64(idx))
			if err != nil{
				return nil,err
			}

			return common.ToHexString(bytes), nil

		case ccntmract.InteropInterface:
			return common.ToHexString(res), nil
		case ccntmract.Array:
			var idx int
			if len(res) == 4 {
				idx = int(binary.LittleEndian.Uint32(res))
			}else if len(res) == 8 {
				idx = int(binary.LittleEndian.Uint64(res))
			}
			bytes,err := mem.GetPointerMemory(uint64(idx))
			if err != nil{
				return nil,err
			}
			tl,_ := mem.MemPoints[uint64(idx)]
			switch tl.Ptype {
			case memory.P_INT32:
				tmp := make([]int,tl.Length / 4)
				for i:= 0 ;i < tl.Length / 4;i++{
					tmp[i] = int(binary.LittleEndian.Uint32(bytes[i*4:(i+1)*4]))
				}
				return tmp,nil
			case memory.P_INT64:
				tmp := make([]int64,tl.Length / 8)
				for i:= 0 ;i < tl.Length / 8;i++{
					tmp[i] = int64(binary.LittleEndian.Uint64(bytes[i*8:(i+1)*8]))
				}
				return tmp,nil
			case memory.P_FLOAT32:
				tmp := make([]float32,tl.Length / 4)
				for i:= 0 ;i < tl.Length / 4;i++{
					tmp[i] = util.ByteToFloat32(bytes[i*4:(i+1)*4])
				}
				return tmp,nil
			case memory.P_FLOAT64:
				tmp := make([]float64,tl.Length / 8)
				for i:= 0 ;i < tl.Length / 8;i++{
					tmp[i] = util.ByteToFloat64(bytes[i*8:(i+1)*8])
				}
				return tmp,nil
			}
		default:
			return common.ToHexString(res), nil
		}

		return res, nil
	default:
		return nil, errors.NewErr("not a support vm")
	}

}

func (sc *SmartCcntmract) InvokeResult() (interface{}, error) {
	switch sc.VMType {
	case types.NEOVM:
		engine := sc.Engine.(*neovm.ExecutionEngine)
		if engine.GetEvaluationStackCount() > 0 && neovm.Peek(engine).GetStackItem() != nil {
			switch sc.ReturnType {
			case ccntmract.Boolean:
				return neovm.PopBoolean(engine), nil
			case ccntmract.Integer:
				log.Error(reflect.TypeOf(neovm.Peek(engine).GetStackItem().GetByteArray()))
				return neovm.PopBigInt(engine).Int64(), nil
			case ccntmract.ByteArray:
				return common.ToHexString(neovm.PopByteArray(engine)), nil
			case ccntmract.String:
				return string(neovm.PopByteArray(engine)), nil
			case ccntmract.Hash160, ccntmract.Hash256:
				return common.ToHexString(neovm.PopByteArray(engine)), nil
			case ccntmract.PublicKey:
				return common.ToHexString(neovm.PopByteArray(engine)), nil
			case ccntmract.InteropInterface:
				if neovm.PeekInteropInterface(engine) != nil {
					return common.ToHexString(neovm.PopInteropInterface(engine).ToArray()), nil
				}
				return nil, nil
			case ccntmract.Array:
				var states []interface{}
				arr := neovm.PeekArray(engine)
				for _, v := range arr {
					states = append(states, scommon.ConvertReturnTypes(v)...)
				}
				return states, nil
			default:
				return common.ToHexString(neovm.PopByteArray(engine)), nil
			}
		}
	}
	return nil, nil
}