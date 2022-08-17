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

	"github.com/cntmio/cntmology-crypto/keypair"
	scommon "github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/log"
	"github.com/cntmio/cntmology/core/payload"
	"github.com/cntmio/cntmology/core/signature"
	"github.com/cntmio/cntmology/core/store"
	"github.com/cntmio/cntmology/core/store/common"
	"github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/errors"
	"github.com/cntmio/cntmology/smartccntmract/ccntmext"
	"github.com/cntmio/cntmology/smartccntmract/event"
	"github.com/cntmio/cntmology/smartccntmract/storage"
	vm "github.com/cntmio/cntmology/vm/neovm"
	ntypes "github.com/cntmio/cntmology/vm/neovm/types"
)

var (
	// Register all service for smart ccntmract execute
	ServiceMap = map[string]Service{
		ATTRIBUTE_GETUSAGE_NAME:              {Execute: AttributeGetUsage, Validator: validatorAttribute},
		ATTRIBUTE_GETDATA_NAME:               {Execute: AttributeGetData, Validator: validatorAttribute},
		BLOCK_GETTRANSACTIONCOUNT_NAME:       {Execute: BlockGetTransactionCount, Validator: validatorBlock},
		BLOCK_GETTRANSACTIONS_NAME:           {Execute: BlockGetTransactions, Validator: validatorBlock},
		BLOCK_GETTRANSACTION_NAME:            {Execute: BlockGetTransaction, Validator: validatorBlockTransaction},
		BLOCKCHAIN_GETHEIGHT_NAME:            {Execute: BlockChainGetHeight},
		BLOCKCHAIN_GETHEADER_NAME:            {Execute: BlockChainGetHeader, Validator: validatorBlockChainHeader},
		BLOCKCHAIN_GETBLOCK_NAME:             {Execute: BlockChainGetBlock, Validator: validatorBlockChainBlock},
		BLOCKCHAIN_GETTRANSACTION_NAME:       {Execute: BlockChainGetTransaction, Validator: validatorBlockChainTransaction},
		BLOCKCHAIN_GETCcntmRACT_NAME:          {Execute: BlockChainGetCcntmract, Validator: validatorBlockChainCcntmract},
		BLOCKCHAIN_GETTRANSACTIONHEIGHT_NAME: {Execute: BlockChainGetTransactionHeight},
		HEADER_GETINDEX_NAME:                 {Execute: HeaderGetIndex, Validator: validatorHeader},
		HEADER_GETHASH_NAME:                  {Execute: HeaderGetHash, Validator: validatorHeader},
		HEADER_GETVERSION_NAME:               {Execute: HeaderGetVersion, Validator: validatorHeader},
		HEADER_GETPREVHASH_NAME:              {Execute: HeaderGetPrevHash, Validator: validatorHeader},
		HEADER_GETTIMESTAMP_NAME:             {Execute: HeaderGetTimestamp, Validator: validatorHeader},
		HEADER_GETCONSENSUSDATA_NAME:         {Execute: HeaderGetConsensusData, Validator: validatorHeader},
		HEADER_GETNEXTCONSENSUS_NAME:         {Execute: HeaderGetNextConsensus, Validator: validatorHeader},
		HEADER_GETMERKLEROOT_NAME:            {Execute: HeaderGetMerkleRoot, Validator: validatorHeader},
		TRANSACTION_GETHASH_NAME:             {Execute: TransactionGetHash, Validator: validatorTransaction},
		TRANSACTION_GETTYPE_NAME:             {Execute: TransactionGetType, Validator: validatorTransaction},
		TRANSACTION_GETATTRIBUTES_NAME:       {Execute: TransactionGetAttributes, Validator: validatorTransaction},
		CcntmRACT_CREATE_NAME:                 {Execute: CcntmractCreate},
		CcntmRACT_MIGRATE_NAME:                {Execute: CcntmractMigrate},
		CcntmRACT_GETSTORAGECcntmEXT_NAME:      {Execute: CcntmractGetStorageCcntmext},
		CcntmRACT_DESTROY_NAME:                {Execute: CcntmractDestory},
		CcntmRACT_GETSCRIPT_NAME:              {Execute: CcntmractGetCode, Validator: validatorGetCode},
		RUNTIME_GETTIME_NAME:                 {Execute: RuntimeGetTime},
		RUNTIME_CHECKWITNESS_NAME:            {Execute: RuntimeCheckWitness, Validator: validatorCheckWitness},
		RUNTIME_NOTIFY_NAME:                  {Execute: RuntimeNotify, Validator: validatorNotify},
		RUNTIME_LOG_NAME:                     {Execute: RuntimeLog, Validator: validatorLog},
		RUNTIME_GETTRIGGER_NAME:              {Execute: RuntimeGetTrigger},
		RUNTIME_SERIALIZE_NAME:               {Execute: RuntimeSerialize, Validator: validatorSerialize},
		RUNTIME_DESERIALIZE_NAME:             {Execute: RuntimeDeserialize, Validator: validatorDeserialize},
		NATIVE_INVOKE_NAME:                   {Execute: NativeInvoke},
		STORAGE_GET_NAME:                     {Execute: StorageGet},
		STORAGE_PUT_NAME:                     {Execute: StoragePut},
		STORAGE_DELETE_NAME:                  {Execute: StorageDelete},
		STORAGE_GETCcntmEXT_NAME:              {Execute: StorageGetCcntmext},
		STORAGE_GETREADONLYCcntmEXT_NAME:      {Execute: StorageGetReadOnlyCcntmext},
		STORAGECcntmEXT_ASREADONLY_NAME:       {Execute: StorageCcntmextAsReadOnly},
		GETSCRIPTCcntmAINER_NAME:              {Execute: GetCodeCcntmainer},
		GETEXECUTINGSCRIPTHASH_NAME:          {Execute: GetExecutingAddress},
		GETCALLINGSCRIPTHASH_NAME:            {Execute: GetCallingAddress},
		GETENTRYSCRIPTHASH_NAME:              {Execute: GetEntryAddress},
	}
)

