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
	"bytes"
	"encoding/hex"
	"encoding/json"
	clisvrcom "github.com/cntmio/cntmology/cmd/sigsvr/common"
	cliutil "github.com/cntmio/cntmology/cmd/utils"
	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/log"
	httpcom "github.com/cntmio/cntmology/http/base/common"
)

type SigNeoVMInvokeTxAbiReq struct {
	GasPrice    uint64          `json:"gas_price"`
	GasLimit    uint64          `json:"gas_limit"`
	Address     string          `json:"address"`
	Method      string          `json:"method"`
	Params      []string        `json:"params"`
	CcntmractAbi json.RawMessage `json:"ccntmract_abi"`
}

type SigNeoVMInvokeTxAbiRsp struct {
	SignedTx string `json:"signed_tx"`
}

func SigNeoVMInvokeAbiTx(req *clisvrcom.CliRpcRequest, resp *clisvrcom.CliRpcResponse) {
	rawReq := &SigNeoVMInvokeTxAbiReq{}
	err := json.Unmarshal(req.Params, rawReq)
	if err != nil {
		log.Infof("SigNeoVMInvokeAbiTx json.Unmarshal SigNeoVMInvokeTxAbiReq:%s error:%s", req.Params, err)
		resp.ErrorCode = clisvrcom.CLIERR_INVALID_PARAMS
		return
	}
	ccntmractAbi, err := cliutil.NewNeovmCcntmractAbi(rawReq.CcntmractAbi)
	if err != nil {
		resp.ErrorCode = clisvrcom.CLIERR_ABI_UNMATCH
		resp.ErrorInfo = err.Error()
		return
	}
	funcAbi := ccntmractAbi.GetFunc(rawReq.Method)
	if funcAbi == nil {
		resp.ErrorCode = clisvrcom.CLIERR_ABI_NOT_FOUND
		return
	}
	invokParams, err := cliutil.ParseNeovmFunc(rawReq.Params, funcAbi)
	if err != nil {
		resp.ErrorCode = clisvrcom.CLIERR_ABI_UNMATCH
		resp.ErrorInfo = err.Error()
		return
	}
	ccntmAddr, err := common.AddressFromHexString(rawReq.Address)
	if err != nil {
		log.Infof("Cli Qid:%s SigNeoVMInvokeAbiTx AddressParseFromBytes:%s error:%s", req.Qid, rawReq.Address, err)
		resp.ErrorCode = clisvrcom.CLIERR_INVALID_PARAMS
		return
	}
	tx, err := httpcom.NewNeovmInvokeTransaction(rawReq.GasPrice, rawReq.GasLimit, ccntmAddr, invokParams)
	if err != nil {
		log.Infof("Cli Qid:%s SigNeoVMInvokeAbiTx InvokeNeoVMCcntmractTx error:%s", req.Qid, err)
		resp.ErrorCode = clisvrcom.CLIERR_INVALID_PARAMS
		return
	}
	signer := clisvrcom.DefAccount
	err = cliutil.SignTransaction(signer, tx)
	if err != nil {
		log.Infof("Cli Qid:%s SigNeoVMInvokeAbiTx SignTransaction error:%s", req.Qid, err)
		resp.ErrorCode = clisvrcom.CLIERR_INTERNAL_ERR
		return
	}
	buf := bytes.NewBuffer(nil)
	err = tx.Serialize(buf)
	if err != nil {
		log.Infof("Cli Qid:%s SigNeoVMInvokeAbiTx tx Serialize error:%s", req.Qid, err)
		resp.ErrorCode = clisvrcom.CLIERR_INTERNAL_ERR
		return
	}
	resp.Result = &SigNeoVMInvokeTxAbiRsp{
		SignedTx: hex.EncodeToString(buf.Bytes()),
	}
}