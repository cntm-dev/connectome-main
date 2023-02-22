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
package testsuite

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/laizy/bigint"
	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/config"
	"github.com/cntmio/cntmology/common/constants"
	"github.com/cntmio/cntmology/core/states"
	"github.com/cntmio/cntmology/smartccntmract/service/native"
	_ "github.com/cntmio/cntmology/smartccntmract/service/native/init"
	"github.com/cntmio/cntmology/smartccntmract/service/native/cntm"
	"github.com/cntmio/cntmology/smartccntmract/service/native/cntm"
	"github.com/cntmio/cntmology/smartccntmract/service/native/utils"
	"github.com/cntmio/cntmology/smartccntmract/storage"
	"github.com/stretchr/testify/assert"
)

func setOntBalance(db *storage.CacheDB, addr common.Address, value uint64) {
	balanceKey := cntm.GenBalanceKey(utils.OntCcntmractAddress, addr)
	item := utils.GenUInt64StorageItem(value)
	db.Put(balanceKey, item.ToArray())
}

func cntmBalanceOf(native *native.NativeService, addr common.Address) int {
	sink := common.NewZeroCopySink(nil)
	utils.EncodeAddress(sink, addr)
	native.Input = sink.Bytes()
	buf, _ := cntm.OntBalanceOf(native)
	val := common.BigIntFromNeoBytes(buf)
	return int(val.Uint64())
}
func cntmBalanceOfV2(native *native.NativeService, addr common.Address) uint64 {
	sink := common.NewZeroCopySink(nil)
	utils.EncodeAddress(sink, addr)
	native.Input = sink.Bytes()
	buf, _ := cntm.OntBalanceOfV2(native)
	val := common.BigIntFromNeoBytes(buf)
	return val.Uint64()
}

func setOngBalance(db *storage.CacheDB, addr common.Address, value uint64) {
	balanceKey := cntm.GenBalanceKey(utils.OngCcntmractAddress, addr)
	item := utils.GenUInt64StorageItem(value)
	db.Put(balanceKey, item.ToArray())
}

func cntmBalanceOf(native *native.NativeService, addr common.Address) uint64 {
	origin := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	native.CcntmextRef.CurrentCcntmext().CcntmractAddress = utils.OngCcntmractAddress
	defer func() {
		native.CcntmextRef.CurrentCcntmext().CcntmractAddress = origin
	}()
	sink := common.NewZeroCopySink(nil)
	utils.EncodeAddress(sink, addr)
	native.Input = sink.Bytes()
	buf, _ := cntm.OngBalanceOf(native)
	val := common.BigIntFromNeoBytes(buf)
	return val.Uint64()
}

func cntmBalanceOfV2(native *native.NativeService, addr common.Address) bigint.Int {
	origin := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	native.CcntmextRef.CurrentCcntmext().CcntmractAddress = utils.OngCcntmractAddress
	defer func() {
		native.CcntmextRef.CurrentCcntmext().CcntmractAddress = origin
	}()
	sink := common.NewZeroCopySink(nil)
	utils.EncodeAddress(sink, addr)
	native.Input = sink.Bytes()
	buf, _ := cntm.OngBalanceOfV2(native)
	val := common.BigIntFromNeoBytes(buf)
	return bigint.New(val)
}

func cntmAllowance(native *native.NativeService, from, to common.Address) uint64 {
	origin := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	native.CcntmextRef.CurrentCcntmext().CcntmractAddress = utils.OngCcntmractAddress
	defer func() {
		native.CcntmextRef.CurrentCcntmext().CcntmractAddress = origin
	}()
	sink := common.NewZeroCopySink(nil)
	utils.EncodeAddress(sink, from)
	utils.EncodeAddress(sink, to)
	native.Input = sink.Bytes()
	buf, _ := cntm.OngAllowance(native)
	val := common.BigIntFromNeoBytes(buf)
	return val.Uint64()
}

func cntmAllowanceV2(native *native.NativeService, from, to common.Address) bigint.Int {
	origin := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	native.CcntmextRef.CurrentCcntmext().CcntmractAddress = utils.OngCcntmractAddress
	defer func() {
		native.CcntmextRef.CurrentCcntmext().CcntmractAddress = origin
	}()
	sink := common.NewZeroCopySink(nil)
	utils.EncodeAddress(sink, from)
	utils.EncodeAddress(sink, to)
	native.Input = sink.Bytes()
	buf, _ := cntm.OngAllowanceV2(native)
	val := common.BigIntFromNeoBytes(buf)
	return bigint.New(val)
}

