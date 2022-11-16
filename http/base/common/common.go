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

// Package common privides functions for http handler call
package common

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/cntmio/cntmology-crypto/keypair"
	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/constants"
	"github.com/cntmio/cntmology/common/log"
	"github.com/cntmio/cntmology/common/serialization"
	"github.com/cntmio/cntmology/core/ledger"
	"github.com/cntmio/cntmology/core/payload"
	"github.com/cntmio/cntmology/core/types"
	cutils "github.com/cntmio/cntmology/core/utils"
	cntmErrors "github.com/cntmio/cntmology/errors"
	bactor "github.com/cntmio/cntmology/http/base/actor"
	"github.com/cntmio/cntmology/smartccntmract/event"
	"github.com/cntmio/cntmology/smartccntmract/service/native/cntm"
	"github.com/cntmio/cntmology/smartccntmract/service/native/utils"
	cstate "github.com/cntmio/cntmology/smartccntmract/states"
	"github.com/cntmio/cntmology/vm/neovm"
	"strings"
	"time"
)

const MAX_SEARCH_HEIGHT uint32 = 100
const MAX_REQUEST_BODY_SIZE = 1 << 20

type BalanceOfRsp struct {
	Ont string `json:"cntm"`
	Ong string `json:"cntm"`
}

type MerkleProof struct {
	Type             string
	TransactionsRoot string
	BlockHeight      uint32
	CurBlockRoot     string
	CurBlockHeight   uint32
	TargetHashes     []string
}

type LogEventArgs struct {
	TxHash          string
	CcntmractAddress string
	Message         string
}

type ExecuteNotify struct {
	TxHash      string
	State       byte
	GasConsumed uint64
	Notify      []NotifyEventInfo
}

type PreExecuteResult struct {
	State  byte
	Gas    uint64
	Result interface{}
	Notify []NotifyEventInfo
}

type NotifyEventInfo struct {
	CcntmractAddress string
	States          interface{}
}

type TxAttributeInfo struct {
	Usage types.TransactionAttributeUsage
	Data  string
}

type AmountMap struct {
	Key   common.Uint256
	Value common.Fixed64
}

type Fee struct {
	Amount common.Fixed64
	Payer  string
}

type Sig struct {
	PubKeys []string
	M       uint16
	SigData []string
}
type Transactions struct {
	Version    byte
	Nonce      uint32
	GasPrice   uint64
	GasLimit   uint64
	Payer      string
	TxType     types.TransactionType
	Payload    PayloadInfo
	Attributes []TxAttributeInfo
	Sigs       []Sig
	Hash       string
	Height     uint32
}

type BlockHead struct {
	Version          uint32
	PrevBlockHash    string
	TransactionsRoot string
	BlockRoot        string
	Timestamp        uint32
	Height           uint32
	ConsensusData    uint64
	ConsensusPayload string
	NextBookkeeper   string

	Bookkeepers []string
	SigData     []string

	Hash string
}

type BlockInfo struct {
	Hash         string
	Size         int
	Header       *BlockHead
	Transactions []*Transactions
}

type NodeInfo struct {
	NodeState   uint   // node status
	NodePort    uint16 // The nodes's port
	ID          uint64 // The nodes's id
	NodeTime    int64
	NodeVersion uint32   // The network protocol the node used
	NodeType    uint64   // The services the node supplied
	Relay       bool     // The relay capability of the node (merge into capbility flag)
	Height      uint32   // The node latest block height
	TxnCnt      []uint32 // The transactions in pool
	//RxTxnCnt uint64 // The transaction received by this node
}

type ConsensusInfo struct {
	// TODO
}

type TXNAttrInfo struct {
	Height  uint32
	Type    int
	ErrCode int
}

type TXNEntryInfo struct {
	State []TXNAttrInfo // the result from each validator
}

func GetLogEvent(obj *event.LogEventArgs) (map[string]bool, LogEventArgs) {
	hash := obj.TxHash
	addr := obj.CcntmractAddress.ToHexString()
	ccntmractAddrs := map[string]bool{addr: true}
	return ccntmractAddrs, LogEventArgs{hash.ToHexString(), addr, obj.Message}
}

