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

package ledgerstore

import (
	"crypto/sha256"
	"fmt"
	"github.com/Ontology/common"
	"github.com/Ontology/core/payload"
	"github.com/Ontology/core/types"
	"github.com/Ontology/crypto"
	"testing"
	"time"
)

func TestVersion(t *testing.T) {
	err := testBlockStore.NewBatch()
	if err != nil {
		t.Errorf("NewBatch error %s", err)
		return
	}
	version := byte(1)
	err = testBlockStore.SaveVersion(version)
	if err != nil {
		t.Errorf("SaveVersion error %s", err)
		return
	}
	err = testBlockStore.CommitTo()
	if err != nil {
		t.Errorf("CommitTo error %s", err)
		return
	}
	v, err := testBlockStore.GetVersion()
	if err != nil {
		t.Errorf("GetVersion error %s", err)
		return
	}
	if version != v {
		t.Errorf("TestVersion failed version %d != %d", v, version)
		return
	}
}

func TestCurrentBlock(t *testing.T) {
	blockHash := common.Uint256(sha256.Sum256([]byte("123456789")))
	blockHeight := uint32(1)
	err := testBlockStore.NewBatch()
	if err != nil {
		t.Errorf("NewBatch error %s", err)
		return
	}
	err = testBlockStore.SaveCurrentBlock(blockHeight, blockHash)
	if err != nil {
		t.Errorf("SaveCurrentBlock error %s", err)
		return
	}
	err = testBlockStore.CommitTo()
	if err != nil {
		t.Errorf("CommitTo error %s", err)
		return
	}
	hash, height, err := testBlockStore.GetCurrentBlock()
	if hash != blockHash {
		t.Errorf("TestCurrentBlock BlockHash %x != %x", hash, blockHash)
		return
	}
	if height != blockHeight {
		t.Errorf("TestCurrentBlock BlockHeight %x != %x", height, blockHeight)
		return
	}
}

func TestBlockHash(t *testing.T) {
	blockHash := common.Uint256(sha256.Sum256([]byte("123456789")))
	blockHeight := uint32(1)
	err := testBlockStore.NewBatch()
	if err != nil {
		t.Errorf("NewBatch error %s", err)
		return
	}
	err = testBlockStore.SaveBlockHash(blockHeight, blockHash)
	if err != nil {
		t.Errorf("SaveBlockHash error %s", err)
		return
	}
	blockHash = common.Uint256(sha256.Sum256([]byte("234567890")))
	blockHeight = uint32(2)
	err = testBlockStore.SaveBlockHash(blockHeight, blockHash)
	if err != nil {
		t.Errorf("SaveBlockHash error %s", err)
		return
	}
	err = testBlockStore.CommitTo()
	if err != nil {
		t.Errorf("CommitTo error %s", err)
		return
	}
	hash, err := testBlockStore.GetBlockHash(blockHeight)
	if err != nil {
		t.Errorf("GetBlockHash error %s", err)
		return
	}
	if hash != blockHash {
		t.Errorf("TestBlockHash failed BlockHash %x != %x", hash, blockHash)
		return
	}
}

func TestSaveTransaction(t *testing.T) {
	bookKeepingPayload := &payload.BookKeeping{
		Nonce: uint64(time.Now().UnixNano()),
	}
	tx := &types.Transaction{
		TxType:     types.BookKeeping,
		Payload:    bookKeepingPayload,
		Attributes: []*types.TxAttribute{},
	}
	blockHeight := uint32(1)
	txHash := tx.Hash()

	exist, err := testBlockStore.CcntmainTransaction(txHash)
	if err != nil {
		t.Errorf("CcntmainTransaction error %s", err)
		return
	}
	if exist {
		t.Errorf("TestSaveTransaction CcntmainTransaction should be false.")
		return
	}

	err = testBlockStore.NewBatch()
	if err != nil {
		t.Errorf("NewBatch error %s", err)
		return
	}
	err = testBlockStore.SaveTransaction(tx, blockHeight)
	if err != nil {
		t.Errorf("SaveTransaction error %s", err)
		return
	}
	err = testBlockStore.CommitTo()
	if err != nil {
		t.Errorf("CommitTo error %s", err)
		return
	}

	tx1, height, err := testBlockStore.GetTransaction(txHash)
	if err != nil {
		t.Errorf("GetTransaction error %s", err)
		return
	}
	if blockHeight != height {
		t.Errorf("TestSaveTransaction failed BlockHeight %d != %d", height, blockHeight)
		return
	}
	if tx1.TxType != tx.TxType {
		t.Errorf("TestSaveTransaction failed TxType %d != %d", tx1.TxType, tx.TxType)
		return
	}
	tx1Hash := tx1.Hash()
	if txHash != tx1Hash {
		t.Errorf("TestSaveTransaction failed TxHash %x != %x", tx1Hash, txHash)
		return
	}

	exist, err = testBlockStore.CcntmainTransaction(txHash)
	if err != nil {
		t.Errorf("CcntmainTransaction error %s", err)
		return
	}
	if !exist {
		t.Errorf("TestSaveTransaction CcntmainTransaction should be true.")
		return
	}
}

