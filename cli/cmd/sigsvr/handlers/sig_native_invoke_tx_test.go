/*
 * Copyright (C) 2018 The cntm Authors
 * This file is part of The cntm library.
 *
 * The cntm is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The cntm is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The cntm.  If not, see <http://www.gnu.org/licenses/>.
 */

package handlers

import (
	"encoding/json"
	"testing"

	"github.com/conntectome/cntm-crypto/keypair"
	"github.com/conntectome/cntm-crypto/signature"
	"github.com/conntectome/cntm/cmd/abi"
	clisvrcom "github.com/conntectome/cntm/cmd/sigsvr/common"
	nutils "github.com/conntectome/cntm/smartcontract/service/native/utils"
)

func TestSigNativeInvokeTx(t *testing.T) {
	defAcc, err := testWallet.GetDefaultAccount(pwd)
	if err != nil {
		t.Errorf("GetDefaultAccount error:%s", err)
		return
	}
	acc1, err := clisvrcom.DefWalletStore.NewAccountData(keypair.PK_ECDSA, keypair.P256, signature.SHA256withECDSA, pwd)
	if err != nil {
		t.Errorf("wallet.NewAccount error:%s", err)
		return
	}
	clisvrcom.DefWalletStore.AddAccountData(acc1)
	invokeReq := &SigNativeInvokeTxReq{
		GasPrice: 0,
		GasLimit: 40000,
		Address:  nutils.CntmContractAddress.ToHexString(),
		Method:   "transfer",
		Version:  0,
		Params: []interface{}{
			[]interface{}{
				[]interface{}{
					defAcc.Address.ToBase58(),
					acc1.Address,
					"10000000000",
				},
			},
		},
	}
	data, err := json.Marshal(invokeReq)
	if err != nil {
		t.Errorf("json.Marshal SigNativeInvokeTxReq error:%s", err)
		return
	}
	req := &clisvrcom.CliRpcRequest{
		Qid:     "t",
		Method:  "signativeinvoketx",
		Params:  data,
		Account: acc1.Address,
		Pwd:     string(pwd),
	}
	rsp := &clisvrcom.CliRpcResponse{}
	abiPath := "../../abi/native_abi_script"
	abi.DefAbiMgr.Init(abiPath)
	SigNativeInvokeTx(req, rsp)
	if rsp.ErrorCode != 0 {
		t.Errorf("SigNativeInvokeTx failed. ErrorCode:%d ErrorInfo:%s", rsp.ErrorCode, rsp.ErrorInfo)
		return
	}
}
