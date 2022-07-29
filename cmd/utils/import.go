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

package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/cntmio/cntmology/common/log"
	"github.com/cntmio/cntmology/common/serialization"
	"github.com/cntmio/cntmology/core/ledger"
	"github.com/cntmio/cntmology/core/types"
	"os"
)

func ImportBlocks(importFile string, targetHeight uint32) error {
	currBlockHeight := ledger.DefLedger.GetCurrentBlockHeight()
	if targetHeight > 0 && currBlockHeight >= targetHeight {
		log.Infof("No blocks to import.")
		return nil
	}

	ifile, err := os.OpenFile(importFile, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer ifile.Close()
	fReader := bufio.NewReader(ifile)

	metadata := NewExportBlockMetadata()
	err = metadata.Deserialize(fReader)
	if err != nil {
		return fmt.Errorf("Metadata deserialize error:%s", err)
	}
	endBlockHeight := metadata.BlockHeight
	if endBlockHeight <= currBlockHeight {
		log.Infof("No blocks to import.\n")
		return nil
	}
	if targetHeight == 0 {
		targetHeight = endBlockHeight
	}
	if targetHeight < endBlockHeight {
		endBlockHeight = targetHeight
	}

	log.Infof("Start import blocks")
	log.Infof("Current block height:%d TotalBlocks:%d", currBlockHeight, endBlockHeight-currBlockHeight)

	for i := uint32(0); i <= endBlockHeight; i++ {
		size, err := serialization.ReadUint32(fReader)
		if err != nil {
			return fmt.Errorf("Read block height:%d error:%s", i, err)
		}
		compressData := make([]byte, size)
		_, err = fReader.Read(compressData)
		if err != nil {
			return fmt.Errorf("Read block data height:%d error:%s", i, err)
		}
		if i <= currBlockHeight {
			ccntminue
		}

		blockData, err := DecompressBlockData(compressData, metadata.CompressType)
		if err != nil {
			return fmt.Errorf("block height:%d decompress error:%s", i, err)
		}

		block := &types.Block{}
		err = block.Deserialize(bytes.NewReader(blockData))
		if err != nil {
			return fmt.Errorf("block height:%d deserialize error:%s", i, err)
		}

		err = ledger.DefLedger.AddBlock(block)
		if err != nil {
			return fmt.Errorf("add block height:%d error:%s", i, err)
		}
	}
	log.Infof("Import block complete, current block height:%d", ledger.DefLedger.GetCurrentBlockHeight())
	return nil
}