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
	"math/big"

	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/config"
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
	native.Register(NAME_NAME, OntName)
	native.Register(SYMBOL_NAME, OntSymbol)
	native.Register(TRANSFER_NAME, OntTransfer)
	native.Register(APPROVE_NAME, OntApprove)
	native.Register(TRANSFERFROM_NAME, OntTransferFrom)
	native.Register(DECIMALS_NAME, OntDecimals)
	native.Register(TOTAL_SUPPLY_NAME, OntTotalSupply)
	native.Register(BALANCEOF_NAME, OntBalanceOf)
	native.Register(ALLOWANCE_NAME, OntAllowance)
	native.Register(TOTAL_ALLOWANCE_NAME, TotalAllowance)

	if native.Height >= config.GetAddDecimalsHeight() || native.PreExec {
		native.Register(BALANCEOF_V2_NAME, OntBalanceOfV2)
		native.Register(ALLOWANCE_V2_NAME, OntAllowanceV2)
		native.Register(TOTAL_ALLOWANCE_V2_NAME, TotalAllowanceV2)
	}

	if native.Height >= config.GetAddDecimalsHeight() {
		native.Register(TRANSFER_V2_NAME, OntTransferV2)
		native.Register(APPROVE_V2_NAME, OntApproveV2)
		native.Register(TRANSFERFROM_V2_NAME, OntTransferFromV2)
		native.Register(DECIMALS_V2_NAME, OntDecimalsV2)
		native.Register(TOTAL_SUPPLY_V2_NAME, OntTotalSupplyV2)
	}

	native.Register(UNBOUND_cntm_TO_GOVERNANCE, UnboundOngToGovernance)
}

func OntInit(native *native.NativeService) ([]byte, error) {
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	amount, err := utils.GetNativeTokenBalance(native.CacheDB, GenTotalSupplyKey(ccntmract))
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	if amount.IsZero() == false {
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
		native.CacheDB.Put(balanceKey, item.ToArray())
		AddNotifications(native, ccntmract, &State{To: addr, Value: val})
	}
	native.CacheDB.Put(GenTotalSupplyKey(ccntmract), utils.GenUInt64StorageItem(constants.cntm_TOTAL_SUPPLY).ToArray())

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
		if v.Value > constants.cntm_TOTAL_SUPPLY {
			return utils.BYTE_FALSE, fmt.Errorf("transfer cntm amount:%d over totalSupply:%d", v.Value, constants.cntm_TOTAL_SUPPLY)
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
	if state.Value > constants.cntm_TOTAL_SUPPLY {
		return utils.BYTE_FALSE, fmt.Errorf("transferFrom cntm amount:%d over totalSupply:%d", state.Value, constants.cntm_TOTAL_SUPPLY)
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
	if state.Value > constants.cntm_TOTAL_SUPPLY {
		return utils.BYTE_FALSE, fmt.Errorf("approve cntm amount:%d over totalSupply:%d", state.Value, constants.cntm_TOTAL_SUPPLY)
	}
	if !native.CcntmextRef.CheckWitness(state.From) {
		return utils.BYTE_FALSE, errors.NewErr("authentication failed!")
	}
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	native.CacheDB.Put(GenApproveKey(ccntmract, state.From, state.To), utils.GenUInt64StorageItem(state.Value).ToArray())
	return utils.BYTE_TRUE, nil
}

func OntApproveV2(native *native.NativeService) ([]byte, error) {
	var state TransferStateV2
	source := common.NewZeroCopySource(native.Input)
	if err := state.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[OntApprove] state deserialize error!")
	}
	if bigint.New(constants.cntm_TOTAL_SUPPLY_V2).LessThan(state.Value.Balance) {
		return utils.BYTE_FALSE, fmt.Errorf("approve cntm amount:%s over totalSupply:%d", state.Value, constants.cntm_TOTAL_SUPPLY)
	}
	if !native.CcntmextRef.CheckWitness(state.From) {
		return utils.BYTE_FALSE, errors.NewErr("authentication failed!")
	}
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	native.CacheDB.Put(GenApproveKey(ccntmract, state.From, state.To), state.Value.MustToStorageItemBytes())
	return utils.BYTE_TRUE, nil
}

func OntName(native *native.NativeService) ([]byte, error) {
	return []byte(constants.cntm_NAME), nil
}

func OntDecimals(native *native.NativeService) ([]byte, error) {
	return common.BigIntToNeoBytes(big.NewInt(int64(constants.cntm_DECIMALS))), nil
}

func OntSymbol(native *native.NativeService) ([]byte, error) {
	return []byte(constants.cntm_SYMBOL), nil
}

func OntTotalSupply(native *native.NativeService) ([]byte, error) {
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	amount, err := utils.GetStorageUInt64(native.CacheDB, GenTotalSupplyKey(ccntmract))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[OntTotalSupply] get totalSupply error!")
	}
	return common.BigIntToNeoBytes(big.NewInt(int64(amount))), nil
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
	balance, err := utils.GetNativeTokenBalance(native.CacheDB, key)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[GetBalanceValue] address parse error!")
	}
	amount := balance.ToBigInt()
	if !scaleDecimal9 {
		amount = balance.ToInteger().BigInt()
	}
	return common.BigIntToNeoBytes(amount), nil
}

