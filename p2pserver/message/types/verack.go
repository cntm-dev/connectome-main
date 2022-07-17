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

package types

import (
	"bytes"
	"fmt"

	"github.com/cntmio/cntmology/common/serialization"
	"github.com/cntmio/cntmology/errors"
	"github.com/cntmio/cntmology/p2pserver/common"
)

type VerACK struct {
	IsConsensus bool
}

//Serialize message payload
func (this VerACK) Serialization() ([]byte, error) {
	p := bytes.NewBuffer([]byte{})
	err := serialization.WriteBool(p, this.IsConsensus)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNetPackFail, fmt.Sprintf("write error. IsConsensus:%v", this.IsConsensus))
	}

	return p.Bytes(), nil
}

func (this VerACK) CmdType() string {
	return common.VERACK_TYPE
}

//Deserialize message payload
func (this *VerACK) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)

	isConsensus, err := serialization.ReadBool(buf)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNetUnPackFail, fmt.Sprintf("read IsConsensus error. buf:%v", buf))
	}

	this.IsConsensus = isConsensus
	return nil
}
