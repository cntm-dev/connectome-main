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
	"bytes"
	"math/big"

	"github.com/cntmio/cntmology/account"
	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/core/genesis"
	cstates "github.com/cntmio/cntmology/core/states"
	scommon "github.com/cntmio/cntmology/core/store/common"
	ctypes "github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/errors"
	"github.com/cntmio/cntmology/smartccntmract/service/native/states"
)

var (
	DECREMENT_INTERVAL = uint32(2000000)
	GENERATION_AMOUNT  = [17]uint32{80, 70, 60, 50, 40, 30, 20, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10}
	GL                 = uint32(len(GENERATION_AMOUNT))
	cntm_TOTAL_SUPPLY   = big.NewInt(1000000000)
)

func OntInit(native *NativeService) error {
	booKeepers := account.GetBookkeepers()

	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	amount, err := getStorageBigInt(native, getTotalSupplyKey(ccntmract))
	if err != nil {
		return err
	}

	if amount != nil && amount.Sign() != 0 {
		return errors.NewErr("Init cntm has been completed!")
	}

	ts := new(big.Int).Div(cntm_TOTAL_SUPPLY, big.NewInt(int64(len(booKeepers))))
	for _, v := range booKeepers {
		address := ctypes.AddressFromPubKey(v)
		native.CloneCache.Add(scommon.ST_STORAGE, append(ccntmract[:], address[:]...), &cstates.StorageItem{Value: ts.Bytes()})
		native.CloneCache.Add(scommon.ST_STORAGE, getTotalSupplyKey(ccntmract), &cstates.StorageItem{Value: ts.Bytes()})
		addNotifications(native, ccntmract, &states.State{To: address, Value: ts})
	}

	return nil
}

func OntTransfer(native *NativeService) error {
	transfers := new(states.Transfers)
	if err := transfers.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Transfer] Transfers deserialize error!")
	}
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	for _, v := range transfers.States {
		fromBalance, toBalance, err := transfer(native, ccntmract, v)
		if err != nil {
			return err
		}

		fromStartHeight, err := getStartHeight(native, ccntmract, v.From)
		if err != nil {
			return err
		}

		toStartHeight, err := getStartHeight(native, ccntmract, v.From)
		if err != nil {
			return err
		}

		if err := grantOng(native, ccntmract, v.From, fromBalance, fromStartHeight); err != nil {
			return err
		}

		if err := grantOng(native, ccntmract, v.To, toBalance, toStartHeight); err != nil {
			return err
		}

		addNotifications(native, ccntmract, v)
	}
	return nil
}

func OntTransferFrom(native *NativeService) error {
	state := new(states.TransferFrom)
	if err := state.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[OntTransferFrom] State deserialize error!")
	}
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	if err := transferFrom(native, ccntmract, state); err != nil {
		return err
	}
	addNotifications(native, ccntmract, &states.State{From: state.From, To: state.To, Value: state.Value})
	return nil
}

func OntApprove(native *NativeService) error {
	state := new(states.State)
	if err := state.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[OngApprove] state deserialize error!")
	}
	if err := isApproveValid(native, state); err != nil {
		return err
	}
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	native.CloneCache.Add(scommon.ST_STORAGE, getApproveKey(ccntmract, state), &cstates.StorageItem{Value: state.Value.Bytes()})
	return nil
}

func grantOng(native *NativeService, ccntmract, address common.Address, balance *big.Int, startHeight uint32) error {
	var amount uint32 = 0
	ustart := startHeight / DECREMENT_INTERVAL
	if ustart < GL {
		istart := startHeight % DECREMENT_INTERVAL
		uend := native.Height / DECREMENT_INTERVAL
		iend := native.Height % DECREMENT_INTERVAL
		if uend >= GL {
			uend = GL
			iend = 0
		}
		if iend == 0 {
			uend--
			iend = DECREMENT_INTERVAL
		}
		for {
			if ustart >= uend {
				break
			}
			amount += (DECREMENT_INTERVAL - istart) * GENERATION_AMOUNT[ustart]
			ustart++
			istart = 0
		}
		amount += (iend - istart) * GENERATION_AMOUNT[ustart]
	}

	args, err := getApproveArgs(native, ccntmract, genesis.OngCcntmractAddress, address, balance, amount)
	if err != nil {
		return err
	}

	if err := native.AppCall(genesis.OngCcntmractAddress, "approve", args); err != nil {
		return err
	}

	native.CloneCache.Add(scommon.ST_STORAGE, getAddressHeightKey(ccntmract, address), getHeightStorageItem(native.Height))
	return nil
}

func getApproveArgs(native *NativeService, ccntmract, cntmCcntmract, address common.Address, balance *big.Int, amount uint32) ([]byte, error) {
	bf := new(bytes.Buffer)
	approve := &states.State{
		From:  ccntmract,
		To:    address,
		Value: new(big.Int).Mul(balance, big.NewInt(int64(amount))),
	}

	stateValue, err := getStorageBigInt(native, getApproveKey(cntmCcntmract, approve))
	if err != nil {
		return nil, err
	}

	approve.Value = new(big.Int).Add(approve.Value, stateValue)

	if err := approve.Serialize(bf); err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}
