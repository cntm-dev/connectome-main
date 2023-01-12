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

package vbft

import (
	"bytes"
	"fmt"

	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/config"
	"github.com/cntmio/cntmology/common/log"
	"github.com/cntmio/cntmology/core/genesis"
	"github.com/cntmio/cntmology/core/ledger"
	"github.com/cntmio/cntmology/core/states"
	gov "github.com/cntmio/cntmology/smartccntmract/service/native"
	govcon "github.com/cntmio/cntmology/smartccntmract/service/native/states"
)

type ChainStore struct {
	db              *ledger.Ledger
	chainedBlockNum uint32
	pendingBlocks   map[uint32]*Block
}

func OpenBlockStore(db *ledger.Ledger) (*ChainStore, error) {
	return &ChainStore{
		db:              db,
		chainedBlockNum: db.GetCurrentBlockHeight(),
		pendingBlocks:   make(map[uint32]*Block),
	}, nil
}

func (self *ChainStore) close() {
	// TODO: any action on ledger actor??
}

func (self *ChainStore) GetChainedBlockNum() uint32 {
	return self.chainedBlockNum
}

func (self *ChainStore) getExecMerkleRoot(blkNum uint32) (common.Uint256, error) {
	if blk, present := self.pendingBlocks[blkNum]; blk != nil && present {
		return blk.execResult.MerkleRoot, nil
	}
	merkleRoot, err := self.db.GetStateMerkleRoot(blkNum)
	if err != nil {
		log.Infof("GetStateMerkleRoot blockNum:%d, error :%s", blkNum, err)
		return common.Uint256{}, fmt.Errorf("GetStateMerkleRoot blockNum:%d, error :%s", blkNum, err)
	} else {
		return merkleRoot, nil
	}
}

func (self *ChainStore) getCrossStatesRoot(blkNum uint32) (common.Uint256, error) {
	if blk, present := self.pendingBlocks[blkNum]; blk != nil && present {
		return blk.execResult.CrossStatesRoot, nil
	}
	statesRoot, err := self.db.GetCrossStatesRoot(blkNum)
	if err != nil {
		log.Infof("getCrossStatesRoot blockNum:%d, error :%s", blkNum, err)
		return common.UINT256_EMPTY, fmt.Errorf("getCrossStatesRoot blockNum:%d, error :%s", blkNum, err)
	} else {
		return statesRoot, nil
	}
}

func (self *ChainStore) getExecWriteSet(blkNum uint32) *overlaydb.MemDB {
	if blk, present := self.pendingBlocks[blkNum]; blk != nil && present {
		return blk.execResult.WriteSet
	}
	return nil
}

func (self *ChainStore) ReloadFromLedger() {
	height := self.db.GetCurrentBlockHeight()
	if height > self.chainedBlockNum {
		// update chainstore height
		self.chainedBlockNum = height
		// remove persisted pending blocks
		newPending := make(map[uint32]*PendingBlock)
		for blkNum, blk := range self.pendingBlocks {
			if blkNum > height {
				newPending[blkNum] = blk
			}
		}
		log.Debug("chainstore ReloadFromLedger pendingBlocks")
		// update pending blocks
		self.pendingBlocks = newPending
	}
}

func (self *ChainStore) AddBlock(block *Block) error {
	if block == nil {
		return fmt.Errorf("try add nil block")
	}

	if block.getBlockNum() <= self.GetChainedBlockNum() {
		log.Warnf("chain store adding chained block(%d, %d)", block.getBlockNum(), self.GetChainedBlockNum())
		return nil
	}

	if block.Block.Header == nil {
		panic("nil block header")
	}
	blkNum := self.GetChainedBlockNum() + 1
	err := self.submitBlock(blkNum - 1)
	if err != nil {
		log.Errorf("chainstore blkNum:%d, SubmitBlock: %s", blkNum-1, err)
	}
	execResult, err := self.db.ExecuteBlock(block.Block)
	if err != nil {
		log.Errorf("chainstore AddBlock GetBlockExecResult: %s", err)
		return fmt.Errorf("chainstore AddBlock GetBlockExecResult: %s", err)
	}
	log.Debugf("execResult:%+v, AddBlock execResult height:%d \n", execResult, block.Block.Header.Height)
	log.Debugf("chainstore addblock pendingBlocks height:%d,block height:%d", blkNum, block.getBlockNum())
	self.pendingBlocks[blkNum] = &PendingBlock{block: block, execResult: &execResult, hasSubmitted: false}

	if self.pid != nil {
		self.pid.Tell(
			&message.BlockConsensusComplete{
				Block: block.Block,
			})
	}
	self.chainedBlockNum = blkNum
	return nil
}