func cntmTransferFromV2(native *native.NativeService, spender, from, to common.Address, amt states.NativeTokenBalance) error {
	origin := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	native.CcntmextRef.CurrentCcntmext().CcntmractAddress = utils.OngCcntmractAddress
	defer func() {
		native.CcntmextRef.CurrentCcntmext().CcntmractAddress = origin
	}()
	native.Tx.SignedAddr = append(native.Tx.SignedAddr, from)
	state := &cntm.TransferFromStateV2{Sender: spender, TransferStateV2: cntm.TransferStateV2{From: from, To: to, Value: amt}}
	native.Input = common.SerializeToBytes(state)

	_, err := cntm.OngTransferFromV2(native)
	return err
}

func cntmTotalAllowance(native *native.NativeService, addr common.Address) int {
	sink := common.NewZeroCopySink(nil)
	utils.EncodeAddress(sink, addr)
	native.Input = sink.Bytes()
	buf, _ := cntm.TotalAllowance(native)
	val := common.BigIntFromNeoBytes(buf)
	return int(val.Uint64())
}

func cntmTotalAllowanceV2(native *native.NativeService, addr common.Address) uint64 {
	sink := common.NewZeroCopySink(nil)
	utils.EncodeAddress(sink, addr)
	native.Input = sink.Bytes()
	buf, _ := cntm.TotalAllowanceV2(native)
	val := common.BigIntFromNeoBytes(buf)
	return val.Uint64()
}

func cntmTransfer(native *native.NativeService, from, to common.Address, value uint64) error {
	native.Tx.SignedAddr = append(native.Tx.SignedAddr, from)
	state := cntm.TransferState{from, to, value}
	native.Input = common.SerializeToBytes(&cntm.TransferStates{States: []cntm.TransferState{state}})
	_, err := cntm.OntTransfer(native)
	return err
}

func cntmTransferV2(native *native.NativeService, from, to common.Address, value uint64) error {
	native.Tx.SignedAddr = append(native.Tx.SignedAddr, from)
	state := &cntm.TransferStateV2{from, to, states.NativeTokenBalance{Balance: bigint.New(value)}}
	native.Input = common.SerializeToBytes(&cntm.TransferStatesV2{States: []*cntm.TransferStateV2{state}})
	_, err := cntm.OntTransferV2(native)
	return err
}

func cntmTransferFrom(native *native.NativeService, sender, from, to common.Address, value uint64) error {
	native.Tx.SignedAddr = append(native.Tx.SignedAddr, from)
	state := &cntm.TransferFrom{sender, cntm.TransferState{from, to, value}}
	native.Input = common.SerializeToBytes(state)
	_, err := cntm.OntTransferFrom(native)
	return err
}

func cntmTransferFromV2(native *native.NativeService, sender, from, to common.Address, value uint64) error {
	native.Tx.SignedAddr = append(native.Tx.SignedAddr, from)
	state := &cntm.TransferFromStateV2{sender, cntm.TransferStateV2{from, to, states.NativeTokenBalance{Balance: bigint.New(value)}}}
	native.Input = common.SerializeToBytes(state)
	_, err := cntm.OntTransferFromV2(native)
	return err
}

func cntmApprove(native *native.NativeService, from, to common.Address, value uint64) error {
	native.Tx.SignedAddr = append(native.Tx.SignedAddr, from)

	native.Input = common.SerializeToBytes(&cntm.TransferState{from, to, value})
	_, err := cntm.OntApprove(native)
	return err
}

func cntmApproveV2(native *native.NativeService, from, to common.Address, value uint64) error {
	native.Tx.SignedAddr = append(native.Tx.SignedAddr, from)

	native.Input = common.SerializeToBytes((&cntm.TransferStateV2{from, to, states.NativeTokenBalance{Balance: bigint.New(value)}}))
	_, err := cntm.OntApproveV2(native)
	return err
}

func unboundGovernanceOng(native *native.NativeService) error {
	_, err := cntm.UnboundOngToGovernance(native)
	return err
}

func TestTransfer(t *testing.T) {
	InvokeNativeCcntmract(t, utils.OntCcntmractAddress, func(native *native.NativeService) ([]byte, error) {
		a, b, c := RandomAddress(), RandomAddress(), RandomAddress()

		setOntBalance(native.CacheDB, a, 10000)
		setOngBalance(native.CacheDB, utils.OntCcntmractAddress, constants.cntm_TOTAL_SUPPLY)

		assert.Equal(t, cntmBalanceOf(native, a), 10000)
		assert.Equal(t, cntmBalanceOf(native, b), 0)
		assert.Equal(t, cntmBalanceOf(native, c), 0)

		assert.Nil(t, cntmTransfer(native, a, b, 10))
		assert.Equal(t, cntmBalanceOf(native, a), 9990)
		assert.Equal(t, cntmBalanceOf(native, b), 10)

		assert.Nil(t, cntmTransfer(native, b, c, 10))
		assert.Equal(t, cntmBalanceOf(native, b), 0)
		assert.Equal(t, cntmBalanceOf(native, c), 10)

		return nil, nil
	})
}

