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

package program

import (
	"fmt"
	"io"

	"github.com/cntmio/cntmology/common/serialization"
)

type Program struct {
	// the ccntmract program code,which will be run on VM or specific environment
	Code []byte

	// the program code's parameter
	Parameter []byte
}

//Serialize the Program
func (p *Program) Serialize(w io.Writer) error {
	err := serialization.WriteVarBytes(w, p.Parameter)
	if err != nil {
		return fmt.Errorf("Execute Program Serialize Code failed: %s", err)
	}
	err = serialization.WriteVarBytes(w, p.Code)
	if err != nil {
		return fmt.Errorf("Execute Program Serialize Parameter failed: %s", err)
	}

	return nil
}

//Deserialize the Program
func (p *Program) Deserialize(w io.Reader) error {
	val, err := serialization.ReadVarBytes(w)
	if err != nil {
		return fmt.Errorf("Execute Program Deserialize Parameter failed: %s", err)
	}
	p.Parameter = val
	p.Code, err = serialization.ReadVarBytes(w)
	if err != nil {
		return fmt.Errorf("Execute Program Deserialize Code failed: %s", err)
	}
	return nil
}
