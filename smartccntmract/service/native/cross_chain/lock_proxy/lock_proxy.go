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
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/config"
	"github.com/cntmio/cntmology/core/genesis"
	"github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/smartccntmract/event"
	"github.com/cntmio/cntmology/smartccntmract/service/native"
	"github.com/cntmio/cntmology/smartccntmract/service/native/cross_chain/cross_chain_manager"
	"github.com/cntmio/cntmology/smartccntmract/service/native/global_params"
	"github.com/cntmio/cntmology/smartccntmract/service/native/cntm"
	"github.com/cntmio/cntmology/smartccntmract/service/native/utils"
)

func InitLockProxy() {
	native.Ccntmracts[utils.LockProxyCcntmractAddress] = RegisterLockProxyCcntmract
}

func RegisterLockProxyCcntmract(native *native.NativeService) {
	native.Register(LOCK_NAME, Lock)
	native.Register(UNLOCK_NAME, Unlock)
	native.Register(BIND_PROXY_NAME, BindProxyHash)
	native.Register(BIND_ASSET_NAME, BindAssetHash)
	native.Register(WITHDRAW_cntm_NAME, Withdrawcntm)
	native.Register(GET_PROXY_HASH_NAME, GetProxyHash)
	native.Register(GET_ASSET_HASH_NAME, GetAssetHash)
	native.Register(GET_CROSSED_AMOUNT_NAME, GetCrossedAmount)
	native.Register(GET_CROSSED_LIMIT_NAME, GetCrossedLimit)
}

func BindProxyHash(native *native.NativeService) ([]byte, error) {
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	source := common.NewZeroCopySource(native.Input)
	var bindParam BindProxyParam
	if err := bindParam.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[BindProxyHash] Deserialize BindProxyParam error:%s", err)
	}
	// get operator from database
	operatorAddress, err := global_params.GetStorageRole(native,
		global_params.GenerateOperatorKey(utils.ParamCcntmractAddress))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[BindProxyHash] get operator error:%s", err)
	}
	//check witness
	if err = utils.ValidateOwner(native, operatorAddress); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[BindProxyHash] checkWitness error:%s", err)
	}
	native.CacheDB.Put(GenBindProxyKey(ccntmract, bindParam.TargetChainId), utils.GenVarBytesStorageItem(bindParam.TargetHash).ToArray())
	if config.DefConfig.Common.EnableEventLog {
		native.Notifications = append(native.Notifications,
			&event.NotifyEventInfo{
				CcntmractAddress: ccntmract,
				States:          []interface{}{BIND_PROXY_NAME, bindParam.TargetChainId, hex.EncodeToString(bindParam.TargetHash)},
			})
	}
	return utils.BYTE_TRUE, nil
}
func BindAssetHash(native *native.NativeService) ([]byte, error) {
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	source := common.NewZeroCopySource(native.Input)
	var bindParam BindAssetParam
	if err := bindParam.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[BindAssetHash] Deserialization BindAssetParam error:%s", err)
	}
	// get operator from database
	operatorAddress, err := global_params.GetStorageRole(native,
		global_params.GenerateOperatorKey(utils.ParamCcntmractAddress))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[BindAssetHash] get operator error:%s", err)
	}
	//check witness
	if err = utils.ValidateOwner(native, operatorAddress); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[BindAssetHash] checkWitness error:%s", err)
	}
	// store the target asset hash
	native.CacheDB.Put(GenBindAssetHashKey(ccntmract, bindParam.SourceAssetHash, bindParam.TargetChainId), utils.GenVarBytesStorageItem(bindParam.TargetAssetHash).ToArray())

	// make sure the new limit is greater than the stored limit
	limitKey := GenCrossedLimitKey(ccntmract, bindParam.SourceAssetHash, bindParam.TargetChainId)
	limit, err := getAmount(native, limitKey)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[BindAssetHash] getCrossedLimit error:%s", err)
	}
	if bindParam.Limit.Cmp(limit) != 1 {
		return utils.BYTE_FALSE, fmt.Errorf("[BindAssetHash] new Limit:%s should be greater than stored Limit:%s", bindParam.Limit.String(), limit.String())
	}
	// if the source asset hash is the target chain asset, increase the crossedAmount value by the limit increment
	if bindParam.IsTargetChainAsset {
		increment := big.NewInt(0).Sub(bindParam.Limit, limit)
		crossedAmountKey := GenCrossedAmountKey(ccntmract, bindParam.SourceAssetHash, bindParam.TargetChainId)
		crossedAmount, err := getAmount(native, crossedAmountKey)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("[BindAssetHash] getCrossedAmount error:%s", err)
		}
		newCrossedAmount := big.NewInt(0).Add(crossedAmount, increment)
		if newCrossedAmount.Cmp(crossedAmount) != 1 {
			return utils.BYTE_FALSE, fmt.Errorf("[BindAssetHash] new crossedAmount:%s is not greater than stored crossed amount:%s", newCrossedAmount.String(), crossedAmount.String())
		}
		native.CacheDB.Put(crossedAmountKey, utils.GenVarBytesStorageItem(newCrossedAmount.Bytes()).ToArray())
	}
	// update the new limit
	native.CacheDB.Put(GenCrossedLimitKey(ccntmract, bindParam.SourceAssetHash, bindParam.TargetChainId), utils.GenVarBytesStorageItem(bindParam.Limit.Bytes()).ToArray())
	if config.DefConfig.Common.EnableEventLog {
		native.Notifications = append(native.Notifications,
			&event.NotifyEventInfo{
				CcntmractAddress: ccntmract,
				States:          []interface{}{BIND_ASSET_NAME, hex.EncodeToString(bindParam.SourceAssetHash[:]), bindParam.TargetChainId, hex.EncodeToString(bindParam.TargetAssetHash), bindParam.Limit.String()},
			})
	}
	return utils.BYTE_TRUE, nil
}

