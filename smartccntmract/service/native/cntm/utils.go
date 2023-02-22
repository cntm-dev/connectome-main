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

package cntm

import (
	"fmt"

	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/config"
	"github.com/cntmio/cntmology/common/constants"
	cstates "github.com/cntmio/cntmology/core/states"
	"github.com/cntmio/cntmology/errors"
	"github.com/cntmio/cntmology/smartccntmract/event"
	"github.com/cntmio/cntmology/smartccntmract/service/native"
	"github.com/cntmio/cntmology/smartccntmract/service/native/utils"
)

// storage key
const (
	UNBOUND_TIME_OFFSET_KEY = "unboundTimeOffset"
	TOTAL_SUPPLY_KEY        = "totalSupply"
)

// methods
const (
	INIT_NAME            = "init"
	TRANSFER_NAME        = "transfer"
	APPROVE_NAME         = "approve"
	TRANSFERFROM_NAME    = "transferFrom"
	NAME_NAME            = "name"
	SYMBOL_NAME          = "symbol"
	DECIMALS_NAME        = "decimals"
	TOTAL_SUPPLY_NAME    = "totalSupply"
	BALANCEOF_NAME       = "balanceOf"
	ALLOWANCE_NAME       = "allowance"
	TOTAL_ALLOWANCE_NAME = "totalAllowance"

	TRANSFER_V2_NAME        = "transferV2"
	APPROVE_V2_NAME         = "approveV2"
	TRANSFERFROM_V2_NAME    = "transferFromV2"
	DECIMALS_V2_NAME        = "decimalsV2"
	TOTAL_SUPPLY_V2_NAME    = "totalSupplyV2"
	BALANCEOF_V2_NAME       = "balanceOfV2"
	ALLOWANCE_V2_NAME       = "allowanceV2"
	TOTAL_ALLOWANCE_V2_NAME = "totalAllowanceV2"

	UNBOUND_cntm_TO_GOVERNANCE = "unboundOngToGovernance"
)

func AddTransferNotifications(native *native.NativeService, ccntmract common.Address, state *TransferStateV2) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	states := []interface{}{TRANSFER_NAME, state.From.ToBase58(), state.To.ToBase58(), state.Value.MustToInteger64()}
	if state.Value.IsFloat() {
		states = append(states, state.Value.FloatPart())
	}
	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			CcntmractAddress: ccntmract,
			States:          states,
		})
}

func GetToUInt64StorageItem(toBalance, value uint64) *cstates.StorageItem {
	sink := common.NewZeroCopySink(nil)
	sink.WriteUint64(toBalance + value)
	return &cstates.StorageItem{Value: sink.Bytes()}
}

func GenTotalSupplyKey(ccntmract common.Address) []byte {
	return append(ccntmract[:], TOTAL_SUPPLY_KEY...)
}

func GenBalanceKey(ccntmract, addr common.Address) []byte {
	return append(ccntmract[:], addr[:]...)
}

func Transfer(native *native.NativeService, ccntmract common.Address, from, to common.Address,
	value cstates.NativeTokenBalance) (oldFrom, oldTo cstates.NativeTokenBalance, err error) {
	if !native.CcntmextRef.CheckWitness(from) {
		return oldFrom, oldTo, errors.NewErr("authentication failed!")
	}

	oldFrom, err = reduceFromBalance(native, GenBalanceKey(ccntmract, from), value)
	if err != nil {
		return
	}

	oldTo, err = increaseToBalance(native, GenBalanceKey(ccntmract, to), value)
	return
}

func GenApproveKey(ccntmract, from, to common.Address) []byte {
	temp := append(ccntmract[:], from[:]...)
	return append(temp, to[:]...)
}

