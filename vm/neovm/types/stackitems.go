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
	"math/big"
	"github.com/Ontology/vm/neovm/interfaces"
)

type StackItems interface {
	Equals(other StackItems) bool
	GetBigInteger() (*big.Int, error)
	GetBoolean() (bool, error)
	GetByteArray() ([]byte, error)
	GetInterface() (interfaces.Interop, error)
	GetArray() ([]StackItems, error)
	GetStruct() ([]StackItems, error)
	GetMap() (map[StackItems]StackItems, error)
}

const (
	ByteArrayType byte = 0x00
	BooleanType   byte = 0x01
	IntegerType   byte = 0x02
	InterfaceType byte = 0x40
	ArrayType     byte = 0x80
	StructType    byte = 0x81
	MapType       byte = 0x82
)
