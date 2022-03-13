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

package genesis

import (
	"bytes"
	"errors"
	"fmt"
	"time"

	"github.com/cntmio/cntmology-crypto/keypair"
	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/config"
	vconfig "github.com/Ontology/consensus/vbft/config"
	"github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/core/utils"
	"github.com/cntmio/cntmology/smartccntmract/states"
	stypes "github.com/cntmio/cntmology/smartccntmract/types"
)

const (
	BlockVersion uint32 = 0
	GenesisNonce uint64 = 2083236893
)

var (
	OntCcntmractAddress, _ = common.AddressParseFromBytes([]byte{0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01})
	OngCcntmractAddress, _ = common.AddressParseFromBytes([]byte{0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02})

	cntmToken   = newGoverningToken()
	cntmToken   = newUtilityToken()
	cntmTokenID = cntmToken.Hash()
	cntmTokenID = cntmToken.Hash()
)

var GenBlockTime = (config.DEFAULT_GEN_BLOCK_TIME * time.Second)

var GenesisBookkeepers []keypair.PublicKey

// GenesisBlockInit returns the genesis block with default consensus bookkeeper list
func GenesisBlockInit(defaultBookkeeper []keypair.PublicKey) (*types.Block, error) {
	//getBookkeeper
	GenesisBookkeepers = defaultBookkeeper
	nextBookkeeper, err := types.AddressFromBookkeepers(defaultBookkeeper)
	if err != nil {
		return nil, errors.New("[Block],GenesisBlockInit err with GetBookkeeperAddress")
	}

	consensusPayload, err := vconfig.GenesisConsensusPayload()
	if err != nil {
		return nil, fmt.Errorf("consensus genesus init failed: %s", err)
	}

	//blockdata
	genesisHeader := &types.Header{
		Version:          BlockVersion,
		PrevBlockHash:    common.Uint256{},
		TransactionsRoot: common.Uint256{},
		Timestamp:        uint32(uint32(time.Date(2017, time.February, 23, 0, 0, 0, 0, time.UTC).Unix())),
		Height:           uint32(0),
		ConsensusData:    GenesisNonce,
		NextBookkeeper:   nextBookkeeper,
		ConsensusPayload: consensusPayload,

		Bookkeepers: nil,
		SigData:     nil,
	}

	//block
	cntm := newGoverningToken()
	cntm := newUtilityToken()

	genesisBlock := &types.Block{
		Header: genesisHeader,
		Transactions: []*types.Transaction{
			cntm,
			cntm,
			newGoverningInit(),
			newUtilityInit(),
		},
	}
	genesisBlock.RebuildMerkleRoot()
	return genesisBlock, nil
}

func newGoverningToken() *types.Transaction {
	tx := utils.NewDeployTransaction(stypes.VmCode{Code: OntCcntmractAddress[:], VmType: stypes.Native}, "cntm", "1.0",
		"Ontology Team", "ccntmact@cntm.io", "Ontology Network cntm Token", true)
	return tx
}

func newUtilityToken() *types.Transaction {
	tx := utils.NewDeployTransaction(stypes.VmCode{Code: OngCcntmractAddress[:], VmType: stypes.Native}, "cntm", "1.0",
		"Ontology Team", "ccntmact@cntm.io", "Ontology Network cntm Token", true)
	return tx
}

func newGoverningInit() *types.Transaction {
	init := states.Ccntmract{
		Address: OntCcntmractAddress,
		Method:  "init",
	}
	bf := new(bytes.Buffer)
	init.Serialize(bf)
	vmCode := stypes.VmCode{
		VmType: stypes.Native,
		Code:   bf.Bytes(),
	}
	tx := utils.NewInvokeTransaction(vmCode)
	return tx
}

func newUtilityInit() *types.Transaction {
	init := states.Ccntmract{
		Address: OngCcntmractAddress,
		Method:  "init",
	}
	bf := new(bytes.Buffer)
	init.Serialize(bf)
	vmCode := stypes.VmCode{
		VmType: stypes.Native,
		Code:   bf.Bytes(),
	}
	tx := utils.NewInvokeTransaction(vmCode)
	return tx
}
