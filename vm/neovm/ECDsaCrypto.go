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

package neovm

import (
	"crypto/sha256"
	"errors"
	"github.com/Ontology/common"
	"github.com/Ontology/common/log"
	"github.com/Ontology/crypto"
	. "github.com/Ontology/errors"
)

type ECDsaCrypto struct {
}

func (c *ECDsaCrypto) Hash160(message []byte) []byte {
	temp := common.ToCodeHash(message)
	return temp[:]
}

func (c *ECDsaCrypto) Hash256(message []byte) []byte {
	temp := sha256.Sum256(message)
	f := sha256.Sum256(temp[:])
	return f[:]
}

func (c *ECDsaCrypto) VerifySignature(message []byte, signature []byte, pubkey []byte) (bool, error) {

	log.Debugf("message: %x", message)
	log.Debugf("signature: %x", signature)
	log.Debugf("pubkey: %x", pubkey)

	pk, err := crypto.DecodePoint(pubkey)
	if err != nil {
		return false, NewDetailErr(errors.New("[ECDsaCrypto], crypto.DecodePoint failed."), ErrNoCode, "")
	}

	err = crypto.Verify(*pk, message, signature)
	if err != nil {
		return false, NewDetailErr(errors.New("[ECDsaCrypto], VerifySignature failed."), ErrNoCode, "")
	}

	return true, nil
}
