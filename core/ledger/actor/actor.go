package actor

import (
	"github.com/Ontology/common"
	"github.com/Ontology/core/ledger"
	"github.com/Ontology/core/types"
	"github.com/Ontology/eventbus/actor"
	"github.com/Ontology/common/log"
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
	DefLedgerPid = actor.Spawn(this.props)
	return DefLedgerPid
}

func (this *LedgerActor) Receive(ctx actor.Ccntmext) {
	switch msg := ctx.Message().(type) {
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
	case *GetBlockHashReq:
		this.handleGetBlockHashReq(ctx, msg)
	case *IsCcntmainBlockReq:
		this.handleIsCcntmainBlockReq(ctx, msg)
	case *GetCcntmractStateReq:
		this.handleGetCcntmractStateReq(ctx, msg)
	case *GetStorageItemReq:
		this.handleGetStorageItemReq(ctx, msg)
	case *GetBookKeeperStateReq:
		this.handleGetBookKeeperStateReq(ctx, msg)
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
	default:
		log.Warnf("LedgerActor cannot deal with type: %v", msg)
	}
}

func (this *LedgerActor) handleAddHeaderReq(ctx actor.Ccntmext, req *AddHeaderReq) {
	err := ledger.DefLedger.AddHeaders([]*types.Header{req.Header})
	hash := req.Header.Hash()
	resp := &AddHeaderRsp{
		BlockHash: &hash,
		Error:     err,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (this *LedgerActor) handleAddHeadersReq(ctx actor.Ccntmext, req *AddHeadersReq) {
	err := ledger.DefLedger.AddHeaders(req.Headers)
	hashes := make([]*common.Uint256, 0, len(req.Headers))
	for _, header := range req.Headers {
		hash := header.Hash()
		hashes = append(hashes, &hash)
	}
	resp := &AddHeadersRsp{
		BlockHashes: hashes,
		Error:       err,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (this *LedgerActor) handleAddBlockReq(ctx actor.Ccntmext, req *AddBlockReq) {
	err := ledger.DefLedger.AddBlock(req.Block)
	hash := req.Block.Hash()
	resp := &AddBlockRsp{
		BlockHash: &hash,
		Error:     err,
	}
	ctx.Sender().Request(resp, ctx.Self())
}

func (this *LedgerActor) handleGetTransactionReq(ctx actor.Ccntmext, req *GetTransactionReq) {
	tx, err := ledger.DefLedger.GetTransaction(req.TxHash)
	resp := GetTransactionRsp{
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

func (this *LedgerActor) handleGetBookKeeperStateReq(ctx actor.Ccntmext, req *GetBookKeeperStateReq) {
	bookKeep, err := ledger.DefLedger.GetBookKeeperState()
	resp := &GetBookKeeperStateRsp{
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

func (this *LedgerActor)handlePreExecuteCcntmractReq(ctx actor.Ccntmext, req *PreExecuteCcntmractReq){
	result, err := ledger.DefLedger.PreExecuteCcntmract(req.Tx)
	resp := PreExecuteCcntmractRsp{
		Result:result,
		Error:err,
	}
	ctx.Sender().Request(resp, ctx.Self())
}