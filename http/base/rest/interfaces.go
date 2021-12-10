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

package rest

import (
	"bytes"
	"github.com/Ontology/common"
	"github.com/Ontology/common/config"
	"github.com/Ontology/core/types"
	cntmerr "github.com/Ontology/errors"
	berr "github.com/Ontology/http/base/error"
	"strconv"
	"github.com/Ontology/core/payload"
	"github.com/Ontology/common/log"
	bcomn "github.com/Ontology/http/base/common"
	bactor "github.com/Ontology/http/base/actor"
	"math/big"
	"github.com/Ontology/core/genesis"
)

const TlsPort int = 443

type ApiServer interface {
	Start() error
	Stop()
}

//Node
func GetGenerateBlockTime(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	resp["Result"] = config.DEFAULTGENBLOCKTIME
	return resp
}
func GetConnectionCount(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	count, err := bactor.GetConnectionCnt()
	if err != nil {
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	resp["Result"] = count
	return resp
}

//Block
func GetBlockHeight(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	height,err := bactor.BlockHeight()
	if err != nil{
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	resp["Result"] = height
	return resp
}
func GetBlockHash(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	param := cmd["Height"].(string)
	if len(param) == 0 {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	height, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	hash, err := bactor.GetBlockHashFromStore(uint32(height))
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	resp["Result"] = common.ToHexString(hash.ToArray())
	return resp
}

func GetBlockTransactions(block *types.Block) interface{} {
	trans := make([]string, len(block.Transactions))
	for i := 0; i < len(block.Transactions); i++ {
		h := block.Transactions[i].Hash()
		trans[i] = common.ToHexString(h.ToArray())
	}
	hash := block.Hash()
	type BlockTransactions struct {
		Hash         string
		Height       uint32
		Transactions []string
	}
	b := BlockTransactions{
		Hash:         common.ToHexString(hash.ToArray()),
		Height:       block.Header.Height,
		Transactions: trans,
	}
	return b
}
func getBlock(hash common.Uint256, getTxBytes bool) (interface{}, int64) {
	block, err := bactor.GetBlockFromStore(hash)
	if err != nil {
		return nil, berr.UNKNOWN_BLOCK
	}
	if block.Header == nil {
		return nil, berr.UNKNOWN_BLOCK
	}
	if getTxBytes {
		w := bytes.NewBuffer(nil)
		block.Serialize(w)
		return common.ToHexString(w.Bytes()), berr.SUCCESS
	}
	return bcomn.GetBlockInfo(block), berr.SUCCESS
}
func GetBlockByHash(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	param := cmd["Hash"].(string)
	if len(param) == 0 {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	var getTxBytes = false
	if raw, ok := cmd["Raw"].(string); ok && raw == "1" {
		getTxBytes = true
	}
	var hash common.Uint256
	hex, err := common.HexToBytes(param)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	if err := hash.Deserialize(bytes.NewReader(hex)); err != nil {
		return ResponsePack(berr.INVALID_TRANSACTION)
	}
	resp["Result"], resp["Error"] = getBlock(hash, getTxBytes)
	return resp
}

func GetBlockHeightByTxHash(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	param := cmd["Hash"].(string)
	if len(param) == 0 {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	var hash common.Uint256
	hex, err := common.HexToBytes(param)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	if err := hash.Deserialize(bytes.NewReader(hex)); err != nil {
		return ResponsePack(berr.INVALID_TRANSACTION)
	}
	height,err := bactor.GetBlockHeightByTxHashFromStore(hash)
	if err != nil {
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	resp["Result"] = height
	return resp
}
func GetBlockTxsByHeight(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)

	param := cmd["Height"].(string)
	if len(param) == 0 {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	height, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	index := uint32(height)
	hash, err := bactor.GetBlockHashFromStore(index)
	if err != nil {
		return ResponsePack(berr.UNKNOWN_BLOCK)
	}
	if hash.CompareTo(common.Uint256{}) == 0{
		return ResponsePack(berr.INVALID_PARAMS)
	}
	block, err := bactor.GetBlockFromStore(hash)
	if err != nil {
		return ResponsePack(berr.UNKNOWN_BLOCK)
	}
	resp["Result"] = GetBlockTransactions(block)
	return resp
}
func GetBlockByHeight(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)

	param := cmd["Height"].(string)
	if len(param) == 0 {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	var getTxBytes bool = false
	if raw, ok := cmd["Raw"].(string); ok && raw == "1" {
		getTxBytes = true
	}
	height, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	index := uint32(height)
	hash, err := bactor.GetBlockHashFromStore(index)
	if err != nil {
		return ResponsePack(berr.UNKNOWN_BLOCK)
	}
	resp["Result"], resp["Error"] = getBlock(hash, getTxBytes)
	return resp
}


//Transaction
func GetTransactionByHash(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)

	str := cmd["Hash"].(string)
	bys, err := common.HexToBytes(str)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	var hash common.Uint256
	err = hash.Deserialize(bytes.NewReader(bys))
	if err != nil {
		return ResponsePack(berr.INVALID_TRANSACTION)
	}
	tx, err := bactor.GetTransaction(hash)
	if err != nil {
		return ResponsePack(berr.UNKNOWN_TRANSACTION)
	}
	if tx == nil {
		return ResponsePack(berr.UNKNOWN_TRANSACTION)
	}
	if raw, ok := cmd["Raw"].(string); ok && raw == "1" {
		w := bytes.NewBuffer(nil)
		tx.Serialize(w)
		resp["Result"] = common.ToHexString(w.Bytes())
		return resp
	}
	tran := bcomn.TransArryByteToHexString(tx)
	resp["Result"] = tran
	return resp
}
func SendRawTransaction(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)

	str, ok := cmd["Data"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	bys, err := common.HexToBytes(str)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	var txn types.Transaction
	if err := txn.Deserialize(bytes.NewReader(bys)); err != nil {
		return ResponsePack(berr.INVALID_TRANSACTION)
	}
	if txn.TxType == types.Invoke {
		if preExec, ok := cmd["PreExec"].(string); ok && preExec == "1" {
			log.Tracef("PreExec SMARTCODE")
			if _, ok := txn.Payload.(*payload.InvokeCode); ok {
				resp["Result"], err = bactor.PreExecuteCcntmract(&txn)
				if err != nil {
					log.Error(err)
					return ResponsePack(berr.SMARTCODE_ERROR)
				}
				return resp
			}
		}
	}
	var hash common.Uint256
	hash = txn.Hash()
	if errCode := bcomn.VerifyAndSendTx(&txn); errCode != cntmerr.ErrNoError {
		resp["Error"] = int64(errCode)
		return resp
	}
	resp["Result"] = common.ToHexString(hash.ToArray())

	if txn.TxType == types.Invoke {
		if userid, ok := cmd["Userid"].(string); ok && len(userid) > 0 {
			resp["Userid"] = userid
		}
	}
	return resp
}

func GetSmartCodeEventByHeight(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)

	param := cmd["Height"].(string)
	if len(param) == 0 {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	height, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	index := uint32(height)
	txs, err := bactor.GetEventNotifyByHeight(index)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	var txhexs []string
	for _, v := range txs {
		txhexs = append(txhexs, common.ToHexString(v.ToArray()))
	}
	resp["Result"] = txhexs
	return resp
}

func GetSmartCodeEventByTxHash(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)

	str := cmd["Hash"].(string)
	bys, err := common.HexToBytes(str)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	var hash common.Uint256
	err = hash.Deserialize(bytes.NewReader(bys))
	if err != nil {
		return ResponsePack(berr.INVALID_TRANSACTION)
	}
	eventInfos, err := bactor.GetEventNotifyByTxHash(hash)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	var evs []map[string]interface{}
	for _, v := range eventInfos {
		evs = append(evs, map[string]interface{}{"CodeHash": v.CodeHash,
			"States": v.States,
			"Ccntmainer": v.Ccntmainer})
	}
	resp["Result"] = evs
	return resp
}

func GetCcntmractState(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	str := cmd["Hash"].(string)
	bys, err := common.HexToBytes(str)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	var hash common.Address
	err = hash.Deserialize(bytes.NewReader(bys))
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	ccntmract, err := bactor.GetCcntmractStateFromStore(hash)
	if err != nil || ccntmract == nil {
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	resp["Result"] = bcomn.TransPayloadToHex(ccntmract)
	return resp
}

func GetStorage(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	str := cmd["Hash"].(string)
	bys, err := common.HexToBytes(str)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	var hash common.Address
	err = hash.Deserialize(bytes.NewReader(bys))
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	key := cmd["Key"].(string)
	item, err := common.HexToBytes(key)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	log.Info("[GetStorage] ",str,key)
	value, err := bactor.GetStorageItem(hash,item)
	if err != nil || value == nil {
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	resp["Result"] = common.ToHexString(value)
	return resp
}
func GetBalance(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	addrBase58 := cmd["Addr"].(string)
	address, err := common.AddressFromBase58(addrBase58)
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	cntm := new(big.Int)
	cntm := new(big.Int)

	cntmBalance, err := bactor.GetStorageItem(genesis.OntCcntmractAddress, address[:])
	if err != nil {
		log.Errorf("GetOntBalanceOf GetStorageItem cntm address:%s error:%s", address, err)
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	if cntmBalance != nil {
		cntm.SetBytes(cntmBalance)
	}
	rsp := &bcomn.BalanceOfRsp{
		Ont: cntm.String(),
		Ong: cntm.String(),
	}
	resp["Result"] = rsp
	return resp
}