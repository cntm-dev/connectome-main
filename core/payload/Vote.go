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

package payload

import (
	"fmt"
	"io"

	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/serialization"
	"github.com/cntmio/cntmology-crypto/keypair"
)

const (
	MaxVoteKeys = 1024
)

type Vote struct {
	PubKeys []keypair.PublicKey // vote node list

	Account common.Address
}

func (self *Vote) Check() bool {
	if len(self.PubKeys) > MaxVoteKeys {
		return false
	}
	return true
}

func (self *Vote) Serialize(w io.Writer) error {
	if err := serialization.WriteUint32(w, uint32(len(self.PubKeys))); err != nil {
		return fmt.Errorf("Vote PubKeys length Serialize failed: %s", err)
	}
	for _, key := range self.PubKeys {
		buf := keypair.SerializePublicKey(key)
		err := serialization.WriteVarBytes(w, buf)
		if err != nil {
			return fmt.Errorf("InvokeCode PubKeys Serialize failed: %s", err)
		}
	}
	if err := self.Account.Serialize(w); err != nil {
		return fmt.Errorf("InvokeCode Account Serialize failed: %s", err)
	}

	return nil
}

func (self *Vote) Deserialize(r io.Reader) error {
	length, err := serialization.ReadUint32(r)
	if err != nil {
		return err
	}
	self.PubKeys = make([]keypair.PublicKey, length)
	for i := 0; i < int(length); i++ {
		buf, err := serialization.ReadVarBytes(r)
		if err != nil {
			return err
		}
		self.PubKeys[i], err = keypair.DeserializePublicKey(buf)
		if err != nil {
			return err
		}
	}

	err = self.Account.Deserialize(r)
	if err != nil {
		return err
	}

	return nil
}
