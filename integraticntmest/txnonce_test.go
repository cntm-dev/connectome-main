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
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	types2 "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/cntmio/cntmology/account"
	utils2 "github.com/cntmio/cntmology/cmd/utils"
	common2 "github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/config"
	"github.com/cntmio/cntmology/core/ledger"
	"github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/smartccntmract/service/native/cntm"
	"github.com/cntmio/cntmology/smartccntmract/service/native/utils"
)

const WingABI = "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"subtractedValue\",\"type\":\"uint256\"}],\"name\":\"decreaseAllowance\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"addedValue\",\"type\":\"uint256\"}],\"name\":\"increaseAllowance\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"mint\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"burn\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

const testCcntmractDir = "./test-ccntmract"

// Mainly test several scenarios
// 1. not enough cntm for transfer, Will transaction nonce be updated
// 2. not enough cntm for deploy ccntmract, Will transaction nonce be updated
// 3. check cntm transfer event, there should be two events of cntm transfer in cntm transfer transaction,
// others, there should be only one event for cntm transfer.
func TestTxNonce(t *testing.T) {
	database, acct := NewLedger()
	gasPrice := uint64(500)
	gasLimit := uint64(200000)

	fromPrivateKey, toEthAddr := prepareEthAcct(database, acct, gasPrice, gasLimit, int64(1*1000000000))
	nonce := int64(0)
	transferAmt := 0.5 * 1000000000
	// enough cntm for fee, enough cntm for transfer
	checkEvmOngTransfer(database, acct, gasPrice, gasLimit, fromPrivateKey, toEthAddr, int64(transferAmt), nonce)

	// enough cntm for fee, not enough cntm for transfer
	checkEvmOngTransfer(database, acct, gasPrice, gasLimit, fromPrivateKey, toEthAddr, int64(transferAmt), nonce+1)

	// not enough cntm for deploy evm ccntmract
	checkDeployEvmCcntmract(database, acct, gasPrice, gasLimit, int64(gasPrice*gasLimit)-1)

}

func checkDeployEvmCcntmract(database *ledger.Ledger, acct *account.Account, gasPrice, gasLimit uint64, cntmAmt int64) {
	privateKey, err := crypto.GenerateKey()
	checkErr(err)
	ethAddr := crypto.PubkeyToAddress(privateKey.PublicKey)
	transferOng(database, gasPrice, gasLimit, acct, common2.Address(ethAddr), cntmAmt)
	code := loadCcntmract(testCcntmractDir + "/wing_eth.evm")
	nonce := int64(0)
	evmTx := NewDeployEvmCcntmract(privateKey, nonce, gasPrice, gasLimit, int64(gasPrice*gasLimit), code, WingABI)
	cntmTx, err := types.TransactionFromEIP155(evmTx)
	checkErr(err)
	genBlock(database, acct, cntmTx)
	acc, err := database.GetEthAccount(ethAddr)
	checkErr(err)
	if acc.Nonce != uint64(nonce+1) {
		panic(fmt.Sprintf("acc.Nonce: %d, nonce+1: %d", acc.Nonce, nonce+1))
	}
}

func checkEvmOngTransfer(database *ledger.Ledger, acct *account.Account, gasPrice, gasLimit uint64, fromPrivateKey *ecdsa.PrivateKey, toEthAddr common.Address, amt int64, nonce int64) {
	fromEthAddr := crypto.PubkeyToAddress(fromPrivateKey.PublicKey)
	before := cntmBalanceOf(database, common2.Address(fromEthAddr))
	tx := evmTransferOng(fromPrivateKey, gasPrice, gasLimit, toEthAddr, nonce, amt)
	genBlock(database, acct, tx)
	evt, err := database.GetEventNotifyByTx(tx.Hash())
	checkErr(err)
	after := cntmBalanceOf(database, common2.Address(fromEthAddr))
	var expect uint64
	if evt.State == 1 {
		expect = evt.GasConsumed + uint64(amt)
	} else {
		expect = evt.GasConsumed
	}
	if before-after != expect {
		panic(fmt.Sprintf("before:%d, after:%d, evt.GasConsumed:%d, transferAmt:%d",
			before, after, evt.GasConsumed, amt))
	}
	ethAcc, err := database.GetEthAccount(fromEthAddr)
	checkErr(err)
	if ethAcc.Nonce != uint64(nonce+1) {
		panic(fmt.Sprintf("ethAcc.Nonce:%d, nonce+1:%d", ethAcc.Nonce, nonce+1))
	}
}

func prepareEthAcct(database *ledger.Ledger, acct *account.Account, gasPrice, gasLimit uint64, cntmAmt int64) (*ecdsa.PrivateKey, common.Address) {
	fromPrivateKey, err := crypto.GenerateKey()
	checkErr(err)
	fromEthAddr := crypto.PubkeyToAddress(fromPrivateKey.PublicKey)
	transferOng(database, gasPrice, gasLimit, acct, common2.Address(fromEthAddr), cntmAmt)

	toPrivateKey, err := crypto.GenerateKey()
	checkErr(err)
	toEthAddr := crypto.PubkeyToAddress(toPrivateKey.PublicKey)
	return fromPrivateKey, toEthAddr
}

func genBlock(database *ledger.Ledger, acct *account.Account, tx *types.Transaction) {
	_, err := database.PreExecuteCcntmract(tx)
	checkErr(err)
	block, _ := makeBlock(acct, []*types.Transaction{tx})
	err = database.AddBlock(block, nil, common2.UINT256_EMPTY)
	checkErr(err)
}

func transferOng(database *ledger.Ledger, gasPrice, gasLimit uint64, acct *account.Account, toAddr common2.Address, amount int64) {
	state := &cntm.State{
		From:  acct.Address,
		To:    toAddr,
		Value: uint64(amount),
	}
	mutable := newNativeTx(utils.OngCcntmractAddress, 0, gasPrice, gasLimit, "transfer", []interface{}{[]*cntm.State{state}})
	err := utils2.SignTransaction(acct, mutable)
	checkErr(err)
	tx, err := mutable.IntoImmutable()
	checkErr(err)
	genBlock(database, acct, tx)
}

func cntmBalanceOf(database *ledger.Ledger, acctAddr common2.Address) uint64 {
	mutable := newNativeTx(utils.OngCcntmractAddress, 0, 0, 0, "balanceOf", []interface{}{acctAddr[:]})
	tx, err := mutable.IntoImmutable()
	checkErr(err)
	res, err := database.PreExecuteCcntmract(tx)
	checkErr(err)
	data, err := hex.DecodeString(res.Result.(string))
	checkErr(err)
	balance := common2.BigIntFromNeoBytes(data)
	return balance.Uint64()
}

func evmTransferOng(testPrivateKey *ecdsa.PrivateKey, gasPrice, gasLimit uint64, toEthAddr common.Address, nonce int64, value int64) *types.Transaction {
	chainId := big.NewInt(int64(config.DefConfig.P2PNode.EVMChainId))
	opts, err := bind.NewKeyedTransactorWithChainID(testPrivateKey, chainId)
	opts.GasPrice = big.NewInt(int64(gasPrice))
	opts.Nonce = big.NewInt(nonce)
	opts.GasLimit = gasLimit
	opts.Value = big.NewInt(value)

	invokeTx := types2.NewTransaction(opts.Nonce.Uint64(), toEthAddr, opts.Value, opts.GasLimit, opts.GasPrice, []byte{})
	signedTx, err := opts.Signer(opts.From, invokeTx)
	checkErr(err)
	tx, err := types.TransactionFromEIP155(signedTx)
	checkErr(err)
	return tx
}