func GetExecuteNotify(obj *event.ExecuteNotify) (map[string]bool, ExecuteNotify) {
	evts := []NotifyEventInfo{}
	var ccntmractAddrs = make(map[string]bool)
	for _, v := range obj.Notify {
		evts = append(evts, NotifyEventInfo{v.CcntmractAddress.ToHexString(), v.States})
		ccntmractAddrs[v.CcntmractAddress.ToHexString()] = true
	}
	txhash := obj.TxHash.ToHexString()
	return ccntmractAddrs, ExecuteNotify{txhash, obj.State, obj.GasConsumed, evts}
}

func ConvertPreExecuteResult(obj *cstate.PreExecResult) PreExecuteResult {
	evts := []NotifyEventInfo{}
	for _, v := range obj.Notify {
		evts = append(evts, NotifyEventInfo{v.CcntmractAddress.ToHexString(), v.States})
	}
	return PreExecuteResult{obj.State, obj.Gas, obj.Result, evts}
}

func TransArryByteToHexString(ptx *types.Transaction) *Transactions {
	trans := new(Transactions)
	trans.TxType = ptx.TxType
	trans.Nonce = ptx.Nonce
	trans.GasLimit = ptx.GasLimit
	trans.GasPrice = ptx.GasPrice
	trans.Payer = ptx.Payer.ToBase58()
	trans.Payload = TransPayloadToHex(ptx.Payload)

	trans.Attributes = make([]TxAttributeInfo, 0)
	trans.Sigs = []Sig{}
	for _, sigdata := range ptx.Sigs {
		sig, _ := sigdata.GetSig()
		e := Sig{M: sig.M}
		for i := 0; i < len(sig.PubKeys); i++ {
			key := keypair.SerializePublicKey(sig.PubKeys[i])
			e.PubKeys = append(e.PubKeys, common.ToHexString(key))
		}
		for i := 0; i < len(sig.SigData); i++ {
			e.SigData = append(e.SigData, common.ToHexString(sig.SigData[i]))
		}
		trans.Sigs = append(trans.Sigs, e)
	}

	mhash := ptx.Hash()
	trans.Hash = mhash.ToHexString()
	return trans
}

func SendTxToPool(txn *types.Transaction) (cntmErrors.ErrCode, string) {
	if errCode, desc := bactor.AppendTxToPool(txn); errCode != cntmErrors.ErrNoError {
		log.Warn("TxnPool verify error:", errCode.Error())
		return errCode, desc
	}
	return cntmErrors.ErrNoError, ""
}

func GetBlockInfo(block *types.Block) BlockInfo {
	hash := block.Hash()
	var bookkeepers = []string{}
	var sigData = []string{}
	for i := 0; i < len(block.Header.SigData); i++ {
		s := common.ToHexString(block.Header.SigData[i])
		sigData = append(sigData, s)
	}
	for i := 0; i < len(block.Header.Bookkeepers); i++ {
		e := block.Header.Bookkeepers[i]
		key := keypair.SerializePublicKey(e)
		bookkeepers = append(bookkeepers, common.ToHexString(key))
	}

	blockHead := &BlockHead{
		Version:          block.Header.Version,
		PrevBlockHash:    block.Header.PrevBlockHash.ToHexString(),
		TransactionsRoot: block.Header.TransactionsRoot.ToHexString(),
		BlockRoot:        block.Header.BlockRoot.ToHexString(),
		Timestamp:        block.Header.Timestamp,
		Height:           block.Header.Height,
		ConsensusData:    block.Header.ConsensusData,
		ConsensusPayload: common.ToHexString(block.Header.ConsensusPayload),
		NextBookkeeper:   block.Header.NextBookkeeper.ToBase58(),
		Bookkeepers:      bookkeepers,
		SigData:          sigData,
		Hash:             hash.ToHexString(),
	}

	trans := make([]*Transactions, len(block.Transactions))
	for i := 0; i < len(block.Transactions); i++ {
		trans[i] = TransArryByteToHexString(block.Transactions[i])
	}

	b := BlockInfo{
		Hash:         hash.ToHexString(),
		Size:         len(block.ToArray()),
		Header:       blockHead,
		Transactions: trans,
	}
	return b
}

