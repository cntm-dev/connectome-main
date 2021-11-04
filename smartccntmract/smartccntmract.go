// Copyright 2017 The Onchain Authors
// This file is part of the Onchain library.
//
// The Onchain library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Onchain library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// alcntm with the Onchain library. If not, see <http://www.gnu.org/licenses/>.

package smartccntmract

import (
	"github.com/Ontology/common"
	"github.com/Ontology/core/ccntmract"
	"github.com/Ontology/smartccntmract/service"
	vmtypes "github.com/Ontology/vm/types"
	"reflect"
	"github.com/Ontology/vm/neovm"
	"github.com/Ontology/errors"
	"github.com/Ontology/common/log"
	"github.com/Ontology/core/store"
	scommon "github.com/Ontology/core/store/common"
	"github.com/Ontology/core/types"
)

type Ccntmext struct {
	LedgerStore store.ILedgerStore
	Code vmtypes.VmCode
	DBCache scommon.IStateStore
	TX *types.Transaction
	Time uint32
}

type SmartCcntmract struct {
	Input          []byte
	VMType         vmtypes.VmType
}

type Engine interface {
	Create(caller common.Uint160, code []byte) ([]byte, error)
	Call(caller common.Uint160, code, input []byte) ([]byte, error)
}

func NewSmartCcntmract(ccntmext *Ccntmext) (*SmartCcntmract, error) {
	var e Engine
	switch ccntmext.Code.VmType {
	case vmtypes.NEOVM:
		stateMachine := service.NewStateMachine(ccntmext.LedgerStore, ccntmext.DBCache, vmtypes.Application, ccntmext.Time)
		e = neovm.NewExecutionEngine(
			ccntmext.TX,
			new(neovm.ECDsaCrypto),
			ccntmext.CacheCodeTable,
			ccntmext.StateMachine,
		)
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
	_, err := sc.Engine.Call(sc.Caller, sc.Code, sc.Input)
	if err != nil {
		return nil, err
	}
	return sc.InvokeResult()
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