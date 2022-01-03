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

package rpc

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/Ontology/common"
	"github.com/Ontology/common/config"
	"github.com/Ontology/common/log"
	"github.com/Ontology/core/genesis"
	"github.com/Ontology/core/payload"
	"github.com/Ontology/core/types"
	cntmErrors "github.com/Ontology/errors"
	bactor "github.com/Ontology/http/base/actor"
	bcomn "github.com/Ontology/http/base/common"
	berr "github.com/Ontology/http/base/error"
)

func GetGenerateBlockTime(params []interface{}) map[string]interface{} {
	return responseSuccess(config.DEFAULT_GEN_BLOCK_TIME)
}

func GetBestBlockHash(params []interface{}) map[string]interface{} {
	hash, err := bactor.CurrentBlockHash()
	if err != nil {
		log.Errorf("GetBestBlockHash error:%s", err)
		return responsePack(berr.INTERNAL_ERROR, false)
	}
	return responseSuccess(common.ToHexString(hash.ToArray()))
}

// Input JSON string examples for getblock method as following:
//   {"jsonrpc": "2.0", "method": "getblock", "params": [1], "id": 0}
//   {"jsonrpc": "2.0", "method": "getblock", "params": ["aabbcc.."], "id": 0}
func GetBlock(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return responsePacking(Err.INVALID_PARAMS, nil)
	}
	var err error
	var hash Uint256
	switch (params[0]).(type) {
	// block height
	case float64:
		index := uint32(params[0].(float64))
		hash, err = GetBlockHashFromStore(index)
		if err != nil {
			return DnaRpcUnknownBlock
		}
		if hash.CompareTo(Uint256{}) == 0{
			return DnaRpcInvalidParameter
		}
		// block hash
	case string:
		str := params[0].(string)
		hex, err := hex.DecodeString(str)
		if err != nil {
			return DnaRpcInvalidParameter
		}
		if err := hash.Deserialize(bytes.NewReader(hex)); err != nil {
			return DnaRpcInvalidTransaction
		}
	default:
		return DnaRpcInvalidParameter
	}

	block, err := GetBlockFromStore(hash)
	if err != nil {
		return DnaRpcUnknownBlock
	}

	blockHead := &BlockHead{
		Version:          block.Header.Version,
		PrevBlockHash:    ToHexString(block.Header.PrevBlockHash.ToArray()),
		TransactionsRoot: ToHexString(block.Header.TransactionsRoot.ToArray()),
		BlockRoot:        ToHexString(block.Header.BlockRoot.ToArray()),
		StateRoot:        ToHexString(block.Header.StateRoot.ToArray()),
		Timestamp:        block.Header.Timestamp,
		Height:           block.Header.Height,
		ConsensusData:    block.Header.ConsensusData,
		NextBookKeeper:   ToHexString(block.Header.NextBookKeeper[:]),
		// TODO replace with bookkeepers and sigdata
		//Program: ProgramInfo{
		//	Code:      ToHexString(block.Header.Program.Code),
		//	Parameter: ToHexString(block.Header.Program.Parameter),
		//},
		Hash: ToHexString(hash.ToArray()),
	}

	trans := make([]*Transactions, len(block.Transactions))
	for i := 0; i < len(block.Transactions); i++ {
		trans[i] = TransArryByteToHexString(block.Transactions[i])
	}

	b := BlockInfo{
		Hash:         ToHexString(hash.ToArray()),
		BlockData:    blockHead,
		Transactions: trans,
	}
	return responsePacking(b, id)
}

func getBlockCount(cmd map[string]interface{}) map[string]interface{} {
	id := cmd["id"]
	count := ledger.DefaultLedger.Blockchain.BlockHeight + 1
	return responsePacking(count, id)
}

func getBlockHash(cmd map[string]interface{}) map[string]interface{} {
	id := cmd["id"]
	index := cmd["params"]
	var hash Uint256
	height, ok := index.(uint32)
	if ok == true {
		hash, _ = ledger.DefaultLedger.Store.GetBlockHash(height)
	}
	hashhex := fmt.Sprintf("%016x", hash)
	return responsePacking(hashhex, id)
}

func getTxn(cmd map[string]interface{}) map[string]interface{} {
	id := cmd["id"]
	params := cmd["params"]
	var hash Uint256

	txid := params.([]interface{})[0].(string)
	hashslice, _ := hex.DecodeString(txid)
	hash.Deserialize(bytes.NewReader(hashslice[0:32]))

	tx, err := ledger.DefaultLedger.Store.GetTransaction(hash)
	if err != nil {
		return responsePacking([]interface{}{-100, "Unknown block"}, id)
	}

	tran := TransArryByteToHexString(tx)
	return responsePacking(tran, id)
}

func getAddrTxn(req *http.Request, cnd map[string]interface{}) map[string]interface{} {
	return nil
}

func getConnectionCount(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	id := cmd["id"]
	count := node.GetConnectionCnt()
	return responsePacking(count, id)
}

func getRawMemPool(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	id := cmd["id"]
	mempoollist := node.GetTxnPool()
	return responsePacking(mempoollist, id)
}

