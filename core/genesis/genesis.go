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
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/cntmio/cntmology-crypto/keypair"
	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/config"
	"github.com/cntmio/cntmology/common/constants"
	"github.com/cntmio/cntmology/consensus/vbft/config"
	"github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/core/utils"
	"github.com/cntmio/cntmology/smartccntmract/service/native/global_params"
	"github.com/cntmio/cntmology/smartccntmract/service/native/governance"
	"github.com/cntmio/cntmology/smartccntmract/service/native/cntm"
	nutils "github.com/cntmio/cntmology/smartccntmract/service/native/utils"
	"github.com/cntmio/cntmology/smartccntmract/service/neovm"
)

const (
	BlockVersion uint32 = 0
	GenesisNonce uint64 = 2083236893
)

var (
	cntmToken   = newGoverningToken()
	cntmToken   = newUtilityToken()
	cntmTokenID = cntmToken.Hash()
	cntmTokenID = cntmToken.Hash()
)

var GenBlockTime = (config.DEFAULT_GEN_BLOCK_TIME * time.Second)

var INIT_PARAM = map[string]string{
	"gasPrice": "0",
}

var GenesisBookkeepers []keypair.PublicKey

// BuildGenesisBlock returns the genesis block with default consensus bookkeeper list
func BuildGenesisBlock(defaultBookkeeper []keypair.PublicKey, genesisConfig *config.GenesisConfig) (*types.Block, error) {
	//getBookkeeper
	GenesisBookkeepers = defaultBookkeeper
	nextBookkeeper, err := types.AddressFromBookkeepers(defaultBookkeeper)
	if err != nil {
		return nil, fmt.Errorf("[Block],BuildGenesisBlock err with GetBookkeeperAddress: %s", err)
	}
	conf := bytes.NewBuffer(nil)
	if genesisConfig.VBFT != nil {
		genesisConfig.VBFT.Serialize(conf)
	}
	govConfig := newGoverConfigInit(conf.Bytes())
	consensusPayload, err := vconfig.GenesisConsensusPayload(govConfig.Hash(), 0)
	if err != nil {
		return nil, fmt.Errorf("consensus genesus init failed: %s", err)
	}
	//blockdata
	genesisHeader := &types.Header{
		Version:          BlockVersion,
		PrevBlockHash:    common.Uint256{},
		TransactionsRoot: common.Uint256{},
		Timestamp:        constants.GENESIS_BLOCK_TIMESTAMP,
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
	param := newParamCcntmract()
	oid := deployOntIDCcntmract()
	auth := deployAuthCcntmract()
	config := newConfig()

	genesisBlock := &types.Block{
		Header: genesisHeader,
		Transactions: []*types.Transaction{
			cntm,
			cntm,
			param,
			oid,
			auth,
			config,
			newGoverningInit(),
			newUtilityInit(),
			newParamInit(),
		},
	}
	genesisBlock.RebuildMerkleRoot()
	return genesisBlock, nil
}

func newGoverningToken() *types.Transaction {
	mutable := utils.NewDeployTransaction(nutils.OntCcntmractAddress[:], "cntm", "1.0",
		"Ontology Team", "ccntmact@cntm.io", "Ontology Network cntm Token", true)
	tx, err := mutable.IntoImmutable()
	if err != nil {
		panic("constract genesis governing token transaction error ")
	}
	return tx
}

func newUtilityToken() *types.Transaction {
	mutable := utils.NewDeployTransaction(nutils.OngCcntmractAddress[:], "cntm", "1.0",
		"Ontology Team", "ccntmact@cntm.io", "Ontology Network cntm Token", true)
	tx, err := mutable.IntoImmutable()
	if err != nil {
		panic("constract genesis utility token transaction error ")
	}
	return tx
}

func newParamCcntmract() *types.Transaction {
	mutable := utils.NewDeployTransaction(nutils.ParamCcntmractAddress[:],
		"ParamConfig", "1.0", "Ontology Team", "ccntmact@cntm.io",
		"Chain Global Environment Variables Manager ", true)
	tx, err := mutable.IntoImmutable()
	if err != nil {
		panic("constract genesis param transaction error ")
	}
	return tx
}

func newConfig() *types.Transaction {
	mutable := utils.NewDeployTransaction(nutils.GovernanceCcntmractAddress[:], "CONFIG", "1.0",
		"Ontology Team", "ccntmact@cntm.io", "Ontology Network Consensus Config", true)
	tx, err := mutable.IntoImmutable()
	if err != nil {
		panic("constract genesis config transaction error ")
	}
	return tx
}

func deployAuthCcntmract() *types.Transaction {
	mutable := utils.NewDeployTransaction(nutils.AuthCcntmractAddress[:], "AuthCcntmract", "1.0",
		"Ontology Team", "ccntmact@cntm.io", "Ontology Network Authorization Ccntmract", true)
	tx, err := mutable.IntoImmutable()
	if err != nil {
		panic("constract genesis auth transaction error ")
	}
	return tx
}

func deployOntIDCcntmract() *types.Transaction {
	mutable := utils.NewDeployTransaction(nutils.OntIDCcntmractAddress[:], "OID", "1.0",
		"Ontology Team", "ccntmact@cntm.io", "Ontology Network cntm ID", true)
	tx, err := mutable.IntoImmutable()
	if err != nil {
		panic("constract genesis cntmid transaction error ")
	}
	return tx
}

func newGoverningInit() *types.Transaction {
	bookkeepers, _ := config.DefConfig.GetBookkeepers()

	var addr common.Address
	if len(bookkeepers) == 1 {
		addr = types.AddressFromPubKey(bookkeepers[0])
	} else {
		m := (5*len(bookkeepers) + 6) / 7
		temp, err := types.AddressFromMultiPubKeys(bookkeepers, m)
		if err != nil {
			panic(fmt.Sprint("wrcntm bookkeeper config, caused by", err))
		}
		addr = temp
	}

	distribute := []struct {
		addr  common.Address
		value uint64
	}{{addr, constants.cntm_TOTAL_SUPPLY}}

	args := bytes.NewBuffer(nil)
	nutils.WriteVarUint(args, uint64(len(distribute)))
	for _, part := range distribute {
		nutils.WriteAddress(args, part.addr)
		nutils.WriteVarUint(args, part.value)
	}

	mutable := utils.BuildNativeTransaction(nutils.OntCcntmractAddress, cntm.INIT_NAME, args.Bytes())
	tx, err := mutable.IntoImmutable()
	if err != nil {
		panic("constract genesis governing token transaction error ")
	}
	return tx
}

func newUtilityInit() *types.Transaction {
	mutable := utils.BuildNativeTransaction(nutils.OngCcntmractAddress, cntm.INIT_NAME, []byte{})
	tx, err := mutable.IntoImmutable()
	if err != nil {
		panic("constract genesis governing token transaction error ")
	}

	return tx
}

func newParamInit() *types.Transaction {
	params := new(global_params.Params)
	var s []string
	for k := range INIT_PARAM {
		s = append(s, k)
	}

	neovm.GAS_TABLE.Range(func(key, value interface{}) bool {
		INIT_PARAM[key.(string)] = strconv.FormatUint(value.(uint64), 10)
		s = append(s, key.(string))
		return true
	})

	sort.Strings(s)
	for _, v := range s {
		params.SetParam(global_params.Param{Key: v, Value: INIT_PARAM[v]})
	}
	bf := new(bytes.Buffer)
	params.Serialize(bf)

	bookkeepers, _ := config.DefConfig.GetBookkeepers()
	var addr common.Address
	if len(bookkeepers) == 1 {
		addr = types.AddressFromPubKey(bookkeepers[0])
	} else {
		m := (5*len(bookkeepers) + 6) / 7
		temp, err := types.AddressFromMultiPubKeys(bookkeepers, m)
		if err != nil {
			panic(fmt.Sprint("wrcntm bookkeeper config, caused by", err))
		}
		addr = temp
	}
	nutils.WriteAddress(bf, addr)

	mutable := utils.BuildNativeTransaction(nutils.ParamCcntmractAddress, global_params.INIT_NAME, bf.Bytes())
	tx, err := mutable.IntoImmutable()
	if err != nil {
		panic("constract genesis governing token transaction error ")
	}
	return tx
}

func newGoverConfigInit(config []byte) *types.Transaction {
	mutable := utils.BuildNativeTransaction(nutils.GovernanceCcntmractAddress, governance.INIT_CONFIG, config)
	tx, err := mutable.IntoImmutable()
	if err != nil {
		panic("constract genesis governing token transaction error ")
	}
	return tx
}
