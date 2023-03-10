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
	"github.com/Ontology/vm/neovm/interfaces"
	"math/big"
)

type Array struct {
	_array []StackItems
}

func NewArray(value []StackItems) *Array {
	var this Array
	this._array = value
	return &this
}

func (this *Array) Equals(other StackItems) bool {
	return this == other
}

func (this *Array) GetBigInteger() (*big.Int, error) {
	return nil, fmt.Errorf("%s", "Not support array to integer")
}

func (this *Array) GetBoolean() (bool, error) {
	return false, fmt.Errorf("%s", "Not support array to boolean")
}

func (this *Array) GetByteArray() ([]byte, error) {
	return nil, fmt.Errorf("%s", "Not support array to byte array")
}

func (this *Array) GetInterface() (interfaces.Interop, error) {
	return nil, fmt.Errorf("%s", "Not support array to interface")
}

func (this *Array) GetArray() ([]StackItems, error) {
	return this._array, nil
}

func (this *Array) GetStruct() ([]StackItems, error) {
	return this._array, nil
}

func (this *Array) GetMap() (map[StackItems]StackItems, error) {
	return nil, fmt.Errorf("%s", "Not support array to map")
}

func (this *Array) Add(item StackItems) {
	this._array = append(this._array, item)
}

func (this *Array) Count() int {
	return len(this._array)
}
