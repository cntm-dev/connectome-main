/*
 * Copyright (C) 2018 The cntm Authors
 * This file is part of The cntm library.
 *
 * The cntm is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The cntm is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * alcntg with The cntm.  If not, see <http://www.gnu.org/licenses/>.
 */

package genesis

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/conntectome/cntm-crypto/keypair"
	"github.com/conntectome/cntm/common"
	"github.com/conntectome/cntm/common/config"
	"github.com/conntectome/cntm/common/constants"
	vconfig "github.com/conntectome/cntm/consensus/Cbft/config"
	"github.com/conntectome/cntm/core/payload"
	"github.com/conntectome/cntm/core/types"
	"github.com/conntectome/cntm/core/utils"
	"github.com/conntectome/cntm/smartcontract/service/native/global_params"
	"github.com/conntectome/cntm/smartcontract/service/native/governance"
	"github.com/conntectome/cntm/smartcontract/service/native/cntm"
	nutils "github.com/conntectome/cntm/smartcontract/service/native/utils"
	"github.com/conntectome/cntm/smartcontract/service/cntmvm"
)

const (
	BlockVersion uint32 = 0
	GenesisNonce uint64 = 2083236893
)

var (
	CNTMToken   = newGoverningToken()
	CNTGToken   = newUtilityToken()
	CNTMTokenID = CNTMToken.Hash()
	CNTGTokenID = CNTGToken.Hash()
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
	if genesisConfig.Cbft != nil {
		err := genesisConfig.Cbft.Serialization(conf)
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
	cntg := newUtilityToken()
	param := newParamCcntmract()
	oid := deploycntmidCcntmract()
	auth := deployAuthCcntmract()
	govConfigTx := newGovConfigTx()

	genesisBlock := &types.Block{
		Header: genesisHeader,
		Transactions: []*types.Transaction{
			cntm,
			cntg,
			param,
			oid,
			auth,
			govConfigTx,
			newGoverningInit(),
			newUtilityInit(),
			newParamInit(),
			govConfig,
		},
	}
	genesisBlock.RebuildMerkleRoot()
	return genesisBlock, nil
}

func newGoverningToken() *types.Transaction {
	mutable, err := utils.NewDeployTransaction(nutils.CntmCcntmractAddress[:], "CNTM", "1.0",
		"Cntm Team", "ccntmact@cntm.io", "Cntm Network CNTM Token", payload.CNTMVM_TYPE)
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
	mutable, err := utils.NewDeployTransaction(nutils.CntgCcntmractAddress[:], "CNTG", "1.0",
		"Cntm Team", "ccntmact@cntm.io", "Cntm Network CNTG Token", payload.CNTMVM_TYPE)
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
		"ParamConfig", "1.0", "Cntm Team", "ccntmact@cntm.io",
		"Chain Global Environment Variables Manager ", payload.CNTMVM_TYPE)
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
		"Cntm Team", "ccntmact@cntm.io", "Cntm Network Consensus Config", payload.CNTMVM_TYPE)
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
		"Cntm Team", "ccntmact@cntm.io", "Cntm Network Authorization Ccntmract", payload.CNTMVM_TYPE)
	if err != nil {
		panic("[NewDeployTransaction] construct genesis governing token transaction error ")
	}
	tx, err := mutable.IntoImmutable()
	if err != nil {
		panic("construct genesis auth transaction error ")
	}
	return tx
}

func deploycntmidCcntmract() *types.Transaction {
	mutable, err := utils.NewDeployTransaction(nutils.cntmidCcntmractAddress[:], "OID", "1.0",
		"Cntm Team", "ccntmact@cntm.io", "Cntm Network CNTM ID", payload.CNTMVM_TYPE)
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
			panic(fmt.Sprint("wrcntg bookkeeper config, caused by", err))
		}
		addr = temp
	}

	distribute := []struct {
		addr  common.Address
		value uint64
	}{{addr, constants.CNTM_TOTAL_SUPPLY}}

	args := common.NewZeroCopySink(nil)
	nutils.EncodeVarUint(args, uint64(len(distribute)))
	for _, part := range distribute {
		nutils.EncodeAddress(args, part.addr)
		nutils.EncodeVarUint(args, part.value)
	}

	mutable := utils.BuildNativeTransaction(nutils.CntmCcntmractAddress, cntm.INIT_NAME, args.Bytes())
	tx, err := mutable.IntoImmutable()
	if err != nil {
		panic("construct genesis governing token transaction error ")
	}
	return tx
}

func newUtilityInit() *types.Transaction {
	mutable := utils.BuildNativeTransaction(nutils.CntgCcntmractAddress, cntm.INIT_NAME, []byte{})
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

	for k, v := range cntmvm.INIT_GAS_TABLE {
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
			panic(fmt.Sprint("wrcntg bookkeeper config, caused by", err))
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
