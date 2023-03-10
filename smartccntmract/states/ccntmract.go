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

	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/smartccntmract/event"
)

// Invoke smart ccntmract struct
// Param Version: invoke smart ccntmract version, default 0
// Param Address: invoke on blockchain smart ccntmract by address
// Param Method: invoke smart ccntmract method, default ""
// Param Args: invoke smart ccntmract arguments
type CcntmractInvokeParam struct {
	Version byte
	Address common.Address
	Method  string
	Args    []byte
}

// Serialize ccntmract
func (this *Ccntmract) Serialize(w io.Writer) error {
	if err := serialization.WriteByte(w, this.Version); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Ccntmract] Version serialize error!")
	}
	if err := this.Address.Serialize(w); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Ccntmract] Address serialize error!")
	}
	if err := serialization.WriteVarBytes(w, []byte(this.Method)); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Ccntmract] Method serialize error!")
	}
	if err := serialization.WriteVarBytes(w, this.Args); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Ccntmract] Args serialize error!")
	}
	return nil
}

// Deserialize ccntmract
func (this *Ccntmract) Deserialize(r io.Reader) error {
	var err error
	this.Version, err = serialization.ReadByte(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Ccntmract] Version deserialize error!")
	}

	if err := this.Address.Deserialize(r); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Ccntmract] Address deserialize error!")
	}

	method, err := serialization.ReadVarBytes(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Ccntmract] Method deserialize error!")
	}
	this.Method = string(method)

	this.Args, err = serialization.ReadVarBytes(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Ccntmract] Args deserialize error!")
	}
	return nil
}

type PreExecResult struct {
	State  byte
	Gas    uint64
	Result interface{}
	Notify []*event.NotifyEventInfo
}
