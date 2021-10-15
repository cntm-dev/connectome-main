package proc

import (
	"fmt"
	"github.com/Ontology/common"
	"github.com/Ontology/common/log"
	tx "github.com/Ontology/core/types"
	"github.com/Ontology/eventbus/actor"
	tc "github.com/Ontology/txnpool/common"
)

func NewTxnActor(s *TXNPoolServer) *TxnActor {
	a := &TxnActor{}
	a.setServer(s)
	return a
}

func NewTxnPoolActor(s *TXNPoolServer) *TxnPoolActor {
	a := &TxnPoolActor{}
	a.setServer(s)
	return a
}

func NewVerifyRspActor(s *TXNPoolServer) *VerifyRspActor {
	a := &VerifyRspActor{}
	a.setServer(s)
	return a
}

// TxnActor: Handle the low priority msg from P2P and API
type GetTxnReq struct {
	Hash common.Uint256
}

type GetTxnRsp struct {
	Txn *tx.Transaction
}

type CheckTxnReq struct {
	Hash common.Uint256
}

type CheckTxnRsp struct {
	Ok bool
}

type GetTxnStatusReq struct {
	Hash common.Uint256
}

type GetTxnStatusRsp struct {
	TxnStatus *tc.TXNEntry
}

type GetTxnStats struct {
}

type GetTxnStatsRsp struct {
	count *[]uint64
}

type TxnActor struct {
	server *TXNPoolServer
}

func (ta *TxnActor) Receive(ccntmext actor.Ccntmext) {
	switch msg := ccntmext.Message().(type) {
	case *actor.Started:
		log.Info("Server started and be ready to receive txn")
	case *actor.Stopping:
		log.Info("Server stopping")
	case *actor.Restarting:
		log.Info("Server Restarting")
	case *tx.Transaction:
		log.Info("Server Receives txn message")
		ta.server.increaseStats(tc.RcvStats)
		if txn := ta.server.getTransaction(msg.Hash()); txn != nil {
			log.Info(fmt.Sprintf("Transaction %x already in the txn pool",
				msg.Hash()))
			ta.server.increaseStats(tc.DuplicateStats)
		} else {
			ta.server.assginTXN2Worker(msg)
		}
	case *GetTxnReq:
		res := ta.server.getTransaction(msg.Hash)
		ccntmext.Sender().Request(&GetTxnRsp{Txn: res}, ccntmext.Self())
	case *GetTxnStats:
		res := ta.server.getStats()
		ccntmext.Sender().Request(&GetTxnStatsRsp{count: res}, ccntmext.Self())
	case *CheckTxnReq:
		res := ta.server.CheckTxn(msg.Hash)
		ccntmext.Sender().Request(&CheckTxnRsp{Ok: res}, ccntmext.Self())
	case *GetTxnStatusReq:
		res := ta.server.GetTxnStatusReq(msg.Hash)
		ccntmext.Sender().Request(&GetTxnStatusRsp{TxnStatus: res}, ccntmext.Self())
	default:
		log.Info("Unknown msg type", msg)
	}
}

func (ta *TxnActor) setServer(s *TXNPoolServer) {
	ta.server = s
}

// TxnPoolActor: Handle the high priority request from Consensus
type GetTxnPoolReq struct {
	ByCount bool
}

type GetTxnPoolRsp struct {
	TxnPool []*tc.TXNEntry
}

type GetPendingTxnReq struct {
	ByCount bool
}

type GetPendingTxnRsp struct {
	Txs []*tx.Transaction
}

type GetUnverifiedTxsReq struct {
	Txs []*tx.Transaction
}

type GetUnverifiedTxsRsp struct {
	Txs []*tx.Transaction
}

type CleanTxnPoolReq struct {
	TxnPool []*tx.Transaction
}

type TxnPoolActor struct {
	server *TXNPoolServer
}

func (tpa *TxnPoolActor) Receive(ccntmext actor.Ccntmext) {
	switch msg := ccntmext.Message().(type) {
	case *actor.Started:
		log.Info("Server started and be ready to receive txn")
	case *actor.Stopping:
		log.Info("Server stopping")
	case *actor.Restarting:
		log.Info("Server Restarting")
	case *GetTxnPoolReq:
		res := tpa.server.GetTxnPool(msg.ByCount)
		ccntmext.Sender().Request(&GetTxnPoolRsp{TxnPool: res}, ccntmext.Self())
	case *CleanTxnPoolReq:
		tpa.server.CleanTransactionList(msg.TxnPool)
	case *GetPendingTxnReq:
		res := tpa.server.GetPendingTxs(msg.ByCount)
		ccntmext.Sender().Request(&GetPendingTxnRsp{Txs: res}, ccntmext.Self())
	case *GetUnverifiedTxsReq:
		res := tpa.server.GetUnverifiedTxs(msg.Txs)
		ccntmext.Sender().Request(&GetUnverifiedTxsRsp{Txs: res}, ccntmext.Self())
	default:
		log.Info("Unknown msg type", msg)
	}
}

func (tpa *TxnPoolActor) setServer(s *TXNPoolServer) {
	tpa.server = s
}

// VerifyRspActor: Handle the response from the validators
type VerifyRspActor struct {
	server *TXNPoolServer
}

func (vpa *VerifyRspActor) Receive(ccntmext actor.Ccntmext) {
	switch msg := ccntmext.Message().(type) {
	case *actor.Started:
		log.Info("Server started and be ready to receive txn")
	case *actor.Stopping:
		log.Info("Server stopping")
	case *actor.Restarting:
		log.Info("Server Restarting")
	case *tc.VerifyRsp:
		log.Info("Server Receives verify rsp message")
		vpa.server.assignRsp2Worker(msg)
	default:
		log.Info("Unknown msg type", msg)
	}
}

func (vpa *VerifyRspActor) setServer(s *TXNPoolServer) {
	vpa.server = s
}
