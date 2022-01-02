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
	"strings"
	"github.com/Ontology/common"
	"github.com/Ontology/common/log"
	"github.com/Ontology/core/payload"
	"github.com/Ontology/core/states"
	"github.com/Ontology/core/store"
	"github.com/Ontology/core/types"
	"github.com/Ontology/errors"
	. "github.com/Ontology/smartccntmract/common"
	"github.com/Ontology/smartccntmract/event"
	trigger "github.com/Ontology/smartccntmract/types"
	vm "github.com/Ontology/vm/neovm"
	vmtypes "github.com/Ontology/vm/neovm/types"
	"github.com/cntmio/cntmology-crypto/keypair"
)

var (
	ERR_DB_NOT_FOUND = "leveldb: not found"
	LOG = "Log"
)

type StateReader struct {
	serviceMap    map[string]func(*vm.ExecutionEngine) (bool, error)
	trigger       trigger.TriggerType
	Notifications []*event.NotifyEventInfo
	ldgerStore    store.LedgerStore
}

func NewStateReader(ldgerStore store.LedgerStore, trigger trigger.TriggerType) *StateReader {
	var stateReader StateReader
	stateReader.ldgerStore = ldgerStore
	stateReader.serviceMap = make(map[string]func(*vm.ExecutionEngine) (bool, error), 0)
	stateReader.trigger = trigger

	stateReader.Register("Neo.Runtime.GetTrigger", stateReader.RuntimeGetTrigger)
	stateReader.Register("Neo.Runtime.GetTime", stateReader.RuntimeGetTime)
	stateReader.Register("Neo.Runtime.CheckWitness", stateReader.RuntimeCheckWitness)
	stateReader.Register("Neo.Runtime.Notify", stateReader.RuntimeNotify)
	stateReader.Register("Neo.Runtime.Log", stateReader.RuntimeLog)

	stateReader.Register("Neo.Blockchain.GetHeight", stateReader.BlockChainGetHeight)
	stateReader.Register("Neo.Blockchain.GetHeader", stateReader.BlockChainGetHeader)
	stateReader.Register("Neo.Blockchain.GetBlock", stateReader.BlockChainGetBlock)
	stateReader.Register("Neo.Blockchain.GetTransaction", stateReader.BlockChainGetTransaction)
	stateReader.Register("Neo.Blockchain.GetCcntmract", stateReader.GetCcntmract)

	stateReader.Register("Neo.Header.GetHash", stateReader.HeaderGetHash)
	stateReader.Register("Neo.Header.GetVersion", stateReader.HeaderGetVersion)
	stateReader.Register("Neo.Header.GetPrevHash", stateReader.HeaderGetPrevHash)
	stateReader.Register("Neo.Header.GetMerkleRoot", stateReader.HeaderGetMerkleRoot)
	stateReader.Register("Neo.Header.GetIndex", stateReader.HeaderGetIndex)
	stateReader.Register("Neo.Header.GetTimestamp", stateReader.HeaderGetTimestamp)
	stateReader.Register("Neo.Header.GetConsensusData", stateReader.HeaderGetConsensusData)
	stateReader.Register("Neo.Header.GetNextConsensus", stateReader.HeaderGetNextConsensus)

	stateReader.Register("Neo.Block.GetTransactionCount", stateReader.BlockGetTransactionCount)
	stateReader.Register("Neo.Block.GetTransactions", stateReader.BlockGetTransactions)
	stateReader.Register("Neo.Block.GetTransaction", stateReader.BlockGetTransaction)

	stateReader.Register("Neo.Transaction.GetHash", stateReader.TransactionGetHash)
	stateReader.Register("Neo.Transaction.GetType", stateReader.TransactionGetType)
	stateReader.Register("Neo.Transaction.GetAttributes", stateReader.TransactionGetAttributes)

	stateReader.Register("Neo.Attribute.GetUsage", stateReader.AttributeGetUsage)
	stateReader.Register("Neo.Attribute.GetData", stateReader.AttributeGetData)

	stateReader.Register("Neo.Storage.GetScript", stateReader.CcntmractGetCode)
	stateReader.Register("Neo.Storage.GetCcntmext", stateReader.StorageGetCcntmext)
	stateReader.Register("Neo.Storage.Get", stateReader.StorageGet)

	return &stateReader
}

