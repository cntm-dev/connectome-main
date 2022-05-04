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
	"bytes"
	"testing"

	"github.com/cntmio/cntmology/smartccntmract/types"
	"github.com/stretchr/testify/assert"
)

func TestInvokeCode_Serialize(t *testing.T) {
	code := InvokeCode{
		Code: types.VmCode{
			VmType: types.NEOVM,
			Code:   []byte{1, 2, 3},
		},
	}

	buf := bytes.NewBuffer(nil)
	code.Serialize(buf)
	bs := buf.Bytes()
	var code2 InvokeCode
	code2.Deserialize(buf)
	assert.Equal(t, code, code2)

	buf = bytes.NewBuffer(bs[:len(bs)-2])
	err := code.Deserialize(buf)

	assert.NotNil(t, err)
}