func Lock(native *native.NativeService) ([]byte, error) {
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	cntmCcntmract := utils.OntCcntmractAddress
	cntmCcntmract := utils.OngCcntmractAddress
	source := common.NewZeroCopySource(native.Input)

	var lockParam LockParam
	if err := lockParam.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[Lock] ccntmract params deserialization error:%s", err)
	}

	if lockParam.Value == 0 {
		return utils.BYTE_FALSE, nil
	}
	// currently, only support cntm and cntm lock operation
	if lockParam.SourceAssetHash != cntmCcntmract && lockParam.SourceAssetHash != cntmCcntmract {
		return utils.BYTE_FALSE, fmt.Errorf("[Lock] only support cntm/cntm lock, expect:%s or %s, but got:%s", hex.EncodeToString(cntmCcntmract[:]), hex.EncodeToString(cntmCcntmract[:]), hex.EncodeToString(lockParam.SourceAssetHash[:]))
	}

	// transfer cntm or cntm from FromAddress to lockCcntmract
	state := cntm.State{
		From:  lockParam.FromAddress,
		To:    ccntmract,
		Value: lockParam.Value,
	}
	transferInput := getTransferInput(state)
	if _, err := native.NativeCall(lockParam.SourceAssetHash, cntm.TRANSFER_NAME, transferInput); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[Lock] NativeCall ccntmract:%s 'transfer(%s, %s, %d)' error:%s", hex.EncodeToString(lockParam.SourceAssetHash[:]), lockParam.FromAddress.ToBase58(), hex.EncodeToString(ccntmract[:]), lockParam.Value, err)
	}

	// make sure new crossed amount is strictly greater than old crossed amount and no less than the limit
	crossedAmount, err := getAmount(native, GenCrossedAmountKey(ccntmract, lockParam.SourceAssetHash, lockParam.ToChainID))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[Lock] getCrossedAmount error:%s", err)
	}
	limit, err := getAmount(native, GenCrossedLimitKey(ccntmract, lockParam.SourceAssetHash, lockParam.ToChainID))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[Lock] getCrossedLimit error:%s", err)
	}
	newCrossedAmount := big.NewInt(0).Add(crossedAmount, big.NewInt(0).SetUint64(lockParam.Value))
	if newCrossedAmount.Cmp(crossedAmount) != 1 || newCrossedAmount.Cmp(limit) == 1 {
		return utils.BYTE_FALSE, fmt.Errorf("[Lock] Is new crossedAmount GREATER than old crossedAmount?:%t, Is new crossedAmount SMALLER than limit?:%t", newCrossedAmount.Cmp(crossedAmount) == 1, newCrossedAmount.Cmp(limit) != 1)
	}
	// increase the new crossed amount by Value
	native.CacheDB.Put(GenCrossedAmountKey(ccntmract, lockParam.SourceAssetHash, lockParam.ToChainID), utils.GenVarBytesStorageItem(newCrossedAmount.Bytes()).ToArray())

	// get target chain proxy hash from storage
	targetProxyHashBs, err := utils.GetStorageVarBytes(native, GenBindProxyKey(ccntmract, lockParam.ToChainID))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[Lock] get bind proxy ccntmract hash with chainID:%d error:%s", lockParam.ToChainID, err)
	}
	if len(targetProxyHashBs) == 0 {
		return utils.BYTE_FALSE, fmt.Errorf("[Lock] get bind proxy ccntmract hash with chainID:%d ccntmractHash empty", lockParam.ToChainID)
	}
	// get target asset hash from storage
	targetAssetHashBs, err := utils.GetStorageVarBytes(native, GenBindAssetHashKey(ccntmract, lockParam.SourceAssetHash, lockParam.ToChainID))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[Lock] get bind asset ccntmract hash:%s with chainID:%d error:%s", hex.EncodeToString(lockParam.SourceAssetHash[:]), lockParam.ToChainID, err)
	}
	args := Args{
		TargetAssetHash: targetAssetHashBs,
		ToAddress:       lockParam.ToAddress,
		Value:           lockParam.Value,
	}
	sink := common.NewZeroCopySink(nil)
	args.Serialization(sink)
	input := getCreateTxArgs(lockParam.ToChainID, targetProxyHashBs, UNLOCK_NAME, sink.Bytes())
	if _, err = native.NativeCall(utils.CrossChainCcntmractAddress, cross_chain_manager.CREATE_CROSS_CHAIN_TX, input); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[Lock] NativeCall %s createCrossChainTx 'createTx', error:%s", hex.EncodeToString(utils.CrossChainCcntmractAddress[:]), err)
	}

	AddLockNotifications(native, ccntmract, lockParam.SourceAssetHash, lockParam.ToChainID, targetProxyHashBs, targetAssetHashBs, lockParam.FromAddress, lockParam.ToAddress, lockParam.Value)
	return utils.BYTE_TRUE, nil
}

