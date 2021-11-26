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

package signature

import (
	"bytes"
	"crypto/sha256"
	"io"

	"github.com/Ontology/common"
	"github.com/Ontology/core/ccntmract/program"
	"github.com/Ontology/crypto"
	. "github.com/Ontology/errors"
	"github.com/Ontology/vm/neovm/interfaces"
)

//SignableData describe the data need be signed.
type SignableData interface {
	interfaces.ICodeCcntmainer

	////Get the the SignableData's program hashes
	GetProgramHashes() ([]common.Address, error)

	SetPrograms([]*program.Program)

	GetPrograms() []*program.Program

	//TODO: add SerializeUnsigned
	SerializeUnsigned(io.Writer) error
}

func SignBySigner(data SignableData, signer Signer) ([]byte, error) {
	return sign(data, signer.PrivKey())
}

func getHashData(data SignableData) []byte {
	buf := new(bytes.Buffer)
	data.SerializeUnsigned(buf)
	return buf.Bytes()
}

func sign(data SignableData, privKey []byte) ([]byte, error) {
	temp := sha256.Sum256(getHashData(data))
	hash := sha256.Sum256(temp[:])

	signature, err := crypto.Sign(privKey, hash[:])
	if err != nil {
		return nil, NewDetailErr(err, ErrNoCode, "[Signature],Sign failed.")
	}
	return signature, nil
}
