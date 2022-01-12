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

func (this *LedgerActor) Start() *actor.PID {
	this.props = actor.FromProducer(func() actor.Actor { return this })
	var err error
	DefLedgerPid, err = actor.SpawnNamed(this.props, "LedgerActor")
	if err != nil {
		panic(fmt.Errorf("LedgerActor SpawnNamed error:%s", err))
	}
	return DefLedgerPid
}

func (this *LedgerActor) Receive(ctx actor.Ccntmext) {
	switch msg := ctx.Message().(type) {
	case *actor.Started:
	case *actor.Stop:
	case *AddHeaderReq:
		this.handleAddHeaderReq(ctx, msg)
	case *AddHeadersReq:
		this.handleAddHeadersReq(ctx, msg)
	case *AddBlockReq:
		this.handleAddBlockReq(ctx, msg)
	case *GetTransactionReq:
		this.handleGetTransactionReq(ctx, msg)
	case *GetBlockByHashReq:
		this.handleGetBlockByHashReq(ctx, msg)
	case *GetBlockByHeightReq:
		this.handleGetBlockByHeightReq(ctx, msg)
	case *GetHeaderByHashReq:
		this.handleGetHeaderByHashReq(ctx, msg)
	case *GetHeaderByHeightReq:
		this.handleGetHeaderByHeightReq(ctx, msg)
	case *GetCurrentBlockHashReq:
		this.handleGetCurrentBlockHashReq(ctx, msg)
	case *GetCurrentBlockHeightReq:
		this.handleGetCurrentBlockHeightReq(ctx, msg)
	case *GetCurrentHeaderHeightReq:
		this.handleGetCurrentHeaderHeightReq(ctx, msg)
	case *GetCurrentHeaderHashReq:
		this.handleGetCurrentHeaderHashReq(ctx, msg)
	case *GetBlockHashReq:
		this.handleGetBlockHashReq(ctx, msg)
	case *IsCcntmainBlockReq:
		this.handleIsCcntmainBlockReq(ctx, msg)
	case *GetCcntmractStateReq:
		this.handleGetCcntmractStateReq(ctx, msg)
	case *GetStorageItemReq:
		this.handleGetStorageItemReq(ctx, msg)
	case *GetBookkeeperStateReq:
		this.handleGetBookkeeperStateReq(ctx, msg)
	case *GetCurrentStateRootReq:
		this.handleGetCurrentStateRootReq(ctx, msg)
	case *IsCcntmainTransactionReq:
		this.handleIsCcntmainTransactionReq(ctx, msg)
	case *GetTransactionWithHeightReq:
		this.handleGetTransactionWithHeightReq(ctx, msg)
	case *GetBlockRootWithNewTxRootReq:
		this.handleGetBlockRootWithNewTxRootReq(ctx, msg)
	case *PreExecuteCcntmractReq:
		this.handlePreExecuteCcntmractReq(ctx, msg)
	case *GetEventNotifyByTxReq:
		this.handleGetEventNotifyByTx(ctx, msg)
	case *GetEventNotifyByBlockReq:
		this.handleGetEventNotifyByBlock(ctx, msg)
	default:
		log.Warnf("LedgerActor cannot deal with type: %v %v", msg, reflect.TypeOf(msg))
	}
}