func Unlock(native *native.NativeService) ([]byte, error) {
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	// this method cannot be invoked by anybody except CrossChainCcntmractAddress
	if !native.CcntmextRef.CheckWitness(utils.CrossChainCcntmractAddress) {
		return utils.BYTE_FALSE, fmt.Errorf("[Unlock] can ONLY be invoked by CrossChainCcntmractAddress:%s Ccntmract, checkwitness failed!", hex.EncodeToString(utils.CrossChainCcntmractAddress[:]))
	}
	cntmCcntmract := utils.OntCcntmractAddress
	cntmCcntmract := utils.OngCcntmractAddress

	var unlockParam UnlockParam
	if err := unlockParam.Deserialization(common.NewZeroCopySource(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[Unlock] ccntmract params deserialization error:%s", err)
	}

	var args Args
	if err := args.Deserialization(common.NewZeroCopySource(unlockParam.ArgsBs)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[Unlock] deserialize args error:%s", err)
	}
	// only recognize the params from proxy ccntmract already bound with chainId in current proxy ccntmract
	proxyCcntmractHash, err := utils.GetStorageVarBytes(native, GenBindProxyKey(ccntmract, unlockParam.FromChainId))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[Unlock] get bind proxy ccntmract hash with chainID:%d error:%s", unlockParam.FromChainId, err)
	}
	if !bytes.Equal(proxyCcntmractHash, unlockParam.FromCcntmractHashBs) {
		return utils.BYTE_FALSE, fmt.Errorf("[Unlock] passed in ccntmractHash NOT equal stored ccntmractHash with chainID:%d, expect:%s, got:%s", unlockParam.FromChainId, hex.EncodeToString(proxyCcntmractHash), hex.EncodeToString(unlockParam.FromCcntmractHashBs))
	}

	// currently, only support cntm and cntm unlock operation
	if !bytes.Equal(args.TargetAssetHash, cntmCcntmract[:]) && !bytes.Equal(args.TargetAssetHash, cntmCcntmract[:]) {
		return utils.BYTE_FALSE, fmt.Errorf("[Unlock] target asset hash, Is cntm ccntmract?:%t, Is cntm ccntmract?:%t, Args.TargetAssetHash:%s", bytes.Equal(args.TargetAssetHash, cntmCcntmract[:]), bytes.Equal(args.ToAddress, cntmCcntmract[:]), hex.EncodeToString(args.TargetAssetHash))
	}

	assetAddress, err := common.AddressParseFromBytes(args.TargetAssetHash)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[Unlock] parse from Args.TargetAssetHash to ccntmract address format error:%s", err)
	}
	toAddress, err := common.AddressParseFromBytes(args.ToAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[Unlock] parse from Args.ToAddress to cntmology address format error:%s", err)
	}
	if args.Value == 0 {
		return utils.BYTE_TRUE, nil
	}
	// unlock cntm or cntm from current proxy ccntmract into toAddress
	transferInput := getTransferInput(cntm.State{ccntmract, toAddress, args.Value})
	if _, err = native.NativeCall(assetAddress, cntm.TRANSFER_NAME, transferInput); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[Unlock] NativeCall ccntmract:%s 'transfer(%s, %s, %d)' error:%s", hex.EncodeToString(assetAddress[:]), hex.EncodeToString(ccntmract[:]), toAddress.ToBase58(), args.Value, err)
	}

	// make sure new crossed amount is strictly less than old crossed amount and no less than the limit
	crossedAmount, err := getAmount(native, GenCrossedAmountKey(ccntmract, assetAddress, unlockParam.FromChainId))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[Unlock] getCrossedAmount error:%s", err)
	}
	newCrossedAmount := big.NewInt(0).Sub(crossedAmount, big.NewInt(0).SetUint64(args.Value))
	if newCrossedAmount.Cmp(crossedAmount) != -1 {
		return utils.BYTE_FALSE, fmt.Errorf("[Unlock] new crossedAmount:%s should be less than old crossedAmount:%s", newCrossedAmount.String(), crossedAmount.String())
	}
	// decrease the new crossed amount by Value
	native.CacheDB.Put(GenCrossedAmountKey(ccntmract, assetAddress, unlockParam.FromChainId), utils.GenVarBytesStorageItem(newCrossedAmount.Bytes()).ToArray())

	AddUnLockNotifications(native, ccntmract, unlockParam.FromChainId, unlockParam.FromCcntmractHashBs, assetAddress, toAddress, args.Value)

	return utils.BYTE_TRUE, nil
}

