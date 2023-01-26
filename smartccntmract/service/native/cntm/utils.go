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

const (
	UNBOUND_TIME_OFFSET       = "unboundTimeOffset"
	TOTAL_SUPPLY_NAME         = "totalSupply"
	INIT_NAME                 = "init"
	TRANSFER_NAME             = "transfer"
	APPROVE_NAME              = "approve"
	TRANSFERFROM_NAME         = "transferFrom"
	NAME_NAME                 = "name"
	SYMBOL_NAME               = "symbol"
	DECIMALS_NAME             = "decimals"
	TOTALSUPPLY_NAME          = "totalSupply"
	BALANCEOF_NAME            = "balanceOf"
	ALLOWANCE_NAME            = "allowance"
	TOTAL_ALLOWANCE_NAME      = "totalAllowance"
	UNBOUND_cntm_TO_GOVERNANCE = "unboundOngToGovernance"
)

func AddNotifications(native *native.NativeService, ccntmract common.Address, state *State) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			CcntmractAddress: ccntmract,
			States:          []interface{}{TRANSFER_NAME, state.From.ToBase58(), state.To.ToBase58(), state.Value},
		})
}
func GetToUInt64StorageItem(toBalance, value uint64) *cstates.StorageItem {
	sink := common.NewZeroCopySink(nil)
	sink.WriteUint64(toBalance + value)
	return &cstates.StorageItem{Value: sink.Bytes()}
}

func GenTotalSupplyKey(ccntmract common.Address) []byte {
	return append(ccntmract[:], TOTAL_SUPPLY_NAME...)
}

func GenBalanceKey(ccntmract, addr common.Address) []byte {
	return append(ccntmract[:], addr[:]...)
}

func Transfer(native *native.NativeService, ccntmract common.Address, state *State) (uint64, uint64, error) {
	if !native.CcntmextRef.CheckWitness(state.From) {
		return 0, 0, errors.NewErr("authentication failed!")
	}

	fromBalance, err := fromTransfer(native, GenBalanceKey(ccntmract, state.From), state.Value)
	if err != nil {
		return 0, 0, err
	}

	toBalance, err := toTransfer(native, GenBalanceKey(ccntmract, state.To), state.Value)
	if err != nil {
		return 0, 0, err
	}
	return fromBalance, toBalance, nil
}

func GenApproveKey(ccntmract, from, to common.Address) []byte {
	temp := append(ccntmract[:], from[:]...)
	return append(temp, to[:]...)
}

func TransferedFrom(native *native.NativeService, currentCcntmract common.Address, state *TransferFrom) (uint64, uint64, error) {
	if native.Time <= config.GetOntHolderUnboundDeadline()+constants.GENESIS_BLOCK_TIMESTAMP {
		if !native.CcntmextRef.CheckWitness(state.Sender) {
			return 0, 0, errors.NewErr("authentication failed!")
		}
	} else {
		if state.Sender != state.To && !native.CcntmextRef.CheckWitness(state.Sender) {
			return 0, 0, errors.NewErr("authentication failed!")
		}
	}

	if err := fromApprove(native, genTransferFromKey(currentCcntmract, state), state.Value); err != nil {
		return 0, 0, err
	}

	fromBalance, err := fromTransfer(native, GenBalanceKey(currentCcntmract, state.From), state.Value)
	if err != nil {
		return 0, 0, err
	}

	toBalance, err := toTransfer(native, GenBalanceKey(currentCcntmract, state.To), state.Value)
	if err != nil {
		return 0, 0, err
	}
	return fromBalance, toBalance, nil
}

func getUnboundOffset(native *native.NativeService, ccntmract, address common.Address) (uint32, error) {
	offset, err := utils.GetStorageUInt32(native, genAddressUnboundOffsetKey(ccntmract, address))
	if err != nil {
		return 0, err
	}
	return offset, nil
}

func getGovernanceUnboundOffset(native *native.NativeService, ccntmract common.Address) (uint32, error) {
	offset, err := utils.GetStorageUInt32(native, genGovernanceUnboundOffsetKey(ccntmract))
	if err != nil {
		return 0, err
	}
	return offset, nil
}

func genTransferFromKey(ccntmract common.Address, state *TransferFrom) []byte {
	temp := append(ccntmract[:], state.From[:]...)
	return append(temp, state.Sender[:]...)
}

func fromApprove(native *native.NativeService, fromApproveKey []byte, value uint64) error {
	approveValue, err := utils.GetStorageUInt64(native, fromApproveKey)
	if err != nil {
		return err
	}
	if approveValue < value {
		return fmt.Errorf("[TransferFrom] approve balance insufficient! have %d, got %d", approveValue, value)
	} else if approveValue == value {
		native.CacheDB.Delete(fromApproveKey)
	} else {
		native.CacheDB.Put(fromApproveKey, utils.GenUInt64StorageItem(approveValue-value).ToArray())
	}
	return nil
}

func fromTransfer(native *native.NativeService, fromKey []byte, value uint64) (uint64, error) {
	fromBalance, err := utils.GetStorageUInt64(native, fromKey)
	if err != nil {
		return 0, err
	}
	if fromBalance < value {
		addr, _ := common.AddressParseFromBytes(fromKey[20:])
		return 0, fmt.Errorf("[Transfer] balance insufficient. ccntmract:%s, account:%s,balance:%d, transfer amount:%d",
			native.CcntmextRef.CurrentCcntmext().CcntmractAddress.ToHexString(), addr.ToBase58(), fromBalance, value)
	} else if fromBalance == value {
		native.CacheDB.Delete(fromKey)
	} else {
		native.CacheDB.Put(fromKey, utils.GenUInt64StorageItem(fromBalance-value).ToArray())
	}
	return fromBalance, nil
}

func toTransfer(native *native.NativeService, toKey []byte, value uint64) (uint64, error) {
	toBalance, err := utils.GetStorageUInt64(native, toKey)
	if err != nil {
		return 0, err
	}
	native.CacheDB.Put(toKey, GetToUInt64StorageItem(toBalance, value).ToArray())
	return toBalance, nil
}

func genAddressUnboundOffsetKey(ccntmract, address common.Address) []byte {
	temp := append(ccntmract[:], UNBOUND_TIME_OFFSET...)
	return append(temp, address[:]...)
}

func genGovernanceUnboundOffsetKey(ccntmract common.Address) []byte {
	temp := append(ccntmract[:], UNBOUND_TIME_OFFSET...)
	return temp
}