func TestTotalAllowance(t *testing.T) {
	InvokeNativeCcntmract(t, utils.OntCcntmractAddress, func(native *native.NativeService) ([]byte, error) {
		a, b, c := RandomAddress(), RandomAddress(), RandomAddress()
		setOntBalance(native.CacheDB, a, 10000)
		setOngBalance(native.CacheDB, utils.OntCcntmractAddress, constants.cntm_TOTAL_SUPPLY)

		assert.Equal(t, cntmBalanceOf(native, a), 10000)
		assert.Equal(t, cntmBalanceOf(native, b), 0)
		assert.Equal(t, cntmBalanceOf(native, c), 0)

		assert.Nil(t, cntmApprove(native, a, b, 10))
		assert.Equal(t, cntmTotalAllowance(native, a), 10)
		assert.Equal(t, cntmTotalAllowance(native, b), 0)

		assert.Nil(t, cntmApprove(native, a, c, 100))
		assert.Equal(t, cntmTotalAllowance(native, a), 110)
		assert.Equal(t, cntmTotalAllowance(native, c), 0)

		assert.Nil(t, cntmTransferFrom(native, c, a, c, 100))
		return nil, nil
	})
}

func TestGovernanceUnbound(t *testing.T) {
	InvokeNativeCcntmract(t, utils.OntCcntmractAddress, func(native *native.NativeService) ([]byte, error) {
		testAddr, _ := common.AddressParseFromBytes([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF})
		setOntBalance(native.CacheDB, testAddr, constants.cntm_TOTAL_SUPPLY)
		setOngBalance(native.CacheDB, utils.OntCcntmractAddress, constants.cntm_TOTAL_SUPPLY)

		native.Time = constants.GENESIS_BLOCK_TIMESTAMP + 1

		assert.Nil(t, cntmTransfer(native, testAddr, testAddr, 1))
		assert.Equal(t, cntmAllowance(native, utils.OntCcntmractAddress, testAddr), uint64(5000000000))

		return nil, nil
	})

	InvokeNativeCcntmract(t, utils.OntCcntmractAddress, func(native *native.NativeService) ([]byte, error) {
		gov := utils.GovernanceCcntmractAddress
		testAddr, _ := common.AddressParseFromBytes([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF})
		setOntBalance(native.CacheDB, testAddr, constants.cntm_TOTAL_SUPPLY)
		setOngBalance(native.CacheDB, utils.OntCcntmractAddress, constants.cntm_TOTAL_SUPPLY)

		native.Time = constants.GENESIS_BLOCK_TIMESTAMP + 18*constants.UNBOUND_TIME_INTERVAL

		assert.Nil(t, cntmTransfer(native, testAddr, testAddr, 1))
		assert.Nil(t, unboundGovernanceOng(native))
		assert.EqualValues(t, cntmBalanceOf(native, gov)+cntmBalanceOf(native, testAddr), constants.cntm_TOTAL_SUPPLY)

		return nil, nil
	})

	InvokeNativeCcntmract(t, utils.OntCcntmractAddress, func(native *native.NativeService) ([]byte, error) {
		gov := utils.GovernanceCcntmractAddress
		testAddr, _ := common.AddressParseFromBytes([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF})
		setOntBalance(native.CacheDB, testAddr, constants.cntm_TOTAL_SUPPLY)
		setOngBalance(native.CacheDB, utils.OntCcntmractAddress, constants.cntm_TOTAL_SUPPLY)

		native.Time = constants.GENESIS_BLOCK_TIMESTAMP + 18*constants.UNBOUND_TIME_INTERVAL

		assert.Nil(t, unboundGovernanceOng(native))
		assert.Nil(t, cntmTransfer(native, testAddr, testAddr, 1))
		assert.EqualValues(t, cntmBalanceOf(native, gov)+cntmBalanceOf(native, testAddr), constants.cntm_TOTAL_SUPPLY)

		return nil, nil
	})

	InvokeNativeCcntmract(t, utils.OntCcntmractAddress, func(native *native.NativeService) ([]byte, error) {
		gov := utils.GovernanceCcntmractAddress
		testAddr, _ := common.AddressParseFromBytes([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF})
		setOntBalance(native.CacheDB, testAddr, constants.cntm_TOTAL_SUPPLY)
		setOngBalance(native.CacheDB, utils.OntCcntmractAddress, constants.cntm_TOTAL_SUPPLY)

		native.Time = constants.GENESIS_BLOCK_TIMESTAMP + 1
		assert.Nil(t, cntmTransfer(native, testAddr, testAddr, 1))
		native.Time = constants.GENESIS_BLOCK_TIMESTAMP + 10000
		assert.Nil(t, cntmTransfer(native, testAddr, testAddr, 1))
		native.Time = config.GetOntHolderUnboundDeadline() - 100
		assert.Nil(t, cntmTransfer(native, testAddr, testAddr, 1))

		native.Time = constants.GENESIS_BLOCK_TIMESTAMP + 18*constants.UNBOUND_TIME_INTERVAL

		assert.Nil(t, unboundGovernanceOng(native))
		assert.Nil(t, cntmTransfer(native, testAddr, testAddr, 1))
		assert.EqualValues(t, cntmBalanceOf(native, gov)+cntmBalanceOf(native, testAddr), constants.cntm_TOTAL_SUPPLY)

		return nil, nil
	})
}