func (s *StateReader) Register(methodName string, handler func(*vm.ExecutionEngine) (bool, error)) bool {
	s.serviceMap[methodName] = handler
	return true
}

func (s *StateReader) GetServiceMap() map[string]func(*vm.ExecutionEngine) (bool, error) {
	return s.serviceMap
}

func (s *StateReader) RuntimeGetTrigger(e *vm.ExecutionEngine) (bool, error) {
	vm.PushData(e, int(s.trigger))
	return true, nil
}

func (s *StateReader) RuntimeGetTime(e *vm.ExecutionEngine) (bool, error) {
	hash := s.ldgerStore.GetCurrentBlockHash()
	header, err := s.ldgerStore.GetHeaderByHash(hash)
	if err != nil {
		return false, errors.NewDetailErr(err, errors.ErrNoCode, "[RuntimeGetTime] GetHeader error!.")
	}

	vm.PushData(e, header.Timestamp)
	return true, nil
}

func (s *StateReader) RuntimeNotify(e *vm.ExecutionEngine) (bool, error) {
	item := vm.PopStackItem(e)
	ccntmainer := e.GetCodeCcntmainer()
	if ccntmainer == nil {
		log.Error("[RuntimeNotify] Get ccntmainer fail!")
		return false, errors.NewErr("[CreateAsset] Get ccntmainer fail!")
	}
	tran, ok := ccntmainer.(*types.Transaction)
	if !ok {
		log.Error("[RuntimeNotify] Ccntmainer not transaction!")
		return false, errors.NewErr("[CreateAsset] Ccntmainer not transaction!")
	}
	ccntmext, err := e.CurrentCcntmext()
	if err != nil {
		return false, err
	}
	hash, err := ccntmext.GetCodeHash()
	if err != nil {
		return false, err
	}
	txid := tran.Hash()
	s.Notifications = append(s.Notifications, &event.NotifyEventInfo{TxHash: txid, CodeHash: hash, States: ConvertReturnTypes(item)})
	return true, nil
}

func (s *StateReader) RuntimeLog(e *vm.ExecutionEngine) (bool, error) {
	item := vm.PopByteArray(e)
	ccntmainer := e.GetCodeCcntmainer()
	if ccntmainer == nil {
		log.Error("[RuntimeLog] Get ccntmainer fail!")
		return false, errors.NewErr("[CreateAsset] Get ccntmainer fail!")
	}
	tran, ok := ccntmainer.(*types.Transaction)
	if !ok {
		log.Error("[RuntimeLog] Ccntmainer not transaction!")
		return false, errors.NewErr("[CreateAsset] Ccntmainer not transaction!")
	}
	ccntmext, err := e.CurrentCcntmext()
	if err != nil {
		return false, err
	}
	hash, err := ccntmext.GetCodeHash()
	if err != nil {
		return false, err
	}
	event.PushSmartCodeEvent(tran.Hash(), 0, LOG, event.LogEventArgs{tran.Hash(), hash, string(item)})
	return true, nil
}

func (s *StateReader) CheckWitness(engine *vm.ExecutionEngine, address common.Address) (bool, error) {
	tx := engine.GetCodeCcntmainer().(*types.Transaction)
	addresses := tx.GetSignatureAddresses()
	return ccntmains(addresses, address), nil
}

