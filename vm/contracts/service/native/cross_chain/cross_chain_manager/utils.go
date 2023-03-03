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

package cross_chain_manager

import (
	"encoding/hex"
	"fmt"

	"github.com/conntectome/cntm/common"
	"github.com/conntectome/cntm/common/config"
	cstates "github.com/conntectome/cntm/core/states"
	"github.com/conntectome/cntm/merkle"
	"github.com/conntectome/cntm/smartcontract/event"
	"github.com/conntectome/cntm/smartcontract/service/native"
	ccom "github.com/conntectome/cntm/smartcontract/service/native/cross_chain/common"
	"github.com/conntectome/cntm/smartcontract/service/native/utils"
)

func putCrossChainID(native *native.NativeService, crossChainID uint64) error {
	contract := utils.CrossChainContractAddress
	crossChainIDBytes, err := utils.GetUint64Bytes(crossChainID)
	if err != nil {
		return fmt.Errorf("putCrossChainID, get crossChainIDBytes error: %v", err)
	}
	native.CacheDB.Put(utils.ConcatKey(contract, []byte(CROSS_CHAIN_ID)), cstates.GenRawStorageItem(crossChainIDBytes))
	return nil
}

func getCrossChainID(native *native.NativeService) (uint64, error) {
	contract := utils.CrossChainContractAddress
	var crossChainID uint64 = 0
	value, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(CROSS_CHAIN_ID)))
	if err != nil {
		return 0, fmt.Errorf("getCrossChainID, native.CacheDB.Get error: %v", err)
	}
	if value != nil {
		crossChainIDBytes, err := cstates.GetValueFromRawStorageItem(value)
		if err != nil {
			return 0, fmt.Errorf("getCrossChainID, deserialize from raw storage item err:%v", err)
		}
		crossChainID, err = utils.GetBytesUint64(crossChainIDBytes)
		if err != nil {
			return 0, fmt.Errorf("getCrossChainID, get chainIDBytes error: %v", err)
		}
	}
	return crossChainID, nil
}

func putDoneTx(native *native.NativeService, crossChainID []byte, chainID uint64) error {
	contract := utils.CrossChainContractAddress
	chainIDBytes, err := utils.GetUint64Bytes(chainID)
	if err != nil {
		return fmt.Errorf("putDoneTx, get chainIDBytes error: %v", err)
	}
	native.CacheDB.Put(utils.ConcatKey(contract, []byte(DONE_TX), chainIDBytes, crossChainID),
		cstates.GenRawStorageItem(crossChainID))
	return nil
}

func checkDoneTx(native *native.NativeService, crossChainID []byte, chainID uint64) error {
	contract := utils.CrossChainContractAddress
	chainIDBytes, err := utils.GetUint64Bytes(chainID)
	if err != nil {
		return fmt.Errorf("checkDoneTx, get chainIDBytes error: %v", err)
	}
	value, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(DONE_TX), chainIDBytes, crossChainID))
	if err != nil {
		return fmt.Errorf("checkDoneTx, native.CacheDB.Get error: %v", err)
	}
	if value != nil {
		return fmt.Errorf("checkDoneTx, tx already done")
	}
	return nil
}

func putRequest(native *native.NativeService, crossChainIDBytes, chainIDBytes, request []byte) error {
	contract := utils.CrossChainContractAddress
	utils.PutBytes(native, utils.ConcatKey(contract, []byte(REQUEST), chainIDBytes, crossChainIDBytes), request)
	return nil
}

