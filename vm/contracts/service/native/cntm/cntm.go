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
 * along with The cntm.  If not, see <http://www.gnu.org/licenses/>.
 */

package cntm

import (
	"fmt"
	"math/big"

	"github.com/conntectome/cntm/common"
	"github.com/conntectome/cntm/common/constants"
	"github.com/conntectome/cntm/common/log"
	"github.com/conntectome/cntm/errors"
	"github.com/conntectome/cntm/smartcontract/service/native"
	"github.com/conntectome/cntm/smartcontract/service/native/utils"
)

const (
	TRANSFER_FLAG byte = 1
	APPROVE_FLAG  byte = 2
)

func InitCntm() {
	native.Ccntmracts[utils.CntmCcntmractAddress] = RegisterCntmCcntmract
}

func RegisterCntmCcntmract(native *native.NativeService) {
	native.Register(INIT_NAME, CntmInit)
	native.Register(TRANSFER_NAME, CntmTransfer)
	native.Register(APPROVE_NAME, CntmApprove)
	native.Register(TRANSFERFROM_NAME, CntmTransferFrom)
	native.Register(NAME_NAME, CntmName)
	native.Register(SYMBOL_NAME, CntmSymbol)
	native.Register(DECIMALS_NAME, CntmDecimals)
	native.Register(TOTALSUPPLY_NAME, CntmTotalSupply)
	native.Register(BALANCEOF_NAME, CntmBalanceOf)
	native.Register(ALLOWANCE_NAME, CntmAllowance)
}

func CntmInit(native *native.NativeService) ([]byte, error) {
	contract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	amount, err := utils.GetStorageUInt64(native, GenTotalSupplyKey(contract))
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	if amount > 0 {
		return utils.BYTE_FALSE, errors.NewErr("Init cntm has been completed!")
	}

	distribute := make(map[common.Address]uint64)
	source := common.NewZeroCopySource(native.Input)
	buf, _, irregular, eof := source.NextVarBytes()
	if eof {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadVarBytes, contract params deserialize error!")
	}
	if irregular {
		return utils.BYTE_FALSE, common.ErrIrregularData
	}
	input := common.NewZeroCopySource(buf)
	num, err := utils.DecodeVarUint(input)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("read number error:%v", err)
	}
	sum := uint64(0)
	overflow := false
	for i := uint64(0); i < num; i++ {
		addr, err := utils.DecodeAddress(input)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("read address error:%v", err)
		}
		value, err := utils.DecodeVarUint(input)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("read value error:%v", err)
		}
		sum, overflow = common.SafeAdd(sum, value)
		if overflow {
			return utils.BYTE_FALSE, errors.NewErr("wrong config. overflow detected")
		}
		distribute[addr] += value
	}
	if sum != constants.CNTM_TOTAL_SUPPLY {
		return utils.BYTE_FALSE, fmt.Errorf("wrong config. total supply %d != %d", sum, constants.CNTM_TOTAL_SUPPLY)
	}

	for addr, val := range distribute {
		balanceKey := GenBalanceKey(contract, addr)
		item := utils.GenUInt64StorageItem(val)
		native.CacheDB.Put(balanceKey, item.ToArray())
		AddNotifications(native, contract, &State{To: addr, Value: val})
	}
	native.CacheDB.Put(GenTotalSupplyKey(contract), utils.GenUInt64StorageItem(constants.CNTM_TOTAL_SUPPLY).ToArray())

	return utils.BYTE_TRUE, nil
}

func CntmTransfer(native *native.NativeService) ([]byte, error) {
	var transfers Transfers
	source := common.NewZeroCopySource(native.Input)
	if err := transfers.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[Transfer] Transfers deserialize error!")
	}
	contract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	for _, v := range transfers.States {
		if v.Value == 0 {
			ccntminue
		}
		if v.Value > constants.CNTM_TOTAL_SUPPLY {
			return utils.BYTE_FALSE, fmt.Errorf("transfer cntm amount:%d over totalSupply:%d", v.Value, constants.CNTM_TOTAL_SUPPLY)
		}
		fromBalance, toBalance, err := Transfer(native, contract, &v)
		if err != nil {
			return utils.BYTE_FALSE, err
		}

		if err := grantCntg(native, contract, v.From, fromBalance); err != nil {
			return utils.BYTE_FALSE, err
		}

		if err := grantCntg(native, contract, v.To, toBalance); err != nil {
			return utils.BYTE_FALSE, err
		}

		AddNotifications(native, contract, &v)
	}
	return utils.BYTE_TRUE, nil
}

