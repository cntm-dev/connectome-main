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
	"testing"

	"crypto/rand"

	"github.com/cntmio/cntmology/common"
	"github.com/stretchr/testify/assert"
)

func TestStorageKey_Deserialize_Serialize(t *testing.T) {
	var addr common.Address
	rand.Read(addr[:])

	storage := StorageKey{
		CcntmractAddress: addr,
		Key:             []byte{1, 2, 3},
	}

	sink := common.NewZeroCopySink(nil)
	storage.Serialization(sink)
	bs := sink.Bytes()

	var storage2 StorageKey
	source := common.NewZeroCopySource(sink.Bytes())
	storage2.Deserialization(source)
	assert.Equal(t, storage, storage2)

	buf := common.NewZeroCopySource(bs[:len(bs)-1])
	err := storage2.Deserialization(buf)
	assert.NotNil(t, err)
}
