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

	"github.com/Ontology/common/log"
	tx "github.com/Ontology/core/types"
	"github.com/Ontology/errors"
	"github.com/Ontology/eventbus/actor"
	"github.com/Ontology/events/message"
	tc "github.com/Ontology/txnpool/common"
	"github.com/Ontology/validator/types"
)

func NewTxActor(s *TXPoolServer) *TxActor {
	a := &TxActor{}
	a.setServer(s)
	return a
}

func NewTxPoolActor(s *TXPoolServer) *TxPoolActor {
	a := &TxPoolActor{}
	a.setServer(s)
	return a
}

func NewVerifyRspActor(s *TXPoolServer) *VerifyRspActor {
	a := &VerifyRspActor{}
	a.setServer(s)
	return a
}

// TxnActor: Handle the low priority msg from P2P and API
type TxActor struct {
	server *TXPoolServer
}

// Handle a new transaction
func (ta *TxActor) handleTransaction(sender, self *actor.PID,
	txn *tx.Transaction) {
	ta.server.increaseStats(tc.RcvStats)

	if ta.server.getTransaction(txn.Hash()) != nil {
		log.Info(fmt.Sprintf("Transaction %x already in the txn pool",
			txn.Hash()))

		ta.server.increaseStats(tc.DuplicateStats)

		if sender == nil {
			return
		}
		rsp := &tc.TxRsp{
			Hash:    txn.Hash(),
			ErrCode: errors.ErrNoError,
		}
		sender.Request(rsp, self)
	} else if ta.server.getTransactionCount() >= tc.MAXCAPACITY {
		log.Info("Transaction pool is full", txn.Hash())

		ta.server.increaseStats(tc.FailureStats)

		if sender == nil {
			return
		}

		rsp := &tc.TxRsp{
			Hash:    txn.Hash(),
			ErrCode: errors.ErrUnknown,
		}
		sender.Request(rsp, self)
	} else {
		ta.server.assginTXN2Worker(txn, sender)
	}
}

func (ta *TxActor) Receive(ccntmext actor.Ccntmext) {
	switch msg := ccntmext.Message().(type) {
	case *actor.Started:
		log.Info("txpool-tx actor started and be ready to receive tx msg")

	case *actor.Stopping:
		log.Info("txpool-tx actor stopping")

	case *actor.Restarting:
		log.Info("txpool-tx actor Restarting")

	case *tx.Transaction:
		sender := ccntmext.Sender()

		log.Info("txpool-tx actor Receives tx from ", sender)

		ta.handleTransaction(sender, ccntmext.Self(), msg)

	case *tc.GetTxnReq:
		sender := ccntmext.Sender()

		log.Info("txpool-tx actor Receives getting tx req from ", sender)

		res := ta.server.getTransaction(msg.Hash)
		if sender != nil {
			sender.Request(&tc.GetTxnRsp{Txn: res},
				ccntmext.Self())
		}

	case *tc.GetTxnStats:
		sender := ccntmext.Sender()

		log.Info("txpool-tx actor Receives getting tx stats from ", sender)

		res := ta.server.getStats()
		if sender != nil {
			sender.Request(&tc.GetTxnStatsRsp{Count: res},
				ccntmext.Self())
		}

	case *tc.CheckTxnReq:
		sender := ccntmext.Sender()

		log.Info("txpool-tx actor Receives checking tx req from ", sender)

		res := ta.server.checkTx(msg.Hash)
		if sender != nil {
			sender.Request(&tc.CheckTxnRsp{Ok: res},
				ccntmext.Self())
		}

	case *tc.GetTxnStatusReq:
		sender := ccntmext.Sender()

		log.Info("txpool-tx actor Receives getting tx status req from ", sender)

		res := ta.server.getTxStatusReq(msg.Hash)
		if sender != nil {
			sender.Request(&tc.GetTxnStatusRsp{Hash: res.Hash,
				TxStatus: res.Attrs}, ccntmext.Self())
		}

	default:
		log.Info("txpool-tx actor: Unknown msg ", msg, "type", reflect.TypeOf(msg))
	}
}

func (ta *TxActor) setServer(s *TXPoolServer) {
	ta.server = s
}

// TxnPoolActor: Handle the high priority request from Consensus
type TxPoolActor struct {
	server *TXPoolServer
}

func (tpa *TxPoolActor) Receive(ccntmext actor.Ccntmext) {
	switch msg := ccntmext.Message().(type) {
	case *actor.Started:
		log.Info("txpool actor started and be ready to receive txPool msg")

	case *actor.Stopping:
		log.Info("txpool actor stopping")

	case *actor.Restarting:
		log.Info("txpool actor Restarting")

	case *tc.GetTxnPoolReq:
		sender := ccntmext.Sender()

		log.Info("txpool actor Receives getting tx pool req from ", sender)

		res := tpa.server.getTxPool(msg.ByCount)
		if sender != nil {
			sender.Request(&tc.GetTxnPoolRsp{TxnPool: res}, ccntmext.Self())
		}

	case *tc.GetPendingTxnReq:
		sender := ccntmext.Sender()

		log.Info("txpool actor Receives getting pedning tx req from ", sender)

		res := tpa.server.getPendingTxs(msg.ByCount)
		if sender != nil {
			sender.Request(&tc.GetPendingTxnRsp{Txs: res}, ccntmext.Self())
		}

	case *tc.VerifyBlockReq:
		sender := ccntmext.Sender()

		log.Info("txpool actor Receives verifying block req from ", sender)

		tpa.server.verifyBlock(msg, sender)

	case *message.SaveBlockCompleteMsg:
		sender := ccntmext.Sender()

		log.Info("txpool actor Receives block complete event from ", sender)

		if msg.Block != nil {
			tpa.server.cleanTransactionList(msg.Block.Transactions)
		}

	default:
		log.Info("txpool actor: Unknown msg ", msg, "type", reflect.TypeOf(msg))
	}
}

func (tpa *TxPoolActor) setServer(s *TXPoolServer) {
	tpa.server = s
}

// VerifyRspActor: Handle the response from the validators
type VerifyRspActor struct {
	server *TXPoolServer
}

func (vpa *VerifyRspActor) Receive(ccntmext actor.Ccntmext) {
	switch msg := ccntmext.Message().(type) {
	case *actor.Started:
		log.Info("txpool-verify actor: started and be ready to receive validator's msg")

	case *actor.Stopping:
		log.Info("txpool-verify actor: stopping")

	case *actor.Restarting:
		log.Info("txpool-verify actor: Restarting")

	case *types.RegisterValidator:
		log.Infof("txpool-verify actor:: validator %v connected", msg.Sender)
		vpa.server.registerValidator(msg)

	case *types.UnRegisterValidator:
		log.Infof("txpool-verify actor:: validator %d:%v disconnected", msg.Type, msg.Id)

		vpa.server.unRegisterValidator(msg.Type, msg.Id)

	case *types.CheckResponse:
		log.Info("txpool-verify actor:: Receives verify rsp message")

		vpa.server.assignRsp2Worker(msg)

	default:
		log.Info("txpool-verify actor:Unknown msg ", msg, "type", reflect.TypeOf(msg))
	}
}

func (vpa *VerifyRspActor) setServer(s *TXPoolServer) {
	vpa.server = s
}