func GetBalance(address common.Address) (*BalanceOfRsp, error) {
	cntm, err := GetCcntmractBalance(0, utils.OntCcntmractAddress, address)
	if err != nil {
		return nil, fmt.Errorf("get cntm balance error:%s", err)
	}
	cntm, err := GetCcntmractBalance(0, utils.OngCcntmractAddress, address)
	if err != nil {
		return nil, fmt.Errorf("get cntm balance error:%s", err)
	}
	return &BalanceOfRsp{
		Ont: fmt.Sprintf("%d", cntm),
		Ong: fmt.Sprintf("%d", cntm),
	}, nil
}

func GetGrantOng(addr common.Address) (string, error) {
	key := append([]byte(cntm.UNBOUND_TIME_OFFSET), addr[:]...)
	value, err := ledger.DefLedger.GetStorageItem(utils.OntCcntmractAddress, key)
	if err != nil {
		value = []byte{0, 0, 0, 0}
	}
	v, err := serialization.ReadUint32(bytes.NewBuffer(value))
	if err != nil {
		return fmt.Sprintf("%v", 0), err
	}
	cntm, err := GetCcntmractBalance(0, utils.OntCcntmractAddress, addr)
	if err != nil {
		return fmt.Sprintf("%v", 0), err
	}
	boundcntm := utils.CalcUnbindOng(cntm, v, uint32(time.Now().Unix())-constants.GENESIS_BLOCK_TIMESTAMP)
	return fmt.Sprintf("%v", boundcntm), nil
}

func GetAllowance(asset string, from, to common.Address) (string, error) {
	var ccntmractAddr common.Address
	switch strings.ToLower(asset) {
	case "cntm":
		ccntmractAddr = utils.OntCcntmractAddress
	case "cntm":
		ccntmractAddr = utils.OngCcntmractAddress
	default:
		return "", fmt.Errorf("unsupport asset")
	}
	allowance, err := GetCcntmractAllowance(0, ccntmractAddr, from, to)
	if err != nil {
		return "", fmt.Errorf("get allowance error:%s", err)
	}
	return fmt.Sprintf("%v", allowance), nil
}

func GetCcntmractBalance(cVersion byte, ccntmractAddr, accAddr common.Address) (uint64, error) {
	mutable, err := NewNativeInvokeTransaction(0, 0, ccntmractAddr, cVersion, "balanceOf", []interface{}{accAddr[:]})
	if err != nil {
		return 0, fmt.Errorf("NewNativeInvokeTransaction error:%s", err)
	}
	tx, err := mutable.IntoImmutable()
	if err != nil {
		return 0, err
	}
	result, err := bactor.PreExecuteCcntmract(tx)
	if err != nil {
		return 0, fmt.Errorf("PrepareInvokeCcntmract error:%s", err)
	}
	if result.State == 0 {
		return 0, fmt.Errorf("prepare invoke failed")
	}
	data, err := hex.DecodeString(result.Result.(string))
	if err != nil {
		return 0, fmt.Errorf("hex.DecodeString error:%s", err)
	}

	balance := common.BigIntFromNeoBytes(data)
	return balance.Uint64(), nil
}

func GetCcntmractAllowance(cVersion byte, ccntmractAddr, fromAddr, toAddr common.Address) (uint64, error) {
	type allowanceStruct struct {
		From common.Address
		To   common.Address
	}
	mutable, err := NewNativeInvokeTransaction(0, 0, ccntmractAddr, cVersion, "allowance",
		[]interface{}{&allowanceStruct{
			From: fromAddr,
			To:   toAddr,
		}})
	if err != nil {
		return 0, fmt.Errorf("NewNativeInvokeTransaction error:%s", err)
	}

	tx, err := mutable.IntoImmutable()
	if err != nil {
		return 0, err
	}

	result, err := bactor.PreExecuteCcntmract(tx)
	if err != nil {
		return 0, fmt.Errorf("PrepareInvokeCcntmract error:%s", err)
	}
	if result.State == 0 {
		return 0, fmt.Errorf("prepare invoke failed")
	}
	data, err := hex.DecodeString(result.Result.(string))
	if err != nil {
		return 0, fmt.Errorf("hex.DecodeString error:%s", err)
	}
	allowance := common.BigIntFromNeoBytes(data)
	return allowance.Uint64(), nil
}

