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
	scommon "github.com/Ontology/core/store/common"
	"github.com/Ontology/errors"
	"github.com/Ontology/core/genesis"
	ctypes "github.com/Ontology/core/types"
	"math/big"
	"github.com/Ontology/smartccntmract/service/native/states"
	cstates "github.com/Ontology/core/states"
	"bytes"
	"github.com/Ontology/account"
	"github.com/Ontology/common"
)

var (
	decrementInterval = uint32(2000000)
	generationAmount = [17]uint32{80, 70, 60, 50, 40, 30, 20, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10}
	gl = uint32(len(generationAmount))
	cntmTotalSupply = big.NewInt(1000000000)
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

	ts := new(big.Int).Div(cntmTotalSupply, big.NewInt(int64(len(booKeepers))))
	for _, v := range booKeepers {
		address := ctypes.AddressFromPubKey(v)
		native.CloneCache.Add(scommon.ST_Storage, append(ccntmract[:], address[:]...), &cstates.StorageItem{Value: ts.Bytes()})
		native.CloneCache.Add(scommon.ST_Storage, getTotalSupplyKey(ccntmract), &cstates.StorageItem{Value: ts.Bytes()})
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
		fromBalance, toBalance, err := transfer(native, ccntmract, v); if err != nil {
			return err
		}

		fromStartHeight, err := getStartHeight(native, ccntmract, v.From); if err != nil {
			return err
		}

		toStartHeight, err := getStartHeight(native, ccntmract, v.From); if err != nil {
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
	native.CloneCache.Add(scommon.ST_Storage, getApproveKey(ccntmract, state), &cstates.StorageItem{Value: state.Value.Bytes()})
	return nil
}

func grantOng(native *NativeService, ccntmract, address common.Address, balance *big.Int, startHeight uint32) error {
	var amount uint32 = 0
	ustart := startHeight / decrementInterval
	if ustart < gl {
		istart := startHeight % decrementInterval
		uend := native.Height / decrementInterval
		iend := native.Height % decrementInterval
		if uend >= gl {
			uend = gl
			iend = 0
		}
		if iend == 0 {
			uend--
			iend = decrementInterval
		}
		for {
			if ustart >= uend {
				break
			}
			amount += (decrementInterval - istart) * generationAmount[ustart]
			ustart++
			istart = 0
		}
		amount += (iend - istart) * generationAmount[ustart]
	}

	args, err := getApproveArgs(native, ccntmract, genesis.OngCcntmractAddress, address, balance, amount); if err != nil {
		return err
	}

	if err := native.AppCall(genesis.OngCcntmractAddress, "approve", args); err != nil {
		return err
	}

	native.CloneCache.Add(scommon.ST_Storage, getAddressHeightKey(ccntmract, address), getHeightStorageItem(native.Height))
	return nil
}

func getApproveArgs(native *NativeService, ccntmract, cntmCcntmract, address common.Address, balance *big.Int, amount uint32) ([]byte, error) {
	bf := new(bytes.Buffer)
	approve := &states.State {
		From: ccntmract,
		To: address,
		Value: new(big.Int).Mul(balance, big.NewInt(int64(amount))),
	}

	stateValue, err := getStorageBigInt(native, getApproveKey(cntmCcntmract, approve)); if err != nil {
		return nil, err
	}

	approve.Value = new(big.Int).Add(approve.Value, stateValue)

	if err := approve.Serialize(bf); err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}

