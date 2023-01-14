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

package lock_proxy

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/config"
	"github.com/cntmio/cntmology/smartccntmract/event"
	"github.com/cntmio/cntmology/smartccntmract/service/native"
	"github.com/cntmio/cntmology/smartccntmract/service/native/cross_chain/cross_chain_manager"
	"github.com/cntmio/cntmology/smartccntmract/service/native/cntm"
	"github.com/cntmio/cntmology/smartccntmract/service/native/utils"
)

const (
	LOCK_NAME               = "lock"
	UNLOCK_NAME             = "unlock"
	BIND_PROXY_NAME         = "bindProxy"
	BIND_ASSET_NAME         = "bindAsset"
	WITHDRAW_cntm_NAME       = "withdrawcntm"
	GET_PROXY_HASH_NAME     = "getProxyHash"
	GET_ASSET_HASH_NAME     = "getAssetHash"
	GET_CROSSED_LIMIT_NAME  = "getCrossedLimit"
	GET_CROSSED_AMOUNT_NAME = "getCrossedAmount"

	TARGET_ASSET_HASH_PEFIX = "TargetAssetHash"
	CROSS_LIMIT_PREFIX      = "AssetCrossLimit"
	CROSS_AMOUNT_PREFIX     = "AssetCrossedAmount"
)

func AddLockNotifications(native *native.NativeService, ccntmract, sourceAssetAddress common.Address, toChainId uint64, toCcntmract []byte, targetAssetHash []byte, fromAddress common.Address, toAddress []byte, amount uint64) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			CcntmractAddress: ccntmract,
			States:          []interface{}{LOCK_NAME, hex.EncodeToString(sourceAssetAddress[:]), toChainId, hex.EncodeToString(toCcntmract), hex.EncodeToString(targetAssetHash), fromAddress.ToBase58(), hex.EncodeToString(toAddress), amount},
		})
}
func AddUnLockNotifications(native *native.NativeService, ccntmract common.Address, fromChainId uint64, fromProxyCcntmract []byte, targetAssetHash common.Address, toAddress common.Address, amount uint64) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			CcntmractAddress: ccntmract,
			States:          []interface{}{UNLOCK_NAME, fromChainId, hex.EncodeToString(fromProxyCcntmract), hex.EncodeToString(targetAssetHash[:]), toAddress.ToBase58(), amount},
		})
}

func getCreateTxArgs(toChainID uint64, ccntmractHashBytes []byte, method string, argsBytes []byte) []byte {
	createCrossChainTxParam := &cross_chain_manager.CreateCrossChainTxParam{
		ToChainID:         toChainID,
		ToCcntmractAddress: ccntmractHashBytes,
		Method:            method,
		Args:              argsBytes,
	}
	sink := common.NewZeroCopySink(nil)
	createCrossChainTxParam.Serialization(sink)
	return sink.Bytes()
}

func getTransferInput(state cntm.State) []byte {
	var transfers cntm.Transfers
	transfers.States = []cntm.State{state}
	sink := common.NewZeroCopySink(nil)
	transfers.Serialization(sink)
	return sink.Bytes()
}

func GenBindProxyKey(ccntmract common.Address, chainId uint64) []byte {
	sink := common.NewZeroCopySink(nil)
	sink.WriteUint64(chainId)
	chainIdBytes := sink.Bytes()
	temp := append(ccntmract[:], []byte(BIND_PROXY_NAME)...)
	return append(temp, chainIdBytes...)
}

func GenBindAssetHashKey(ccntmract, assetCcntmract common.Address, chainId uint64) []byte {
	sink := common.NewZeroCopySink(nil)
	sink.WriteUint64(chainId)
	chainIdBytes := sink.Bytes()
	temp := append(ccntmract[:], []byte(BIND_ASSET_NAME)...)
	temp = append(temp, []byte(TARGET_ASSET_HASH_PEFIX)...)
	temp = append(temp, assetCcntmract[:]...)
	return append(temp, chainIdBytes...)
}

func GenCrossedLimitKey(ccntmract, assetCcntmract common.Address, chainId uint64) []byte {
	sink := common.NewZeroCopySink(nil)
	sink.WriteUint64(chainId)
	chainIdBytes := sink.Bytes()
	temp := append(ccntmract[:], []byte(BIND_ASSET_NAME)...)
	temp = append(temp, []byte(CROSS_LIMIT_PREFIX)...)
	temp = append(temp, assetCcntmract[:]...)
	return append(temp, chainIdBytes...)
}

func GenCrossedAmountKey(ccntmract, sourceCcntmract common.Address, chainId uint64) []byte {
	sink := common.NewZeroCopySink(nil)
	sink.WriteUint64(chainId)
	chainIdBytes := sink.Bytes()
	temp := append(ccntmract[:], []byte(CROSS_AMOUNT_PREFIX)...)
	temp = append(temp, sourceCcntmract[:]...)
	return append(temp, chainIdBytes...)
}

func getAmount(native *native.NativeService, storgedKey []byte) (*big.Int, error) {
	valueBs, err := utils.GetStorageVarBytes(native, storgedKey)
	if err != nil {
		return nil, fmt.Errorf("getAmount, error:%s", err)
	}
	//value := common.BigIntFromNeoBytes(valueBs)
	value := big.NewInt(0).SetBytes(valueBs)
	return value, nil
}

func getAllowanceInput() []byte {
	sink := common.NewZeroCopySink(nil)
	sink.WriteAddress(utils.OntCcntmractAddress)
	sink.WriteAddress(utils.LockProxyCcntmractAddress)

	return sink.Bytes()
}

func getTransferFromInput(toAddress common.Address, value uint64) []byte {
	transferFromState := cntm.TransferFrom{
		Sender: utils.LockProxyCcntmractAddress,
		From:   utils.OntCcntmractAddress,
		To:     toAddress,
		Value:  value,
	}
	sink := common.NewZeroCopySink(nil)
	transferFromState.Serialization(sink)
	return sink.Bytes()
}
