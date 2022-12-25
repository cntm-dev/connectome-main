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
package wasmvm

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"reflect"

	"github.com/go-interpreter/wagon/exec"
	"github.com/go-interpreter/wagon/wasm"
	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/log"
	"github.com/cntmio/cntmology/common/serialization"
	"github.com/cntmio/cntmology/core/payload"
	"github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/errors"
	"github.com/cntmio/cntmology/smartccntmract/event"
	native2 "github.com/cntmio/cntmology/smartccntmract/service/native"
	"github.com/cntmio/cntmology/smartccntmract/service/native/utils"
	"github.com/cntmio/cntmology/smartccntmract/service/util"
	"github.com/cntmio/cntmology/smartccntmract/states"
	"github.com/cntmio/cntmology/vm/crossvm_codec"
	neotypes "github.com/cntmio/cntmology/vm/neovm/types"
)

type CcntmractType byte

const (
	NATIVE_CcntmRACT CcntmractType = iota
	NEOVM_CcntmRACT
	WASMVM_CcntmRACT
	UNKOWN_CcntmRACT
)

type Runtime struct {
	Service    *WasmVmService
	Input      []byte
	Output     []byte
	CallOutPut []byte
}

func TimeStamp(proc *exec.Process) uint64 {
	self := proc.HostData().(*Runtime)
	self.checkGas(TIME_STAMP_GAS)
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
	selfaddr := self.Service.CcntmextRef.CurrentCcntmext().CcntmractAddress
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
	if self.Service.CcntmextRef.CallingCcntmext() != nil {
		calleraddr := self.Service.CcntmextRef.CallingCcntmext().CcntmractAddress
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
	entryAddress := self.Service.CcntmextRef.EntryCcntmext().CcntmractAddress
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

	if self.Service.CcntmextRef.CheckWitness(address) {
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

	log.Debugf("[WasmCcntmract]Debug:%s\n", bs)
}

func Notify(proc *exec.Process, ptr uint32, len uint32) {
	self := proc.HostData().(*Runtime)
	bs, err := ReadWasmMemory(proc, ptr, len)
	if err != nil {
		panic(err)
	}

	list, err := crossvm_codec.DeserializeInput(bs)
	if err != nil {
		panic(err)
	}

	notify := &event.NotifyEventInfo{self.Service.CcntmextRef.CurrentCcntmext().CcntmractAddress, list}
	notifys := make([]*event.NotifyEventInfo, 1)
	notifys[0] = notify
	self.Service.CcntmextRef.PushNotifications(notifys)
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

	panic(fmt.Errorf("[RaiseException]Ccntmract RaiseException:%s\n", bs))
}

func CallCcntmract(proc *exec.Process, ccntmractAddr uint32, inputPtr uint32, inputLen uint32) uint32 {
	self := proc.HostData().(*Runtime)

	self.checkGas(CALL_CcntmRACT_GAS)
	var ccntmractAddress common.Address
	_, err := proc.ReadAt(ccntmractAddress[:], int64(ccntmractAddr))
	if err != nil {
		panic(err)
	}

	inputs, err := ReadWasmMemory(proc, inputPtr, inputLen)
	if err != nil {
		panic(err)
	}

	ccntmracttype, err := self.getCcntmractType(ccntmractAddress)
	if err != nil {
		panic(err)
	}

	var result []byte

	switch ccntmracttype {
	case NATIVE_CcntmRACT:
		bf := bytes.NewBuffer(inputs)
		ver, err := serialization.ReadByte(bf)
		if err != nil {
			panic(err)
		}

		method, err := serialization.ReadString(bf)
		if err != nil {
			panic(err)
		}

		args, err := serialization.ReadVarBytes(bf)
		if err != nil {
			panic(err)
		}

		ccntmract := states.CcntmractInvokeParam{
			Version: ver,
			Address: ccntmractAddress,
			Method:  method,
			Args:    args,
		}

		self.checkGas(NATIVE_INVOKE_GAS)
		native := &native2.NativeService{
			CacheDB:     self.Service.CacheDB,
			InvokeParam: ccntmract,
			Tx:          self.Service.Tx,
			Height:      self.Service.Height,
			Time:        self.Service.Time,
			CcntmextRef:  self.Service.CcntmextRef,
			ServiceMap:  make(map[string]native2.Handler),
		}

		tmpRes, err := native.Invoke()
		if err != nil {
			panic(errors.NewErr("[nativeInvoke]AppCall failed:" + err.Error()))
		}

		result = tmpRes

	case WASMVM_CcntmRACT:
		conParam := states.WasmCcntmractParam{Address: ccntmractAddress, Args: inputs}
		sink := common.NewZeroCopySink(nil)
		conParam.Serialization(sink)

		newservice, err := self.Service.CcntmextRef.NewExecuteEngine(sink.Bytes(), types.InvokeWasm)
		if err != nil {
			panic(err)
		}

		tmpRes, err := newservice.Invoke()
		if err != nil {
			panic(err)
		}

		result = tmpRes.([]byte)

	case NEOVM_CcntmRACT:

		parambytes, err := util.CreateNeoInvokeParam(ccntmractAddress, inputs)
		if err != nil {
			panic(err)
		}

		neoservice, err := self.Service.CcntmextRef.NewExecuteEngine(parambytes, types.InvokeNeo)
		if err != nil {
			panic(err)
		}

		tmp, err := neoservice.Invoke()
		if err != nil {
			panic(err)
		}

		if tmp != nil {
			val := tmp.(*neotypes.VmValue)
			source := common.NewZeroCopySink([]byte{byte(crossvm_codec.VERSION)})

			err = neotypes.BuildResultFromNeo(*val, source)
			if err != nil {
				panic(err)
			}
			result = source.Bytes()
		}

	default:
		panic(errors.NewErr("Not a supported ccntmract type"))
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
			Host: reflect.ValueOf(TimeStamp),
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
			Host: reflect.ValueOf(CallCcntmract),
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
			Host: reflect.ValueOf(CcntmractCreate),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{ //20
			Sig:  &m.Types.Entries[9],
			Host: reflect.ValueOf(CcntmractMigrate),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{ //21
			Sig:  &m.Types.Entries[10],
			Host: reflect.ValueOf(CcntmractDelete),
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
			"cntmio_timestamp": {
				FieldStr: "cntmio_timestamp",
				Kind:     wasm.ExternalFunction,
				Index:    0,
			},
			"cntmio_block_height": {
				FieldStr: "cntmio_block_height",
				Kind:     wasm.ExternalFunction,
				Index:    1,
			},
			"cntmio_input_length": {
				FieldStr: "cntmio_input_length",
				Kind:     wasm.ExternalFunction,
				Index:    2,
			},
			"cntmio_call_output_length": {
				FieldStr: "cntmio_call_output_length",
				Kind:     wasm.ExternalFunction,
				Index:    3,
			},
			"cntmio_self_address": {
				FieldStr: "cntmio_self_address",
				Kind:     wasm.ExternalFunction,
				Index:    4,
			},
			"cntmio_caller_address": {
				FieldStr: "cntmio_caller_address",
				Kind:     wasm.ExternalFunction,
				Index:    5,
			},
			"cntmio_entry_address": {
				FieldStr: "cntmio_entry_address",
				Kind:     wasm.ExternalFunction,
				Index:    6,
			},
			"cntmio_get_input": {
				FieldStr: "cntmio_get_input",
				Kind:     wasm.ExternalFunction,
				Index:    7,
			},
			"cntmio_get_call_output": {
				FieldStr: "cntmio_get_call_output",
				Kind:     wasm.ExternalFunction,
				Index:    8,
			},
			"cntmio_check_witness": {
				FieldStr: "cntmio_check_witness",
				Kind:     wasm.ExternalFunction,
				Index:    9,
			},
			"cntmio_current_blockhash": {
				FieldStr: "cntmio_current_blockhash",
				Kind:     wasm.ExternalFunction,
				Index:    10,
			},
			"cntmio_current_txhash": {
				FieldStr: "cntmio_current_txhash",
				Kind:     wasm.ExternalFunction,
				Index:    11,
			},
			"cntmio_return": {
				FieldStr: "cntmio_return",
				Kind:     wasm.ExternalFunction,
				Index:    12,
			},
			"cntmio_notify": {
				FieldStr: "cntmio_notify",
				Kind:     wasm.ExternalFunction,
				Index:    13,
			},
			"cntmio_debug": {
				FieldStr: "cntmio_debug",
				Kind:     wasm.ExternalFunction,
				Index:    14,
			},
			"cntmio_call_ccntmract": {
				FieldStr: "cntmio_call_ccntmract",
				Kind:     wasm.ExternalFunction,
				Index:    15,
			},
			"cntmio_storage_read": {
				FieldStr: "cntmio_storage_read",
				Kind:     wasm.ExternalFunction,
				Index:    16,
			},
			"cntmio_storage_write": {
				FieldStr: "cntmio_storage_write",
				Kind:     wasm.ExternalFunction,
				Index:    17,
			},
			"cntmio_storage_delete": {
				FieldStr: "cntmio_storage_delete",
				Kind:     wasm.ExternalFunction,
				Index:    18,
			},
			"cntmio_ccntmract_create": {
				FieldStr: "cntmio_ccntmract_create",
				Kind:     wasm.ExternalFunction,
				Index:    19,
			},
			"cntmio_ccntmract_migrate": {
				FieldStr: "cntmio_ccntmract_migrate",
				Kind:     wasm.ExternalFunction,
				Index:    20,
			},
			"cntmio_ccntmract_destroy": {
				FieldStr: "cntmio_ccntmract_destroy",
				Kind:     wasm.ExternalFunction,
				Index:    21,
			},
			"cntmio_panic": {
				FieldStr: "cntmio_panic",
				Kind:     wasm.ExternalFunction,
				Index:    22,
			},
			"cntmio_sha256": {
				FieldStr: "cntmio_sha256",
				Kind:     wasm.ExternalFunction,
				Index:    23,
			},
		},
	}

	return m
}

func (self *Runtime) getCcntmractType(addr common.Address) (CcntmractType, error) {
	if utils.IsNativeCcntmract(addr) {
		return NATIVE_CcntmRACT, nil
	}

	dep, err := self.Service.CacheDB.GetCcntmract(addr)
	if err != nil {
		return UNKOWN_CcntmRACT, err
	}
	if dep == nil {
		return UNKOWN_CcntmRACT, errors.NewErr("ccntmract is not exist.")
	}
	if dep.VmType == payload.WASMVM_TYPE {
		return WASMVM_CcntmRACT, nil
	}

	return NEOVM_CcntmRACT, nil

}

func (self *Runtime) checkGas(gaslimit uint64) {
	gas := self.Service.vm.AvaliableGas
	if *gas.GasLimit >= gaslimit {
		*gas.GasLimit -= gaslimit
	} else {
		panic(errors.NewErr("[wasm_Service]Insufficient gas limit"))
	}
}

func serializeStorageKey(ccntmractAddress common.Address, key []byte) []byte {
	bf := new(bytes.Buffer)

	bf.Write(ccntmractAddress[:])
	bf.Write(key)

	return bf.Bytes()
}