var (
	ERR_CHECK_STACK_SIZE  = errors.NewErr("[NeoVmService] vm over max stack size!")
	ERR_EXECUTE_CODE      = errors.NewErr("[NeoVmService] vm execute code invalid!")
	ERR_GAS_INSUFFICIENT  = errors.NewErr("[NeoVmService] gas insufficient")
	VM_EXEC_STEP_EXCEED   = errors.NewErr("[NeoVmService] vm execute step exceed!")
	CcntmRACT_NOT_EXIST    = errors.NewErr("[NeoVmService] Get ccntmract code from db fail")
	DEPLOYCODE_TYPE_ERROR = errors.NewErr("[NeoVmService] DeployCode type error!")
)

type (
	Execute   func(service *NeoVmService, engine *vm.ExecutionEngine) error
	Validator func(engine *vm.ExecutionEngine) error
)

type Service struct {
	Execute   Execute
	Validator Validator
}

// NeoVmService is a struct for smart ccntmract provide interop service
type NeoVmService struct {
	Store         store.LedgerStore
	CloneCache    *storage.CloneCache
	CcntmextRef    ccntmext.CcntmextRef
	Notifications []*event.NotifyEventInfo
	Code          []byte
	Tx            *types.Transaction
	Time          uint32
	Height        uint32
	Engine        *vm.ExecutionEngine
}

