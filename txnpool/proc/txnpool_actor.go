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

package proc

import (
	"fmt"
	"reflect"

	"github.com/cntmio/cntmology-eventbus/actor"
	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/config"
	"github.com/cntmio/cntmology/common/log"
	"github.com/cntmio/cntmology/core/ledger"
	tx "github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/errors"
	"github.com/cntmio/cntmology/events/message"
	hComm "github.com/cntmio/cntmology/http/base/common"
	"github.com/cntmio/cntmology/smartccntmract/service/native/utils"
	"github.com/cntmio/cntmology/smartccntmract/service/neovm"
	tc "github.com/cntmio/cntmology/txnpool/common"
)

// NewTxActor creates an actor to handle the transaction-based messages from
// network and http
func NewTxActor(s *TXPoolServer) *TxActor {
	a := &TxActor{server: s}
	return a
}

// NewTxPoolActor creates an actor to handle the messages from the consensus
func NewTxPoolActor(s *TXPoolServer) *TxPoolActor {
	a := &TxPoolActor{server: s}
	return a
}

// isBalanceEnough checks if the tranactor has enough to cover gas cost
func isBalanceEnough(address common.Address, gas uint64) bool {
	balance, _, err := hComm.GetCcntmractBalance(0, []common.Address{utils.OngCcntmractAddress}, address, false)
	if err != nil {
		log.Debugf("failed to get ccntmract balance %s err %v", address.ToHexString(), err)
		return false
	}
	return balance[0] >= gas
}

func replyTxResult(sender tc.SenderType, txResultCh chan *tc.TxResult, hash common.Uint256, err errors.ErrCode, desc string) {
	if sender == tc.HttpSender && txResultCh != nil {
		result := &tc.TxResult{
			Err:  err,
			Hash: hash,
			Desc: desc,
		}
		select {
		case txResultCh <- result:
		default:
			log.Debugf("handleTransaction: duplicated result")
		}
	}
}

// preExecCheck checks whether preExec pass
func preExecCheck(txn *tx.Transaction) (bool, string) {
	result, err := ledger.DefLedger.PreExecuteCcntmract(txn)
	if err != nil {
		log.Debugf("preExecCheck: failed to preExecuteCcntmract tx %x err %v",
			txn.Hash(), err)
	}
	if txn.GasLimit < result.Gas {
		log.Debugf("preExecCheck: transaction's gasLimit %d is less than preExec gasLimit %d",
			txn.GasLimit, result.Gas)
		return false, fmt.Sprintf("transaction's gasLimit %d is less than preExec gasLimit %d",
			txn.GasLimit, result.Gas)
	}
	gas, overflow := common.SafeMul(txn.GasPrice, result.Gas)
	if overflow {
		log.Debugf("preExecCheck: gasPrice %d preExec gasLimit %d overflow",
			txn.GasPrice, result.Gas)
		return false, fmt.Sprintf("gasPrice %d preExec gasLimit %d overflow",
			txn.GasPrice, result.Gas)
	}
	if !isBalanceEnough(txn.Payer, gas) {
		log.Debugf("preExecCheck: transactor %s has no balance enough to cover gas cost %d",
			txn.Payer.ToHexString(), gas)
		return false, fmt.Sprintf("transactor %s has no balance enough to cover gas cost %d",
			txn.Payer.ToHexString(), gas)
	}
	return true, ""
}

// TxnActor: Handle the low priority msg from P2P and API
type TxActor struct {
	server *TXPoolServer
}

// handles a transaction from network and http
func (ta *TxActor) handleTransaction(sender tc.SenderType, txn *tx.Transaction, txResultCh chan *tc.TxResult) {
	if len(txn.ToArray()) > tc.MAX_TX_SIZE {
		log.Debugf("handleTransaction: reject a transaction due to size over 1M")
		replyTxResult(sender, txResultCh, txn.Hash(), errors.ErrUnknown, "size is over 1M")
		return
	}

	if ta.server.getTransaction(txn.Hash()) != nil {
		log.Debugf("handleTransaction: transaction %x already in the txn pool", txn.Hash())

		replyTxResult(sender, txResultCh, txn.Hash(), errors.ErrDuplicateInput,
			fmt.Sprintf("transaction %x is already in the tx pool", txn.Hash()))
		return
	}
	if ta.server.getTransactionCount() >= tc.MAX_CAPACITY {
		log.Debugf("handleTransaction: transaction pool is full for tx %x", txn.Hash())

		replyTxResult(sender, txResultCh, txn.Hash(), errors.ErrTxPoolFull, "transaction pool is full")
		return
	}
	if _, overflow := common.SafeMul(txn.GasLimit, txn.GasPrice); overflow {
		log.Debugf("handleTransaction: gasLimit %v, gasPrice %v overflow", txn.GasLimit, txn.GasPrice)
		replyTxResult(sender, txResultCh, txn.Hash(), errors.ErrUnknown,
			fmt.Sprintf("gasLimit %d * gasPrice %d overflow", txn.GasLimit, txn.GasPrice))
		return
	}

	gasLimitConfig := config.DefConfig.Common.MinGasLimit
	gasPriceConfig := ta.server.getGasPrice()
	if txn.GasLimit < gasLimitConfig || txn.GasPrice < gasPriceConfig {
		log.Debugf("handleTransaction: invalid gasLimit %v, gasPrice %v", txn.GasLimit, txn.GasPrice)
		replyTxResult(sender, txResultCh, txn.Hash(), errors.ErrUnknown,
			fmt.Sprintf("Please input gasLimit >= %d and gasPrice >= %d", gasLimitConfig, gasPriceConfig))
		return
	}

	if txn.TxType == tx.Deploy && txn.GasLimit < neovm.CcntmRACT_CREATE_GAS {
		log.Debugf("handleTransaction: deploy tx invalid gasLimit %v, gasPrice %v", txn.GasLimit, txn.GasPrice)
		replyTxResult(sender, txResultCh, txn.Hash(), errors.ErrUnknown,
			fmt.Sprintf("Deploy tx gaslimit should >= %d", neovm.CcntmRACT_CREATE_GAS))
		return
	}

	if !ta.server.disablePreExec {
		if ok, desc := preExecCheck(txn); !ok {
			log.Debugf("handleTransaction: preExecCheck tx %x failed", txn.Hash())
			replyTxResult(sender, txResultCh, txn.Hash(), errors.ErrUnknown, desc)
			return
		}
		log.Debugf("handleTransaction: preExecCheck tx %x passed", txn.Hash())
	}
	<-ta.server.slots
	ta.server.assignTxToWorker(txn, sender, txResultCh)
}

