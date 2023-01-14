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
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/cntmio/cntmology-crypto/keypair"
	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/config"
	"github.com/cntmio/cntmology/common/constants"
	vconfig "github.com/cntmio/cntmology/consensus/vbft/config"
	"github.com/cntmio/cntmology/core/payload"
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

var GenBlockTime = config.DEFAULT_GEN_BLOCK_TIME * time.Second

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
	conf := common.NewZeroCopySink(nil)
	if genesisConfig.VBFT != nil {
		err := genesisConfig.VBFT.Serialization(conf)
		if err != nil {
			return nil, err
		}
	}
	govConfig := newGoverConfigInit(conf.Bytes())
	consensusPayload, err := vconfig.GenesisConsensusPayload(govConfig.Hash(), 0)
	if err != nil {
		return nil, fmt.Errorf("consensus genesis init failed: %s", err)
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
	govConfigTx := newGovConfigTx()

	genesisBlock := &types.Block{
		Header: genesisHeader,
		Transactions: []*types.Transaction{
			cntm,
			cntm,
			param,
			oid,
			auth,
			govConfigTx,
			newGoverningInit(),
			newUtilityInit(),
			newParamInit(),
		},
	}
	genesisBlock.RebuildMerkleRoot()
	return genesisBlock, nil
}

func newGoverningToken() *types.Transaction {
	mutable, err := utils.NewDeployTransaction(nutils.OntCcntmractAddress[:], "cntm", "1.0",
		"Ontology Team", "ccntmact@cntm.io", "Ontology Network cntm Token", payload.NEOVM_TYPE)
	if err != nil {
		panic("[NewDeployTransaction] construct genesis governing token transaction error ")
	}
	tx, err := mutable.IntoImmutable()
	if err != nil {
		panic("construct genesis governing token transaction error ")
	}
	return tx
}

func newUtilityToken() *types.Transaction {
	mutable, err := utils.NewDeployTransaction(nutils.OngCcntmractAddress[:], "cntm", "1.0",
		"Ontology Team", "ccntmact@cntm.io", "Ontology Network cntm Token", payload.NEOVM_TYPE)
	if err != nil {
		panic("[NewDeployTransaction] construct genesis governing token transaction error ")
	}
	tx, err := mutable.IntoImmutable()
	if err != nil {
		panic("construct genesis utility token transaction error ")
	}
	return tx
}

func newParamCcntmract() *types.Transaction {
	mutable, err := utils.NewDeployTransaction(nutils.ParamCcntmractAddress[:],
		"ParamConfig", "1.0", "Ontology Team", "ccntmact@cntm.io",
		"Chain Global Environment Variables Manager ", payload.NEOVM_TYPE)
	if err != nil {
		panic("[NewDeployTransaction] construct genesis governing token transaction error ")
	}
	tx, err := mutable.IntoImmutable()
	if err != nil {
		panic("construct genesis param transaction error ")
	}
	return tx
}

func newGovConfigTx() *types.Transaction {
	mutable, err := utils.NewDeployTransaction(nutils.GovernanceCcntmractAddress[:], "CONFIG", "1.0",
		"Ontology Team", "ccntmact@cntm.io", "Ontology Network Consensus Config", payload.NEOVM_TYPE)
	if err != nil {
		panic("[NewDeployTransaction] construct genesis governing token transaction error ")
	}
	tx, err := mutable.IntoImmutable()
	if err != nil {
		panic("construct genesis config transaction error ")
	}
	return tx
}

func deployAuthCcntmract() *types.Transaction {
	mutable, err := utils.NewDeployTransaction(nutils.AuthCcntmractAddress[:], "AuthCcntmract", "1.0",
		"Ontology Team", "ccntmact@cntm.io", "Ontology Network Authorization Ccntmract", payload.NEOVM_TYPE)
	if err != nil {
		panic("[NewDeployTransaction] construct genesis governing token transaction error ")
	}
	tx, err := mutable.IntoImmutable()
	if err != nil {
		panic("construct genesis auth transaction error ")
	}
	return tx
}

func deployOntIDCcntmract() *types.Transaction {
	mutable, err := utils.NewDeployTransaction(nutils.OntIDCcntmractAddress[:], "OID", "1.0",
		"Ontology Team", "ccntmact@cntm.io", "Ontology Network cntm ID", payload.NEOVM_TYPE)
	if err != nil {
		panic("[NewDeployTransaction] construct genesis governing token transaction error ")
	}
	tx, err := mutable.IntoImmutable()
	if err != nil {
		panic("construct genesis cntmid transaction error ")
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

	args := common.NewZeroCopySink(nil)
	nutils.EncodeVarUint(args, uint64(len(distribute)))
	for _, part := range distribute {
		nutils.EncodeAddress(args, part.addr)
		nutils.EncodeVarUint(args, part.value)
	}

	mutable := utils.BuildNativeTransaction(nutils.OntCcntmractAddress, cntm.INIT_NAME, args.Bytes())
	tx, err := mutable.IntoImmutable()
	if err != nil {
		panic("construct genesis governing token transaction error ")
	}
	return tx
}

func newUtilityInit() *types.Transaction {
	mutable := utils.BuildNativeTransaction(nutils.OngCcntmractAddress, cntm.INIT_NAME, []byte{})
	tx, err := mutable.IntoImmutable()
	if err != nil {
		panic("construct genesis utility token transaction error ")
	}

	return tx
}

func newParamInit() *types.Transaction {
	params := new(global_params.Params)
	var s []string
	for k := range INIT_PARAM {
		s = append(s, k)
	}

	for k, v := range neovm.INIT_GAS_TABLE {
		INIT_PARAM[k] = strconv.FormatUint(v, 10)
		s = append(s, k)
	}

	sort.Strings(s)
	for _, v := range s {
		params.SetParam(global_params.Param{Key: v, Value: INIT_PARAM[v]})
	}
	sink := common.NewZeroCopySink(nil)
	params.Serialization(sink)

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
	nutils.EncodeAddress(sink, addr)

	mutable := utils.BuildNativeTransaction(nutils.ParamCcntmractAddress, global_params.INIT_NAME, sink.Bytes())
	tx, err := mutable.IntoImmutable()
	if err != nil {
		panic("construct genesis governing token transaction error ")
	}
	return tx
}

func newGoverConfigInit(config []byte) *types.Transaction {
	mutable := utils.BuildNativeTransaction(nutils.GovernanceCcntmractAddress, governance.INIT_CONFIG, config)
	tx, err := mutable.IntoImmutable()
	if err != nil {
		panic("construct genesis governing token transaction error ")
	}
	return tx
}