func (s *StateReader) RuntimeCheckWitness(e *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(e) < 1 {
		return false, errors.NewErr("[RuntimeCheckWitness] Too few input parameters ")
	}
	data := vm.PopByteArray(e)
	var addr common.Address
	if len(data) == 20 {
		temp, err := common.AddressParseFromBytes(data)
		if err != nil {
			return false, err
		}
		addr = temp
	} else {
		publicKey, err := keypair.DeserializePublicKey(data)
		if err != nil {
			return false, fmt.Errorf("[RuntimeCheckWitness] data invalid: %s", err)
		}
		addr = types.AddressFromPubKey(publicKey)
	}

	result, err := s.CheckWitness(e, addr)
	if err != nil {
		return false, err
	}
	vm.PushData(e, result)
	return true, nil
}

func (s *StateReader) BlockChainGetHeight(e *vm.ExecutionEngine) (bool, error) {
	var i uint32
	i = s.ldgerStore.GetCurrentBlockHeight()
	vm.PushData(e, i)
	return true, nil
}

func (s *StateReader) BlockChainGetHeader(e *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(e) < 1 {
		return false, errors.NewErr("[BlockChainGetHeader] Too few input parameters ")
	}
	data := vm.PopByteArray(e)
	var (
		header *types.Header
		err error
	)
	l := len(data)
	if l <= 5 {
		b := new(big.Int)
		height := uint32(b.SetBytes(common.BytesReverse(data)).Int64())
		hash := s.ldgerStore.GetBlockHash(height)
		header, err = s.ldgerStore.GetHeaderByHash(hash)
		if err != nil {
			return false, errors.NewDetailErr(err, errors.ErrNoCode, "[BlockChainGetHeader] GetHeader error!.")
		}

	} else if l == 32 {
		hash, _ := common.Uint256ParseFromBytes(data)
		header, err = s.ldgerStore.GetHeaderByHash(hash)
		if err != nil {
			return false, errors.NewDetailErr(err, errors.ErrNoCode, "[BlockChainGetHeader] GetHeader error!.")
		}
	} else {
		return false, errors.NewErr("[BlockChainGetHeader] data invalid.")
	}
	vm.PushData(e, header)
	return true, nil
}

func (s *StateReader) BlockChainGetBlock(e *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(e) < 1 {
		return false, errors.NewErr("[BlockChainGetBlock] Too few input parameters ")
	}
	data := vm.PopByteArray(e)
	var (
		block *types.Block
	)
	l := len(data)
	if l <= 5 {
		b := new(big.Int)
		height := uint32(b.SetBytes(common.BytesReverse(data)).Int64())
		var err error
		block, err = s.ldgerStore.GetBlockByHeight(height)
		if err != nil {
			return false, errors.NewDetailErr(err, errors.ErrNoCode, "[BlockChainGetBlock] GetBlock error!.")
		}
	} else if l == 32 {
		hash, err := common.Uint256ParseFromBytes(data)
		if err != nil {
			return false, err
		}
		block, err = s.ldgerStore.GetBlockByHash(hash)
		if err != nil {
			return false, errors.NewDetailErr(err, errors.ErrNoCode, "[BlockChainGetBlock] GetBlock error!.")
		}
	} else {
		return false, errors.NewErr("[BlockChainGetBlock] data invalid.")
	}
	vm.PushData(e, block)
	return true, nil
}

func (s *StateReader) BlockChainGetTransaction(e *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(e) < 1 {
		return false, errors.NewErr("[BlockChainGetTransaction] Too few input parameters ")
	}
	d := vm.PopByteArray(e)
	hash, err := common.Uint256ParseFromBytes(d)
	if err != nil {
		return false, err
	}
	t, _, err := s.ldgerStore.GetTransaction(hash)
	if err != nil {
		return false, errors.NewDetailErr(err, errors.ErrNoCode, "[BlockChainGetTransaction] GetTransaction error!")
	}

	vm.PushData(e, t)
	return true, nil
}

func (s *StateReader) GetCcntmract(e *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(e) < 1 {
		return false, errors.NewErr("[GetCcntmract] Too few input parameters ")
	}
	hashByte := vm.PopByteArray(e)
	hash, err := common.AddressParseFromBytes(hashByte)
	if err != nil {
		return false, err
	}
	item, err := s.ldgerStore.GetCcntmractState(hash)
	if err != nil {
		return false, errors.NewDetailErr(err, errors.ErrNoCode, "[GetCcntmract] GetAsset error!")
	}
	vm.PushData(e, item)
	return true, nil
}

