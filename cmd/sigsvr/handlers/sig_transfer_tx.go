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

package handlers

import (
	"encoding/hex"
	"encoding/json"
	clisvrcom "github.com/cntmio/cntmology/cmd/sigsvr/common"
	cliutil "github.com/cntmio/cntmology/cmd/utils"
	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/log"
	"strconv"
)

type SigTransferTransactionReq struct {
	GasPrice uint64 `json:"gas_price"`
	GasLimit uint64 `json:"gas_limit"`
	Asset    string `json:"asset"`
	From     string `json:"from"`
	To       string `json:"to"`
	Amount   string `json:"amount"`
	Payer    string `json:"payer"`
}

type SinTransferTransactionRsp struct {
	SignedTx string `json:"signed_tx"`
}

func SigTransferTransaction(req *clisvrcom.CliRpcRequest, resp *clisvrcom.CliRpcResponse) {
	rawReq := &SigTransferTransactionReq{}
	err := json.Unmarshal(req.Params, rawReq)
	if err != nil {
		resp.ErrorCode = clisvrcom.CLIERR_INVALID_PARAMS
		return
	}
	amount, err := strconv.ParseInt(rawReq.Amount, 10, 64)
	if err != nil {
		resp.ErrorCode = clisvrcom.CLIERR_INVALID_PARAMS
		resp.ErrorInfo = "amount should be string type"
		return
	}
	if amount < 0 {
		resp.ErrorCode = clisvrcom.CLIERR_INVALID_PARAMS
		return
	}
	mutable, err := cliutil.TransferTx(rawReq.GasPrice, rawReq.GasLimit, rawReq.Asset, rawReq.From, rawReq.To, uint64(amount))
	if err != nil {
		log.Infof("Cli Qid:%s SigTransferTransaction TransferTx error:%s", req.Qid, err)
		resp.ErrorCode = clisvrcom.CLIERR_INVALID_PARAMS
		return
	}
	if rawReq.Payer != "" {
		payerAddress, err := common.AddressFromBase58(rawReq.Payer)
		if err != nil {
			log.Infof("Cli Qid:%s SigTransferTransaction AddressFromBase58 error:%s", req.Qid, err)
			resp.ErrorCode = clisvrcom.CLIERR_INVALID_PARAMS
			return
		}
		mutable.Payer = payerAddress
	}

	signer, err := req.GetAccount()
	if err != nil {
		log.Infof("Cli Qid:%s SigTransferTransaction GetAccount:%s", req.Qid, err)
		resp.ErrorCode = clisvrcom.CLIERR_ACCOUNT_UNLOCK
		return
	}
	if signer == nil {
		resp.ErrorCode = clisvrcom.CLIERR_ACCOUNT_UNLOCK
		return
	}
	err = cliutil.SignTransaction(signer, mutable)
	if err != nil {
		log.Infof("Cli Qid:%s SigTransferTransaction SignTransaction error:%s", req.Qid, err)
		resp.ErrorCode = clisvrcom.CLIERR_INTERNAL_ERR
		return
	}
	tx, err := mutable.IntoImmutable()
	if err != nil {
		log.Infof("Cli Qid:%s SigTransferTransaction tx IntoInmmutable error:%s", req.Qid, err)
		resp.ErrorCode = clisvrcom.CLIERR_INTERNAL_ERR
		return
	}
	sink := common.ZeroCopySink{}
	tx.Serialization(&sink)
	resp.Result = &SinTransferTransactionRsp{
		SignedTx: hex.EncodeToString(sink.Bytes()),
	}
}
