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
package wasmvm

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"reflect"

	"github.com/conntectome/cntm/common"
	"github.com/conntectome/cntm/common/log"
	"github.com/conntectome/cntm/core/payload"
	"github.com/conntectome/cntm/core/types"
	"github.com/conntectome/cntm/errors"
	"github.com/conntectome/cntm/smartcontract/event"
	native2 "github.com/conntectome/cntm/smartcontract/service/native"
	"github.com/conntectome/cntm/smartcontract/service/native/utils"
	"github.com/conntectome/cntm/smartcontract/service/util"
	"github.com/conntectome/cntm/smartcontract/states"
	"github.com/conntectome/cntm/vm/crossvm_codec"
	cntmtypes "github.com/conntectome/cntm/vm/cntmvm/types"
	"github.com/conntectome/wagon/exec"
	"github.com/conntectome/wagon/wasm"
)

type ContractType byte

const (
	NATIVE_CCNTMRACT ContractType = iota
	CNTMVM_CCNTMRACT
	WASMVM_CCNTMRACT
	UNKOWN_CCNTMRACT
)

type Runtime struct {
	Service    *WasmVmService
	Input      []byte
	Output     []byte
	CallOutPut []byte
}

func Timestamp(proc *exec.Process) uint64 {
	self := proc.HostData().(*Runtime)
	self.checkGas(TIMESTAMP_GAS)
	return uint64(self.Service.Time)
}

func BlockHeight(proc *exec.Process) uint32 {
	self := proc.HostData().(*Runtime)
	self.checkGas(BLOCK_HEGHT_GAS)
	return self.Service.Height
}

func SelfAddress(proc *exec.Process, dst uint32) {
	self := proc.HostData().(*Runtime)
	self.checkGas(SELF_ADDRESS_GAS)
	selfaddr := self.Service.ContextRef.CurrentContext().ContractAddress
	_, err := proc.WriteAt(selfaddr[:], int64(dst))
	if err != nil {
		panic(err)
	}
}

func Sha256(proc *exec.Process, src uint32, slen uint32, dst uint32) {
	self := proc.HostData().(*Runtime)
	cost := uint64((slen/1024)+1) * SHA256_GAS
	self.checkGas(cost)

	bs, err := ReadWasmMemory(proc, src, slen)
	if err != nil {
		panic(err)
	}

	sh := sha256.New()
	sh.Write(bs[:])
	hash := sh.Sum(nil)

	_, err = proc.WriteAt(hash[:], int64(dst))
	if err != nil {
		panic(err)
	}
}

func CallerAddress(proc *exec.Process, dst uint32) {
	self := proc.HostData().(*Runtime)
	self.checkGas(CALLER_ADDRESS_GAS)
	if self.Service.ContextRef.CallingContext() != nil {
		calleraddr := self.Service.ContextRef.CallingContext().ContractAddress
		_, err := proc.WriteAt(calleraddr[:], int64(dst))
		if err != nil {
			panic(err)
		}
	} else {
		_, err := proc.WriteAt(common.ADDRESS_EMPTY[:], int64(dst))
		if err != nil {
			panic(err)
		}
	}

}

func EntryAddress(proc *exec.Process, dst uint32) {
	self := proc.HostData().(*Runtime)
	self.checkGas(ENTRY_ADDRESS_GAS)
	entryAddress := self.Service.ContextRef.EntryContext().ContractAddress
	_, err := proc.WriteAt(entryAddress[:], int64(dst))
	if err != nil {
		panic(err)
	}
}

func Checkwitness(proc *exec.Process, dst uint32) uint32 {
	self := proc.HostData().(*Runtime)
	self.checkGas(CHECKWITNESS_GAS)
	var addr common.Address
	_, err := proc.ReadAt(addr[:], int64(dst))
	if err != nil {
		panic(err)
	}

	address, err := common.AddressParseFromBytes(addr[:])
	if err != nil {
		panic(err)
	}

	if self.Service.ContextRef.CheckWitness(address) {
		return 1
	}
	return 0
}

func Ret(proc *exec.Process, ptr uint32, len uint32) {
	self := proc.HostData().(*Runtime)
	bs, err := ReadWasmMemory(proc, ptr, len)
	if err != nil {
		panic(err)
	}

	self.Output = bs
	proc.Terminate()
}

func Debug(proc *exec.Process, ptr uint32, len uint32) {
	bs, err := ReadWasmMemory(proc, ptr, len)
	if err != nil {
		//do not panic on debug
		return
	}

	debugLog(bs)
}