func MakeFromCntmProof(native *native.NativeService, params *CreateCrossChainTxParam) error {
	//get cross chain ID
	crossChainID, err := getCrossChainID(native)
	if err != nil {
		return fmt.Errorf("MakeFromCntmProof, getCrossChainID error:%s", err)
	}
	err = putCrossChainID(native, crossChainID+1)
	if err != nil {
		return fmt.Errorf("MakeFromCntmProof, putCrossChainID error:%s", err)
	}
	crossChainIDBytes, err := utils.GetUint64Bytes(crossChainID)
	if err != nil {
		return fmt.Errorf("MakeFromCntmProof, utils.GetUint64Bytes error:%s", err)
	}
	//record cross chain tx
	txHash := native.Tx.Hash()
	merkleValue := &ccom.MakeTxParam{
		TxHash:              txHash.ToArray(),
		CrossChainID:        crossChainIDBytes,
		FromContractAddress: native.ContextRef.CallingContext().ContractAddress[:],
		ToChainID:           params.ToChainID,
		ToContractAddress:   params.ToContractAddress,
		Method:              params.Method,
		Args:                params.Args,
	}
	sink := common.NewZeroCopySink(nil)
	merkleValue.Serialization(sink)
	chainIDBytes, err := utils.GetUint64Bytes(params.ToChainID)
	if err != nil {
		return fmt.Errorf("MakeFromCntmProof, get chainIDBytes error: %v", err)
	}
	err = putRequest(native, crossChainIDBytes, chainIDBytes, sink.Bytes())
	if err != nil {
		return fmt.Errorf("MakeFromCntmProof, putRequest error:%s", err)
	}
	native.PushCrossState(sink.Bytes())
	key := hex.EncodeToString(utils.ConcatBytes([]byte(REQUEST), chainIDBytes, crossChainIDBytes))
	args := hex.EncodeToString(params.Args)
	notifyMakeFromCntmProof(native, hex.EncodeToString(merkleValue.TxHash), params.ToChainID, key,
		hex.EncodeToString(merkleValue.FromContractAddress), args)
	return nil
}

func VerifyToCntmTx(native *native.NativeService, proof []byte, fromChainid uint64, header *ccom.Header) (*ccom.ToMerkleValue, error) {
	v, err := merkle.MerkleProve(proof, header.CrossStateRoot)
	if err != nil {
		return nil, fmt.Errorf("VerifyToCntmTx, merkle.MerkleProve verify merkle proof error: %v", err)
	}

	s := common.NewZeroCopySource(v)
	merkleValue := new(ccom.ToMerkleValue)
	if err := merkleValue.Deserialization(s); err != nil {
		return nil, fmt.Errorf("VerifyToCntmTx, deserialize merkleValue error:%s", err)
	}

	//record done cross chain tx
	err = checkDoneTx(native, merkleValue.MakeTxParam.CrossChainID, fromChainid)
	if err != nil {
		return nil, fmt.Errorf("VerifyToCntmTx, checkDoneTx error:%s", err)
	}
	err = putDoneTx(native, merkleValue.MakeTxParam.CrossChainID, fromChainid)
	if err != nil {
		return nil, fmt.Errorf("VerifyToCntmTx, putDoneTx error:%s", err)
	}

	notifyVerifyToCntmProof(native, hex.EncodeToString(merkleValue.TxHash), hex.EncodeToString(merkleValue.MakeTxParam.TxHash),
		fromChainid, hex.EncodeToString(merkleValue.MakeTxParam.ToContractAddress))
	return merkleValue, nil
}

func notifyMakeFromCntmProof(native *native.NativeService, txHash string, toChainID uint64, key string, contract, args string) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			ContractAddress: utils.CrossChainContractAddress,
			States:          []interface{}{MAKE_FROM_CNTM_PROOF, txHash, toChainID, native.Height, key, contract, args},
		})
}

func notifyVerifyToCntmProof(native *native.NativeService, txHash, rawTxHash string, fromChainID uint64, contract string) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			ContractAddress: utils.CrossChainContractAddress,
			States:          []interface{}{VERIFY_TO_CNTM_PROOF, txHash, rawTxHash, fromChainID, native.Height, contract},
		})
}

func getUnlockArgs(args []byte, fromContractAddress []byte, fromChainID uint64) []byte {
	sink := common.NewZeroCopySink(nil)
	utils.EncodeVarBytes(sink, args)
	utils.EncodeVarBytes(sink, fromContractAddress)
	utils.EncodeVarUint(sink, fromChainID)
	return sink.Bytes()
}