func TransferedFrom(native *native.NativeService, currentCcntmract common.Address, state *TransferFromStateV2) (oldFrom cstates.NativeTokenBalance, oldTo cstates.NativeTokenBalance, err error) {
	if native.Time <= config.GetOntHolderUnboundDeadline()+constants.GENESIS_BLOCK_TIMESTAMP {
		if !native.CcntmextRef.CheckWitness(state.Sender) {
			err = errors.NewErr("authentication failed!")
			return
		}
	} else {
		if state.Sender != state.To && !native.CcntmextRef.CheckWitness(state.Sender) {
			err = errors.NewErr("authentication failed!")
			return
		}
	}

	if err = fromApprove(native, genTransferFromKey(currentCcntmract, state.From, state.Sender), state.Value); err != nil {
		return
	}

	oldFrom, err = reduceFromBalance(native, GenBalanceKey(currentCcntmract, state.From), state.Value)
	if err != nil {
		return
	}

	oldTo, err = increaseToBalance(native, GenBalanceKey(currentCcntmract, state.To), state.Value)
	if err != nil {
		return
	}
	return oldFrom, oldTo, nil
}

func getUnboundOffset(native *native.NativeService, ccntmract, address common.Address) (uint32, error) {
	offset, err := utils.GetStorageUInt32(native.CacheDB, genAddressUnboundOffsetKey(ccntmract, address))
	if err != nil {
		return 0, err
	}
	return offset, nil
}

func getGovernanceUnboundOffset(native *native.NativeService, ccntmract common.Address) (uint32, error) {
	offset, err := utils.GetStorageUInt32(native.CacheDB, genGovernanceUnboundOffsetKey(ccntmract))
	if err != nil {
		return 0, err
	}
	return offset, nil
}

func genTransferFromKey(ccntmract common.Address, from, sender common.Address) []byte {
	temp := append(ccntmract[:], from[:]...)
	return append(temp, sender[:]...)
}

func fromApprove(native *native.NativeService, fromApproveKey []byte, value cstates.NativeTokenBalance) error {
	approveValue, err := utils.GetNativeTokenBalance(native.CacheDB, fromApproveKey)
	if err != nil {
		return err
	}
	newApprove, err := approveValue.Sub(value)
	if err != nil {
		return fmt.Errorf("[TransferFrom] approve balance insufficient: %v", err)
	}
	if newApprove.IsZero() {
		native.CacheDB.Delete(fromApproveKey)
	} else {
		native.CacheDB.Put(fromApproveKey, newApprove.MustToStorageItemBytes())
	}
	return nil
}

func reduceFromBalance(native *native.NativeService, fromKey []byte, value cstates.NativeTokenBalance) (cstates.NativeTokenBalance, error) {
	fromBalance, err := utils.GetNativeTokenBalance(native.CacheDB, fromKey)
	if err != nil {
		return cstates.NativeTokenBalance{}, err
	}
	newFromBalance, err := fromBalance.Sub(value)
	if err != nil {
		addr, _ := common.AddressParseFromBytes(fromKey[20:])
		return cstates.NativeTokenBalance{}, fmt.Errorf("[Transfer] balance insufficient. ccntmract:%s, account:%s, fromBalance:%s,value:%s, err: %v",
			native.CcntmextRef.CurrentCcntmext().CcntmractAddress.ToHexString(), addr.ToBase58(), fromBalance.String(), value.String(), err)
	}

	if newFromBalance.IsZero() {
		native.CacheDB.Delete(fromKey)
	} else {
		native.CacheDB.Put(fromKey, newFromBalance.MustToStorageItemBytes())
	}

	return fromBalance, nil
}

func increaseToBalance(native *native.NativeService, toKey []byte, value cstates.NativeTokenBalance) (cstates.NativeTokenBalance, error) {
	toBalance, err := utils.GetNativeTokenBalance(native.CacheDB, toKey)
	if err != nil {
		return cstates.NativeTokenBalance{}, err
	}
	native.CacheDB.Put(toKey, toBalance.Add(value).MustToStorageItemBytes())
	return toBalance, nil
}

func genAddressUnboundOffsetKey(ccntmract, address common.Address) []byte {
	temp := append(ccntmract[:], UNBOUND_TIME_OFFSET_KEY...)
	return append(temp, address[:]...)
}

func genGovernanceUnboundOffsetKey(ccntmract common.Address) []byte {
	temp := append(ccntmract[:], UNBOUND_TIME_OFFSET_KEY...)
	return temp
}
