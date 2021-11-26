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
	"crypto/sha256"
	"errors"
	"github.com/Ontology/common/log"
	. "github.com/Ontology/errors"
	"io"
	"math/big"

	"github.com/itchyny/base58-go"
	"fmt"
)

const AddrLen int = 20

type Address [AddrLen]byte

func (u *Address) ToArray() []byte {
	x := append([]byte{}, u[:]...)
	return x
}


func (self *Address) ToHexString() string {
	return fmt.Sprintf("%x", self[:])
}

func (self *Address) Serialize(w io.Writer) error {
	_, err := w.Write(self[:])
	return err
}

func (self *Address) Deserialize(r io.Reader) error {
	n, err := r.Read(self[:])
	if n != len(self[:]) || err != nil {
		return errors.New("deserialize Address error")
	}
	return nil
}


func (f *Address) ToBase58() string {
	data := append([]byte{0x41}, f[:]...)
	temp := sha256.Sum256(data)
	temps := sha256.Sum256(temp[:])
	data = append(data, temps[0:4]...)

	bi := new(big.Int).SetBytes(data).String()
	encoded, _ := base58.BitcoinEncoding.Encode([]byte(bi))
	return string(encoded)
}

func Uint160ParseFromBytes(f []byte) (Address, error) {
	if len(f) != AddrLen {
		return Address{}, NewDetailErr(errors.New("[Common]: Uint160ParseFromBytes err, len != 20"), ErrNoCode, "")
	}

	var hash [20]uint8
	for i := 0; i < 20; i++ {
		hash[i] = f[i]
	}
	return Address(hash), nil
}

func AddressFromBase58(encoded string) (Address, error) {
	decoded, err := base58.BitcoinEncoding.Decode([]byte(encoded))
	if err != nil {
		return Address{}, err
	}

	x, _ := new(big.Int).SetString(string(decoded), 10)
	log.Tracef("[ToAddress] x: ", x.Bytes())

	ph, err := Uint160ParseFromBytes(x.Bytes()[1:21])
	if err != nil {
		return Address{}, err
	}

	log.Tracef("[AddressToProgramHash] programhash: %x", ph[:])

	addr := ph.ToBase58()

	log.Tracef("[AddressToProgramHash] encoded: %s", addr)

	if addr != encoded {
		return Address{}, errors.New("[AddressFromBase58]: decode encoded verify failed.")
	}

	return ph, nil
}
