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

package states

import (
	"io"

	"github.com/cntmio/cntmology-crypto/keypair"
	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/serialization"
)

type VoteState struct {
	StateBase
	PublicKeys []keypair.PublicKey
	Count      common.Fixed64
}

func (this *VoteState) Serialize(w io.Writer) error {
	this.StateBase.Serialize(w)
	serialization.WriteUint32(w, uint32(len(this.PublicKeys)))
	for _, v := range this.PublicKeys {
		buf := keypair.SerializePublicKey(v)
		err := serialization.WriteVarBytes(w, buf)
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *VoteState) Deserialize(r io.Reader) error {
	if this == nil {
		this = new(VoteState)
	}
	err := this.StateBase.Deserialize(r)
	if err != nil {
		return err
	}
	n, err := serialization.ReadUint32(r)
	if err != nil {
		return err
	}
	for i := 0; i < int(n); i++ {
		buf, err := serialization.ReadVarBytes(r)
		if err != nil {
			return err
		}
		pk, err := keypair.DeserializePublicKey(buf)
		if err != nil {
			return err
		}
		this.PublicKeys = append(this.PublicKeys, pk)
	}
	return nil
}