//************************ version 2 ***************************

func TestTransferV2(t *testing.T) {
	InvokeNativeCcntmract(t, utils.OntCcntmractAddress, func(native *native.NativeService) ([]byte, error) {
		a, b, c := RandomAddress(), RandomAddress(), RandomAddress()
		setOntBalance(native.CacheDB, a, 10000)
		// default networkid is mainnet, need set cntm balance for cntm ccntmract
		setOngBalance(native.CacheDB, utils.OntCcntmractAddress, constants.cntm_TOTAL_SUPPLY)

		assert.Equal(t, cntmBalanceOfV2(native, a), uint64(10000*states.ScaleFactor))
		assert.Equal(t, cntmBalanceOfV2(native, b), uint64(0))
		assert.Equal(t, cntmBalanceOfV2(native, c), uint64(0))
		assert.Equal(t, cntmBalanceOf(native, a), 10000)

		assert.Nil(t, cntmTransferV2(native, a, b, 10*states.ScaleFactor))
		assert.Equal(t, cntmBalanceOfV2(native, a), uint64(9990*states.ScaleFactor))
		assert.Equal(t, cntmBalanceOfV2(native, b), uint64(10*states.ScaleFactor))
		assert.Equal(t, cntmBalanceOf(native, a), 9990)
		assert.Equal(t, cntmBalanceOf(native, b), 10)

		assert.Nil(t, cntmTransferV2(native, b, c, 10*states.ScaleFactor))
		assert.Equal(t, cntmBalanceOfV2(native, b), uint64(0))
		assert.Equal(t, cntmBalanceOfV2(native, c), uint64(10*states.ScaleFactor))

		return nil, nil
	})
}

func TestTotalAllowanceV2(t *testing.T) {
	InvokeNativeCcntmract(t, utils.OntCcntmractAddress, func(native *native.NativeService) ([]byte, error) {
		a, b, c := RandomAddress(), RandomAddress(), RandomAddress()
		setOntBalance(native.CacheDB, a, 10000)
		setOngBalance(native.CacheDB, utils.OntCcntmractAddress, constants.cntm_TOTAL_SUPPLY)

		assert.Equal(t, cntmBalanceOfV2(native, a), uint64(10000*states.ScaleFactor))
		assert.Equal(t, cntmBalanceOfV2(native, b), uint64(0))
		assert.Equal(t, cntmBalanceOfV2(native, c), uint64(0))

		assert.Nil(t, cntmApproveV2(native, a, b, 10*states.ScaleFactor))
		assert.Equal(t, cntmTotalAllowanceV2(native, a), uint64(10*states.ScaleFactor))
		assert.Equal(t, cntmTotalAllowanceV2(native, b), uint64(0))

		assert.Nil(t, cntmApproveV2(native, a, c, uint64(100*states.ScaleFactor)))
		assert.Equal(t, cntmTotalAllowanceV2(native, a), uint64(110*states.ScaleFactor))
		assert.Equal(t, cntmTotalAllowanceV2(native, c), uint64(0))
		fmt.Println(cntmBalanceOfV2(native, a))

		assert.Nil(t, cntmTransferFromV2(native, c, a, c, uint64(100*states.ScaleFactor)))
		return nil, nil
	})
}

