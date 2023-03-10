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

package test

import (
	"testing"

	"github.com/conntectome/cntm/common"
	"github.com/conntectome/cntm/smartcontract/service/native/cross_chain/cross_chain_manager"
	"github.com/stretchr/testify/assert"
)

func TestCreateCrossChainTxParam(t *testing.T) {
	param := cross_chain_manager.CreateCrossChainTxParam{
		ToChainID:         1,
		ToContractAddress: []byte{1, 2, 3, 4},
		Method:            "test",
		Args:              []byte{1, 2, 3, 4},
	}
	sink := common.NewZeroCopySink(nil)
	param.Serialization(sink)

	var p cross_chain_manager.CreateCrossChainTxParam
	err := p.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)

	assert.Equal(t, p, param)
}

func TestProcessCrossChainTxParam(t *testing.T) {
	param := cross_chain_manager.ProcessCrossChainTxParam{
		Address:     common.ADDRESS_EMPTY,
		FromChainID: 1,
		Height:      2,
		Proof:       "test",
		Header:      []byte{1, 2, 3, 4},
	}

	sink := common.NewZeroCopySink(nil)
	param.Serialization(sink)

	var p cross_chain_manager.ProcessCrossChainTxParam
	err := p.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)

	assert.Equal(t, param, p)
}

func TestCntgUnlockParam(t *testing.T) {
	param := cross_chain_manager.CntgUnlockParam{
		FromChainID: 1,
		Address:     common.ADDRESS_EMPTY,
		Amount:      1,
	}
	sink := common.NewZeroCopySink(nil)
	param.Serialization(sink)

	var p cross_chain_manager.CntgUnlockParam
	err := p.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	assert.NoError(t, err)
	assert.Equal(t, param, p)
}
