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

package test

import (
	"testing"

	"github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/smartccntmract"
	"github.com/cntmio/cntmology/vm/neovm"
	"github.com/cntmio/cntmology/vm/neovm/errors"
	"github.com/stretchr/testify/assert"
)

func TestAppendOverFlow(t *testing.T) {
	// define 1024 len array
	byteCode := []byte{
		byte(0x04), //neovm.PUSHBYTES4
		byte(0x00),
		byte(0x04),
		byte(0x00),
		byte(0x00),
		byte(neovm.NEWARRAY),
		byte(neovm.PUSH2),
		byte(neovm.APPEND),
	}

	config := &smartccntmract.Config{
		Time:   10,
		Height: 10,
	}
	sc := smartccntmract.SmartCcntmract{
		Config:  config,
		Gas:     200,
		CacheDB: nil,
	}
	engine, _ := sc.NewExecuteEngine(byteCode, types.InvokeNeo)
	_, err := engine.Invoke()
	assert.EqualError(t, err, "[NeoVmService] vm execution error!: "+errors.ERR_OVER_MAX_ARRAY_SIZE.Error())
}
