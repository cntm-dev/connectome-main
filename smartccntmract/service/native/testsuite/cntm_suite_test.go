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
	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/config"
	"github.com/cntmio/cntmology/common/constants"
	"github.com/cntmio/cntmology/smartccntmract/service/native"
	_ "github.com/cntmio/cntmology/smartccntmract/service/native/init"
	"github.com/cntmio/cntmology/smartccntmract/service/native/cntm"
	"github.com/cntmio/cntmology/smartccntmract/service/native/cntm"
	"github.com/cntmio/cntmology/smartccntmract/service/native/utils"
	"github.com/cntmio/cntmology/smartccntmract/storage"
	"github.com/stretchr/testify/assert"

	"testing"
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

func setOngBalance(db *storage.CacheDB, addr common.Address, value uint64) {
	balanceKey := cntm.GenBalanceKey(utils.OngCcntmractAddress, addr)
	item := utils.GenUInt64StorageItem(value)
	db.Put(balanceKey, item.ToArray())
}

func cntmBalanceOf(native *native.NativeService, addr common.Address) uint64 {
	native.CcntmextRef.CurrentCcntmext().CcntmractAddress = utils.OngCcntmractAddress
	sink := common.NewZeroCopySink(nil)
	utils.EncodeAddress(sink, addr)
	native.Input = sink.Bytes()
	buf, _ := cntm.OngBalanceOf(native)
	val := common.BigIntFromNeoBytes(buf)
	return val.Uint64()
}

func cntmAllowance(native *native.NativeService, from, to common.Address) uint64 {
	native.CcntmextRef.CurrentCcntmext().CcntmractAddress = utils.OngCcntmractAddress
	sink := common.NewZeroCopySink(nil)
	utils.EncodeAddress(sink, from)
	utils.EncodeAddress(sink, to)
	native.Input = sink.Bytes()
	buf, _ := cntm.OngAllowance(native)
	val := common.BigIntFromNeoBytes(buf)
	return val.Uint64()
}

func cntmTotalAllowance(native *native.NativeService, addr common.Address) int {
	sink := common.NewZeroCopySink(nil)
	utils.EncodeAddress(sink, addr)
	native.Input = sink.Bytes()
	buf, _ := cntm.TotalAllowance(native)
	val := common.BigIntFromNeoBytes(buf)
	return int(val.Uint64())
}

func cntmTransfer(native *native.NativeService, from, to common.Address, value uint64) error {
	native.Tx.SignedAddr = append(native.Tx.SignedAddr, from)

	state := cntm.State{from, to, value}
	native.Input = common.SerializeToBytes(&cntm.Transfers{States: []cntm.State{state}})

	_, err := cntm.OntTransfer(native)
	return err
}

func cntmApprove(native *native.NativeService, from, to common.Address, value uint64) error {
	native.Tx.SignedAddr = append(native.Tx.SignedAddr, from)

	native.Input = common.SerializeToBytes(&cntm.State{from, to, value})

	_, err := cntm.OntApprove(native)
	return err
}

func unboundGovernanceOng(native *native.NativeService) error {
	_, err := cntm.UnboundOngToGovernance(native)
	return err
}

func TestTransfer(t *testing.T) {
	InvokeNativeCcntmract(t, utils.OntCcntmractAddress, func(native *native.NativeService) ([]byte, error) {
		a := RandomAddress()
		b := RandomAddress()
		c := RandomAddress()
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
		a := RandomAddress()
		b := RandomAddress()
		c := RandomAddress()
		setOntBalance(native.CacheDB, a, 10000)

		assert.Equal(t, cntmBalanceOf(native, a), 10000)
		assert.Equal(t, cntmBalanceOf(native, b), 0)
		assert.Equal(t, cntmBalanceOf(native, c), 0)

		assert.Nil(t, cntmApprove(native, a, b, 10))
		assert.Equal(t, cntmTotalAllowance(native, a), 10)
		assert.Equal(t, cntmTotalAllowance(native, b), 0)

		assert.Nil(t, cntmApprove(native, a, c, 100))
		assert.Equal(t, cntmTotalAllowance(native, a), 110)
		assert.Equal(t, cntmTotalAllowance(native, c), 0)

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
		assert.Equal(t, cntmBalanceOf(native, gov)+cntmBalanceOf(native, testAddr), constants.cntm_TOTAL_SUPPLY)

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
		assert.Equal(t, cntmBalanceOf(native, gov)+cntmBalanceOf(native, testAddr), constants.cntm_TOTAL_SUPPLY)

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
		assert.Equal(t, cntmBalanceOf(native, gov)+cntmBalanceOf(native, testAddr), constants.cntm_TOTAL_SUPPLY)

		return nil, nil
	})
}
