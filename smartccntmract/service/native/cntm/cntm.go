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
	"math/big"

	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/constants"
	"github.com/cntmio/cntmology/common/log"
	"github.com/cntmio/cntmology/common/serialization"
	scommon "github.com/cntmio/cntmology/core/store/common"
	"github.com/cntmio/cntmology/errors"
	"github.com/cntmio/cntmology/smartccntmract/service/native"
	"github.com/cntmio/cntmology/smartccntmract/service/native/utils"
)

const (
	TRANSFER_FLAG byte = 1
	APPROVE_FLAG  byte = 2
)

func InitOnt() {
	native.Ccntmracts[utils.OntCcntmractAddress] = RegisterOntCcntmract
}

func RegisterOntCcntmract(native *native.NativeService) {
	native.Register(INIT_NAME, OntInit)
	native.Register(TRANSFER_NAME, OntTransfer)
	native.Register(APPROVE_NAME, OntApprove)
	native.Register(TRANSFERFROM_NAME, OntTransferFrom)
	native.Register(NAME_NAME, OntName)
	native.Register(SYMBOL_NAME, OntSymbol)
	native.Register(DECIMALS_NAME, OntDecimals)
	native.Register(TOTALSUPPLY_NAME, OntTotalSupply)
	native.Register(BALANCEOF_NAME, OntBalanceOf)
	native.Register(ALLOWANCE_NAME, OntAllowance)
}

func OntInit(native *native.NativeService) ([]byte, error) {
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	amount, err := utils.GetStorageUInt64(native, GenTotalSupplyKey(ccntmract))
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	if amount > 0 {
		return utils.BYTE_FALSE, errors.NewErr("Init cntm has been completed!")
	}

	distribute := make(map[common.Address]uint64)
	buf, err := serialization.ReadVarBytes(bytes.NewBuffer(native.Input))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadVarBytes, ccntmract params deserialize error!")
	}
	input := bytes.NewBuffer(buf)
	num, err := utils.ReadVarUint(input)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("read number error:%v", err)
	}
	sum := uint64(0)
	overflow := false
	for i := uint64(0); i < num; i++ {
		addr, err := utils.ReadAddress(input)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("read address error:%v", err)
		}
		value, err := utils.ReadVarUint(input)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("read value error:%v", err)
		}
		sum, overflow = common.SafeAdd(sum, value)
		if overflow {
			return utils.BYTE_FALSE, errors.NewErr("wrcntm config. overflow detected")
		}
		distribute[addr] += value
	}
	if sum != constants.cntm_TOTAL_SUPPLY {
		return utils.BYTE_FALSE, fmt.Errorf("wrcntm config. total supply %d != %d", sum, constants.cntm_TOTAL_SUPPLY)
	}

	for addr, val := range distribute {
		balanceKey := GenBalanceKey(ccntmract, addr)
		item := utils.GenUInt64StorageItem(val)
		native.CloneCache.Add(scommon.ST_STORAGE, balanceKey, item)
		AddNotifications(native, ccntmract, &State{To: addr, Value: val})
	}
	native.CloneCache.Add(scommon.ST_STORAGE, GenTotalSupplyKey(ccntmract), utils.GenUInt64StorageItem(constants.cntm_TOTAL_SUPPLY))

	return utils.BYTE_TRUE, nil
}

func OntTransfer(native *native.NativeService) ([]byte, error) {
	transfers := new(Transfers)
	if err := transfers.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[Transfer] Transfers deserialize error!")
	}
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	for _, v := range transfers.States {
		if v.Value == 0 {
			ccntminue
		}
		fromBalance, toBalance, err := Transfer(native, ccntmract, v)
		if err != nil {
			return utils.BYTE_FALSE, err
		}

		if err := grantOng(native, ccntmract, v.From, fromBalance); err != nil {
			return utils.BYTE_FALSE, err
		}

		if err := grantOng(native, ccntmract, v.To, toBalance); err != nil {
			return utils.BYTE_FALSE, err
		}

		AddNotifications(native, ccntmract, v)
	}
	return utils.BYTE_TRUE, nil
}

func OntTransferFrom(native *native.NativeService) ([]byte, error) {
	state := new(TransferFrom)
	if err := state.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[OntTransferFrom] State deserialize error!")
	}
	if state.Value == 0 {
		return utils.BYTE_FALSE, nil
	}
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	fromBalance, toBalance, err := TransferedFrom(native, ccntmract, state)
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	if err := grantOng(native, ccntmract, state.From, fromBalance); err != nil {
		return utils.BYTE_FALSE, err
	}
	if err := grantOng(native, ccntmract, state.To, toBalance); err != nil {
		return utils.BYTE_FALSE, err
	}
	AddNotifications(native, ccntmract, &State{From: state.From, To: state.To, Value: state.Value})
	return utils.BYTE_TRUE, nil
}

