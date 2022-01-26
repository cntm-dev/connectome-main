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

package actor

import (
	"fmt"
	"reflect"

	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/log"
	"github.com/cntmio/cntmology/core/ledger"
	"github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology-eventbus/actor"
)

var DefLedgerPid *actor.PID

type LedgerActor struct {
	props *actor.Props
}

func NewLedgerActor() *LedgerActor {
	return &LedgerActor{}
}

func (self *LedgerActor) Start() *actor.PID {
	self.props = actor.FromProducer(func() actor.Actor { return self })
	var err error
	DefLedgerPid, err = actor.SpawnNamed(self.props, "LedgerActor")
	if err != nil {
		panic(fmt.Errorf("LedgerActor SpawnNamed error:%s", err))
	}
	return DefLedgerPid
}

func (self *LedgerActor) Receive(ctx actor.Ccntmext) {
	switch msg := ctx.Message().(type) {
	case *actor.Started:
	case *actor.Stop:
	case *AddHeaderReq:
		self.handleAddHeaderReq(ctx, msg)
	case *AddHeadersReq:
		self.handleAddHeadersReq(ctx, msg)
	case *AddBlockReq:
		self.handleAddBlockReq(ctx, msg)
	case *GetTransactionReq:
		self.handleGetTransactionReq(ctx, msg)
	case *GetBlockByHashReq:
		self.handleGetBlockByHashReq(ctx, msg)
	case *GetBlockByHeightReq:
		self.handleGetBlockByHeightReq(ctx, msg)
	case *GetHeaderByHashReq:
		self.handleGetHeaderByHashReq(ctx, msg)
	case *GetHeaderByHeightReq:
		self.handleGetHeaderByHeightReq(ctx, msg)
	case *GetCurrentBlockHashReq:
		self.handleGetCurrentBlockHashReq(ctx, msg)
	case *GetCurrentBlockHeightReq:
		self.handleGetCurrentBlockHeightReq(ctx, msg)
	case *GetCurrentHeaderHeightReq:
		self.handleGetCurrentHeaderHeightReq(ctx, msg)
	case *GetCurrentHeaderHashReq:
		self.handleGetCurrentHeaderHashReq(ctx, msg)
	case *GetBlockHashReq:
		self.handleGetBlockHashReq(ctx, msg)
	case *IsCcntmainBlockReq:
		self.handleIsCcntmainBlockReq(ctx, msg)
	case *GetCcntmractStateReq:
		self.handleGetCcntmractStateReq(ctx, msg)
	case *GetMerkleProofReq:
		self.handleGetMerkleProofReq(ctx, msg)
	case *GetStorageItemReq:
		self.handleGetStorageItemReq(ctx, msg)
	case *GetBookkeeperStateReq:
		self.handleGetBookkeeperStateReq(ctx, msg)
	case *GetCurrentStateRootReq:
		self.handleGetCurrentStateRootReq(ctx, msg)
	case *IsCcntmainTransactionReq:
		self.handleIsCcntmainTransactionReq(ctx, msg)
	case *GetTransactionWithHeightReq:
		self.handleGetTransactionWithHeightReq(ctx, msg)
	case *GetBlockRootWithNewTxRootReq:
		self.handleGetBlockRootWithNewTxRootReq(ctx, msg)
	case *PreExecuteCcntmractReq:
		self.handlePreExecuteCcntmractReq(ctx, msg)
	case *GetEventNotifyByTxReq:
		self.handleGetEventNotifyByTx(ctx, msg)
	case *GetEventNotifyByBlockReq:
		self.handleGetEventNotifyByBlock(ctx, msg)
	default:
		log.Warnf("LedgerActor cannot deal with type: %v %v", msg, reflect.TypeOf(msg))
	}
}

func (self *LedgerActor) handleAddHeaderReq(ctx actor.Ccntmext, req *AddHeaderReq) {
	err := ledger.DefLedger.AddHeaders([]*types.Header{req.Header})
	if ctx.Sender() != nil {
		hash := req.Header.Hash()
		resp := &AddHeaderRsp{
			BlockHash: hash,
			Error:     err,
		}
		ctx.Sender().Request(resp, ctx.Self())
	}
}