func GetGasPrice() (map[string]interface{}, error) {
	start := bactor.GetCurrentBlockHeight()
	var gasPrice uint64 = 0
	var height uint32 = 0
	var end uint32 = 0
	if start > MAX_SEARCH_HEIGHT {
		end = start - MAX_SEARCH_HEIGHT
	}
	for i := start; i >= end; i-- {
		head, err := bactor.GetHeaderByHeight(i)
		if err == nil && head.TransactionsRoot != common.UINT256_EMPTY {
			height = i
			blk, err := bactor.GetBlockByHeight(i)
			if err != nil {
				return nil, err
			}
			for _, v := range blk.Transactions {
				gasPrice += v.GasPrice
			}
			gasPrice = gasPrice / uint64(len(blk.Transactions))
			break
		}
	}
	result := map[string]interface{}{"gasprice": gasPrice, "height": height}
	return result, nil
}

func GetBlockTransactions(block *types.Block) interface{} {
	trans := make([]string, len(block.Transactions))
	for i := 0; i < len(block.Transactions); i++ {
		t := block.Transactions[i].Hash()
		trans[i] = t.ToHexString()
	}
	hash := block.Hash()
	type BlockTransactions struct {
		Hash         string
		Height       uint32
		Transactions []string
	}
	b := BlockTransactions{
		Hash:         hash.ToHexString(),
		Height:       block.Header.Height,
		Transactions: trans,
	}
	return b
}

//NewNativeInvokeTransaction return native ccntmract invoke transaction
func NewNativeInvokeTransaction(gasPirce, gasLimit uint64, ccntmractAddress common.Address, version byte,
	method string, params []interface{}) (*types.MutableTransaction, error) {
	invokeCode, err := cutils.BuildNativeInvokeCode(ccntmractAddress, version, method, params)
	if err != nil {
		return nil, err
	}
	return NewSmartCcntmractTransaction(gasPirce, gasLimit, invokeCode)
}

func NewNeovmInvokeTransaction(gasPrice, gasLimit uint64, ccntmractAddress common.Address, params []interface{}) (*types.MutableTransaction, error) {
	invokeCode, err := BuildNeoVMInvokeCode(ccntmractAddress, params)
	if err != nil {
		return nil, err
	}
	return NewSmartCcntmractTransaction(gasPrice, gasLimit, invokeCode)
}

func NewSmartCcntmractTransaction(gasPrice, gasLimit uint64, invokeCode []byte) (*types.MutableTransaction, error) {
	invokePayload := &payload.InvokeCode{
		Code: invokeCode,
	}
	tx := &types.MutableTransaction{
		GasPrice: gasPrice,
		GasLimit: gasLimit,
		TxType:   types.Invoke,
		Nonce:    uint32(time.Now().Unix()),
		Payload:  invokePayload,
		Sigs:     nil,
	}
	return tx, nil
}

//BuildNeoVMInvokeCode build NeoVM Invoke code for params
func BuildNeoVMInvokeCode(smartCcntmractAddress common.Address, params []interface{}) ([]byte, error) {
	builder := neovm.NewParamsBuilder(new(bytes.Buffer))
	err := cutils.BuildNeoVMParam(builder, params)
	if err != nil {
		return nil, err
	}
	args := append(builder.ToArray(), 0x67)
	args = append(args, smartCcntmractAddress[:]...)
	return args, nil
}

func GetAddress(str string) (common.Address, error) {
	var address common.Address
	var err error
	if len(str) == common.ADDR_LEN*2 {
		address, err = common.AddressFromHexString(str)
	} else {
		address, err = common.AddressFromBase58(str)
	}
	return address, err
}