func notify(service *WasmVmService, bs []byte) error {
	if len(bs) >= cntmtypes.MAX_NOTIFY_LENGTH {
		return errors.NewErr("notify length over the uplimit")
	}

	notify := &event.NotifyEventInfo{ContractAddress: service.ContextRef.CurrentContext().ContractAddress}
	val := crossvm_codec.DeserializeNotify(bs)
	notify.States = val

	notifys := make([]*event.NotifyEventInfo, 1)
	notifys[0] = notify
	service.ContextRef.PushNotifications(notifys)
	return nil
}

func Notify(proc *exec.Process, ptr uint32, l uint32) {
	self := proc.HostData().(*Runtime)
	bs, err := ReadWasmMemory(proc, ptr, l)
	if err != nil {
		panic(err)
	}

	err = notify(self.Service, bs)
	if err != nil {
		panic(err)
	}
}

func InputLength(proc *exec.Process) uint32 {
	self := proc.HostData().(*Runtime)
	return uint32(len(self.Input))
}

func GetInput(proc *exec.Process, dst uint32) {
	self := proc.HostData().(*Runtime)
	_, err := proc.WriteAt(self.Input, int64(dst))
	if err != nil {
		panic(err)
	}
}

func CallOutputLength(proc *exec.Process) uint32 {
	self := proc.HostData().(*Runtime)
	return uint32(len(self.CallOutPut))
}

func GetCallOut(proc *exec.Process, dst uint32) {
	self := proc.HostData().(*Runtime)
	_, err := proc.WriteAt(self.CallOutPut, int64(dst))
	if err != nil {
		panic(err)
	}
}

func GetCurrentTxHash(proc *exec.Process, ptr uint32) uint32 {
	self := proc.HostData().(*Runtime)
	self.checkGas(CURRENT_TX_HASH_GAS)

	txhash := self.Service.Tx.Hash()

	length, err := proc.WriteAt(txhash[:], int64(ptr))
	if err != nil {
		panic(err)
	}

	return uint32(length)
}

func RaiseException(proc *exec.Process, ptr uint32, len uint32) {
	bs, err := ReadWasmMemory(proc, ptr, len)
	if err != nil {
		//do not panic on debug
		return
	}

	panic(fmt.Errorf("[RaiseException]Contract RaiseException:%s\n", bs))
}

func CallContract(proc *exec.Process, contractAddr uint32, inputPtr uint32, inputLen uint32) uint32 {
	self := proc.HostData().(*Runtime)

	self.checkGas(CALL_CCNTMRACT_GAS)
	var contractAddress common.Address
	_, err := proc.ReadAt(contractAddress[:], int64(contractAddr))
	if err != nil {
		panic(err)
	}

	inputs, err := ReadWasmMemory(proc, inputPtr, inputLen)
	if err != nil {
		panic(err)
	}

	result, err := callContractInner(self.Service, contractAddress, inputs)
	if err != nil {
		panic(err)
	}
	self.CallOutPut = result
	return uint32(len(self.CallOutPut))
}