func (s *StateReader) HeaderGetHash(e *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(e) < 1 {
		return false, errors.NewErr("[HeaderGetHash] Too few input parameters ")
	}
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.NewErr("[HeaderGetHash] Pop blockdata nil!")
	}
	var data *types.Header
	if b, ok := d.(*types.Block); ok {
		data = b.Header
	} else if h, ok := d.(*types.Header); ok {
		data = h
	} else {
		return false, errors.NewErr("[HeaderGetHash] Wrcntm type!")
	}
	h := data.Hash()
	vm.PushData(e, h.ToArray())
	return true, nil
}

func (s *StateReader) HeaderGetVersion(e *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(e) < 1 {
		return false, errors.NewErr("[HeaderGetVersion] Too few input parameters ")
	}
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.NewErr("[HeaderGetVersion] Pop blockdata nil!")
	}
	var data *types.Header
	if b, ok := d.(*types.Block); ok {
		data = b.Header
	} else if h, ok := d.(*types.Header); ok {
		data = h
	} else {
		return false, errors.NewErr("[HeaderGetVersion] Wrcntm type!")
	}
	vm.PushData(e, data.Version)
	return true, nil
}

func (s *StateReader) HeaderGetPrevHash(e *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(e) < 1 {
		return false, errors.NewErr("[HeaderGetPrevHash] Too few input parameters ")
	}
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.NewErr("[HeaderGetPrevHash] Pop blockdata nil!")
	}
	var data *types.Header
	if b, ok := d.(*types.Block); ok {
		data = b.Header
	} else if h, ok := d.(*types.Header); ok {
		data = h
	} else {
		return false, errors.NewErr("[HeaderGetPrevHash] Wrcntm type!")
	}
	vm.PushData(e, data.PrevBlockHash.ToArray())
	return true, nil
}

func (s *StateReader) HeaderGetMerkleRoot(e *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(e) < 1 {
		return false, errors.NewErr("[HeaderGetMerkleRoot] Too few input parameters ")
	}
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.NewErr("[HeaderGetMerkleRoot] Pop blockdata nil!")
	}
	var data *types.Header
	if b, ok := d.(*types.Block); ok {
		data = b.Header
	} else if h, ok := d.(*types.Header); ok {
		data = h
	} else {
		return false, errors.NewErr("[HeaderGetMerkleRoot] Wrcntm type!")
	}
	vm.PushData(e, data.TransactionsRoot.ToArray())
	return true, nil
}

func (s *StateReader) HeaderGetIndex(e *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(e) < 1 {
		return false, errors.NewErr("[HeaderGetIndex] Too few input parameters ")
	}
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.NewErr("[HeaderGetIndex] Pop blockdata nil!")
	}
	var data *types.Header
	if b, ok := d.(*types.Block); ok {
		data = b.Header
	} else if h, ok := d.(*types.Header); ok {
		data = h
	} else {
		return false, errors.NewErr("[HeaderGetIndex] Wrcntm type!")
	}
	vm.PushData(e, data.Height)
	return true, nil
}

func (s *StateReader) HeaderGetTimestamp(e *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(e) < 1 {
		return false, errors.NewErr("[HeaderGetTimestamp] Too few input parameters ")
	}
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.NewErr("[HeaderGetTimestamp] Pop blockdata nil!")
	}
	var data *types.Header
	if b, ok := d.(*types.Block); ok {
		data = b.Header
	} else if h, ok := d.(*types.Header); ok {
		data = h
	} else {
		return false, errors.NewErr("[HeaderGetTimestamp] Wrcntm type!")
	}
	vm.PushData(e, data.Timestamp)
	return true, nil
}

