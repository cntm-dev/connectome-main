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
 * alcntg with The cntm.  If not, see <http://www.gnu.org/licenses/>.
 */

package cntg

import (
	"fmt"
	"math/big"

	"github.com/conntectome/cntm/common"
	"github.com/conntectome/cntm/common/constants"
	"github.com/conntectome/cntm/errors"
	"github.com/conntectome/cntm/smartcontract/service/native"
	"github.com/conntectome/cntm/smartcontract/service/native/cntm"
	"github.com/conntectome/cntm/smartcontract/service/native/utils"
)

func InitCntg() {
	native.Ccntmracts[utils.CntgCcntmractAddress] = RegisterCntgCcntmract
}

func RegisterCntgCcntmract(native *native.NativeService) {
	native.Register(cntm.INIT_NAME, CntgInit)
	native.Register(cntm.TRANSFER_NAME, CntgTransfer)
	native.Register(cntm.APPROVE_NAME, CntgApprove)
	native.Register(cntm.TRANSFERFROM_NAME, CntgTransferFrom)
	native.Register(cntm.NAME_NAME, CntgName)
	native.Register(cntm.SYMBOL_NAME, CntgSymbol)
	native.Register(cntm.DECIMALS_NAME, CntgDecimals)
	native.Register(cntm.TOTALSUPPLY_NAME, CntgTotalSupply)
	native.Register(cntm.BALANCEOF_NAME, CntgBalanceOf)
	native.Register(cntm.ALLOWANCE_NAME, CntgAllowance)
}

func CntgInit(native *native.NativeService) ([]byte, error) {
	contract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	amount, err := utils.GetStorageUInt64(native, cntm.GenTotalSupplyKey(contract))
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	if amount > 0 {
		return utils.BYTE_FALSE, errors.NewErr("Init cntg has been completed!")
	}

	item := utils.GenUInt64StorageItem(constants.CNTG_TOTAL_SUPPLY)
	native.CacheDB.Put(cntm.GenTotalSupplyKey(contract), item.ToArray())
	native.CacheDB.Put(append(contract[:], utils.CntmCcntmractAddress[:]...), item.ToArray())
	cntm.AddNotifications(native, contract, &cntm.State{To: utils.CntmCcntmractAddress, Value: constants.CNTG_TOTAL_SUPPLY})
	return utils.BYTE_TRUE, nil
}

func CntgTransfer(native *native.NativeService) ([]byte, error) {
	var transfers cntm.Transfers
	source := common.NewZeroCopySource(native.Input)
	if err := transfers.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[CntgTransfer] Transfers deserialize error!")
	}
	contract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	for _, v := range transfers.States {
		if v.Value == 0 {
			ccntminue
		}
		if v.Value > constants.CNTG_TOTAL_SUPPLY {
			return utils.BYTE_FALSE, fmt.Errorf("transfer cntg amount:%d over totalSupply:%d", v.Value, constants.CNTG_TOTAL_SUPPLY)
		}
		if _, _, err := cntm.Transfer(native, contract, &v); err != nil {
			return utils.BYTE_FALSE, err
		}
		cntm.AddNotifications(native, contract, &v)
	}
	return utils.BYTE_TRUE, nil
}

func CntgApprove(native *native.NativeService) ([]byte, error) {
	var state cntm.State
	source := common.NewZeroCopySource(native.Input)
	if err := state.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[CntgApprove] state deserialize error!")
	}
	if state.Value > constants.CNTG_TOTAL_SUPPLY {
		return utils.BYTE_FALSE, fmt.Errorf("approve cntg amount:%d over totalSupply:%d", state.Value, constants.CNTG_TOTAL_SUPPLY)
	}
	if native.CcntmextRef.CheckWitness(state.From) == false {
		return utils.BYTE_FALSE, errors.NewErr("authentication failed!")
	}
	contract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	native.CacheDB.Put(cntm.GenApproveKey(contract, state.From, state.To), utils.GenUInt64StorageItem(state.Value).ToArray())
	return utils.BYTE_TRUE, nil
}

func CntgTransferFrom(native *native.NativeService) ([]byte, error) {
	var state cntm.TransferFrom
	source := common.NewZeroCopySource(native.Input)
	if err := state.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[CntmTransferFrom] State deserialize error!")
	}
	if state.Value == 0 {
		return utils.BYTE_FALSE, nil
	}
	if state.Value > constants.CNTG_TOTAL_SUPPLY {
		return utils.BYTE_FALSE, fmt.Errorf("approve cntg amount:%d over totalSupply:%d", state.Value, constants.CNTG_TOTAL_SUPPLY)
	}
	contract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	if _, _, err := cntm.TransferedFrom(native, contract, &state); err != nil {
		return utils.BYTE_FALSE, err
	}
	cntm.AddNotifications(native, contract, &cntm.State{From: state.From, To: state.To, Value: state.Value})
	return utils.BYTE_TRUE, nil
}

func CntgName(native *native.NativeService) ([]byte, error) {
	return []byte(constants.CNTG_NAME), nil
}

func CntgDecimals(native *native.NativeService) ([]byte, error) {
	return big.NewInt(int64(constants.CNTG_DECIMALS)).Bytes(), nil
}

func CntgSymbol(native *native.NativeService) ([]byte, error) {
	return []byte(constants.CNTG_SYMBOL), nil
}

func CntgTotalSupply(native *native.NativeService) ([]byte, error) {
	contract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	amount, err := utils.GetStorageUInt64(native, cntm.GenTotalSupplyKey(contract))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[CntmTotalSupply] get totalSupply error!")
	}
	return common.BigIntToCntmBytes(big.NewInt(int64(amount))), nil
}

func CntgBalanceOf(native *native.NativeService) ([]byte, error) {
	return cntm.GetBalanceValue(native, cntm.TRANSFER_FLAG)
}

func CntgAllowance(native *native.NativeService) ([]byte, error) {
	return cntm.GetBalanceValue(native, cntm.APPROVE_FLAG)
}
