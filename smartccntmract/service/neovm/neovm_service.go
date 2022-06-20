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
	"math/big"

	"github.com/cntmio/cntmology/core/store"
	"github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/errors"
	"github.com/cntmio/cntmology/smartccntmract/ccntmext"
	"github.com/cntmio/cntmology/smartccntmract/event"
	"github.com/cntmio/cntmology/smartccntmract/states"
	"github.com/cntmio/cntmology/smartccntmract/storage"
	vm "github.com/cntmio/cntmology/vm/neovm"
	vmtype "github.com/cntmio/cntmology/vm/neovm/types"
)

const (
	MAX_STACK_SIZE          = 2 * 1024
	MAX_ARRAY_SIZE          = 1024
	MAX_SIZE_FOR_BIGINTEGER = 32
)

var (
	// Register all service for smart ccntmract execute
	ServiceMap = map[string]Service{
		"Neo.Attribute.GetUsage":                        {Execute: AttributeGetUsage, Validator: validatorAttribute},
		"Neo.Attribute.GetData":                         {Execute: AttributeGetData, Validator: validatorAttribute},
		"Neo.Block.GetTransactionCount":                 {Execute: BlockGetTransactionCount, Validator: validatorBlock},
		"Neo.Block.GetTransactions":                     {Execute: BlockGetTransactions, Validator: validatorBlock},
		"Neo.Block.GetTransaction":                      {Execute: BlockGetTransaction, Validator: validatorBlockTransaction},
		"Neo.Blockchain.GetHeight":                      {Execute: BlockChainGetHeight},
		"Neo.Blockchain.GetHeader":                      {Execute: BlockChainGetHeader, Validator: validatorBlockChainHeader},
		"Neo.Blockchain.GetBlock":                       {Execute: BlockChainGetBlock, Validator: validatorBlockChainBlock},
		"Neo.Blockchain.GetTransaction":                 {Execute: BlockChainGetTransaction, Validator: validatorBlockChainTransaction},
		"Neo.Blockchain.GetCcntmract":                    {Execute: BlockChainGetCcntmract, Validator: validatorBlockChainCcntmract},
		"Neo.Header.GetIndex":                           {Execute: HeaderGetIndex, Validator: validatorHeader},
		"Neo.Header.GetHash":                            {Execute: HeaderGetHash, Validator: validatorHeader},
		"Neo.Header.GetVersion":                         {Execute: HeaderGetVersion, Validator: validatorHeader},
		"Neo.Header.GetPrevHash":                        {Execute: HeaderGetPrevHash, Validator: validatorHeader},
		"Neo.Header.GetTimestamp":                       {Execute: HeaderGetTimestamp, Validator: validatorHeader},
		"Neo.Header.GetConsensusData":                   {Execute: HeaderGetConsensusData, Validator: validatorHeader},
		"Neo.Header.GetNextConsensus":                   {Execute: HeaderGetNextConsensus, Validator: validatorHeader},
		"Neo.Header.GetMerkleRoot":                      {Execute: HeaderGetMerkleRoot, Validator: validatorHeader},
		"Neo.Transaction.GetHash":                       {Execute: TransactionGetHash, Validator: validatorTransaction},
		"Neo.Transaction.GetType":                       {Execute: TransactionGetType, Validator: validatorTransaction},
		"Neo.Transaction.GetAttributes":                 {Execute: TransactionGetAttributes, Validator: validatorTransaction},
		"Neo.Ccntmract.Create":                           {Execute: CcntmractCreate},
		"Neo.Ccntmract.Migrate":                          {Execute: CcntmractMigrate},
		"Neo.Ccntmract.GetStorageCcntmext":                {Execute: CcntmractGetStorageCcntmext},
		"Neo.Ccntmract.Destroy":                          {Execute: CcntmractDestory},
		"Neo.Ccntmract.GetScript":                        {Execute: CcntmractGetCode, Validator: validatorGetCode},
		"Neo.Runtime.GetTime":                           {Execute: RuntimeGetTime},
		"Neo.Runtime.CheckWitness":                      {Execute: RuntimeCheckWitness, Validator: validatorCheckWitness},
		"Neo.Runtime.Notify":                            {Execute: RuntimeNotify, Validator: validatorNotify},
		"Neo.Runtime.Log":                               {Execute: RuntimeLog, Validator: validatorLog},
		"Neo.Runtime.CheckSig":                          {Execute: RuntimeCheckSig, Validator: validatorCheckSig},
		"Neo.Storage.Get":                               {Execute: StorageGet},
		"Neo.Storage.Put":                               {Execute: StoragePut},
		"Neo.Storage.Delete":                            {Execute: StorageDelete},
		"Neo.Storage.GetCcntmext":                        {Execute: StorageGetCcntmext},
		"System.ExecutionEngine.GetScriptCcntmainer":     {Execute: GetCodeCcntmainer},
		"System.ExecutionEngine.GetExecutingScriptHash": {Execute: GetExecutingAddress},
		"System.ExecutionEngine.GetCallingScriptHash":   {Execute: GetCallingAddress},
		"System.ExecutionEngine.GetEntryScriptHash":     {Execute: GetEntryAddress},
	}
)