func Withdrawcntm(native *native.NativeService) ([]byte, error) {
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	source := common.NewZeroCopySource(native.Input)
	toAddress, eof := source.NextAddress()
	if eof {
		return utils.BYTE_FALSE, fmt.Errorf("[Withdrawcntm] input DecodeAddress toAddress error!")
	}
	operatorAddress, err := types.AddressFromBookkeepers(genesis.GenesisBookkeepers)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[Withdrawcntm] get operator error: %v", err)
	}
	//check witness
	if err = utils.ValidateOwner(native, operatorAddress); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[Withdrawcntm] checkWitness error: %v", err)
	}
	// query unbound cntm or allowance
	allowanceInput := getAllowanceInput()
	allowanceRes, err := native.NativeCall(utils.OngCcntmractAddress, cntm.ALLOWANCE_NAME, allowanceInput)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[Withdrawcntm] invoke cntm ccntmract get allowance error:%s", err)
	}
	allowance := common.BigIntFromNeoBytes(allowanceRes)
	// transfer cntm to toAddress
	transferFromInput := getTransferFromInput(toAddress, allowance.Uint64())
	if _, err = native.NativeCall(utils.OngCcntmractAddress, cntm.TRANSFERFROM_NAME, transferFromInput); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[Withdrawcntm] invoke cntm ccntmract, transferFrom(lockProxy, cntmCcntmract, toAddress, unboundOngAmount) err:%s", err)
	}
	if config.DefConfig.Common.EnableEventLog {
		native.Notifications = append(native.Notifications,
			&event.NotifyEventInfo{
				CcntmractAddress: ccntmract,
				States:          []interface{}{WITHDRAW_cntm_NAME, toAddress.ToBase58(), allowance.Uint64()},
			})
	}
	return utils.BYTE_TRUE, nil
}