func CntmTransferFrom(native *native.NativeService) ([]byte, error) {
	var state TransferFrom
	source := common.NewZeroCopySource(native.Input)
	if err := state.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[CntmTransferFrom] State deserialize error!")
	}
	if state.Value == 0 {
		return utils.BYTE_FALSE, nil
	}
	if state.Value > constants.CNTM_TOTAL_SUPPLY {
		return utils.BYTE_FALSE, fmt.Errorf("transferFrom cntm amount:%d over totalSupply:%d", state.Value, constants.CNTM_TOTAL_SUPPLY)
	}
	contract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	fromBalance, toBalance, err := TransferedFrom(native, contract, &state)
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	if err := grantCntg(native, contract, state.From, fromBalance); err != nil {
		return utils.BYTE_FALSE, err
	}
	if err := grantCntg(native, contract, state.To, toBalance); err != nil {
		return utils.BYTE_FALSE, err
	}
	AddNotifications(native, contract, &State{From: state.From, To: state.To, Value: state.Value})
	return utils.BYTE_TRUE, nil
}

func CntmApprove(native *native.NativeService) ([]byte, error) {
	var state State
	source := common.NewZeroCopySource(native.Input)
	if err := state.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[CntgApprove] state deserialize error!")
	}
	if state.Value > constants.CNTM_TOTAL_SUPPLY {
		return utils.BYTE_FALSE, fmt.Errorf("approve cntm amount:%d over totalSupply:%d", state.Value, constants.CNTM_TOTAL_SUPPLY)
	}
	if native.CcntmextRef.CheckWitness(state.From) == false {
		return utils.BYTE_FALSE, errors.NewErr("authentication failed!")
	}
	contract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	native.CacheDB.Put(GenApproveKey(contract, state.From, state.To), utils.GenUInt64StorageItem(state.Value).ToArray())
	return utils.BYTE_TRUE, nil
}

func CntmName(native *native.NativeService) ([]byte, error) {
	return []byte(constants.CNTM_NAME), nil
}

func CntmDecimals(native *native.NativeService) ([]byte, error) {
	return common.BigIntToCntmBytes(big.NewInt(int64(constants.CNTM_DECIMALS))), nil
}

func CntmSymbol(native *native.NativeService) ([]byte, error) {
	return []byte(constants.CNTM_SYMBOL), nil
}

func CntmTotalSupply(native *native.NativeService) ([]byte, error) {
	contract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	amount, err := utils.GetStorageUInt64(native, GenTotalSupplyKey(contract))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[CntmTotalSupply] get totalSupply error!")
	}
	return common.BigIntToCntmBytes(big.NewInt(int64(amount))), nil
}

func CntmBalanceOf(native *native.NativeService) ([]byte, error) {
	return GetBalanceValue(native, TRANSFER_FLAG)
}

func CntmAllowance(native *native.NativeService) ([]byte, error) {
	return GetBalanceValue(native, APPROVE_FLAG)
}

func GetBalanceValue(native *native.NativeService, flag byte) ([]byte, error) {
	source := common.NewZeroCopySource(native.Input)
	from, err := utils.DecodeAddress(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[GetBalanceValue] get from address error!")
	}
	contract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	var key []byte
	if flag == APPROVE_FLAG {
		to, err := utils.DecodeAddress(source)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[GetBalanceValue] get from address error!")
		}
		key = GenApproveKey(contract, from, to)
	} else if flag == TRANSFER_FLAG {
		key = GenBalanceKey(contract, from)
	}
	amount, err := utils.GetStorageUInt64(native, key)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[GetBalanceValue] address parse error!")
	}
	return common.BigIntToCntmBytes(big.NewInt(int64(amount))), nil
}

func grantCntg(native *native.NativeService, contract, address common.Address, balance uint64) error {
	startOffset, err := getUnboundOffset(native, contract, address)
	if err != nil {
		return err
	}
	if native.Time <= constants.GENESIS_BLOCK_TIMESTAMP {
		return nil
	}
	endOffset := native.Time - constants.GENESIS_BLOCK_TIMESTAMP
	if endOffset < startOffset {
		if native.PreExec {
			return nil
		}
		errstr := fmt.Sprintf("grant Cntg error: wrong timestamp endOffset: %d < startOffset: %d", endOffset, startOffset)
		log.Error(errstr)
		return errors.NewErr(errstr)
	} else if endOffset == startOffset {
		return nil
	}

	if balance != 0 {
		value := utils.CalcUnbindCntg(balance, startOffset, endOffset)

		args, err := getApproveArgs(native, contract, utils.CntgCcntmractAddress, address, value)
		if err != nil {
			return err
		}

		if _, err := native.NativeCall(utils.CntgCcntmractAddress, "approve", args); err != nil {
			return err
		}
	}

	native.CacheDB.Put(genAddressUnboundOffsetKey(contract, address), utils.GenUInt32StorageItem(endOffset).ToArray())
	return nil
}

func getApproveArgs(native *native.NativeService, contract, ongCcntmract, address common.Address, value uint64) ([]byte, error) {
	bf := common.NewZeroCopySink(nil)
	approve := State{
		From:  contract,
		To:    address,
		Value: value,
	}

	stateValue, err := utils.GetStorageUInt64(native, GenApproveKey(ongCcntmract, approve.From, approve.To))
	if err != nil {
		return nil, err
	}

	approve.Value += stateValue
	approve.Serialization(bf)
	return bf.Bytes(), nil
}