var (
	ERR_CHECK_STACK_SIZE    = errors.NewErr("[NeoVmService] vm over max stack size!")
	ERR_CHECK_ARRAY_SIZE    = errors.NewErr("[NeoVmService] vm over max array size!")
	ERR_CHECK_BIGINTEGER    = errors.NewErr("[NeoVmService] vm over max biginteger size!")
	ERR_CURRENT_CcntmEXT_NIL = errors.NewErr("[NeoVmService] neovm service current ccntmext doesn't exist!")
	ERR_EXECUTE_CODE        = errors.NewErr("[NeoVmService] vm execute code invalid!")
	ERR_GAS_INSUFFICIENT    = errors.NewErr("[NeoVmService] gas insufficient")
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
}

// Invoke a smart ccntmract
func (this *NeoVmService) Invoke() (interface{}, error) {
	engine := vm.NewExecutionEngine()
	if len(this.Code) == 0 {
		return nil, ERR_EXECUTE_CODE
	}
	engine.PushCcntmext(vm.NewExecutionCcntmext(engine, this.Code))
	for {
		if len(engine.Ccntmexts) == 0 || engine.Ccntmext == nil {
			break
		}
		if engine.Ccntmext.GetInstructionPointer() >= len(engine.Ccntmext.Code) {
			break
		}
		if err := engine.ExecuteCode(); err != nil {
			return nil, err
		}
		if engine.Ccntmext.GetInstructionPointer() < len(engine.Ccntmext.Code) {
			if ok := checkStackSize(engine); !ok {
				return nil, ERR_CHECK_STACK_SIZE
			}
			if ok := checkArraySize(engine); !ok {
				return nil, ERR_CHECK_ARRAY_SIZE
			}
			if ok := checkBigIntegers(engine); !ok {
				return nil, ERR_CHECK_BIGINTEGER
			}
		}
		if !this.CcntmextRef.CheckUseGas(0) {
			return nil, ERR_GAS_INSUFFICIENT
		}
		switch engine.OpCode {
		case vm.SYSCALL:
			if err := this.SystemCall(engine); err != nil {
				return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[NeoVmService] service system call error!")
			}
		case vm.APPCALL, vm.TAILCALL:
			c := new(states.Ccntmract)
			if err := c.Deserialize(engine.Ccntmext.OpReader.Reader()); err != nil {
				return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[NeoVmService] get ccntmract parameters error!")
			}
			result, err := this.CcntmextRef.AppCall(c.Address, c.Method, c.Code, c.Args)
			if err != nil {
				return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[NeoVmService] service app call error!")
			}
			if result != nil {
				vm.PushData(engine, result)
			}
		default:
			if err := engine.StepInto(); err != nil {
				return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[NeoVmService] vm execute error!")
			}
		}
	}
	this.CcntmextRef.PushNotifications(this.Notifications)
	if engine.EvaluationStack.Count() != 0 {
		return engine.EvaluationStack.Peek(0), nil
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
			size = len(vm.PeekStackItem(engine).GetArray())
		}
	}
	size += engine.EvaluationStack.Count() + engine.AltStack.Count()
	if uint32(size) > MAX_STACK_SIZE {
		return false
	}
	return true
}

