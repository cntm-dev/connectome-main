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

package ccntmract

import (
	. "github.com/Ontology/common"
	pg "github.com/Ontology/core/ccntmract/program"
	"github.com/Ontology/crypto"
	. "github.com/Ontology/errors"
	vm "github.com/Ontology/vm/neovm"
	"math/big"
	"sort"
)

//create a Single Singature ccntmract for owner
func CreateSignatureCcntmract(ownerPubKey *crypto.PubKey) (*Ccntmract, error) {
	temp, err := ownerPubKey.EncodePoint(true)
	if err != nil {
		return nil, NewDetailErr(err, ErrNoCode, "[Ccntmract],CreateSignatureCcntmract failed.")
	}
	signatureRedeemScript, err := CreateSignatureRedeemScript(ownerPubKey)
	if err != nil {
		return nil, NewDetailErr(err, ErrNoCode, "[Ccntmract],CreateSignatureCcntmract failed.")
	}
	hash := ToCodeHash(temp)
	signatureRedeemScriptHashToCodeHash := ToCodeHash(signatureRedeemScript)
	return &Ccntmract{
		Code:            signatureRedeemScript,
		Parameters:      []CcntmractParameterType{Signature},
		ProgramHash:     signatureRedeemScriptHashToCodeHash,
		OwnerPubkeyHash: hash,
	}, nil
}

func CreateSignatureRedeemScript(pubkey *crypto.PubKey) ([]byte, error) {
	temp, err := pubkey.EncodePoint(true)
	if err != nil {
		return nil, NewDetailErr(err, ErrNoCode, "[Ccntmract],CreateSignatureRedeemScript failed.")
	}
	sb := pg.NewProgramBuilder()
	sb.PushData(temp)
	sb.AddOp(vm.CHECKSIG)
	return sb.ToArray(), nil
}

//create a Multi Singature ccntmract for owner  。
func CreateMultiSigCcntmract(publicKeyHash Address, m int, publicKeys []*crypto.PubKey) (*Ccntmract, error) {

	params := make([]CcntmractParameterType, m)
	for i, _ := range params {
		params[i] = Signature
	}
	MultiSigRedeemScript, err := CreateMultiSigRedeemScript(m, publicKeys)
	if err != nil {
		return nil, NewDetailErr(err, ErrNoCode, "[Ccntmract],CreateSignatureRedeemScript failed.")
	}
	signatureRedeemScriptHashToCodeHash := ToCodeHash(MultiSigRedeemScript)
	return &Ccntmract{
		Code:            MultiSigRedeemScript,
		Parameters:      params,
		ProgramHash:     signatureRedeemScriptHashToCodeHash,
		OwnerPubkeyHash: publicKeyHash,
	}, nil
}

func CreateMultiSigRedeemScript(m int, pubkeys []*crypto.PubKey) ([]byte, error) {
	if !(m >= 1 && m <= len(pubkeys) && len(pubkeys) <= 24) {
		return nil, nil //TODO: add panic
	}

	sb := pg.NewProgramBuilder()
	sb.PushNumber(big.NewInt(int64(m)))

	//sort pubkey
	sort.Sort(crypto.PubKeySlice(pubkeys))

	for _, pubkey := range pubkeys {
		temp, err := pubkey.EncodePoint(true)
		if err != nil {
			return nil, NewDetailErr(err, ErrNoCode, "[Ccntmract],CreateSignatureCcntmract failed.")
		}
		sb.PushData(temp)
	}

	sb.PushNumber(big.NewInt(int64(len(pubkeys))))
	sb.AddOp(vm.CHECKMULTISIG)
	return sb.ToArray(), nil
}
