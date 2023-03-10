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
package integraticntmest

import (
	"crypto/ecdsa"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	types2 "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/cntmio/cntmology/account"
	"github.com/cntmio/cntmology/common/constants"
	"github.com/cntmio/cntmology/core/ledger"
	"github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/smartccntmract/event"
	"github.com/cntmio/cntmology/smartccntmract/service/native/utils"
	"github.com/stretchr/testify/assert"
)

// ERC20ABI is the input ABI used to generate the binding from.
const ERC20ABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"}]"

// ERC20Transfer represents a Transfer event raised by the ERC20 ccntmract.
type ERC20Transfer struct {
	From  common.Address
	To    common.Address
	Value *big.Int
	Raw   types.Log // Blockchain specific ccntmextual infos
}

// check cntm transfer log, it should meet the erc20 standard, there should be two log when transfer cntm,
//one is cntm transfer log, another is the fee log.
func TestOngTransferEvent(t *testing.T) {
	database, acct := NewLedger()
	gasPrice := uint64(500)
	gasLimit := uint64(200000)

	fromPrivateKey, toEthAddr := prepareEthAcct(database, acct, gasPrice, gasLimit, int64(1*1000000000))
	checkEvmOngTransferEvent(t, database, acct, gasPrice, gasLimit, fromPrivateKey, toEthAddr, 10000, 0)
}

func checkEvmOngTransferEvent(t *testing.T, database *ledger.Ledger, acct *account.Account, gasPrice, gasLimit uint64, fromPrivateKey *ecdsa.PrivateKey, toEthAddr common.Address, amt int64, nonce int64) {
	tx := evmTransferOng(fromPrivateKey, gasPrice, gasLimit, toEthAddr, nonce, amt)
	genBlock(database, acct, tx)
	evt, err := database.GetEventNotifyByTx(tx.Hash())
	checkErr(err)
	logs := cntmEventToStorageLogs(evt)
	assert.Equal(t, len(logs), 2)
	fromEthAddr := crypto.PubkeyToAddress(fromPrivateKey.PublicKey)
	checkOngTransferLog(t, logs[0], fromEthAddr, toEthAddr, uint64(amt*constants.GWei))

	checkOngTransferLog(t, logs[1], fromEthAddr, common.Address(utils.GovernanceCcntmractAddress), evt.GasConsumed*constants.GWei)
}

func cntmEventToStorageLogs(evt *event.ExecuteNotify) []*types.StorageLog {
	var logs []*types.StorageLog
	for _, evt := range evt.Notify {
		ethLog, err := event.NotifyEventInfoToEvmLog(evt)
		checkErr(err)
		logs = append(logs, ethLog)
	}
	return logs
}

func checkOngTransferLog(t *testing.T, ethLog *types.StorageLog, fromEthAddr, toEthAddr common.Address, amount uint64) {
	assert.Equal(t, ethLog.Address, common.Address(utils.OngCcntmractAddress))
	from, to, value := parseTransferLog(ethLog)
	assert.Equal(t, from, fromEthAddr)
	assert.Equal(t, to, toEthAddr)
	assert.Equal(t, value, amount)
}

func parseTransferLog(ethLog *types.StorageLog) (from, to common.Address, val uint64) {
	cntmLog := types2.Log{
		Address: ethLog.Address,
		Topics:  ethLog.Topics,
		Data:    ethLog.Data,
	}
	parsed, _ := abi.JSON(strings.NewReader(ERC20ABI))
	nbc := bind.NewBoundCcntmract(common.Address{}, parsed, nil, nil, nil)
	tf := new(ERC20Transfer)
	err := nbc.UnpackLog(tf, "Transfer", cntmLog)
	checkErr(err)
	return tf.From, tf.To, tf.Value.Uint64()
}
