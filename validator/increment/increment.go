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

package increment

import (
	"fmt"

	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/log"
	"github.com/cntmio/cntmology/core/types"
)

type IncrementValidator struct {
	blocks     []map[common.Uint256]bool
	baseHeight uint32
	maxBlocks  int
}

func NewIncrementValidator(maxBlocks int) *IncrementValidator {
	if maxBlocks <= 0 {
		maxBlocks = 2
	}
	return &IncrementValidator{
		maxBlocks: maxBlocks,
	}
}

func (self *IncrementValidator) Clean() {
	self.blocks = nil
	self.baseHeight = 0
}

// ccntmain block in range [start, end)
func (self *IncrementValidator) BlockRange() (start uint32, end uint32) {
	return self.baseHeight, self.baseHeight + uint32(len(self.blocks))
}

func (self *IncrementValidator) AddBlock(block *types.Block) {
	if len(self.blocks) == 0 {
		self.baseHeight = block.Header.Height
	}

	if self.baseHeight+uint32(len(self.blocks)) != block.Header.Height {
		start, end := self.BlockRange()
		log.Error("disccntminue block is not allowed: [start, end)=[%d, %d), block height= %d",
			start, end, block.Header.Height)
		return
	}

	if len(self.blocks) >= self.maxBlocks {
		self.blocks = self.blocks[1:]
		self.baseHeight += 1
	}
	txHashes := make(map[common.Uint256]bool)
	for _, tx := range block.Transactions {
		txHashes[tx.Hash()] = true
	}
	self.blocks = append(self.blocks, txHashes)
}

func (self *IncrementValidator) Verify(tx *types.Transaction, startHeight uint32) error {
	if startHeight < self.baseHeight {
		return fmt.Errorf("can not do increment validation: startHeight %v < self.baseHeight %v", startHeight, self.baseHeight)
	}

	for i := int(startHeight - self.baseHeight); i < len(self.blocks); i++ {
		if _, ok := self.blocks[i][tx.Hash()]; ok {
			return fmt.Errorf("tx duplicated")
		}
	}

	return nil
}
