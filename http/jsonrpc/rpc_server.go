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

// Package jsonrpc privides a function to start json rpc server
package jsonrpc

import (
	"net/http"
	"strconv"

	"fmt"
	cfg "github.com/cntmio/cntmology/common/config"
	"github.com/cntmio/cntmology/common/log"
	"github.com/cntmio/cntmology/http/base/rpc"
)

func StartRPCServer() error {
	log.Debug()
	http.HandleFunc("/", rpc.Handle)

	rpc.HandleFunc("getbestblockhash", rpc.GetBestBlockHash)
	rpc.HandleFunc("getblock", rpc.GetBlock)
	rpc.HandleFunc("getblockcount", rpc.GetBlockCount)
	rpc.HandleFunc("getblockhash", rpc.GetBlockHash)
	rpc.HandleFunc("getconnectioncount", rpc.GetConnectionCount)
	//HandleFunc("getrawmempool", GetRawMemPool)

	rpc.HandleFunc("getrawtransaction", rpc.GetRawTransaction)
	rpc.HandleFunc("sendrawtransaction", rpc.SendRawTransaction)
	rpc.HandleFunc("getstorage", rpc.GetStorage)
	rpc.HandleFunc("getversion", rpc.GetNodeVersion)
	rpc.HandleFunc("getnetworkid", rpc.GetNetworkId)

	rpc.HandleFunc("getccntmractstate", rpc.GetCcntmractState)
	rpc.HandleFunc("getmempooltxcount", rpc.GetMemPoolTxCount)
	rpc.HandleFunc("getmempooltxstate", rpc.GetMemPoolTxState)
	rpc.HandleFunc("getsmartcodeevent", rpc.GetSmartCodeEvent)
	rpc.HandleFunc("getblockheightbytxhash", rpc.GetBlockHeightByTxHash)

	rpc.HandleFunc("getbalance", rpc.GetBalance)
	rpc.HandleFunc("getallowance", rpc.GetAllowance)
	rpc.HandleFunc("getmerkleproof", rpc.GetMerkleProof)
	rpc.HandleFunc("getblocktxsbyheight", rpc.GetBlockTxsByHeight)
	rpc.HandleFunc("getgasprice", rpc.GetGasPrice)
	rpc.HandleFunc("getunboundcntm", rpc.GetUnboundOng)

	err := http.ListenAndServe(":"+strconv.Itoa(int(cfg.DefConfig.Rpc.HttpJsonPort)), nil)
	if err != nil {
		return fmt.Errorf("ListenAndServe error:%s", err)
	}
	return nil
}
