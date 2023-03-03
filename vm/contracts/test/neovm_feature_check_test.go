/*
 * Copyright (C) 2018 The cntm Authors
 * This file is part of The cntm library.
 *
 * The cntm is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The cntm is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The cntm.  If not, see <http://www.gnu.org/licenses/>.
 */

package test

import (
	"testing"

	"github.com/conntectome/cntm/common/config"
	"github.com/conntectome/cntm/core/types"
	"github.com/conntectome/cntm/smartcontract"
	"github.com/conntectome/cntm/vm/cntmvm"
	"github.com/conntectome/cntm/vm/cntmvm/errors"
	"github.com/stretchr/testify/assert"
)

func TestHeight(t *testing.T) {
	byteCode0 := []byte{
		byte(cntmvm.NEWMAP),
		byte(cntmvm.PUSH0),
		byte(cntmvm.HASKEY),
	}

	byteCode1 := []byte{
		byte(cntmvm.NEWMAP),
		byte(cntmvm.KEYS),
	}

	byteCode2 := []byte{
		byte(cntmvm.NEWMAP),
		byte(cntmvm.VALUES),
	}

	bytecode := [...][]byte{byteCode0, byteCode1, byteCode2}

	disableHeight := config.GetOpcodeUpdateCheckHeight(config.DefConfig.P2PNode.NetworkId)
	heights := []uint32{10, disableHeight, disableHeight + 1}

	for _, height := range heights {
		config := &smartcontract.Config{Time: 10, Height: height}
		sc := smartcontract.SmartContract{Config: config, Gas: 100}
		expected := "[CntmVmService] vm execution error!: " + errors.ERR_NOT_SUPPORT_OPCODE.Error()
		if height > disableHeight {
			expected = ""
		}
		for i := 0; i < 3; i++ {
			engine, err := sc.NewExecuteEngine(bytecode[i], types.InvokeCntm)
			assert.Nil(t, err)

			_, err = engine.Invoke()
			if len(expected) > 0 {
				assert.EqualError(t, err, expected)
			} else {
				assert.Nil(t, err)
			}
		}
	}
}
