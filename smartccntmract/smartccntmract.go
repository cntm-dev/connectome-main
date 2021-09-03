package smartccntmract

import (
	"github.com/Ontology/common"
	"github.com/Ontology/core/ccntmract"
	sig "github.com/Ontology/core/signature"
	"github.com/Ontology/smartccntmract/service"
	"github.com/Ontology/smartccntmract/types"
	"github.com/Ontology/vm/neovm"
	"github.com/Ontology/vm/neovm/interfaces"
	"math/big"
	"github.com/Ontology/core/store"
	"github.com/Ontology/errors"
	"github.com/Ontology/common/log"
	"reflect"
)

type SmartCcntmract struct {
	Engine         Engine
	Code           []byte
	Input          []byte
	ParameterTypes []ccntmract.CcntmractParameterType
	Caller         common.Uint160
	CodeHash       common.Uint160
	VMType         types.VmType
	ReturnType     ccntmract.CcntmractParameterType
}

type Ccntmext struct {
	VmType         types.VmType
	Caller         common.Uint160
	StateMachine   *service.StateMachine
	DBCache        store.IStateStore
	Code           []byte
	Input          []byte
	CodeHash       common.Uint160
	Time           *big.Int
	BlockNumber    *big.Int
	CacheCodeTable interfaces.ICodeTable
	SignableData   sig.SignableData
	Gas            common.Fixed64
	ReturnType     ccntmract.CcntmractParameterType
	ParameterTypes []ccntmract.CcntmractParameterType
}

type Engine interface {
	Create(caller common.Uint160, code []byte) ([]byte, error)
	Call(caller common.Uint160, code, input []byte) ([]byte, error)
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
	log.Error("[InvokeCcntmract] code:", sc.Code)
	log.Error("[InvokeCcntmract] input:", sc.Input)
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
		log.Error("[InvokeResult]", engine.GetEvaluationStackCount(), reflect.TypeOf(neovm.Peek(engine).GetStackItem()), sc.ReturnType)
		if engine.GetEvaluationStackCount() > 0 && neovm.Peek(engine).GetStackItem() != nil {
			switch sc.ReturnType {
			case ccntmract.Boolean:
				return neovm.PopBoolean(engine), nil
			case ccntmract.Integer:
				log.Error(reflect.TypeOf(neovm.Peek(engine).GetStackItem().GetByteArray()))
				return neovm.PopBigInt(engine).Int64(), nil
			case ccntmract.ByteArray:
				return common.ToHexString(neovm.PopByteArray(engine)), nil
			//bs := neovm.PopByteArray(engine)
			//return common.BytesToInt(bs), nil
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
				var strs []string
				for _, v := range neovm.PopArray(engine) {
					strs = append(strs, common.ToHexString(v.GetByteArray()))
				}
				return strs, nil
			default:
				return common.ToHexString(neovm.PopByteArray(engine)), nil
			}
		}
	}
	return nil, nil
}