func getRawTransaction(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	id := cmd["id"]
	params := cmd["params"]
	txid := params.([]interface{})[0].(string)
	txidSlice, _ := hex.DecodeString(txid)
	var txidArr Uint256
	txidArr.Deserialize(bytes.NewReader(txidSlice[0:32]))
	verbose := params.([]interface{})[1].(bool)
	tx := node.GetTransaction(txidArr)
	txBuffer := bytes.NewBuffer([]byte{})
	tx.Serialize(txBuffer)
	if verbose == true {
		t := TxInfo{
			Hash: txid,
			Hex:  hex.EncodeToString(txBuffer.Bytes()),
			Tx:   tx,
		}
		response := responsePacking(t, id)
		return response
	}

	return responsePacking(txBuffer.Bytes(), id)
}

func getTxout(cmd map[string]interface{}) map[string]interface{} {
	id := cmd["id"]
	//params := cmd["params"]
	//txid := params.([]interface{})[0].(string)
	//var n int = params.([]interface{})[1].(int)
	var txout tx.TxOutput // := tx.GetTxOut() //TODO
	high := uint32(txout.Value >> 32)
	low := uint32(txout.Value)
	to := TxoutInfo{
		High:  high,
		Low:   low,
		Txout: txout,
	}
	return responsePacking(to, id)
}

func sendRawTransaction(cmd map[string]interface{}) map[string]interface{} {
	id := cmd["id"]
	params := cmd["params"]
	hexValue := params.([]interface{})[0].(string)

	hexSlice, err := hex.DecodeString(hexValue)
	if err != nil {
		log.Error("Decode raw transaction error")
		return responsePacking(false, id)
	}
	var txTransaction tx.Transaction
	if err := txTransaction.Deserialize(bytes.NewReader(hexSlice[:])); err != nil {
		log.Error("Deserialize raw transaction error")
		return responsePacking(false, id)
	}
	if err := SendTx(&txTransaction); err != nil {
		return responsePacking(false, id)
	}
	return responsePacking(true, id)
}

func submitBlock(cmd map[string]interface{}) map[string]interface{} {
	id := cmd["id"]
	hexValue := cmd["params"].(string)
	hexSlice, _ := hex.DecodeString(hexValue)
	var txTransaction tx.Transaction
	txTransaction.Deserialize(bytes.NewReader(hexSlice[:]))
	err := node.Xmit(&txTransaction)
	response := responsePacking(err, id)
	return response
}

func getNeighbor(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	id := cmd["id"]
	addr, _ := node.GetNeighborAddrs()
	return responsePacking(addr, id)
}

func getNodeState(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	id := cmd["id"]
	n := NodeInfo{
		State:    node.GetState(),
		Time:     node.GetTime(),
		Port:     node.GetPort(),
		ID:       node.GetID(),
		Version:  node.Version(),
		Services: node.Services(),
		Relay:    node.GetRelay(),
		Height:   node.GetHeight(),
		TxnCnt:   node.GetTxnCnt(),
		RxTxnCnt: node.GetRxTxnCnt(),
	}
	return DnaRpc(n)
}

func startConsensus(params []interface{}) map[string]interface{} {
	if err := consensusSrv.Start(); err != nil {
		return DnaRpcFailed
	}
	return DnaRpcSuccess
}

func stopConsensus(params []interface{}) map[string]interface{} {
	if err := consensusSrv.Halt(); err != nil {
		return DnaRpcFailed
	}
	return DnaRpcSuccess
}

func sendSampleTransaction(params []interface{}) map[string]interface{} {
	panic("need reimplementation")
	return nil

	/*
		if len(params) < 1 {
			return DnaRpcNil
		}
		var txType string
		switch params[0].(type) {
		case string:
			txType = params[0].(string)
		default:
			return DnaRpcInvalidParameter
		}

		issuer, err := account.NewAccount()
		if err != nil {
			return DnaRpc("Failed to create account")
		}
		admin := issuer

		rbuf := make([]byte, RANDBYTELEN)
		rand.Read(rbuf)
		switch string(txType) {
		case "perf":
			num := 1
			if len(params) == 2 {
				switch params[1].(type) {
				case float64:
					num = int(params[1].(float64))
				}
			}
			for i := 0; i < num; i++ {
				regTx := NewRegTx(ToHexString(rbuf), i, admin, issuer)
				SignTx(admin, regTx)
				VerifyAndSendTx(regTx)
			}
			return DnaRpc(fmt.Sprintf("%d transaction(s) was sent", num))
		default:
			return DnaRpc("Invalid transacion type")
		}
	*/
}

func setDebugInfo(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return DnaRpcInvalidParameter
	}
	switch params[0].(type) {
	case float64:
		level := params[0].(float64)
		if err := log.Log.SetDebugLevel(int(level)); err != nil {
			return DnaRpcInvalidParameter
		}
	default:
		return DnaRpcInvalidParameter
	}
	return DnaRpcSuccess
}

func getVersion(params []interface{}) map[string]interface{} {
	return DnaRpc(config.Version)
}

