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

package native

import (
	"fmt"
	"math/big"

	"github.com/Ontology/common"
	cstates "github.com/Ontology/core/states"
	scommon "github.com/Ontology/core/store/common"
	"github.com/Ontology/errors"
	"github.com/Ontology/smartccntmract/event"
	"github.com/Ontology/smartccntmract/service/native/states"
)

var (
	ADDRESS_HEIGHT    = []byte("addressHeight")
	TRANSFER_NAME     = "transfer"
	TOTAL_SUPPLY_NAME = []byte("totalSupply")
)

func getAddressHeightKey(ccntmract, address common.Address) []byte {
	temp := append(ADDRESS_HEIGHT, address[:]...)
	return append(ccntmract[:], temp...)
}

func getHeightStorageItem(height uint32) *cstates.StorageItem {
	return &cstates.StorageItem{Value: big.NewInt(int64(height)).Bytes()}
}

func getAmountStorageItem(value *big.Int) *cstates.StorageItem {
	return &cstates.StorageItem{Value: value.Bytes()}
}

func getToAmountStorageItem(toBalance, value *big.Int) *cstates.StorageItem {
	return &cstates.StorageItem{Value: new(big.Int).Add(toBalance, value).Bytes()}
}

func getTotalSupplyKey(ccntmract common.Address) []byte {
	return append(ccntmract[:], TOTAL_SUPPLY_NAME...)
}

func getTransferKey(ccntmract, from common.Address) []byte {
	return append(ccntmract[:], from[:]...)
}

func getApproveKey(ccntmract common.Address, state *states.State) []byte {
	temp := append(ccntmract[:], state.From[:]...)
	return append(temp, state.To[:]...)
}

func getTransferFromKey(ccntmract common.Address, state *states.TransferFrom) []byte {
	temp := append(ccntmract[:], state.From[:]...)
	return append(temp, state.Sender[:]...)
}

func isTransferValid(native *NativeService, state *states.State) error {
	if state.Value.Sign() < 0 {
		return errors.NewErr("Transfer amount invalid!")
	}

	if native.CcntmextRef.CheckWitness(state.From) == false {
		return errors.NewErr("[Sender] Authentication failed!")
	}
	return nil
}

func transfer(native *NativeService, ccntmract common.Address, state *states.State) (*big.Int, *big.Int, error) {
	if err := isTransferValid(native, state); err != nil {
		return nil, nil, err
	}

	fromBalance, err := fromTransfer(native, getTransferKey(ccntmract, state.From), state.Value)
	if err != nil {
		return nil, nil, err
	}

	toBalance, err := toTransfer(native, getTransferKey(ccntmract, state.To), state.Value)
	if err != nil {
		return nil, nil, err
	}
	return fromBalance, toBalance, nil
}

func transferFrom(native *NativeService, currentCcntmract common.Address, state *states.TransferFrom) error {
	if err := isTransferFromValid(native, state); err != nil {
		return err
	}

	if err := fromApprove(native, getTransferFromKey(currentCcntmract, state), state.Value); err != nil {
		return err
	}

	if _, err := fromTransfer(native, getTransferKey(currentCcntmract, state.From), state.Value); err != nil {
		return err
	}

	if _, err := toTransfer(native, getTransferKey(currentCcntmract, state.To), state.Value); err != nil {
		return err
	}
	return nil
}

func isTransferFromValid(native *NativeService, state *states.TransferFrom) error {
	if state.Value.Sign() < 0 {
		return errors.NewErr("TransferFrom amount invalid!")
	}

	if native.CcntmextRef.CheckWitness(state.Sender) == false {
		return errors.NewErr("[Sender] Authentication failed!")
	}
	return nil
}

func isApproveValid(native *NativeService, state *states.State) error {
	if state.Value.Sign() < 0 {
		return errors.NewErr("Approve amount invalid!")
	}
	if native.CcntmextRef.CheckWitness(state.From) == false {
		return errors.NewErr("[Sender] Authentication failed!")
	}
	return nil
}

func fromApprove(native *NativeService, fromApproveKey []byte, value *big.Int) error {
	approveValue, err := getStorageBigInt(native, fromApproveKey)
	if err != nil {
		return err
	}
	approveBalance := new(big.Int).Sub(approveValue, value)
	sign := approveBalance.Sign()
	if sign < 0 {
		return fmt.Errorf("[TransferFrom] approve balance insufficient! have %d, got %d", approveValue.Int64(), value.Int64())
	} else if sign == 0 {
		native.CloneCache.Delete(scommon.ST_STORAGE, fromApproveKey)
	} else {
		native.CloneCache.Add(scommon.ST_STORAGE, fromApproveKey, getAmountStorageItem(approveBalance))
	}
	return nil
}

func fromTransfer(native *NativeService, fromKey []byte, value *big.Int) (*big.Int, error) {
	fromBalance, err := getStorageBigInt(native, fromKey)
	if err != nil {
		return nil, err
	}
	balance := new(big.Int).Sub(fromBalance, value)
	sign := balance.Sign()
	if sign < 0 {
		return nil, errors.NewErr("[Transfer] balance insufficient!")
	} else if sign == 0 {
		native.CloneCache.Delete(scommon.ST_STORAGE, fromKey)
	} else {
		native.CloneCache.Add(scommon.ST_STORAGE, fromKey, getAmountStorageItem(balance))
	}
	return fromBalance, nil
}

func toTransfer(native *NativeService, toKey []byte, value *big.Int) (*big.Int, error) {
	toBalance, err := getStorageBigInt(native, toKey)
	if err != nil {
		return nil, err
	}
	native.CloneCache.Add(scommon.ST_STORAGE, toKey, getToAmountStorageItem(toBalance, value))
	return toBalance, nil
}

func getStartHeight(native *NativeService, ccntmract, from common.Address) (uint32, error) {
	startHeight, err := getStorageBigInt(native, getAddressHeightKey(ccntmract, from))
	if err != nil {
		return 0, err
	}
	return uint32(startHeight.Int64()), nil
}

func getStorageBigInt(native *NativeService, key []byte) (*big.Int, error) {
	balance, err := native.CloneCache.Get(scommon.ST_STORAGE, key)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[getBalance] storage error!")
	}
	if balance == nil {
		return big.NewInt(0), nil
	}
	item, ok := balance.(*cstates.StorageItem)
	if !ok {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[getBalance] get amount error!")
	}
	return new(big.Int).SetBytes(item.Value), nil
}

func addNotifications(native *NativeService, ccntmract common.Address, state *states.State) {
	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			TxHash:   native.Tx.Hash(),
			CodeHash: ccntmract,
			States:   []interface{}{TRANSFER_NAME, state.From, state.To, state.Value},
		})
}
