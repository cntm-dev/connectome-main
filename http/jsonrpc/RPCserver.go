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

package jsonrpc

import (
	. "github.com/Ontology/common/config"
	"github.com/Ontology/common/log"
	. "github.com/Ontology/http/base/rpc"
	"net/http"
	"strconv"
)




func StartRPCServer() {
	log.Debug()
	http.HandleFunc("/", Handle)

	HandleFunc("getbestblockhash", GetBestBlockHash)
	HandleFunc("getblock", GetBlock)
	HandleFunc("getblockcount", GetBlockCount)
	HandleFunc("getblockhash", GetBlockHash)
	HandleFunc("getconnectioncount", GetConnectionCount)
	HandleFunc("getrawmempool", GetRawMemPool)

	HandleFunc("getrawtransaction", GetRawTransaction)
	HandleFunc("sendrawtransaction", SendRawTransaction)
	HandleFunc("getstorage", GetStorage)
	HandleFunc("getversion", GetNodeVersion)

	HandleFunc("getblocksysfee", GetSystemFee)
	HandleFunc("getccntmractstate", GetCcntmractState)
	HandleFunc("getmempooltxstate", GetMemPoolTxState)
	HandleFunc("getsmartcodeevent", GetSmartCodeEvent)
	HandleFunc("getblockheightbytxhash", GetBlockHeightByTxHash)

	HandleFunc("getbalance", GetBalance)

	err := http.ListenAndServe(":"+strconv.Itoa(Parameters.HttpJsonPort), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err.Error())
	}
}
