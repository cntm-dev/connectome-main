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

package cross_chain_manager

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/smartccntmract/service/native"
	ccom "github.com/cntmio/cntmology/smartccntmract/service/native/cross_chain/common"
	"github.com/cntmio/cntmology/smartccntmract/service/native/cross_chain/header_sync"
	"github.com/cntmio/cntmology/smartccntmract/service/native/utils"
	"github.com/cntmio/cntmology/vm/neovm/types"
)

const (
	CREATE_CROSS_CHAIN_TX  = "createCrossChainTx"
	PROCESS_CROSS_CHAIN_TX = "processCrossChainTx"
	MAKE_FROM_cntm_PROOF    = "makeFromOntProof"
	VERIFY_TO_cntm_PROOF    = "verifyToOntProof"

	//key prefix
	DONE_TX        = "doneTx"
	REQUEST        = "request"
	CROSS_CHAIN_ID = "crossChainID"

	//cntm chain id
	cntm_CHAIN_ID = 3
)

//Init governance ccntmract address
func InitCrossChain() {
	native.Ccntmracts[utils.CrossChainCcntmractAddress] = RegisterCrossChainCcntmract
}

//Register methods of governance ccntmract
func RegisterCrossChainCcntmract(native *native.NativeService) {
	native.Register(CREATE_CROSS_CHAIN_TX, CreateCrossChainTx)
	native.Register(PROCESS_CROSS_CHAIN_TX, ProcessCrossChainTx)
}

func CreateCrossChainTx(native *native.NativeService) ([]byte, error) {
	params := new(CreateCrossChainTxParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("CreateCrossChainTx, ccntmract params deserialize error: %v", err)
	}

	err := MakeFromOntProof(native, params)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("CreateCrossChainTx, MakeOntProof error: %v", err)
	}
	return utils.BYTE_TRUE, nil
}

func ProcessCrossChainTx(native *native.NativeService) ([]byte, error) {
	params := new(ProcessCrossChainTxParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, ccntmract params deserialize error: %v", err)
	}

	//get block header
	header, err := header_sync.GetHeaderByHeight(native, params.FromChainID, params.Height)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, %d, %d", params.FromChainID, params.Height)
	}
	if header == nil {
		header2 := new(ccom.Header)
		err := header2.Deserialization(common.NewZeroCopySource(params.Header))
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, deserialize header error: %v", err)
		}
		if err := header_sync.ProcessHeader(native, header2, params.Header); err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, error: %s", err)
		}
		header = header2
	}

	proof, err := hex.DecodeString(params.Proof)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, proof hex.DecodeString error: %v", err)
	}
	merkleValue, err := VerifyToOntTx(native, proof, params.FromChainID, header)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, VerifyOntTx error: %v", err)
	}

	if merkleValue.MakeTxParam.ToChainID != cntm_CHAIN_ID {
		return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, to chain id is not cntm")
	}

	//call cross chain function
	dest, err := common.AddressParseFromBytes(merkleValue.MakeTxParam.ToCcntmractAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, common.AddressParseFromBytes error: %v", err)
	}
	functionName := merkleValue.MakeTxParam.Method
	fromCcntmractAddress := merkleValue.MakeTxParam.FromCcntmractAddress
	args := merkleValue.MakeTxParam.Args

	var res interface{}
	if bytes.Equal(merkleValue.MakeTxParam.ToCcntmractAddress, utils.LockProxyCcntmractAddress[:]) {
		argsBytes := getUnlockArgs(args, fromCcntmractAddress, merkleValue.FromChainID)
		_, err = native.NativeCall(utils.LockProxyCcntmractAddress, functionName, argsBytes)
		if err != nil {
			return utils.BYTE_FALSE, err
		}
	} else {
		res, err = ccom.CrossChainNeoVMCall(native, dest, functionName, args, fromCcntmractAddress, merkleValue.FromChainID)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, native.NeoVMCall error: %v", err)
		}
		v, err := res.(*types.VmValue).AsBigInt()
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, result error")
		}
		if v.Cmp(new(big.Int).SetUint64(0)) == 0 {
			return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, res of neo vm call is false")
		}
	}
	return utils.BYTE_TRUE, nil
}