func getTotalAllowance(native *native.NativeService, from common.Address) (result cstates.NativeTokenBalance, err error) {
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	iter := native.CacheDB.NewIterator(utils.ConcatKey(ccntmract, from[:]))
	defer iter.Release()
	r := cstates.NativeTokenBalanceFromInteger(0)
	for has := iter.First(); has; has = iter.Next() {
		if bytes.Equal(iter.Key(), utils.ConcatKey(ccntmract, from[:])) {
			ccntminue
		}
		item := new(cstates.StorageItem)
		err = item.Deserialization(common.NewZeroCopySource(iter.Value()))
		if err != nil {
			return result, errors.NewDetailErr(err, errors.ErrNoCode, "[TotalAllowance] instance isn't StorageItem!")
		}
		balance, err := cstates.NativeTokenBalanceFromStorageItem(item)
		if err != nil {
			return result, errors.NewDetailErr(err, errors.ErrNoCode, "[TotalAllowance] get token allowance from storage value error!")
		}
		r = r.Add(balance)
	}

	return r, nil
}

func TotalAllowance(native *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(native.Input)
	from, err := utils.DecodeAddress(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[TotalAllowance] get from address error!")
	}
	r, err := getTotalAllowance(native, from)
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	return common.BigIntToNeoBytes(r.ToInteger().BigInt()), nil
}

func TotalAllowanceV2(native *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(native.Input)
	from, err := utils.DecodeAddress(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[TotalAllowance] get from address error!")
	}
	r, err := getTotalAllowance(native, from)
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	return common.BigIntToNeoBytes(r.ToBigInt()), nil
}

func UnboundOngToGovernance(native *native.NativeService) ([]byte, error) {
	err := unboundOngToGovernance(native)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("unboundOngToGovernance error: %s", err)
	}
	return utils.BYTE_TRUE, nil
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
		if native.PreExec {
			return nil
		}
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

	native.CacheDB.Put(genAddressUnboundOffsetKey(ccntmract, address), utils.GenUInt32StorageItem(endOffset).ToArray())
	return nil
}

func getApproveArgs(native *native.NativeService, ccntmract, cntmCcntmract, address common.Address, value uint64) ([]byte, error) {
	bf := new(bytes.Buffer)
	approve := &State{
		From:  ccntmract,
		To:    address,
		Value: value,
	}

	stateValue, err := utils.GetStorageUInt64(native.CacheDB, GenApproveKey(cntmCcntmract, approve.From, approve.To))
	if err != nil {
		return nil, 0, err
	}

	approve.Value += stateValue
	approve.Serialization(bf)
	return bf.Bytes(), approve.Value, nil
}

func getTransferArgs(ccntmract, address common.Address, value uint64) ([]byte, error) {
	bf := common.NewZeroCopySink(nil)
	state := State{
		From:  ccntmract,
		To:    address,
		Value: value,
	}
	transfers := Transfers{[]State{state}}

	transfers.Serialization(bf)
	return bf.Bytes(), nil
}

func getTransferFromArgs(sender, from, to common.Address, value uint64) ([]byte, error) {
	sink := common.NewZeroCopySink(nil)
	param := TransferFrom{
		Sender: sender,
		From:   from,
		To:     to,
		Value:  value,
	}

	param.Serialization(sink)
	return sink.Bytes(), nil
}
