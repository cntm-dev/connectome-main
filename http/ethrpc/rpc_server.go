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
package ethrpc

import (
	"net/http"
	"strconv"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
	cfg "github.com/cntmio/cntmology/common/config"
	"github.com/cntmio/cntmology/http/ethrpc/eth"
	"github.com/cntmio/cntmology/http/ethrpc/net"
	"github.com/cntmio/cntmology/http/ethrpc/utils"
	"github.com/cntmio/cntmology/http/ethrpc/web3"
	tp "github.com/cntmio/cntmology/txnpool/proc"
)

func StartEthServer(txpool *tp.TXPoolServer) error {
	log.Root().SetHandler(utils.OntLogHandler())
	ethAPI := eth.NewEthereumAPI(txpool)
	server := rpc.NewServer()
	err := server.RegisterName("eth", ethAPI)
	if err != nil {
		return err
	}
	netRpcService := net.NewPublicNetAPI()
	err = server.RegisterName("net", netRpcService)
	if err != nil {
		return err
	}
	web3API := web3.NewAPI()
	err = server.RegisterName("web3", web3API)
	if err != nil {
		return err
	}
	err = http.ListenAndServe(":"+strconv.Itoa(int(cfg.DefConfig.Rpc.EthJsonPort)), server)
	if err != nil {
		return err
	}
	return nil
}