// Receive implements the actor interface
func (ta *TxActor) Receive(ccntmext actor.Ccntmext) {
	switch msg := ccntmext.Message().(type) {
	case *actor.Started:
		log.Info("txpool-tx actor started and be ready to receive tx msg")

	case *actor.Stopping:
		log.Warn("txpool-tx actor stopping")

	case *actor.Restarting:
		log.Warn("txpool-tx actor restarting")

	case *tc.TxReq:
		sender := msg.Sender

		log.Debugf("txpool-tx actor receives tx from %v ", sender.Sender())

		ta.handleTransaction(sender, msg.Tx, msg.TxResultCh)

	case *tc.GetTxnReq:
		sender := ccntmext.Sender()

		log.Debugf("txpool-tx actor receives getting tx req from %v", sender)

		res := ta.server.getTransaction(msg.Hash)
		if sender != nil {
			sender.Request(&tc.GetTxnRsp{Txn: res}, ccntmext.Self())
		}

	case *tc.GetTxnStatusReq:
		sender := ccntmext.Sender()

		log.Debugf("txpool-tx actor receives getting tx status req from %v", sender)

		res := ta.server.getTxStatusReq(msg.Hash)
		if sender != nil {
			if res == nil {
				sender.Request(&tc.GetTxnStatusRsp{Hash: msg.Hash,
					TxStatus: nil}, ccntmext.Self())
			} else {
				sender.Request(&tc.GetTxnStatusRsp{Hash: res.Hash,
					TxStatus: res.Attrs}, ccntmext.Self())
			}
		}

	case *tc.GetTxnCountReq:
		sender := ccntmext.Sender()

		log.Debugf("txpool-tx actor receives getting tx count req from %v", sender)

		res := ta.server.getTxCount()
		if sender != nil {
			sender.Request(&tc.GetTxnCountRsp{Count: res},
				ccntmext.Self())
		}
	case *tc.GetPendingTxnHashReq:
		sender := ccntmext.Sender()

		log.Debugf("txpool-tx actor receives getting pedning tx hash req from %v", sender)

		res := ta.server.getTxHashList()
		if sender != nil {
			sender.Request(&tc.GetPendingTxnHashRsp{TxHashs: res}, ccntmext.Self())
		}

	default:
		log.Debugf("txpool-tx actor: unknown msg %v type %v", msg, reflect.TypeOf(msg))
	}
}

// TxnPoolActor: Handle the high priority request from Consensus
type TxPoolActor struct {
	server *TXPoolServer
}

// Receive implements the actor interface
func (tpa *TxPoolActor) Receive(ccntmext actor.Ccntmext) {
	switch msg := ccntmext.Message().(type) {
	case *actor.Started:
		log.Info("txpool actor started and be ready to receive txPool msg")

	case *actor.Stopping:
		log.Warn("txpool actor stopping")

	case *actor.Restarting:
		log.Warn("txpool actor Restarting")

	case *tc.GetTxnPoolReq:
		sender := ccntmext.Sender()

		log.Debugf("txpool actor receives getting tx pool req from %v", sender)

		res := tpa.server.getTxPool(msg.ByCount, msg.Height)
		if sender != nil {
			sender.Request(&tc.GetTxnPoolRsp{TxnPool: res}, ccntmext.Self())
		}

	case *tc.VerifyBlockReq:
		sender := ccntmext.Sender()

		log.Debugf("txpool actor receives verifying block req from %v", sender)

		tpa.server.verifyBlock(msg, sender)

	case *message.SaveBlockCompleteMsg:
		sender := ccntmext.Sender()

		log.Debugf("txpool actor receives block complete event from %v", sender)

		if msg.Block != nil {
			tpa.server.cleanTransactionList(msg.Block.Transactions, msg.Block.Header.Height)
		}

	default:
		log.Debugf("txpool actor: unknown msg %v type %v", msg, reflect.TypeOf(msg))
	}
}
