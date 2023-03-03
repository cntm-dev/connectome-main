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
package testsuite

import (
	"github.com/conntectome/cntm/common"
	"github.com/conntectome/cntm/smartcontract/service/native"
	_ "github.com/conntectome/cntm/smartcontract/service/native/init"
	"github.com/conntectome/cntm/smartcontract/service/native/cntm"
	"github.com/conntectome/cntm/smartcontract/service/native/utils"
	"github.com/conntectome/cntm/smartcontract/storage"
	"github.com/stretchr/testify/assert"

	"testing"
)

func setCntmBalance(db *storage.CacheDB, addr common.Address, value uint64) {
	balanceKey := cntm.GenBalanceKey(utils.CntmCcntmractAddress, addr)
	item := utils.GenUInt64StorageItem(value)
	db.Put(balanceKey, item.ToArray())
}

func cntmBalanceOf(native *native.NativeService, addr common.Address) int {
	sink := common.NewZeroCopySink(nil)
	utils.EncodeAddress(sink, addr)
	native.Input = sink.Bytes()
	buf, _ := cntm.CntmBalanceOf(native)
	val := common.BigIntFromCntmBytes(buf)
	return int(val.Uint64())
}

func cntmTransfer(native *native.NativeService, from, to common.Address, value uint64) error {
	native.Tx.SignedAddr = append(native.Tx.SignedAddr, from)

	state := cntm.State{from, to, value}
	native.Input = common.SerializeToBytes(&cntm.Transfers{States: []cntm.State{state}})

	_, err := cntm.CntmTransfer(native)
	return err
}

func TestTransfer(t *testing.T) {
	InvokeNativeCcntmract(t, utils.CntmCcntmractAddress, func(native *native.NativeService) ([]byte, error) {
		a := RandomAddress()
		b := RandomAddress()
		c := RandomAddress()
		setCntmBalance(native.CacheDB, a, 10000)

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
