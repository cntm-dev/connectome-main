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
	"bytes"
	"fmt"

	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/serialization"
	cstates "github.com/cntmio/cntmology/core/states"
	scommon "github.com/cntmio/cntmology/core/store/common"
	"github.com/cntmio/cntmology/errors"
	"github.com/cntmio/cntmology/smartccntmract/event"
	"github.com/cntmio/cntmology/smartccntmract/service/native"
	"github.com/cntmio/cntmology/smartccntmract/service/native/utils"
)

var (
	ADDRESS_HEIGHT    = []byte("addressHeight")
	TRANSFER_NAME     = "transfer"
	TOTAL_SUPPLY_NAME = []byte("totalSupply")
)

func AddNotifications(native *native.NativeService, ccntmract common.Address, state *State) {
	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			CcntmractAddress: ccntmract,
			States:          []interface{}{TRANSFER_NAME, state.From.ToBase58(), state.To.ToBase58(), state.Value},
		})
}

func IsTransferFromValid(native *native.NativeService, state *TransferFrom) error {
	if native.CcntmextRef.CheckWitness(state.Sender) == false {
		return errors.NewErr("[IsTransferFromValid] Authentication failed!")
	}
	return nil
}

func IsApproveValid(native *native.NativeService, state *State) error {
	if native.CcntmextRef.CheckWitness(state.From) == false {
		return errors.NewErr("[IsApproveValid] Authentication failed!")
	}
	return nil
}

func IsTransferValid(native *native.NativeService, state *State) error {
	if !native.CcntmextRef.CheckWitness(state.From) {
		return errors.NewErr("[IsTransferValid] Authentication failed!")
	}
	return nil
}

func GetToUInt64StorageItem(toBalance, value uint64) *cstates.StorageItem {
	bf := new(bytes.Buffer)
	serialization.WriteUint64(bf, toBalance+value)
	return &cstates.StorageItem{Value: bf.Bytes()}
}

func GetTotalSupplyKey(ccntmract common.Address) []byte {
	return append(ccntmract[:], TOTAL_SUPPLY_NAME...)
}

func GetTransferKey(ccntmract, from common.Address) []byte {
	return append(ccntmract[:], from[:]...)
}

func Transfer(native *native.NativeService, ccntmract common.Address, state *State) (uint64, uint64, error) {
	if err := IsTransferValid(native, state); err != nil {
		return 0, 0, err
	}

	fromBalance, err := fromTransfer(native, GetTransferKey(ccntmract, state.From), state.Value)
	if err != nil {
		return 0, 0, err
	}

	toBalance, err := toTransfer(native, GetTransferKey(ccntmract, state.To), state.Value)
	if err != nil {
		return 0, 0, err
	}
	return fromBalance, toBalance, nil
}

func GetApproveKey(ccntmract, from, to common.Address) []byte {
	temp := append(ccntmract[:], from[:]...)
	return append(temp, to[:]...)
}

func TransferedFrom(native *native.NativeService, currentCcntmract common.Address, state *TransferFrom) error {
	if err := IsTransferFromValid(native, state); err != nil {
		return err
	}

	if err := fromApprove(native, getTransferFromKey(currentCcntmract, state), state.Value); err != nil {
		return err
	}

	if _, err := fromTransfer(native, GetTransferKey(currentCcntmract, state.From), state.Value); err != nil {
		return err
	}

	if _, err := toTransfer(native, GetTransferKey(currentCcntmract, state.To), state.Value); err != nil {
		return err
	}
	return nil
}

func getStartHeight(native *native.NativeService, ccntmract, address common.Address) (uint32, error) {
	startHeight, err := utils.GetStorageUInt32(native, getAddressHeightKey(ccntmract, address))
	if err != nil {
		return 0, err
	}
	return startHeight, nil
}

func getTransferFromKey(ccntmract common.Address, state *TransferFrom) []byte {
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
		native.CloneCache.Delete(scommon.ST_STORAGE, fromApproveKey)
	} else {
		native.CloneCache.Add(scommon.ST_STORAGE, fromApproveKey, utils.GetUInt64StorageItem(approveValue-value))
	}
	return nil
}

func fromTransfer(native *native.NativeService, fromKey []byte, value uint64) (uint64, error) {
	fromBalance, err := utils.GetStorageUInt64(native, fromKey)
	if err != nil {
		return 0, err
	}
	if fromBalance < value {
		return 0, errors.NewErr("[Transfer] balance insufficient!")
	} else if fromBalance == value {
		native.CloneCache.Delete(scommon.ST_STORAGE, fromKey)
	} else {
		native.CloneCache.Add(scommon.ST_STORAGE, fromKey, utils.GetUInt64StorageItem(fromBalance-value))
	}
	return fromBalance, nil
}

func toTransfer(native *native.NativeService, toKey []byte, value uint64) (uint64, error) {
	toBalance, err := utils.GetStorageUInt64(native, toKey)
	if err != nil {
		return 0, err
	}
	native.CloneCache.Add(scommon.ST_STORAGE, toKey, GetToUInt64StorageItem(toBalance, value))
	return toBalance, nil
}

func getAddressHeightKey(ccntmract, address common.Address) []byte {
	temp := append(ADDRESS_HEIGHT, address[:]...)
	return append(ccntmract[:], temp...)
}