func (self *ChainStore) submitBlock(blkNum uint32) error {
	if blkNum == 0 {
		return nil
	}
	if submitBlk, present := self.pendingBlocks[blkNum]; submitBlk != nil && submitBlk.hasSubmitted == false && present {
		err := self.db.SubmitBlock(submitBlk.block.Block, submitBlk.block.CrossChainMsg, *submitBlk.execResult)
		if err != nil {
			return fmt.Errorf("ledger add submitBlk (%d, %d, %d) failed: %s", blkNum, self.GetChainedBlockNum(), self.db.GetCurrentBlockHeight(), err)
		}
		if _, present := self.pendingBlocks[blkNum-1]; present {
			delete(self.pendingBlocks, blkNum-1)
		}
		submitBlk.hasSubmitted = true
	}
	return nil
}

func (self *ChainStore) getBlock(blockNum uint32) (*Block, error) {
	if blk, present := self.pendingBlocks[blockNum]; present {
		return blk.block, nil
	}
	block, err := self.db.GetBlockByHeight(blockNum)
	if err != nil {
		return nil, err
	}
	prevMerkleRoot := common.Uint256{}
	var crossChainMsg *types.CrossChainMsg
	if blockNum > 1 {
		prevMerkleRoot, err = self.db.GetStateMerkleRoot(blockNum - 1)
		if err != nil {
			log.Errorf("GetStateMerkleRoot blockNum:%d, error :%s", blockNum, err)
			return nil, fmt.Errorf("GetStateMerkleRoot blockNum:%d, error :%s", blockNum, err)
		}
		crossChainMsg, err = self.db.GetCrossChainMsg(blockNum - 1)
		if err != nil {
			log.Errorf("GetCrossChainMsg blockNum:%d, error :%s", blockNum, err)
			return nil, fmt.Errorf("v blockNum:%d, error :%s", blockNum, err)
		}
	}
	return initVbftBlock(block, crossChainMsg, prevMerkleRoot)
}

func (self *ChainStore) GetVbftConfigInfo() (*config.VBFTConfig, error) {
	storageKey := &states.StorageKey{
		CodeHash: genesis.GovernanceCcntmractAddress,
		Key:      append([]byte(gov.VBFT_CONFIG)),
	}
	vbft, err := ledger.DefLedger.GetStorageItem(storageKey.CodeHash, storageKey.Key)
	if err != nil {
		return nil, err
	}
	cfg := &govcon.Configuration{}
	if err := json.Unmarshal(vbft, cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %s", err)
	}
	chainconfig := &config.VBFTConfig{
		N:                    cfg.N,
		C:                    cfg.C,
		K:                    cfg.K,
		L:                    cfg.L,
		BlockMsgDelay:        cfg.BlockMsgDelay,
		HashMsgDelay:         cfg.HashMsgDelay,
		PeerHandshakeTimeout: cfg.PeerHandshakeTimeout,
		MaxBlockChangeView:   cfg.MaxBlockChangeView,
	}
	return chainconfig, nil
}

func (self *ChainStore) GetPeersConfig() ([]*config.VBFTPeerStakeInfo, error) {
	goveranceview, err := self.GetGovernanceView()
	if err != nil {
		return nil, err
	}
	storageKey := &states.StorageKey{
		CodeHash: genesis.GovernanceCcntmractAddress,
		Key:      append([]byte(gov.PEER_POOL), goveranceview.View.Bytes()...),
	}
	peers, err := ledger.DefLedger.GetStorageItem(storageKey.CodeHash, storageKey.Key)
	if err != nil {
		return nil, err
	}
	peerMap := &govcon.PeerPoolMap{
		PeerPoolMap: make(map[string]*govcon.PeerPool),
	}
	if err := json.Unmarshal(peers, peerMap); err != nil {
		return nil, fmt.Errorf("unmarshal peersconfig: %s", err)
	}
	var peerstakes []*config.VBFTPeerStakeInfo
	for _, id := range peerMap.PeerPoolMap {
		config := &config.VBFTPeerStakeInfo{
			Index:      uint32(id.Index),
			PeerPubkey: id.PeerPubkey,
			InitPos:    id.InitPos + id.TotalPos,
		}
		peerstakes = append(peerstakes, config)
	}
	return peerstakes, nil
}

func (self *ChainStore) isForceUpdate() (bool, error) {
	goveranceview, err := self.GetGovernanceView()
	if err != nil {
		return false, err
	}
	return goveranceview.VoteCommit, nil
}

func (self *ChainStore) GetGovernanceView() (*govcon.GovernanceView, error) {
	storageKey := &states.StorageKey{
		CodeHash: genesis.GovernanceCcntmractAddress,
		Key:      append([]byte(gov.GOVERNANCE_VIEW)),
	}
	force, err := ledger.DefLedger.GetStorageItem(storageKey.CodeHash, storageKey.Key)
	if err != nil {
		return nil, err
	}
	config := &govcon.GovernanceView{}
	if err := json.Unmarshal(force, config); err != nil {
		return nil, fmt.Errorf("unmarshal GovernanceView config: %s", err)
	}
	return config, nil
}