func (this *LedgerActor) handleAddHeaderReq(ctx actor.Ccntmext, req *AddHeaderReq) {
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

func (this *LedgerActor) handleAddHeadersReq(ctx actor.Ccntmext, req *AddHeadersReq) {
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

func (this *LedgerActor) handleAddBlockReq(ctx actor.Ccntmext, req *AddBlockReq) {
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

func (this *LedgerActor) handleGetTransactionReq(ctx actor.Ccntmext, req *GetTransactionReq) {
	tx, err := ledger.DefLedger.GetTransaction(req.TxHash)
	resp := &GetTransactionRsp{
		Error: err,
		Tx:    tx,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (this *LedgerActor) handleGetBlockByHashReq(ctx actor.Ccntmext, req *GetBlockByHashReq) {
	block, err := ledger.DefLedger.GetBlockByHash(req.BlockHash)
	resp := &GetBlockByHashRsp{
		Error: err,
		Block: block,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (this *LedgerActor) handleGetBlockByHeightReq(ctx actor.Ccntmext, req *GetBlockByHeightReq) {
	block, err := ledger.DefLedger.GetBlockByHeight(req.Height)
	resp := &GetBlockByHeightRsp{
		Error: err,
		Block: block,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (this *LedgerActor) handleGetHeaderByHashReq(ctx actor.Ccntmext, req *GetHeaderByHashReq) {
	header, err := ledger.DefLedger.GetHeaderByHash(req.BlockHash)
	resp := &GetHeaderByHashRsp{
		Error:  err,
		Header: header,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (this *LedgerActor) handleGetHeaderByHeightReq(ctx actor.Ccntmext, req *GetHeaderByHeightReq) {
	header, err := ledger.DefLedger.GetHeaderByHeight(req.Height)
	resp := &GetHeaderByHeightRsp{
		Error:  err,
		Header: header,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (this *LedgerActor) handleGetCurrentBlockHashReq(ctx actor.Ccntmext, req *GetCurrentBlockHashReq) {
	curBlockHash := ledger.DefLedger.GetCurrentBlockHash()
	resp := &GetCurrentBlockHashRsp{
		BlockHash: curBlockHash,
		Error:     nil,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (this *LedgerActor) handleGetCurrentBlockHeightReq(ctx actor.Ccntmext, req *GetCurrentBlockHeightReq) {
	curBlockHeight := ledger.DefLedger.GetCurrentBlockHeight()
	resp := &GetCurrentBlockHeightRsp{
		Height: curBlockHeight,
		Error:  nil,
	}
	ctx.Sender().Request(resp, ctx.Sender())
}

func (this *LedgerActor) handleGetCurrentHeaderHeightReq(ctx actor.Ccntmext, req *GetCurrentHeaderHeightReq) {
	curHeaderHeight := ledger.DefLedger.GetCurrentHeaderHeight()
	resp := &GetCurrentHeaderHeightRsp{
		Height: curHeaderHeight,
		Error:  nil,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (this *LedgerActor) handleGetCurrentHeaderHashReq(ctx actor.Ccntmext, req *GetCurrentHeaderHashReq) {
	curHeaderHash := ledger.DefLedger.GetCurrentHeaderHash()
	resp := &GetCurrentHeaderHashRsp{
		BlockHash: curHeaderHash,
		Error:     nil,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (this *LedgerActor) handleGetBlockHashReq(ctx actor.Ccntmext, req *GetBlockHashReq) {
	hash := ledger.DefLedger.GetBlockHash(req.Height)
	resp := &GetBlockHashRsp{
		BlockHash: hash,
		Error:     nil,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (this *LedgerActor) handleIsCcntmainBlockReq(ctx actor.Ccntmext, req *IsCcntmainBlockReq) {
	con, err := ledger.DefLedger.IsCcntmainBlock(req.BlockHash)
	resp := &IsCcntmainBlockRsp{
		IsCcntmain: con,
		Error:     err,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (this *LedgerActor) handleGetCcntmractStateReq(ctx actor.Ccntmext, req *GetCcntmractStateReq) {
	state, err := ledger.DefLedger.GetCcntmractState(req.CcntmractHash)
	resp := &GetCcntmractStateRsp{
		CcntmractState: state,
		Error:         err,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (this *LedgerActor) handleGetBlockRootWithNewTxRootReq(ctx actor.Ccntmext, req *GetBlockRootWithNewTxRootReq) {
	newRoot := ledger.DefLedger.GetBlockRootWithNewTxRoot(req.TxRoot)
	resp := &GetBlockRootWithNewTxRootRsp{
		NewTxRoot: newRoot,
		Error:     nil,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (this *LedgerActor) handleGetTransactionWithHeightReq(ctx actor.Ccntmext, req *GetTransactionWithHeightReq) {
	tx, height, err := ledger.DefLedger.GetTransactionWithHeight(req.TxHash)
	resp := &GetTransactionWithHeightRsp{
		Tx:     tx,
		Height: height,
		Error:  err,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (this *LedgerActor) handleGetCurrentStateRootReq(ctx actor.Ccntmext, req *GetCurrentStateRootReq) {
	stateRoot, err := ledger.DefLedger.GetCurrentStateRoot()
	resp := &GetCurrentStateRootRsp{
		StateRoot: stateRoot,
		Error:     err,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (this *LedgerActor) handleGetBookkeeperStateReq(ctx actor.Ccntmext, req *GetBookkeeperStateReq) {
	bookKeep, err := ledger.DefLedger.GetBookkeeperState()
	resp := &GetBookkeeperStateRsp{
		BookKeepState: bookKeep,
		Error:         err,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (this *LedgerActor) handleGetStorageItemReq(ctx actor.Ccntmext, req *GetStorageItemReq) {
	value, err := ledger.DefLedger.GetStorageItem(req.CodeHash, req.Key)
	resp := &GetStorageItemRsp{
		Value: value,
		Error: err,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (this *LedgerActor) handleIsCcntmainTransactionReq(ctx actor.Ccntmext, req *IsCcntmainTransactionReq) {
	isCon, err := ledger.DefLedger.IsCcntmainTransaction(req.TxHash)
	resp := &IsCcntmainTransactionRsp{
		IsCcntmain: isCon,
		Error:     err,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (this *LedgerActor) handlePreExecuteCcntmractReq(ctx actor.Ccntmext, req *PreExecuteCcntmractReq) {
	result, err := ledger.DefLedger.PreExecuteCcntmract(req.Tx)
	resp := &PreExecuteCcntmractRsp{
		Result: result,
		Error:  err,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (this *LedgerActor) handleGetEventNotifyByTx(ctx actor.Ccntmext, req *GetEventNotifyByTxReq) {
	result, err := ledger.DefLedger.GetEventNotifyByTx(req.Tx)
	resp := &GetEventNotifyByTxRsp{
		Notifies: result,
		Error:    err,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (this *LedgerActor) handleGetEventNotifyByBlock(ctx actor.Ccntmext, req *GetEventNotifyByBlockReq) {
	result, err := ledger.DefLedger.GetEventNotifyByBlock(req.Height)
	resp := &GetEventNotifyByBlockRsp{
		TxHashes: result,
		Error:    err,
	}
	ctx.Sender().Request(resp, ctx.Self())
}
