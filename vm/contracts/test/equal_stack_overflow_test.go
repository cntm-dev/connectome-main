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
	"os"
	"testing"

	"github.com/conntectome/cntm/common/log"
	"github.com/conntectome/cntm/core/types"
	. "github.com/conntectome/cntm/smartcontract"
	"github.com/conntectome/cntm/vm/cntmvm"
	"github.com/stretchr/testify/assert"
)

func TestEqualStackOverflow(t *testing.T) {
	log.InitLog(4)
	defer func() {
		os.RemoveAll("./Log")
	}()

	code := []byte{
		byte(cntmvm.PUSH1),    // {1}
		byte(cntmvm.NEWARRAY), // {[]}
		byte(cntmvm.DUP),      // {[],[]}
		byte(cntmvm.DUP),      // {[],[],[]}
		byte(cntmvm.PUSH0),    // {[],[],[],0}
		byte(cntmvm.ROT),      // {[],[],0,[]}
		byte(cntmvm.SETITEM),  // {[[]]}
		byte(cntmvm.DUP),      // {[[]],[[]]}
		byte(cntmvm.EQUAL),
	}

	config := &Config{
		Time:   10,
		Height: 10,
		Tx:     &types.Transaction{},
	}
	sc := SmartContract{
		Config:  config,
		Gas:     10000,
		CacheDB: nil,
	}
	engine, _ := sc.NewExecuteEngine(code, types.InvokeCntm)
	_, err := engine.Invoke()

	assert.Nil(t, err)
}
