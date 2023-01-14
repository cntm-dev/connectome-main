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
	"encoding/hex"
	"fmt"

	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/config"
	cstates "github.com/cntmio/cntmology/core/states"
	"github.com/cntmio/cntmology/merkle"
	"github.com/cntmio/cntmology/smartccntmract/event"
	"github.com/cntmio/cntmology/smartccntmract/service/native"
	ccom "github.com/cntmio/cntmology/smartccntmract/service/native/cross_chain/common"
	"github.com/cntmio/cntmology/smartccntmract/service/native/utils"
)

func putCrossChainID(native *native.NativeService, crossChainID uint64) error {
	ccntmract := utils.CrossChainCcntmractAddress
	crossChainIDBytes, err := utils.GetUint64Bytes(crossChainID)
	if err != nil {
		return fmt.Errorf("putCrossChainID, get crossChainIDBytes error: %v", err)
	}
	native.CacheDB.Put(utils.ConcatKey(ccntmract, []byte(CROSS_CHAIN_ID)), cstates.GenRawStorageItem(crossChainIDBytes))
	return nil
}

func getCrossChainID(native *native.NativeService) (uint64, error) {
	ccntmract := utils.CrossChainCcntmractAddress
	var crossChainID uint64 = 0
	value, err := native.CacheDB.Get(utils.ConcatKey(ccntmract, []byte(CROSS_CHAIN_ID)))
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
	ccntmract := utils.CrossChainCcntmractAddress
	chainIDBytes, err := utils.GetUint64Bytes(chainID)
	if err != nil {
		return fmt.Errorf("putDoneTx, get chainIDBytes error: %v", err)
	}
	native.CacheDB.Put(utils.ConcatKey(ccntmract, []byte(DONE_TX), chainIDBytes, crossChainID),
		cstates.GenRawStorageItem(crossChainID))
	return nil
}

func checkDoneTx(native *native.NativeService, crossChainID []byte, chainID uint64) error {
	ccntmract := utils.CrossChainCcntmractAddress
	chainIDBytes, err := utils.GetUint64Bytes(chainID)
	if err != nil {
		return fmt.Errorf("checkDoneTx, get chainIDBytes error: %v", err)
	}
	value, err := native.CacheDB.Get(utils.ConcatKey(ccntmract, []byte(DONE_TX), chainIDBytes, crossChainID))
	if err != nil {
		return fmt.Errorf("checkDoneTx, native.CacheDB.Get error: %v", err)
	}
	if value != nil {
		return fmt.Errorf("checkDoneTx, tx already done")
	}
	return nil
}

func putRequest(native *native.NativeService, crossChainIDBytes, chainIDBytes, request []byte) error {
	ccntmract := utils.CrossChainCcntmractAddress
	utils.PutBytes(native, utils.ConcatKey(ccntmract, []byte(REQUEST), chainIDBytes, crossChainIDBytes), request)
	return nil
}

func MakeFromOntProof(native *native.NativeService, params *CreateCrossChainTxParam) error {
	//get cross chain ID
	crossChainID, err := getCrossChainID(native)
	if err != nil {
		return fmt.Errorf("MakeFromOntProof, getCrossChainID error:%s", err)
	}
	err = putCrossChainID(native, crossChainID+1)
	if err != nil {
		return fmt.Errorf("MakeFromOntProof, putCrossChainID error:%s", err)
	}
	crossChainIDBytes, err := utils.GetUint64Bytes(crossChainID)
	if err != nil {
		return fmt.Errorf("MakeFromOntProof, utils.GetUint64Bytes error:%s", err)
	}
	//record cross chain tx
	txHash := native.Tx.Hash()
	merkleValue := &ccom.MakeTxParam{
		TxHash:              txHash.ToArray(),
		CrossChainID:        crossChainIDBytes,
		FromCcntmractAddress: native.CcntmextRef.CallingCcntmext().CcntmractAddress[:],
		ToChainID:           params.ToChainID,
		ToCcntmractAddress:   params.ToCcntmractAddress,
		Method:              params.Method,
		Args:                params.Args,
	}
	sink := common.NewZeroCopySink(nil)
	merkleValue.Serialization(sink)
	chainIDBytes, err := utils.GetUint64Bytes(params.ToChainID)
	if err != nil {
		return fmt.Errorf("MakeFromOntProof, get chainIDBytes error: %v", err)
	}
	err = putRequest(native, crossChainIDBytes, chainIDBytes, sink.Bytes())
	if err != nil {
		return fmt.Errorf("MakeFromOntProof, putRequest error:%s", err)
	}
	native.PushCrossState(sink.Bytes())
	key := hex.EncodeToString(utils.ConcatBytes([]byte(REQUEST), chainIDBytes, crossChainIDBytes))
	args := hex.EncodeToString(params.Args)
	notifyMakeFromOntProof(native, hex.EncodeToString(merkleValue.TxHash), params.ToChainID, key,
		hex.EncodeToString(merkleValue.FromCcntmractAddress), args)
	return nil
}

func VerifyToOntTx(native *native.NativeService, proof []byte, fromChainid uint64, header *ccom.Header) (*ccom.ToMerkleValue, error) {
	v, err := merkle.MerkleProve(proof, header.CrossStateRoot)
	if err != nil {
		return nil, fmt.Errorf("VerifyToOntTx, merkle.MerkleProve verify merkle proof error: %v", err)
	}

	s := common.NewZeroCopySource(v)
	merkleValue := new(ccom.ToMerkleValue)
	if err := merkleValue.Deserialization(s); err != nil {
		return nil, fmt.Errorf("VerifyToOntTx, deserialize merkleValue error:%s", err)
	}

	//record done cross chain tx
	err = checkDoneTx(native, merkleValue.MakeTxParam.CrossChainID, fromChainid)
	if err != nil {
		return nil, fmt.Errorf("VerifyToOntTx, checkDoneTx error:%s", err)
	}
	err = putDoneTx(native, merkleValue.MakeTxParam.CrossChainID, fromChainid)
	if err != nil {
		return nil, fmt.Errorf("VerifyToOntTx, putDoneTx error:%s", err)
	}

	notifyVerifyToOntProof(native, hex.EncodeToString(merkleValue.TxHash), hex.EncodeToString(merkleValue.MakeTxParam.TxHash),
		fromChainid, hex.EncodeToString(merkleValue.MakeTxParam.ToCcntmractAddress))
	return merkleValue, nil
}

func notifyMakeFromOntProof(native *native.NativeService, txHash string, toChainID uint64, key string, ccntmract, args string) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			CcntmractAddress: utils.CrossChainCcntmractAddress,
			States:          []interface{}{MAKE_FROM_cntm_PROOF, txHash, toChainID, native.Height, key, ccntmract, args},
		})
}

func notifyVerifyToOntProof(native *native.NativeService, txHash, rawTxHash string, fromChainID uint64, ccntmract string) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			CcntmractAddress: utils.CrossChainCcntmractAddress,
			States:          []interface{}{VERIFY_TO_cntm_PROOF, txHash, rawTxHash, fromChainID, native.Height, ccntmract},
		})
}

func getUnlockArgs(args []byte, fromCcntmractAddress []byte, fromChainID uint64) []byte {
	sink := common.NewZeroCopySink(nil)
	utils.EncodeVarBytes(sink, args)
	utils.EncodeVarBytes(sink, fromCcntmractAddress)
	utils.EncodeVarUint(sink, fromChainID)
	return sink.Bytes()
}