func TestHeaderIndexList(t *testing.T) {
	err := testBlockStore.NewBatch()
	if err != nil {
		t.Errorf("NewBatch error %s", err)
		return
	}
	startHeight := uint32(0)
	size := uint32(100)
	indexMap := make(map[uint32]common.Uint256, size)
	indexList := make([]common.Uint256, 0)
	for i := startHeight; i < size; i++ {
		hash := common.Uint256(sha256.Sum256([]byte(fmt.Sprintf("%v", i))))
		indexMap[i] = hash
		indexList = append(indexList, hash)
	}
	err = testBlockStore.SaveHeaderIndexList(startHeight, indexList)
	if err != nil {
		t.Errorf("SaveHeaderIndexList error %s", err)
		return
	}
	startHeight = uint32(100)
	size = uint32(100)
	indexMap = make(map[uint32]common.Uint256, size)
	for i := startHeight; i < size; i++ {
		hash := common.Uint256(sha256.Sum256([]byte(fmt.Sprintf("%v", i))))
		indexMap[i] = hash
		indexList = append(indexList, hash)
	}
	err = testBlockStore.CommitTo()
	if err != nil {
		t.Errorf("CommitTo error %s", err)
		return
	}

	totalMap, err := testBlockStore.GetHeaderIndexList()
	if err != nil {
		t.Errorf("GetHeaderIndexList error %s", err)
		return
	}

	for height, hash := range indexList {
		h, ok := totalMap[uint32(height)]
		if !ok {
			t.Errorf("TestHeaderIndexList failed height:%d hash not exist", height)
			return
		}
		if hash != h {
			t.Errorf("TestHeaderIndexList failed height:%d hash %x != %x", height, hash, h)
			return
		}
	}
}

func TestSaveHeader(t *testing.T) {
	_, pubKey1, _ := crypto.GenKeyPair()
	_, pubKey2, _ := crypto.GenKeyPair()
	bookKeeper, err := types.AddressFromBookKeepers([]*crypto.PubKey{&pubKey1, &pubKey2})
	if err != nil {
		t.Errorf("AddressFromBookKeepers error %s", err)
		return
	}
	header := &types.Header{
		Version:          123,
		PrevBlockHash:    common.Uint256{},
		TransactionsRoot: common.Uint256{},
		Timestamp:        uint32(uint32(time.Date(2017, time.February, 23, 0, 0, 0, 0, time.UTC).Unix())),
		Height:           uint32(1),
		ConsensusData:    123456789,
		NextBookKeeper:   bookKeeper,
	}
	tx1 := &types.Transaction{
		TxType: types.BookKeeping,
		Payload: &payload.BookKeeping{
			Nonce: 123456789,
		},
		Attributes: []*types.TxAttribute{},
	}
	tx2 := &types.Transaction{
		TxType: types.Enrollment,
		Payload: &payload.Enrollment{
			PublicKey: nil,
		},
		Attributes: []*types.TxAttribute{},
	}
	block := &types.Block{
		Header:       header,
		Transactions: []*types.Transaction{tx1, tx2},
	}
	blockHash := block.Hash()
	sysFee := common.Fixed64(1)

	testBlockStore.NewBatch()

	err = testBlockStore.SaveHeader(block, sysFee)
	if err != nil {
		t.Errorf("SaveHeader error %s", err)
		return
	}
	err = testBlockStore.CommitTo()
	if err != nil {
		t.Errorf("CommitTo error %s", err)
		return
	}

	h, err := testBlockStore.GetHeader(blockHash)
	if err != nil {
		t.Errorf("GetHeader error %s", err)
		return
	}

	headerHash := h.Hash()
	if blockHash != headerHash {
		t.Errorf("TestSaveHeader failed HeaderHash %x != %x", headerHash, blockHash)
		return
	}

	if header.Height != h.Height {
		t.Errorf("TestSaveHeader failed Height %d != %d", h.Height, header.Height)
		return
	}

	fee, err := testBlockStore.GetSysFeeAmount(blockHash)
	if err != nil {
		t.Errorf("TestSaveHeader SysFee %d != %d", fee, sysFee)
		return
	}
}