func (s *StateReader) HeaderGetConsensusData(e *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(e) < 1 {
		return false, errors.NewErr("[HeaderGetConsensusData] Too few input parameters ")
	}
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.NewErr("[HeaderGetConsensusData] Pop blockdata nil!")
	}
	var data *types.Header
	if b, ok := d.(*types.Block); ok {
		data = b.Header
	} else if h, ok := d.(*types.Header); ok {
		data = h
	} else {
		return false, errors.NewErr("[HeaderGetConsensusData] Wrcntm type!")
	}
	vm.PushData(e, data.ConsensusData)
	return true, nil
}

func (s *StateReader) HeaderGetNextConsensus(e *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(e) < 1 {
		return false, errors.NewErr("[HeaderGetNextConsensus] Too few input parameters ")
	}
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.NewErr("[HeaderGetNextConsensus] Pop blockdata nil!")
	}
	var data *types.Header
	if b, ok := d.(*types.Block); ok {
		data = b.Header
	} else if h, ok := d.(*types.Header); ok {
		data = h
	} else {
		return false, errors.NewErr("[HeaderGetNextConsensus] Wrcntm type!")
	}
	vm.PushData(e, data.NextBookkeeper[:])
	return true, nil
}

func (s *StateReader) BlockGetTransactionCount(e *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(e) < 1 {
		return false, errors.NewErr("[BlockGetTransactionCount] Too few input parameters ")
	}
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.NewErr("[BlockGetTransactionCount] Pop blockdata nil!")
	}
	block, ok := d.(*types.Block)
	if ok == false {
		return false, errors.NewErr("[BlockGetTransactionCount] Wrcntm type!")
	}
	transactions := block.Transactions
	vm.PushData(e, len(transactions))
	return true, nil
}

func (s *StateReader) BlockGetTransactions(e *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(e) < 1 {
		return false, errors.NewErr("[BlockGetTransactions] Too few input parameters ")
	}
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.NewErr("[BlockGetTransactions] Pop blockdata nil!")
	}
	block, ok := d.(*types.Block)
	if ok == false {
		return false, errors.NewErr("[BlockGetTransactions] Wrcntm type!")
	}
	transactions := block.Transactions
	transactionList := make([]vmtypes.StackItems, 0)
	for _, v := range transactions {
		transactionList = append(transactionList, vmtypes.NewInteropInterface(v))
	}
	vm.PushData(e, transactionList)
	return true, nil
}

func (s *StateReader) BlockGetTransaction(e *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(e) < 2 {
		return false, errors.NewErr("[BlockGetTransaction] Too few input parameters ")
	}
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.NewErr("[BlockGetTransaction] Pop transactions nil!")
	}
	index := vm.PopInt(e)
	if index < 0 {
		return false, errors.NewErr("[BlockGetTransaction] Pop index invalid!")
	}

	block, ok := d.(*types.Block)
	if ok == false {
		return false, errors.NewErr("[BlockGetTransactions] Wrcntm type!")
	}
	transactions := block.Transactions
	if index >= len(transactions) {
		return false, errors.NewErr("[BlockGetTransaction] index invalid!")
	}
	vm.PushData(e, transactions[index])
	return true, nil
}

func (s *StateReader) TransactionGetHash(e *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(e) < 1 {
		return false, errors.NewErr("[TransactionGetHash] Too few input parameters ")
	}
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.NewErr("[TransactionGetHash] Pop transaction nil!")
	}

	txn, ok := d.(*types.Transaction)
	if ok == false {
		return false, errors.NewErr("[TransactionGetHash] Wrcntm type!")
	}
	txHash := txn.Hash()
	vm.PushData(e, txHash.ToArray())
	return true, nil
}

func (s *StateReader) TransactionGetType(e *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(e) < 1 {
		return false, errors.NewErr("[TransactionGetType] Too few input parameters ")
	}
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.NewErr("[TransactionGetType] Pop transaction nil!")
	}
	txn, ok := d.(*types.Transaction)
	if ok == false {
		return false, errors.NewErr("[TransactionGetHash] Wrcntm type!")
	}
	txType := txn.TxType
	vm.PushData(e, int(txType))
	return true, nil
}