func NewHostModule() *wasm.Module {
	m := wasm.NewModule()
	paramTypes := make([]wasm.ValueType, 14)
	for i := 0; i < len(paramTypes); i++ {
		paramTypes[i] = wasm.ValueTypeI32
	}

	m.Types = &wasm.SectionTypes{
		Entries: []wasm.FunctionSig{
			//func()uint64    [0]
			{
				Form:        0, // value for the 'func' type constructor
				ReturnTypes: []wasm.ValueType{wasm.ValueTypeI64},
			},
			//func()uint32     [1]
			{
				Form:        0, // value for the 'func' type constructor
				ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32},
			},
			//func(uint32)     [2]
			{
				Form:       0, // value for the 'func' type constructor
				ParamTypes: []wasm.ValueType{wasm.ValueTypeI32},
			},
			//func(uint32)uint32  [3]
			{
				Form:        0, // value for the 'func' type constructor
				ParamTypes:  []wasm.ValueType{wasm.ValueTypeI32},
				ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32},
			},
			//func(uint32,uint32)  [4]
			{
				Form:       0, // value for the 'func' type constructor
				ParamTypes: []wasm.ValueType{wasm.ValueTypeI32, wasm.ValueTypeI32},
			},
			//func(uint32,uint32,uint32)uint32  [5]
			{
				Form:        0, // value for the 'func' type constructor
				ParamTypes:  []wasm.ValueType{wasm.ValueTypeI32, wasm.ValueTypeI32, wasm.ValueTypeI32},
				ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32},
			},
			//func(uint32,uint32,uint32,uint32,uint32)uint32  [6]
			{
				Form:        0, // value for the 'func' type constructor
				ParamTypes:  []wasm.ValueType{wasm.ValueTypeI32, wasm.ValueTypeI32, wasm.ValueTypeI32, wasm.ValueTypeI32, wasm.ValueTypeI32},
				ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32},
			},
			//func(uint32,uint32,uint32,uint32)  [7]
			{
				Form:       0, // value for the 'func' type constructor
				ParamTypes: []wasm.ValueType{wasm.ValueTypeI32, wasm.ValueTypeI32, wasm.ValueTypeI32, wasm.ValueTypeI32},
			},
			//func(uint32,uint32)uint32   [8]
			{
				Form:        0, // value for the 'func' type constructor
				ParamTypes:  []wasm.ValueType{wasm.ValueTypeI32, wasm.ValueTypeI32},
				ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32},
			},
			//func(uint32 * 14)uint32   [9]
			{
				Form:        0, // value for the 'func' type constructor
				ParamTypes:  paramTypes,
				ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32},
			},
			//funct()   [10]
			{
				Form: 0, // value for the 'func' type constructor
			},
			//func(uint32,uint32,uint32)  [11]
			{
				Form:       0, // value for the 'func' type constructor
				ParamTypes: []wasm.ValueType{wasm.ValueTypeI32, wasm.ValueTypeI32, wasm.ValueTypeI32},
			},
		},
	}
	m.FunctionIndexSpace = []wasm.Function{
		{ //0
			Sig:  &m.Types.Entries[0],
			Host: reflect.ValueOf(Timestamp),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{ //1
			Sig:  &m.Types.Entries[1],
			Host: reflect.ValueOf(BlockHeight),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{ //2
			Sig:  &m.Types.Entries[1],
			Host: reflect.ValueOf(InputLength),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{ //3
			Sig:  &m.Types.Entries[1],
			Host: reflect.ValueOf(CallOutputLength),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{ //4
			Sig:  &m.Types.Entries[2],
			Host: reflect.ValueOf(SelfAddress),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{ //5
			Sig:  &m.Types.Entries[2],
			Host: reflect.ValueOf(CallerAddress),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{ //6
			Sig:  &m.Types.Entries[2],
			Host: reflect.ValueOf(EntryAddress),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{ //7
			Sig:  &m.Types.Entries[2],
			Host: reflect.ValueOf(GetInput),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{ //8
			Sig:  &m.Types.Entries[2],
			Host: reflect.ValueOf(GetCallOut),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{ //9
			Sig:  &m.Types.Entries[3],
			Host: reflect.ValueOf(Checkwitness),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{ //10
			Sig:  &m.Types.Entries[3],
			Host: reflect.ValueOf(GetCurrentBlockHash),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{ //11
			Sig:  &m.Types.Entries[3],
			Host: reflect.ValueOf(GetCurrentTxHash),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{ //12
			Sig:  &m.Types.Entries[4],
			Host: reflect.ValueOf(Ret),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{ //13
			Sig:  &m.Types.Entries[4],
			Host: reflect.ValueOf(Notify),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{ //14
			Sig:  &m.Types.Entries[4],
			Host: reflect.ValueOf(Debug),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{ //15
			Sig:  &m.Types.Entries[5],
			Host: reflect.ValueOf(CallContract),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{ //16
			Sig:  &m.Types.Entries[6],
			Host: reflect.ValueOf(StorageRead),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{ //17
			Sig:  &m.Types.Entries[7],
			Host: reflect.ValueOf(StorageWrite),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{ //18
			Sig:  &m.Types.Entries[4],
			Host: reflect.ValueOf(StorageDelete),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{ //19
			Sig:  &m.Types.Entries[9],
			Host: reflect.ValueOf(ContractCreate),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{ //20
			Sig:  &m.Types.Entries[9],
			Host: reflect.ValueOf(ContractMigrate),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{ //21
			Sig:  &m.Types.Entries[10],
			Host: reflect.ValueOf(ContractDestroy),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{ //22
			Sig:  &m.Types.Entries[4],
			Host: reflect.ValueOf(RaiseException),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{ //23
			Sig:  &m.Types.Entries[11],
			Host: reflect.ValueOf(Sha256),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
	}

	m.Export = &wasm.SectionExports{
		Entries: map[string]wasm.ExportEntry{
			"conntectome_timestamp": {
				FieldStr: "conntectome_timestamp",
				Kind:     wasm.ExternalFunction,
				Index:    0,
			},
			"conntectome_block_height": {
				FieldStr: "conntectome_block_height",
				Kind:     wasm.ExternalFunction,
				Index:    1,
			},
			"conntectome_input_length": {
				FieldStr: "conntectome_input_length",
				Kind:     wasm.ExternalFunction,
				Index:    2,
			},
			"conntectome_call_output_length": {
				FieldStr: "conntectome_call_output_length",
				Kind:     wasm.ExternalFunction,
				Index:    3,
			},
			"conntectome_self_address": {
				FieldStr: "conntectome_self_address",
				Kind:     wasm.ExternalFunction,
				Index:    4,
			},
			"conntectome_caller_address": {
				FieldStr: "conntectome_caller_address",
				Kind:     wasm.ExternalFunction,
				Index:    5,
			},
			"conntectome_entry_address": {
				FieldStr: "conntectome_entry_address",
				Kind:     wasm.ExternalFunction,
				Index:    6,
			},
			"conntectome_get_input": {
				FieldStr: "conntectome_get_input",
				Kind:     wasm.ExternalFunction,
				Index:    7,
			},
			"conntectome_get_call_output": {
				FieldStr: "conntectome_get_call_output",
				Kind:     wasm.ExternalFunction,
				Index:    8,
			},
			"conntectome_check_witness": {
				FieldStr: "conntectome_check_witness",
				Kind:     wasm.ExternalFunction,
				Index:    9,
			},
			"conntectome_current_blockhash": {
				FieldStr: "conntectome_current_blockhash",
				Kind:     wasm.ExternalFunction,
				Index:    10,
			},
			"conntectome_current_txhash": {
				FieldStr: "conntectome_current_txhash",
				Kind:     wasm.ExternalFunction,
				Index:    11,
			},
			"conntectome_return": {
				FieldStr: "conntectome_return",
				Kind:     wasm.ExternalFunction,
				Index:    12,
			},
			"conntectome_notify": {
				FieldStr: "conntectome_notify",
				Kind:     wasm.ExternalFunction,
				Index:    13,
			},
			"conntectome_debug": {
				FieldStr: "conntectome_debug",
				Kind:     wasm.ExternalFunction,
				Index:    14,
			},
			"conntectome_call_contract": {
				FieldStr: "conntectome_call_contract",
				Kind:     wasm.ExternalFunction,
				Index:    15,
			},
			"conntectome_storage_read": {
				FieldStr: "conntectome_storage_read",
				Kind:     wasm.ExternalFunction,
				Index:    16,
			},
			"conntectome_storage_write": {
				FieldStr: "conntectome_storage_write",
				Kind:     wasm.ExternalFunction,
				Index:    17,
			},
			"conntectome_storage_delete": {
				FieldStr: "conntectome_storage_delete",
				Kind:     wasm.ExternalFunction,
				Index:    18,
			},
			"conntectome_contract_create": {
				FieldStr: "conntectome_contract_create",
				Kind:     wasm.ExternalFunction,
				Index:    19,
			},
			"conntectome_contract_migrate": {
				FieldStr: "conntectome_contract_migrate",
				Kind:     wasm.ExternalFunction,
				Index:    20,
			},
			"conntectome_contract_destroy": {
				FieldStr: "conntectome_contract_destroy",
				Kind:     wasm.ExternalFunction,
				Index:    21,
			},
			"conntectome_panic": {
				FieldStr: "conntectome_panic",
				Kind:     wasm.ExternalFunction,
				Index:    22,
			},
			"conntectome_sha256": {
				FieldStr: "conntectome_sha256",
				Kind:     wasm.ExternalFunction,
				Index:    23,
			},
		},
	}

	return m
}

func getContractTypeInner(service *WasmVmService, addr common.Address) (ContractType, error) {
	if utils.IsNativeContract(addr) {
		return NATIVE_CCNTMRACT, nil
	}

	dep, err := service.CacheDB.GetContract(addr)
	if err != nil {
		return UNKOWN_CCNTMRACT, err
	}
	if dep == nil {
		return UNKOWN_CCNTMRACT, errors.NewErr("contract is not exist.")
	}
	if dep.VmType() == payload.WASMVM_TYPE {
		return WASMVM_CCNTMRACT, nil
	}

	return CNTMVM_CCNTMRACT, nil
}

func (self *Runtime) getContractType(addr common.Address) (ContractType, error) {
	return getContractTypeInner(self.Service, addr)
}

func checkGasInner(gasLimit *uint64, cost uint64) error {
	if *gasLimit >= cost {
		*gasLimit -= cost
	} else {
		return errors.NewErr("[wasm_Service]Insufficient gas limit")
	}

	return nil
}

func (self *Runtime) checkGas(gaslimit uint64) {
	err := checkGasInner(self.Service.vm.ExecMetrics.GasLimit, gaslimit)
	if err != nil {
		panic(err)
	}
}

func serializeStorageKey(contractAddress common.Address, key []byte) []byte {
	bf := new(bytes.Buffer)

	bf.Write(contractAddress[:])
	bf.Write(key)

	return bf.Bytes()
}

func debugLog(bs []byte) {
	log.Debugf("[WasmContract]Debug:%s\n", bs)
}

func callContractInner(service *WasmVmService, contractAddress common.Address, inputs []byte) ([]byte, error) {
	contracttype, err := getContractTypeInner(service, contractAddress)
	if err != nil {
		return []byte{}, err
	}

	var result []byte

	switch contracttype {
	case NATIVE_CCNTMRACT:
		source := common.NewZeroCopySource(inputs)
		ver, eof := source.NextByte()
		if eof {
			return []byte{}, io.ErrUnexpectedEOF
		}
		method, _, irregular, eof := source.NextString()
		if irregular {
			return []byte{}, common.ErrIrregularData
		}
		if eof {
			return []byte{}, io.ErrUnexpectedEOF
		}

		args, _, irregular, eof := source.NextVarBytes()
		if irregular {
			return []byte{}, common.ErrIrregularData
		}
		if eof {
			return []byte{}, io.ErrUnexpectedEOF
		}

		contract := states.ContractInvokeParam{
			Version: ver,
			Address: contractAddress,
			Method:  method,
			Args:    args,
		}

		err = checkGasInner(service.GasLimit, NATIVE_INVOKE_GAS)
		if err != nil {
			return []byte{}, errors.NewErr("[wasm_Service]Insufficient gas limit")
		}

		native := &native2.NativeService{
			CacheDB:     service.CacheDB,
			InvokeParam: contract,
			Tx:          service.Tx,
			Height:      service.Height,
			Time:        service.Time,
			ContextRef:  service.ContextRef,
			ServiceMap:  make(map[string]native2.Handler),
			PreExec:     service.PreExec,
		}

		tmpRes, err := native.Invoke()
		if err != nil {
			return []byte{}, errors.NewErr("[nativeInvoke]AppCall failed:" + err.Error())
		}

		result = tmpRes

	case WASMVM_CCNTMRACT:
		conParam := states.WasmContractParam{Address: contractAddress, Args: inputs}
		param := common.SerializeToBytes(&conParam)

		newservice, err := service.ContextRef.NewExecuteEngine(param, types.InvokeWasm)
		if err != nil {
			return []byte{}, err
		}

		tmpRes, err := newservice.Invoke()
		if err != nil {
			return []byte{}, err
		}

		result = tmpRes.([]byte)

	case CNTMVM_CCNTMRACT:
		evalstack, err := util.GenerateCntmVMParamEvalStack(inputs)
		if err != nil {
			return []byte{}, err
		}

		cntmservice, err := service.ContextRef.NewExecuteEngine([]byte{}, types.InvokeCntm)
		if err != nil {
			return []byte{}, err
		}

		err = util.SetCntmServiceParamAndEngine(contractAddress, cntmservice, evalstack)
		if err != nil {
			return []byte{}, err
		}

		tmp, err := cntmservice.Invoke()
		if err != nil {
			return []byte{}, err
		}

		if tmp != nil {
			val := tmp.(*cntmtypes.VmValue)
			source := common.NewZeroCopySink([]byte{byte(crossvm_codec.VERSION)})

			err = cntmtypes.BuildResultFromCntm(*val, source)
			if err != nil {
				return []byte{}, err
			}
			result = source.Bytes()
		}

	default:
		return []byte{}, errors.NewErr("Not a supported contract type")
	}

	return result, nil
}
