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

package store

import (
	"github.com/cntmio/cntmology-crypto/keypair"
	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/core/payload"
	"github.com/cntmio/cntmology/core/states"
	"github.com/cntmio/cntmology/core/store/overlaydb"
	"github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/smartccntmract/event"
	cstates "github.com/cntmio/cntmology/smartccntmract/states"
	"github.com/cntmio/cntmology/smartccntmract/storage"
)

type ExecuteResult struct {
	WriteSet        *overlaydb.MemDB
	Hash            common.Uint256
	MerkleRoot      common.Uint256
	CrossStates     []common.Uint256
	CrossStatesRoot common.Uint256
	Notify          []*event.ExecuteNotify
}

// LedgerStore provides func with store package.
type LedgerStore interface {
	InitLedgerStoreWithGenesisBlock(genesisblock *types.Block, defaultBookkeeper []keypair.PublicKey) error
	Close() error
	AddHeaders(headers []*types.Header) error
	AddBlock(block *types.Block, ccMsg *types.CrossChainMsg, stateMerkleRoot common.Uint256) error
	ExecuteBlock(b *types.Block) (ExecuteResult, error)                                       // called by consensus
	SubmitBlock(b *types.Block, crossChainMsg *types.CrossChainMsg, exec ExecuteResult) error // called by consensus
	GetStateMerkleRoot(height uint32) (result common.Uint256, err error)
	GetCurrentBlockHash() common.Uint256
	GetCurrentBlockHeight() uint32
	GetCurrentHeaderHeight() uint32
	GetCurrentHeaderHash() common.Uint256
	GetBlockHash(height uint32) common.Uint256
	GetHeaderByHash(blockHash common.Uint256) (*types.Header, error)
	GetRawHeaderByHash(blockHash common.Uint256) (*types.RawHeader, error)
	GetHeaderByHeight(height uint32) (*types.Header, error)
	GetBlockByHash(blockHash common.Uint256) (*types.Block, error)
	GetBlockByHeight(height uint32) (*types.Block, error)
	GetTransaction(txHash common.Uint256) (*types.Transaction, uint32, error)
	IsCcntmainBlock(blockHash common.Uint256) (bool, error)
	IsCcntmainTransaction(txHash common.Uint256) (bool, error)
	GetBlockRootWithNewTxRoots(startHeight uint32, txRoots []common.Uint256) common.Uint256
	GetMerkleProof(m, n uint32) ([]common.Uint256, error)
	GetCcntmractState(ccntmractHash common.Address) (*payload.DeployCode, error)
	GetBookkeeperState() (*states.BookkeeperState, error)
	GetStorageItem(key *states.StorageKey) (*states.StorageItem, error)
	FindStorageItem(key *states.StorageKey) ([]*states.StorageItem, error)
	PreExecuteCcntmract(tx *types.Transaction) (interface{}, error)
	GetEventNotifyByTx(tx common.Uint256) (*event.ExecuteNotify, error)
	GetEventNotifyByBlock(height uint32) ([]common.Uint256, error)
}
