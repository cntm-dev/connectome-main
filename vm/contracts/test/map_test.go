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
	"fmt"
	"testing"

	"github.com/conntectome/cntm/core/types"
	"github.com/conntectome/cntm/smartcontract"
	"github.com/conntectome/cntm/vm/cntmvm"
	"github.com/stretchr/testify/assert"
)

func TestMap(t *testing.T) {
	byteCode := []byte{
		byte(cntmvm.NEWMAP),
		byte(cntmvm.DUP),   // dup map
		byte(cntmvm.PUSH0), // key (index)
		byte(cntmvm.PUSH0), // key (index)
		byte(cntmvm.SETITEM),

		byte(cntmvm.DUP),   // dup map
		byte(cntmvm.PUSH0), // key (index)
		byte(cntmvm.PUSH1), // value (newItem)
		byte(cntmvm.SETITEM),
	}

	// pick a value out
	byteCode = append(byteCode,
		[]byte{ // extract element
			byte(cntmvm.DUP),   // dup map (items)
			byte(cntmvm.PUSH0), // key (index)

			byte(cntmvm.PICKITEM),
			byte(cntmvm.JMPIF), // dup map (items)
			0x04, 0x00,        // skip a drop?
			byte(cntmvm.DROP),
		}...)

	// count faults vs successful executions
	N := 1024
	faults := 0

	//dbFile := "/tmp/test"
	//os.RemoveAll(dbFile)
	//testLevelDB, err := leveldbstore.NewLevelDBStore(dbFile)
	//if err != nil {
	//	panic(err)
	//}

	for n := 0; n < N; n++ {
		// Setup Execution Environment
		//store := statestore.NewMemDatabase()
		//testBatch := statestore.NewStateStoreBatch(store, testLevelDB)
		config := &smartcontract.Config{
			Time:   10,
			Height: 10,
			Tx:     &types.Transaction{},
		}
		sc := smartcontract.SmartContract{
			Config:  config,
			Gas:     100,
			CacheDB: nil,
		}
		engine, err := sc.NewExecuteEngine(byteCode, types.InvokeCntm)

		_, err = engine.Invoke()
		if err != nil {
			fmt.Println("err:", err)
			faults += 1
		}
	}
	assert.Equal(t, faults, 0)

}