func uploadDataFile(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return DnaRpcNil
	}

	rbuf := make([]byte, 4)
	rand.Read(rbuf)
	tmpname := hex.EncodeToString(rbuf)

	str := params[0].(string)

	data, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return RpcInvalidParameter
	}
	f, err := os.OpenFile(tmpname, os.O_WRONLY|os.O_CREATE, 0664)
	if err != nil {
		return RpcIOError
	}
	defer f.Close()
	f.Write(data)

	refpath, err := AddFileIPFS(tmpname, true)
	if err != nil {
		return RpcAPIError
	}

	return Rpc(refpath)

}
func GetSmartCodeEvent(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return RpcNil
	}

	switch (params[0]).(type) {
	// block height
	case float64:
		height := uint32(params[0].(float64))
		//TODO resp
		return Rpc(map[string]interface{}{"Height": height})
	default:
		return RpcInvalidParameter
	}
	return RpcInvalidParameter
}

func GetBlockHeightByTxHash(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return RpcNil
	}

	switch (params[0]).(type) {
	// tx hash
	case string:
		str := params[0].(string)
		hex, err := hex.DecodeString(str)
		if err != nil {
			return responsePack(berr.INVALID_PARAMS, "")
		}
		var hash common.Uint256
		if err := hash.Deserialize(bytes.NewReader(hex)); err != nil {
			return responsePack(berr.INVALID_PARAMS, "")
		}
		height, err := bactor.GetBlockHeightByTxHashFromStore(hash)
		if err != nil {
			return responsePack(berr.INVALID_PARAMS, "")
		}
		return responseSuccess(height)
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}
	return responsePack(berr.INVALID_PARAMS, "")
}

func GetBalance(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return responsePack(berr.INVALID_PARAMS, "")
	}
	addrBase58, ok := params[0].(string)
	if !ok {
		return responsePack(berr.INVALID_PARAMS, "")
	}
	address, err := common.AddressFromBase58(addrBase58)
	if err != nil {
		return responsePack(berr.INVALID_PARAMS, "")
	}
	cntm := new(big.Int)
	cntm := new(big.Int)

	cntmBalance, err := bactor.GetStorageItem(genesis.OntCcntmractAddress, address[:])
	if err != nil {
		log.Errorf("GetOntBalanceOf GetStorageItem cntm address:%s error:%s", addrBase58, err)
		return responsePack(berr.INTERNAL_ERROR, "internal error")
	}
	if cntmBalance != nil {
		cntm.SetBytes(cntmBalance)
	}
	rsp := &bcomn.BalanceOfRsp{
		Ont: cntm.String(),
		Ong: cntm.String(),
	}
	return responseSuccess(rsp)
}

func RegDataFile(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return responsePacking(Err.INVALID_PARAMS, nil)
	}
	var hash Uint256
	switch params[0].(type) {
	case string:
		str := params[0].(string)
		hex, err := hex.DecodeString(str)
		if err != nil {
			return DnaRpcInvalidParameter
		}
		var txn types.Transaction
		if err := txn.Deserialize(bytes.NewReader(hex)); err != nil {
			return DnaRpcInvalidTransaction
		}

		hash = txn.Hash()
		if errCode := VerifyAndSendTx(&txn); errCode != ErrNoError {
			return DnaRpcInternalError
		}
	default:
		return DnaRpcInvalidParameter
	}
	return DnaRpc(ToHexString(hash.ToArray()))
}

func CatDataRecord(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return DnaRpcNil
	}
	switch params[0].(type) {
	case string:
		str := params[0].(string)
		b, err := hex.DecodeString(str)
		if err != nil {
			return DnaRpcInvalidParameter
		}
		var hash Uint256
		err = hash.Deserialize(bytes.NewReader(b))
		if err != nil {
			return DnaRpcInvalidTransaction
		}
		tx, err := GetTransaction(hash) //ledger.DefaultLedger.Store.GetTransaction(hash)
		if err != nil {
			return DnaRpcUnknownTransaction
		}
		tran := TransArryByteToHexString(tx)
		info := tran.Payload.(*DataFileInfo)
		//ref := string(record.RecordData[:])
		return DnaRpc(info)
	default:
		return DnaRpcInvalidParameter
	}
}

func GetDataFile(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return DnaRpcNil
	}
	switch params[0].(type) {
	case string:
		str := params[0].(string)
		hex, err := hex.DecodeString(str)
		if err != nil {
			return DnaRpcInvalidParameter
		}
		var hash Uint256
		err = hash.Deserialize(bytes.NewReader(hex))
		if err != nil {
			return DnaRpcInvalidTransaction
		}
		tx, err := GetTransaction(hash)
		if err != nil {
			return DnaRpcUnknownTransaction
		}

		tran := TransArryByteToHexString(tx)
		info := tran.Payload.(*DataFileInfo)

		err = GetFileIPFS(info.IPFSPath, info.Filename)
		if err != nil {
			return RpcAPIError
		}
		//TODO: shoud return download address
		return RpcSuccess
	default:
		return RpcInvalidParameter
	}
}
