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
 * along with The cntm.  If not, see <http://www.gnu.org/licenses/>.
 */

package lock_proxy

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/conntectome/cntm/common"
	"github.com/conntectome/cntm/common/constants"
	"github.com/conntectome/cntm/smartcontract/service/native/utils"
	"github.com/stretchr/testify/assert"
)

func TestLockParam_Serialize(t *testing.T) {
	fromAddr, _ := common.AddressFromBase58("709c937270e1d5a490718a2b4a230186bdd06a01")
	toAddrBs, _ := hex.DecodeString("709c937270e1d5a490718a2b4a230186bdd06a02")
	param := LockParam{
		SourceAssetHash: utils.CntmContractAddress,
		ToChainID:       0,
		FromAddress:     fromAddr,
		ToAddress:       toAddrBs,
		Value:           1,
	}
	sink := common.NewZeroCopySink(nil)
	param.Serialization(sink)

	param2 := LockParam{}
	source := common.NewZeroCopySource(sink.Bytes())
	if err := param2.Deserialization(source); err != nil {
		t.Fatal("LockParam deserialize fail!")
	}
	assert.Equal(t, param, param2)
}

func TestUnlockParam_Serialize(t *testing.T) {
	param := UnlockParam{
		ArgsBs:             []byte{1, 2, 3, 0, 100},
		FromContractHashBs: utils.CntmContractAddress[:],
		FromChainId:        2,
	}
	sink := common.NewZeroCopySink(nil)
	param.Serialization(sink)

	param2 := UnlockParam{}
	source := common.NewZeroCopySource(sink.Bytes())
	if err := param2.Deserialization(source); err != nil {
		t.Fatal("LockParam deserialize fail!")
	}
	assert.Equal(t, param, param2)
}

func TestArgs_Serialize(t *testing.T) {
	toAddr, _ := hex.DecodeString("709c937270e1d5a490718a2b4a230186bdd06a02")
	args := Args{
		TargetAssetHash: utils.CntmContractAddress[:],
		ToAddress:       toAddr,
		Value:           100,
	}
	sink := common.NewZeroCopySink(nil)
	args.Serialization(sink)

	args2 := Args{}
	source := common.NewZeroCopySource(sink.Bytes())
	if err := args2.Deserialization(source); err != nil {
		t.Fatal("Args deserialize fail!")
	}
	assert.Equal(t, args, args2)
}

func TestBindAssetParam_Serialize(t *testing.T) {
	bindAssetParam := BindAssetParam{
		SourceAssetHash:    utils.CntmContractAddress,
		TargetChainId:      uint64(0),
		TargetAssetHash:    utils.CntmContractAddress[:],
		Limit:              big.NewInt(int64(constants.CNTM_TOTAL_SUPPLY)),
		IsTargetChainAsset: false,
	}
	sink := common.NewZeroCopySink(nil)
	bindAssetParam.Serialization(sink)

	bindAssetParam2 := BindAssetParam{}
	source := common.NewZeroCopySource(sink.Bytes())
	if err := bindAssetParam2.Deserialization(source); err != nil {
		t.Fatal("BindAssetParam deserialize fail!")
	}
	assert.Equal(t, bindAssetParam, bindAssetParam2)
}