func GetProxyHash(native *native.NativeService) ([]byte, error) {
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	toChainId, err := utils.DecodeVarUint(common.NewZeroCopySource(native.Input))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[GetProxyHash] input DecodeVarUint toChainId error:%s", err)
	}
	proxyHash, err := utils.GetStorageVarBytes(native, GenBindProxyKey(ccntmract, toChainId))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[GetProxyHash] get proxy hash with toChainId:%d error:%s", toChainId, err)
	}
	return proxyHash, nil
}

func GetAssetHash(native *native.NativeService) ([]byte, error) {
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	source := common.NewZeroCopySource(native.Input)
	sourceAssetAddress, err := utils.DecodeAddress(source)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[GetAssetHash] input DecodeAddress sourceAssetAddress error:%s", err)
	}
	toChainId, err := utils.DecodeVarUint(source)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[GetAssetHash] input DecodeVarUint toChainId error:%s", err)
	}
	toAssetHash, err := utils.GetStorageVarBytes(native, GenBindAssetHashKey(ccntmract, sourceAssetAddress, toChainId))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[GetAssetHash] get asset hash with toChainId:%d for sourceAssetAddress:%s error:%s", toChainId, hex.EncodeToString(sourceAssetAddress[:]), err)
	}
	return toAssetHash, nil
}

func GetCrossedAmount(native *native.NativeService) ([]byte, error) {
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	source := common.NewZeroCopySource(native.Input)
	sourceAssetAddress, err := utils.DecodeAddress(source)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[GetCrossedAmount] input DecodeAddress sourceAssetAddress error:%s", err)
	}
	toChainId, err := utils.DecodeVarUint(source)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[GetCrossedAmount] input DecodeVarUint toChainId error:%s", err)
	}

	crossedAmountBs, err := utils.GetStorageVarBytes(native, GenCrossedAmountKey(ccntmract, sourceAssetAddress, toChainId))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[GetCrossedAmount] get crossed amount in big.Int bytes format with toChainId:%d for sourceAssetAddress:%s error:%s", toChainId, sourceAssetAddress.ToHexString(), err)
	}
	return common.BigIntToNeoBytes(big.NewInt(0).SetBytes(crossedAmountBs)), nil
}

func GetCrossedLimit(native *native.NativeService) ([]byte, error) {
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	source := common.NewZeroCopySource(native.Input)
	sourceAssetAddress, err := utils.DecodeAddress(source)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[GetCrossedLimit] input DecodeAddress sourceAssetAddress error:%s", err)
	}
	toChainId, err := utils.DecodeVarUint(source)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[GetCrossedLimit] input DecodeVarUint toChainId error:%s", err)
	}

	crossedLimitBs, err := utils.GetStorageVarBytes(native, GenCrossedLimitKey(ccntmract, sourceAssetAddress, toChainId))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[GetCrossedLimit] get crossed limit in big.Int bytes format with toChainId:%d for sourceAssetAddress:%s error:%s", toChainId, sourceAssetAddress.ToHexString(), err)
	}
	return common.BigIntToNeoBytes(big.NewInt(0).SetBytes(crossedLimitBs)), nil
}