// Invoke a smart ccntmract
func (this *NeoVmService) Invoke() (interface{}, error) {
	if len(this.Code) == 0 {
		return nil, ERR_EXECUTE_CODE
	}
	this.CcntmextRef.PushCcntmext(&ccntmext.Ccntmext{CcntmractAddress: types.AddressFromVmCode(this.Code), Code: this.Code})
	this.Engine.PushCcntmext(vm.NewExecutionCcntmext(this.Engine, this.Code))
	for {
		//check the execution step count
		if !this.CcntmextRef.CheckExecStep() {
			return nil, VM_EXEC_STEP_EXCEED
		}
		if len(this.Engine.Ccntmexts) == 0 || this.Engine.Ccntmext == nil {
			break
		}
		if this.Engine.Ccntmext.GetInstructionPointer() >= len(this.Engine.Ccntmext.Code) {
			break
		}
		if err := this.Engine.ExecuteCode(); err != nil {
			return nil, err
		}
		if this.Engine.Ccntmext.GetInstructionPointer() < len(this.Engine.Ccntmext.Code) {
			if ok := checkStackSize(this.Engine); !ok {
				return nil, ERR_CHECK_STACK_SIZE
			}
		}
		if this.Engine.OpCode >= vm.PUSHBYTES1 && this.Engine.OpCode <= vm.PUSHBYTES75 {
			if !this.CcntmextRef.CheckUseGas(OPCODE_GAS) {
				return nil, ERR_GAS_INSUFFICIENT
			}
		} else {
			if err := this.Engine.ValidateOp(); err != nil {
				return nil, err
			}
			price, err := GasPrice(this.Engine, this.Engine.OpExec.Name)
			if err != nil {
				return nil, err
			}
			if !this.CcntmextRef.CheckUseGas(price) {
				return nil, ERR_GAS_INSUFFICIENT
			}
		}
		switch this.Engine.OpCode {
		case vm.VERIFY:
			if vm.EvaluationStackCount(this.Engine) < 3 {
				return nil, errors.NewErr("[VERIFY] Too few input parameters ")
			}
			pubKey, err := vm.PopByteArray(this.Engine)
			if err != nil {
				return nil, err
			}
			key, err := keypair.DeserializePublicKey(pubKey)
			if err != nil {
				return nil, err
			}
			sig, err := vm.PopByteArray(this.Engine)
			if err != nil {
				return nil, err
			}
			data, err := vm.PopByteArray(this.Engine)
			if err != nil {
				return nil, err
			}
			if err := signature.Verify(key, data, sig); err != nil {
				vm.PushData(this.Engine, false)
			} else {
				vm.PushData(this.Engine, true)
			}
		case vm.SYSCALL:
			if err := this.SystemCall(this.Engine); err != nil {
				return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[NeoVmService] service system call error!")
			}
		case vm.APPCALL, vm.TAILCALL:
			address := this.Engine.Ccntmext.OpReader.ReadBytes(20)
			code, err := this.getCcntmract(address)
			if err != nil {
				return nil, err
			}
			service, err := this.CcntmextRef.NewExecuteEngine(code)
			if err != nil {
				return nil, err
			}
			this.Engine.EvaluationStack.CopyTo(service.(*NeoVmService).Engine.EvaluationStack)
			result, err := service.Invoke()
			if err != nil {
				return nil, err
			}
			if result != nil {
				vm.PushData(this.Engine, result)
			}
		default:
			if err := this.Engine.StepInto(); err != nil {
				return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[NeoVmService] vm execute error!")
			}
		}
	}
	this.CcntmextRef.PopCcntmext()
	this.CcntmextRef.PushNotifications(this.Notifications)
	if this.Engine.EvaluationStack.Count() != 0 {
		return this.Engine.EvaluationStack.Peek(0), nil
	}
	return nil, nil
}

// SystemCall provide register service for smart ccntmract to interaction with blockchain
func (this *NeoVmService) SystemCall(engine *vm.ExecutionEngine) error {
	serviceName := engine.Ccntmext.OpReader.ReadVarString()
	service, ok := ServiceMap[serviceName]
	if !ok {
		return errors.NewErr(fmt.Sprintf("[SystemCall] service not support: %s", serviceName))
	}
	price, err := GasPrice(engine, serviceName)
	if err != nil {
		return err
	}
	if !this.CcntmextRef.CheckUseGas(price) {
		return ERR_GAS_INSUFFICIENT
	}
	if service.Validator != nil {
		if err := service.Validator(engine); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[SystemCall] service validator error!")
		}
	}

	if err := service.Execute(this, engine); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[SystemCall] service execute error!")
	}
	return nil
}

func (this *NeoVmService) getCcntmract(address []byte) ([]byte, error) {
	item, err := this.CloneCache.Store.TryGet(common.ST_CcntmRACT, address)
	if err != nil {
		return nil, errors.NewErr("[getCcntmract] Get ccntmract ccntmext error!")
	}
	log.Infof("invoke ccntmract address:%x", scommon.ToArrayReverse(address))
	if item == nil {
		return nil, CcntmRACT_NOT_EXIST
	}
	ccntmract, ok := item.Value.(*payload.DeployCode)
	if !ok {
		return nil, DEPLOYCODE_TYPE_ERROR
	}
	return ccntmract.Code, nil
}

func checkStackSize(engine *vm.ExecutionEngine) bool {
	size := 0
	if engine.OpCode < vm.PUSH16 {
		size = 1
	} else {
		switch engine.OpCode {
		case vm.DEPTH, vm.DUP, vm.OVER, vm.TUCK:
			size = 1
		case vm.UNPACK:
			if engine.EvaluationStack.Count() == 0 {
				return false
			}
			item := vm.PeekStackItem(engine)
			if a, ok := item.(*ntypes.Array); ok {
				size = a.Count()
			}
			if a, ok := item.(*ntypes.Struct); ok {
				size = a.Count()
			}
		}
	}
	size += engine.EvaluationStack.Count() + engine.AltStack.Count()
	if size > MAX_STACK_SIZE {
		return false
	}
	return true
}