func checkArraySize(engine *vm.ExecutionEngine) bool {
	switch engine.OpCode {
	case vm.PACK:
	case vm.NEWARRAY:
	case vm.NEWSTRUCT:
		if engine.EvaluationStack.Count() == 0 {
			return false
		}
		size := vm.PeekInt(engine)
		if size > MAX_ARRAY_SIZE {
			return false
		}
	}
	return true
}

func checkBigIntegers(engine *vm.ExecutionEngine) bool {
	switch engine.OpCode {
	case vm.INC:
		if engine.EvaluationStack.Count() == 0 {
			return false
		}
		x := vm.PeekBigInteger(engine)
		if !checkBigInteger(x) || !checkBigInteger(new(big.Int).Add(x, big.NewInt(1))) {
			return false
		}
	case vm.DEC:
		if engine.EvaluationStack.Count() == 0 {
			return false
		}
		x := vm.PeekBigInteger(engine)
		if !checkBigInteger(x) || (x.Sign() < 0 && !checkBigInteger(new(big.Int).Sub(x, big.NewInt(1)))) {
			return false
		}
	case vm.ADD:
		if engine.EvaluationStack.Count() < 2 {
			return false
		}
		x2 := vm.PeekBigInteger(engine)
		x1 := vm.PeekNBigInt(1, engine)
		if !checkBigInteger(x1) || !checkBigInteger(x2) || !checkBigInteger(new(big.Int).Add(x1, x2)) {
			return false
		}
	case vm.SUB:
		if engine.EvaluationStack.Count() < 2 {
			return false
		}
		x2 := vm.PeekBigInteger(engine)
		x1 := vm.PeekNBigInt(1, engine)
		if !checkBigInteger(x1) || !checkBigInteger(x2) || !checkBigInteger(new(big.Int).Sub(x1, x2)) {
			return false
		}
	case vm.MUL:
		if engine.EvaluationStack.Count() < 2 {
			return false
		}
		x2 := vm.PeekBigInteger(engine)
		x1 := vm.PeekNBigInt(1, engine)
		lx2 := len(vmtype.ConvertBigIntegerToBytes(x2))
		lx1 := len(vmtype.ConvertBigIntegerToBytes(x1))
		if lx2 > MAX_SIZE_FOR_BIGINTEGER || lx1 > MAX_SIZE_FOR_BIGINTEGER || (lx1+lx2) > MAX_SIZE_FOR_BIGINTEGER {
			return false
		}
	case vm.DIV:
		if engine.EvaluationStack.Count() < 2 {
			return false
		}
		x2 := vm.PeekBigInteger(engine)
		x1 := vm.PeekNBigInt(1, engine)
		if !checkBigInteger(x2) || !checkBigInteger(x1) {
			return false
		}
		if x2.Sign() == 0 {
			return false
		}
	case vm.MOD:
		if engine.EvaluationStack.Count() < 2 {
			return false
		}
		x2 := vm.PeekBigInteger(engine)
		x1 := vm.PeekNBigInt(1, engine)
		if !checkBigInteger(x2) || !checkBigInteger(x1) {
			return false
		}
		if x2.Sign() == 0 {
			return false
		}
	}
	return true
}

func checkBigInteger(value *big.Int) bool {
	if value == nil {
		return false
	}
	if len(vmtype.ConvertBigIntegerToBytes(value)) > MAX_SIZE_FOR_BIGINTEGER {
		return false
	}
	return true
}
