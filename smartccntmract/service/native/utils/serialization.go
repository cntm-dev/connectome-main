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

package utils

import (
	"fmt"
	"io"
	"math/big"

	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/serialization"
)

func WriteVarUint(w io.Writer, value uint64) error {
	if err := serialization.WriteVarBytes(w, common.BigIntToNeoBytes(big.NewInt(int64(value)))); err != nil {
		return fmt.Errorf("serialize value error:%v", err)
	}
	return nil
}

func ReadVarUint(r io.Reader) (uint64, error) {
	value, err := serialization.ReadVarBytes(r)
	if err != nil {
		return 0, fmt.Errorf("deserialize value error:%v", err)
	}
	v := common.BigIntFromNeoBytes(value)
	if v.Cmp(big.NewInt(0)) < 0 {
		return 0, fmt.Errorf("%s", "value should not be a negative number.")
	}
	return v.Uint64(), nil
}

func WriteAddress(w io.Writer, address common.Address) error {
	if err := serialization.WriteVarBytes(w, address[:]); err != nil {
		return fmt.Errorf("serialize value error:%v", err)
	}
	return nil
}

func ReadAddress(r io.Reader) (common.Address, error) {
	from, err := serialization.ReadVarBytes(r)
	if err != nil {
		return common.Address{}, fmt.Errorf("[State] deserialize from error:%v", err)
	}
	return common.AddressParseFromBytes(from)
}