func OntApprove(native *native.NativeService) ([]byte, error) {
	state := new(State)
	if err := state.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[OngApprove] state deserialize error!")
	}
	if state.Value == 0 {
		return utils.BYTE_FALSE, nil
	}
	if native.CcntmextRef.CheckWitness(state.From) == false {
		return utils.BYTE_FALSE, errors.NewErr("authentication failed!")
	}
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	native.CloneCache.Add(scommon.ST_STORAGE, GenApproveKey(ccntmract, state.From, state.To), utils.GenUInt64StorageItem(state.Value))
	return utils.BYTE_TRUE, nil
}

func OntName(native *native.NativeService) ([]byte, error) {
	return []byte(constants.cntm_NAME), nil
}

func OntDecimals(native *native.NativeService) ([]byte, error) {
	return big.NewInt(int64(constants.cntm_DECIMALS)).Bytes(), nil
}

func OntSymbol(native *native.NativeService) ([]byte, error) {
	return []byte(constants.cntm_SYMBOL), nil
}

func OntTotalSupply(native *native.NativeService) ([]byte, error) {
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	amount, err := utils.GetStorageUInt64(native, GenTotalSupplyKey(ccntmract))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[OntTotalSupply] get totalSupply error!")
	}
	return big.NewInt(int64(amount)).Bytes(), nil
}

func OntBalanceOf(native *native.NativeService) ([]byte, error) {
	return GetBalanceValue(native, TRANSFER_FLAG)
}

func OntAllowance(native *native.NativeService) ([]byte, error) {
	return GetBalanceValue(native, APPROVE_FLAG)
}

func GetBalanceValue(native *native.NativeService, flag byte) ([]byte, error) {
	var key []byte
	buf := bytes.NewBuffer(native.Input)
	from, err := utils.ReadAddress(buf)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[GetBalanceValue] get from address error!")
	}
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	if flag == APPROVE_FLAG {
		to, err := utils.ReadAddress(buf)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[GetBalanceValue] get from address error!")
		}
		key = GenApproveKey(ccntmract, from, to)
	} else if flag == TRANSFER_FLAG {
		key = GenBalanceKey(ccntmract, from)
	}
	amount, err := utils.GetStorageUInt64(native, key)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[GetBalanceValue] address parse error!")
	}
	return big.NewInt(int64(amount)).Bytes(), nil
}

func grantOng(native *native.NativeService, ccntmract, address common.Address, balance uint64) error {
	startOffset, err := getUnboundOffset(native, ccntmract, address)
	if err != nil {
		return err
	}
	if native.Time <= constants.GENESIS_BLOCK_TIMESTAMP {
		return nil
	}
	endOffset := native.Time - constants.GENESIS_BLOCK_TIMESTAMP
	if endOffset < startOffset {
		errstr := fmt.Sprintf("grant Ong error: wrcntm timestamp endOffset: %d < startOffset: %d", endOffset, startOffset)
		log.Error(errstr)
		return errors.NewErr(errstr)
	} else if endOffset == startOffset {
		return nil
	}

	if balance != 0 {
		value := utils.CalcUnbindOng(balance, startOffset, endOffset)

		args, err := getApproveArgs(native, ccntmract, utils.OngCcntmractAddress, address, value)
		if err != nil {
			return err
		}

		if _, err := native.NativeCall(utils.OngCcntmractAddress, "approve", args); err != nil {
			return err
		}
	}

	native.CloneCache.Add(scommon.ST_STORAGE, genAddressUnboundOffsetKey(ccntmract, address), utils.GenUInt32StorageItem(endOffset))
	return nil
}

func getApproveArgs(native *native.NativeService, ccntmract, cntmCcntmract, address common.Address, value uint64) ([]byte, error) {
	bf := new(bytes.Buffer)
	approve := &State{
		From:  ccntmract,
		To:    address,
		Value: value,
	}

	stateValue, err := utils.GetStorageUInt64(native, GenApproveKey(cntmCcntmract, approve.From, approve.To))
	if err != nil {
		return nil, err
	}

	approve.Value += stateValue

	if err := approve.Serialize(bf); err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}