func (self *LedgerActor) handleAddHeadersReq(ctx actor.Ccntmext, req *AddHeadersReq) {
	err := ledger.DefLedger.AddHeaders(req.Headers)
	if ctx.Sender() != nil {
		hashes := make([]common.Uint256, 0, len(req.Headers))
		for _, header := range req.Headers {
			hash := header.Hash()
			hashes = append(hashes, hash)
		}
		resp := &AddHeadersRsp{
			BlockHashes: hashes,
			Error:       err,
		}
		ctx.Sender().Request(resp, ctx.Self())
	}
}

func (self *LedgerActor) handleAddBlockReq(ctx actor.Ccntmext, req *AddBlockReq) {
	err := ledger.DefLedger.AddBlock(req.Block)
	if ctx.Sender() != nil {
		hash := req.Block.Hash()
		resp := &AddBlockRsp{
			BlockHash: hash,
			Error:     err,
		}
		ctx.Sender().Request(resp, ctx.Self())
	}
}

func (self *LedgerActor) handleGetTransactionReq(ctx actor.Ccntmext, req *GetTransactionReq) {
	tx, err := ledger.DefLedger.GetTransaction(req.TxHash)
	resp := &GetTransactionRsp{
		Error: err,
		Tx:    tx,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (self *LedgerActor) handleGetBlockByHashReq(ctx actor.Ccntmext, req *GetBlockByHashReq) {
	block, err := ledger.DefLedger.GetBlockByHash(req.BlockHash)
	resp := &GetBlockByHashRsp{
		Error: err,
		Block: block,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (self *LedgerActor) handleGetBlockByHeightReq(ctx actor.Ccntmext, req *GetBlockByHeightReq) {
	block, err := ledger.DefLedger.GetBlockByHeight(req.Height)
	resp := &GetBlockByHeightRsp{
		Error: err,
		Block: block,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (self *LedgerActor) handleGetHeaderByHashReq(ctx actor.Ccntmext, req *GetHeaderByHashReq) {
	header, err := ledger.DefLedger.GetHeaderByHash(req.BlockHash)
	resp := &GetHeaderByHashRsp{
		Error:  err,
		Header: header,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (self *LedgerActor) handleGetHeaderByHeightReq(ctx actor.Ccntmext, req *GetHeaderByHeightReq) {
	header, err := ledger.DefLedger.GetHeaderByHeight(req.Height)
	resp := &GetHeaderByHeightRsp{
		Error:  err,
		Header: header,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (self *LedgerActor) handleGetCurrentBlockHashReq(ctx actor.Ccntmext, req *GetCurrentBlockHashReq) {
	curBlockHash := ledger.DefLedger.GetCurrentBlockHash()
	resp := &GetCurrentBlockHashRsp{
		BlockHash: curBlockHash,
		Error:     nil,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (self *LedgerActor) handleGetCurrentBlockHeightReq(ctx actor.Ccntmext, req *GetCurrentBlockHeightReq) {
	curBlockHeight := ledger.DefLedger.GetCurrentBlockHeight()
	resp := &GetCurrentBlockHeightRsp{
		Height: curBlockHeight,
		Error:  nil,
	}
	ctx.Sender().Request(resp, ctx.Sender())
}

func (self *LedgerActor) handleGetCurrentHeaderHeightReq(ctx actor.Ccntmext, req *GetCurrentHeaderHeightReq) {
	curHeaderHeight := ledger.DefLedger.GetCurrentHeaderHeight()
	resp := &GetCurrentHeaderHeightRsp{
		Height: curHeaderHeight,
		Error:  nil,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (self *LedgerActor) handleGetCurrentHeaderHashReq(ctx actor.Ccntmext, req *GetCurrentHeaderHashReq) {
	curHeaderHash := ledger.DefLedger.GetCurrentHeaderHash()
	resp := &GetCurrentHeaderHashRsp{
		BlockHash: curHeaderHash,
		Error:     nil,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (self *LedgerActor) handleGetBlockHashReq(ctx actor.Ccntmext, req *GetBlockHashReq) {
	hash := ledger.DefLedger.GetBlockHash(req.Height)
	resp := &GetBlockHashRsp{
		BlockHash: hash,
		Error:     nil,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (self *LedgerActor) handleIsCcntmainBlockReq(ctx actor.Ccntmext, req *IsCcntmainBlockReq) {
	con, err := ledger.DefLedger.IsCcntmainBlock(req.BlockHash)
	resp := &IsCcntmainBlockRsp{
		IsCcntmain: con,
		Error:     err,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (self *LedgerActor) handleGetCcntmractStateReq(ctx actor.Ccntmext, req *GetCcntmractStateReq) {
	state, err := ledger.DefLedger.GetCcntmractState(req.CcntmractHash)
	resp := &GetCcntmractStateRsp{
		CcntmractState: state,
		Error:         err,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (self *LedgerActor) handleGetMerkleProofReq(ctx actor.Ccntmext, req *GetMerkleProofReq) {
	state, err := ledger.DefLedger.GetMerkleProof(req.ProofHeight, req.RootHeight)
	resp := &GetMerkleProofRsp{
		Proof: state,
		Error: err,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (self *LedgerActor) handleGetBlockRootWithNewTxRootReq(ctx actor.Ccntmext, req *GetBlockRootWithNewTxRootReq) {
	newRoot := ledger.DefLedger.GetBlockRootWithNewTxRoot(req.TxRoot)
	resp := &GetBlockRootWithNewTxRootRsp{
		NewTxRoot: newRoot,
		Error:     nil,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (self *LedgerActor) handleGetTransactionWithHeightReq(ctx actor.Ccntmext, req *GetTransactionWithHeightReq) {
	tx, height, err := ledger.DefLedger.GetTransactionWithHeight(req.TxHash)
	resp := &GetTransactionWithHeightRsp{
		Tx:     tx,
		Height: height,
		Error:  err,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (self *LedgerActor) handleGetCurrentStateRootReq(ctx actor.Ccntmext, req *GetCurrentStateRootReq) {
	stateRoot, err := ledger.DefLedger.GetCurrentStateRoot()
	resp := &GetCurrentStateRootRsp{
		StateRoot: stateRoot,
		Error:     err,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (self *LedgerActor) handleGetBookkeeperStateReq(ctx actor.Ccntmext, req *GetBookkeeperStateReq) {
	bookKeep, err := ledger.DefLedger.GetBookkeeperState()
	resp := &GetBookkeeperStateRsp{
		BookKeepState: bookKeep,
		Error:         err,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (self *LedgerActor) handleGetStorageItemReq(ctx actor.Ccntmext, req *GetStorageItemReq) {
	value, err := ledger.DefLedger.GetStorageItem(req.CodeHash, req.Key)
	resp := &GetStorageItemRsp{
		Value: value,
		Error: err,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (self *LedgerActor) handleIsCcntmainTransactionReq(ctx actor.Ccntmext, req *IsCcntmainTransactionReq) {
	isCon, err := ledger.DefLedger.IsCcntmainTransaction(req.TxHash)
	resp := &IsCcntmainTransactionRsp{
		IsCcntmain: isCon,
		Error:     err,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (self *LedgerActor) handlePreExecuteCcntmractReq(ctx actor.Ccntmext, req *PreExecuteCcntmractReq) {
	result, err := ledger.DefLedger.PreExecuteCcntmract(req.Tx)
	resp := &PreExecuteCcntmractRsp{
		Result: result,
		Error:  err,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (self *LedgerActor) handleGetEventNotifyByTx(ctx actor.Ccntmext, req *GetEventNotifyByTxReq) {
	result, err := ledger.DefLedger.GetEventNotifyByTx(req.Tx)
	resp := &GetEventNotifyByTxRsp{
		Notifies: result,
		Error:    err,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (self *LedgerActor) handleGetEventNotifyByBlock(ctx actor.Ccntmext, req *GetEventNotifyByBlockReq) {
	result, err := ledger.DefLedger.GetEventNotifyByBlock(req.Height)
	resp := &GetEventNotifyByBlockRsp{
		TxHashes: result,
		Error:    err,
	}
	ctx.Sender().Request(resp, ctx.Self())
}