func TestGovernanceUnboundV2(t *testing.T) {
	InvokeNativeCcntmract(t, utils.OntCcntmractAddress, func(native *native.NativeService) ([]byte, error) {
		testAddr, _ := common.AddressParseFromBytes([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF})
		setOntBalance(native.CacheDB, testAddr, constants.cntm_TOTAL_SUPPLY)
		setOngBalance(native.CacheDB, utils.OntCcntmractAddress, constants.cntm_TOTAL_SUPPLY)

		native.Time = constants.GENESIS_BLOCK_TIMESTAMP + 1
		assert.Nil(t, cntmTransferV2(native, testAddr, testAddr, 1))
		assert.Equal(t, cntmAllowanceV2(native, utils.OntCcntmractAddress, testAddr).String(), big.NewInt(5000000000*states.ScaleFactor).String())
		native.Time = native.Time + 100000
		native.Height = config.GetAddDecimalsHeight()
		assert.Nil(t, cntmTransferV2(native, testAddr, testAddr, constants.cntm_TOTAL_SUPPLY_V2/2))

		assert.Nil(t, cntmTransferFromV2(native, testAddr, utils.OntCcntmractAddress, testAddr, states.NativeTokenBalance{Balance: bigint.New(1)}))
		native.Time = native.Time + 100000
		assert.Nil(t, cntmTransferV2(native, testAddr, testAddr, 1))
		return nil, nil
	})

	InvokeNativeCcntmract(t, utils.OntCcntmractAddress, func(native *native.NativeService) ([]byte, error) {
		gov := utils.GovernanceCcntmractAddress
		testAddr, _ := common.AddressParseFromBytes([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF})
		setOntBalance(native.CacheDB, testAddr, constants.cntm_TOTAL_SUPPLY)
		setOngBalance(native.CacheDB, utils.OntCcntmractAddress, constants.cntm_TOTAL_SUPPLY)

		native.Time = constants.GENESIS_BLOCK_TIMESTAMP + 18*constants.UNBOUND_TIME_INTERVAL

		assert.Nil(t, cntmTransferV2(native, testAddr, testAddr, 1))
		assert.Nil(t, unboundGovernanceOng(native))
		assert.EqualValues(t, cntmBalanceOfV2(native, gov).Add(cntmBalanceOfV2(native, testAddr)).String(), constants.cntm_TOTAL_SUPPLY_V2.String())

		return nil, nil
	})

	InvokeNativeCcntmract(t, utils.OntCcntmractAddress, func(native *native.NativeService) ([]byte, error) {
		gov := utils.GovernanceCcntmractAddress
		testAddr, _ := common.AddressParseFromBytes([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF})
		setOntBalance(native.CacheDB, testAddr, constants.cntm_TOTAL_SUPPLY)
		setOngBalance(native.CacheDB, utils.OntCcntmractAddress, constants.cntm_TOTAL_SUPPLY)

		native.Time = constants.GENESIS_BLOCK_TIMESTAMP + 18*constants.UNBOUND_TIME_INTERVAL

		assert.Nil(t, unboundGovernanceOng(native))
		assert.Nil(t, cntmTransferV2(native, testAddr, testAddr, 1))
		assert.EqualValues(t, cntmBalanceOfV2(native, gov).Add(cntmBalanceOfV2(native, testAddr)).String(), constants.cntm_TOTAL_SUPPLY_V2.String())

		return nil, nil
	})

	InvokeNativeCcntmract(t, utils.OntCcntmractAddress, func(native *native.NativeService) ([]byte, error) {
		gov := utils.GovernanceCcntmractAddress
		testAddr, _ := common.AddressParseFromBytes([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF})
		setOntBalance(native.CacheDB, testAddr, constants.cntm_TOTAL_SUPPLY)
		setOngBalance(native.CacheDB, utils.OntCcntmractAddress, constants.cntm_TOTAL_SUPPLY)

		native.Time = constants.GENESIS_BLOCK_TIMESTAMP + 1
		assert.Nil(t, cntmTransferV2(native, testAddr, testAddr, 1))
		native.Time = constants.GENESIS_BLOCK_TIMESTAMP + 10000
		assert.Nil(t, cntmTransferV2(native, testAddr, testAddr, 1))
		native.Time = config.GetOntHolderUnboundDeadline() - 100
		assert.Nil(t, cntmTransferV2(native, testAddr, testAddr, 1))

		native.Time = constants.GENESIS_BLOCK_TIMESTAMP + 18*constants.UNBOUND_TIME_INTERVAL

		assert.Nil(t, unboundGovernanceOng(native))
		assert.Nil(t, cntmTransferV2(native, testAddr, testAddr, 1))
		assert.EqualValues(t, cntmBalanceOfV2(native, gov).Add(cntmBalanceOfV2(native, testAddr)).String(), constants.cntm_TOTAL_SUPPLY_V2.String())

		return nil, nil
	})
}
