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
	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/core/ledger"
	"github.com/cntmio/cntmology/core/payload"
	"github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/smartccntmract/event"
)

const (
	REQ_TIMEOUT    = 5
	ERR_ACTOR_COMM = "[http] Actor comm error: %v"
)

func GetHeaderByHeight(height uint32) (*types.Header, error) {
	return ledger.DefLedger.GetHeaderByHeight(height)
}
func GetBlockByHeight(height uint32) (*types.Block, error) {
	return ledger.DefLedger.GetBlockByHeight(height)
}
func GetBlockHashFromStore(height uint32) common.Uint256 {
	return ledger.DefLedger.GetBlockHash(height)
}

func CurrentBlockHash() common.Uint256 {
	return ledger.DefLedger.GetCurrentBlockHash()
}

func GetBlockFromStore(hash common.Uint256) (*types.Block, error) {
	return ledger.DefLedger.GetBlockByHash(hash)
}

func GetCurrentBlockHeight() uint32 {
	return ledger.DefLedger.GetCurrentBlockHeight()
}

func GetTransaction(hash common.Uint256) (*types.Transaction, error) {
	return ledger.DefLedger.GetTransaction(hash)
}

func GetStorageItem(codeHash common.Address, key []byte) ([]byte, error) {
	return ledger.DefLedger.GetStorageItem(codeHash, key)
}

func GetCcntmractStateFromStore(hash common.Address) (*payload.DeployCode, error) {
	return ledger.DefLedger.GetCcntmractState(hash)
}

func GetTxnWithHeightByTxHash(hash common.Uint256) (uint32, *types.Transaction, error) {
	tx, height, err := ledger.DefLedger.GetTransactionWithHeight(hash)
	return height, tx, err
}

func PreExecuteCcntmract(tx *types.Transaction) (interface{}, error) {
	return ledger.DefLedger.PreExecuteCcntmract(tx)
}

func GetEventNotifyByTxHash(txHash common.Uint256) ([]*event.NotifyEventInfo, error) {
	return ledger.DefLedger.GetEventNotifyByTx(txHash)
}

func GetEventNotifyByHeight(height uint32) ([]common.Uint256, error) {
	return ledger.DefLedger.GetEventNotifyByBlock(height)
}

func GetMerkleProof(proofHeight uint32, rootHeight uint32) ([]common.Uint256, error) {
	return ledger.DefLedger.GetMerkleProof(proofHeight, rootHeight)
}

