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
	"errors"
	"time"
	"bytes"

	"github.com/Ontology/common"
	"github.com/Ontology/common/config"
	"github.com/Ontology/core/types"
	"github.com/Ontology/core/utils"
	vmtypes "github.com/Ontology/vm/types"
	"github.com/cntmio/cntmology-crypto/keypair"
	"github.com/Ontology/smartccntmract/service/native/states"
)

const (
	BlockVersion      uint32 = 0
	GenesisNonce      uint64 = 2083236893

	OntRegisterAmount = 1000000000
	OngRegisterAmount = 1000000000
)

var (
	OntCcntmractAddress, _ = common.AddressParseFromBytes([]byte{0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01})
	OngCcntmractAddress, _ = common.AddressParseFromBytes([]byte{0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02})

	cntmToken   = NewGoverningToken()
	cntmToken   = NewUtilityToken()
	cntmTokenID = cntmToken.Hash()
	cntmTokenID = cntmToken.Hash()
)

var GenBlockTime = (config.DEFAULTGENBLOCKTIME * time.Second)

var GenesisBookkeepers []keypair.PublicKey

func GenesisBlockInit(defaultBookkeeper []keypair.PublicKey) (*types.Block, error) {
	//getBookKeeper
	GenesisBookkeepers = defaultBookkeeper
	nextBookkeeper, err := types.AddressFromBookkeepers(defaultBookkeeper)
	if err != nil {
		return nil, errors.New("[Block],GenesisBlockInit err with GetBookkeeperAddress")
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

		Bookkeepers: nil,
		SigData:     nil,
	}

	//block
	cntm := NewGoverningToken()
	cntm := NewUtilityToken()

	genesisBlock := &types.Block{
		Header: genesisHeader,
		Transactions: []*types.Transaction{
			cntm,
			cntm,
			NewGoverningInit(),
			NewUtilityInit(),
		},
	}
	return genesisBlock, nil
}

func NewGoverningToken() *types.Transaction {
	tx := utils.NewDeployTransaction(&vmtypes.VmCode{Code: OntCcntmractAddress[:], VmType: vmtypes.Native}, "cntm", "1.0",
		"Ontology Team", "ccntmact@cntm.io", "Ontology Network cntm Token", true)
	return tx
}

func NewUtilityToken() *types.Transaction {
	tx := utils.NewDeployTransaction(&vmtypes.VmCode{Code: OngCcntmractAddress[:], VmType: vmtypes.Native}, "cntm", "1.0",
		"Ontology Team", "ccntmact@cntm.io", "Ontology Network cntm Token", true)
	return tx
}

func NewGoverningInit() *types.Transaction {
	init := states.Ccntmract{
		Address: OntCcntmractAddress,
		Method: "init",
		Args: []byte{},
	}
	bf := new(bytes.Buffer)
	init.Serialize(bf)
	vmCode := vmtypes.VmCode{
		VmType: vmtypes.Native,
		Code: bf.Bytes(),
	}
	tx := utils.NewInvokeTransaction(vmCode)
	return tx
}

func NewUtilityInit() *types.Transaction {
	init := states.Ccntmract{
		Address: OngCcntmractAddress,
		Method: "init",
		Args: []byte{},
	}
	bf := new(bytes.Buffer)
	init.Serialize(bf)
	vmCode := vmtypes.VmCode{
		VmType: vmtypes.Native,
		Code: bf.Bytes(),
	}
	tx := utils.NewInvokeTransaction(vmCode)
	return tx
}
