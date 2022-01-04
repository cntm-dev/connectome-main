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
	"github.com/Ontology/common"
	"github.com/Ontology/core/payload"
	"github.com/Ontology/core/states"
	"github.com/Ontology/core/types"
	"github.com/Ontology/smartccntmract/event"
)

type AddHeaderReq struct {
	Header *types.Header
}

type AddHeaderRsp struct {
	BlockHash common.Uint256
	Error     error
}

type AddHeadersReq struct {
	Headers []*types.Header
}

type AddHeadersRsp struct {
	BlockHashes []common.Uint256
	Error       error
}

type AddBlockReq struct {
	Block *types.Block
}

type AddBlockRsp struct {
	BlockHash common.Uint256
	Error     error
}

type GetTransactionReq struct {
	TxHash common.Uint256
}

type GetTransactionRsp struct {
	Tx    *types.Transaction
	Error error
}

type GetBlockByHashReq struct {
	BlockHash common.Uint256
}

type GetBlockByHashRsp struct {
	Block *types.Block
	Error error
}

type GetBlockByHeightReq struct {
	Height uint32
}

type GetBlockByHeightRsp struct {
	Block *types.Block
	Error error
}

type GetHeaderByHashReq struct {
	BlockHash common.Uint256
}

type GetHeaderByHashRsp struct {
	Header *types.Header
	Error  error
}

type GetHeaderByHeightReq struct {
	Height uint32
}

type GetHeaderByHeightRsp struct {
	Header *types.Header
	Error  error
}

type GetCurrentBlockHashReq struct{}

type GetCurrentBlockHashRsp struct {
	BlockHash common.Uint256
	Error     error
}

type GetCurrentBlockHeightReq struct{}

type GetCurrentBlockHeightRsp struct {
	Height uint32
	Error  error
}

type GetCurrentHeaderHeightReq struct{}

type GetCurrentHeaderHeightRsp struct {
	Height uint32
	Error  error
}

type GetCurrentHeaderHashReq struct{}

type GetCurrentHeaderHashRsp struct {
	BlockHash common.Uint256
	Error     error
}

type GetBlockHashReq struct {
	Height uint32
}

type GetBlockHashRsp struct {
	BlockHash common.Uint256
	Error     error
}

type IsCcntmainBlockReq struct {
	BlockHash common.Uint256
}

type IsCcntmainBlockRsp struct {
	IsCcntmain bool
	Error     error
}

type GetBlockRootWithNewTxRootReq struct {
	TxRoot common.Uint256
}

type GetBlockRootWithNewTxRootRsp struct {
	NewTxRoot common.Uint256
	Error     error
}

type GetTransactionWithHeightReq struct {
	TxHash common.Uint256
}

type GetTransactionWithHeightRsp struct {
	Tx     *types.Transaction
	Height uint32
	Error  error
}

type IsCcntmainTransactionReq struct {
	TxHash common.Uint256
}

type IsCcntmainTransactionRsp struct {
	IsCcntmain bool
	Error     error
}

type GetCurrentStateRootReq struct{}

type GetCurrentStateRootRsp struct {
	StateRoot common.Uint256
	Error     error
}

type GetBookkeeperStateReq struct{}

type GetBookkeeperStateRsp struct {
	BookKeepState *states.BookkeeperState
	Error         error
}

type GetStorageItemReq struct {
	CodeHash common.Address
	Key      []byte
}

type GetStorageItemRsp struct {
	Value []byte
	Error error
}

type GetCcntmractStateReq struct {
	CcntmractHash common.Address
}

type GetCcntmractStateRsp struct {
	CcntmractState *payload.DeployCode
	Error         error
}

type PreExecuteCcntmractReq struct {
	Tx *types.Transaction
}

type PreExecuteCcntmractRsp struct {
	Result []interface{}
	Error  error
}

type GetEventNotifyByTxReq struct {
	Tx common.Uint256
}

type GetEventNotifyByTxRsp struct {
	Notifies []*event.NotifyEventInfo
	Error    error
}

type GetEventNotifyByBlockReq struct {
	Height uint32
}

type GetEventNotifyByBlockRsp struct {
	TxHashes []common.Uint256
	Error    error
}