func TestBlock(t *testing.T) {
	_, pubKey1, _ := crypto.GenKeyPair()
	_, pubKey2, _ := crypto.GenKeyPair()
	bookKeeper, err := types.AddressFromBookKeepers([]*crypto.PubKey{&pubKey1, &pubKey2})
	if err != nil {
		t.Errorf("AddressFromBookKeepers error %s", err)
		return
	}
	header := &types.Header{
		Version:          123,
		PrevBlockHash:    common.Uint256{},
		TransactionsRoot: common.Uint256{},
		Timestamp:        uint32(uint32(time.Date(2017, time.February, 23, 0, 0, 0, 0, time.UTC).Unix())),
		Height:           uint32(2),
		ConsensusData:    1234567890,
		NextBookKeeper:   bookKeeper,
	}
	tx1 := &types.Transaction{
		TxType: types.BookKeeping,
		Payload: &payload.BookKeeping{
			Nonce: 1234567890,
		},
		Attributes: []*types.TxAttribute{},
	}
	tx2 := &types.Transaction{
		TxType: types.BookKeeping,
		Payload: &payload.BookKeeping{
			Nonce: 1234567890,
		},
		Attributes: []*types.TxAttribute{},
	}
	block := &types.Block{
		Header:       header,
		Transactions: []*types.Transaction{tx1, tx2},
	}
	blockHash := block.Hash()
	tx1Hash := tx1.Hash()
	tx2Hash := tx2.Hash()

	testBlockStore.NewBatch()

	err = testBlockStore.SaveBlock(block)
	if err != nil {
		t.Errorf("SaveHeader error %s", err)
		return
	}
	err = testBlockStore.CommitTo()
	if err != nil {
		t.Errorf("CommitTo error %s", err)
		return
	}

	b, err := testBlockStore.GetBlock(blockHash)
	if err != nil {
		t.Errorf("GetBlock error %s", err)
		return
	}

	hash := b.Hash()
	if hash != blockHash {
		t.Errorf("TestBlock failed BlockHash %x != %x ", hash, blockHash)
		return
	}
	exist, err := testBlockStore.CcntmainTransaction(tx1Hash)
	if err != nil {
		t.Errorf("CcntmainTransaction error %s", err)
		return
	}
	if !exist {
		t.Errorf("TestBlock failed transaction %x should exist", tx1Hash)
		return
	}
	exist, err = testBlockStore.CcntmainTransaction(tx2Hash)
	if err != nil {
		t.Errorf("CcntmainTransaction error %s", err)
		return
	}
	if !exist {
		t.Errorf("TestBlock failed transaction %x should exist", tx2Hash)
		return
	}

	if len(block.Transactions) != len(b.Transactions) {
		t.Errorf("TestBlock failed Transaction size %d != %d ", len(b.Transactions), len(block.Transactions))
		return
	}
	if b.Transactions[0].Hash() != tx1Hash {
		t.Errorf("TestBlock failed transaction1 hash %x != %x", b.Transactions[0].Hash(), tx1Hash)
		return
	}
	if b.Transactions[1].Hash() != tx2Hash {
		t.Errorf("TestBlock failed transaction2 hash %x != %x", b.Transactions[1].Hash(), tx2Hash)
		return
	}
}