func (s *StateReader) TransactionGetAttributes(e *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(e) < 1 {
		return false, errors.NewErr("[TransactionGetAttributes] Too few input parameters ")
	}
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.NewErr("[TransactionGetAttributes] Pop transaction nil!")
	}
	txn, ok := d.(*types.Transaction)
	if ok == false {
		return false, errors.NewErr("[TransactionGetAttributes] Wrcntm type!")
	}
	attributes := txn.Attributes
	attributList := make([]vmtypes.StackItems, 0)
	for _, v := range attributes {
		attributList = append(attributList, vmtypes.NewInteropInterface(v))
	}
	vm.PushData(e, attributList)
	return true, nil
}

func (s *StateReader) AttributeGetUsage(e *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(e) < 1 {
		return false, errors.NewErr("[AttributeGetUsage] Too few input parameters ")
	}
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.NewErr("[AttributeGetUsage] Pop txAttribute nil!")
	}
	attribute, ok := d.(*types.TxAttribute)
	if ok == false {
		return false, errors.NewErr("[AttributeGetUsage] Wrcntm type!")
	}
	vm.PushData(e, int(attribute.Usage))
	return true, nil
}

func (s *StateReader) AttributeGetData(e *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(e) < 1 {
		return false, errors.NewErr("[AttributeGetData] Too few input parameters ")
	}
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.NewErr("[AttributeGetData] Pop txAttribute nil!")
	}
	attribute, ok := d.(*types.TxAttribute)
	if ok == false {
		return false, errors.NewErr("[AttributeGetUsage] Wrcntm type!")
	}
	vm.PushData(e, attribute.Data)
	return true, nil
}

func (s *StateReader) CcntmractGetCode(e *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(e) < 1 {
		return false, errors.NewErr("[CcntmractGetCode] Too few input parameters ")
	}
	d := vm.PopInteropInterface(e)
	if d == nil {
		return false, errors.NewErr("[CcntmractGetCode] Pop ccntmractState nil!")
	}
	ccntmractState, ok := d.(*payload.DeployCode)
	if ok == false {
		return false, errors.NewErr("[CcntmractGetCode] Wrcntm type!")
	}
	vm.PushData(e, ccntmractState.Code)
	return true, nil
}

func (s *StateReader) StorageGetCcntmext(e *vm.ExecutionEngine) (bool, error) {
	ccntmext, err := e.CurrentCcntmext()
	if err != nil {
		return false, err
	}
	hash, err := ccntmext.GetCodeHash()
	if err != nil {
		return false, err
	}
	vm.PushData(e, NewStorageCcntmext(hash))
	return true, nil
}

func (s *StateReader) StorageGet(e *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(e) < 2 {
		return false, errors.NewErr("[StorageGet] Too few input parameters ")
	}
	opInterface := vm.PopInteropInterface(e)
	if opInterface == nil {
		return false, errors.NewErr("[StorageGet] Get StorageCcntmext error!")
	}
	ccntmext, ok := opInterface.(*StorageCcntmext)
	if ok == false {
		return false, errors.NewErr("[StorageGet] Wrcntm type!")
	}
	c, err := s.ldgerStore.GetCcntmractState(ccntmext.codeHash)
	if err != nil && !strings.EqualFold(err.Error(), ERR_DB_NOT_FOUND) {
		return false, err
	}
	if c == nil {
		return false, nil
	}
	key := vm.PopByteArray(e)
	item, err := s.ldgerStore.GetStorageItem(&states.StorageKey{CodeHash: ccntmext.codeHash, Key: key})
	if err != nil && !strings.EqualFold(err.Error(), ERR_DB_NOT_FOUND) {
		return false, err
	}
	if item == nil {
		vm.PushData(e, []byte{})
	} else {
		vm.PushData(e, item.Value)
	}
	return true, nil
}
