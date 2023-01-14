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
package common

import (
	"fmt"

	"github.com/cntmio/cntmology/common"
)

type ToMerkleValue struct {
	TxHash      []byte
	FromChainID uint64
	MakeTxParam *MakeTxParam
}

func (this *ToMerkleValue) Deserialization(source *common.ZeroCopySource) error {
	txHash, _, irr, eof := source.NextVarBytes()
	if eof || irr {
		return fmt.Errorf("MerkleValue deserialize txHash error")
	}
	fromChainID, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("MerkleValue deserialize fromChainID error")
	}

	makeTxParam := new(MakeTxParam)
	err := makeTxParam.Deserialization(source)
	if err != nil {
		return fmt.Errorf("MerkleValue deserialize makeTxParam error:%s", err)
	}

	this.TxHash = txHash
	this.FromChainID = fromChainID
	this.MakeTxParam = makeTxParam
	return nil
}

type MakeTxParam struct {
	TxHash              []byte
	CrossChainID        []byte
	FromCcntmractAddress []byte
	ToChainID           uint64
	ToCcntmractAddress   []byte
	Method              string
	Args                []byte
}

func (this *MakeTxParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes(this.TxHash)
	sink.WriteVarBytes(this.CrossChainID)
	sink.WriteVarBytes(this.FromCcntmractAddress)
	sink.WriteUint64(this.ToChainID)
	sink.WriteVarBytes(this.ToCcntmractAddress)
	sink.WriteVarBytes([]byte(this.Method))
	sink.WriteVarBytes(this.Args)
}

func (this *MakeTxParam) Deserialization(source *common.ZeroCopySource) error {
	txHash, _, irr, eof := source.NextVarBytes()
	if eof || irr {
		return fmt.Errorf("MakeTxParam deserialize txHash error")
	}
	crossChainID, _, irr, eof := source.NextVarBytes()
	if eof || irr {
		return fmt.Errorf("MakeTxParam deserialize crossChainID error")
	}
	fromCcntmractAddress, _, irr, eof := source.NextVarBytes()
	if eof || irr {
		return fmt.Errorf("MakeTxParam deserialize fromCcntmractAddress error")
	}
	toChainID, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("MakeTxParam deserialize toChainID error")
	}
	toCcntmractAddress, _, irr, eof := source.NextVarBytes()
	if eof || irr {
		return fmt.Errorf("MakeTxParam deserialize toCcntmractAddress error")
	}
	method, _, irr, eof := source.NextString()
	if eof || irr {
		return fmt.Errorf("MakeTxParam deserialize method error")
	}
	args, _, irr, eof := source.NextVarBytes()
	if eof || irr {
		return fmt.Errorf("MakeTxParam deserialize args error")
	}

	this.TxHash = txHash
	this.CrossChainID = crossChainID
	this.FromCcntmractAddress = fromCcntmractAddress
	this.ToChainID = toChainID
	this.ToCcntmractAddress = toCcntmractAddress
	this.Method = method
	this.Args = args
	return nil
}
