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
	"bytes"
	"crypto/rand"
	"fmt"
	"os"
	"testing"

	"github.com/conntectome/cntm/common"
	"github.com/conntectome/cntm/common/log"
	"github.com/conntectome/cntm/core/types"
	. "github.com/conntectome/cntm/smartcontract"
	cntmvm2 "github.com/conntectome/cntm/smartcontract/service/cntmvm"
	"github.com/conntectome/cntm/vm/cntmvm"
	"github.com/stretchr/testify/assert"
)

func TestRandomCodeCrash(t *testing.T) {
	log.InitLog(4)
	defer func() {
		os.RemoveAll("Log")
	}()

	config := &Config{
		Time:   10,
		Height: 10,
		Tx:     &types.Transaction{},
	}

	var code []byte
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("code %x \n", code)
		}
	}()

	for i := 1; i < 10; i++ {
		fmt.Printf("test round:%d \n", i)
		code := make([]byte, i)
		for j := 0; j < 10; j++ {
			rand.Read(code)

			sc := SmartContract{
				Config:  config,
				Gas:     10000,
				CacheDB: nil,
			}
			engine, _ := sc.NewExecuteEngine(code, types.InvokeCntm)
			engine.Invoke()
		}
	}
}

func TestOpCodeDUP(t *testing.T) {
	log.InitLog(4)
	defer func() {
		os.RemoveAll("Log")
	}()

	config := &Config{
		Time:   10,
		Height: 10,
		Tx:     &types.Transaction{},
	}

	var code = []byte{byte(cntmvm.DUP)}

	sc := SmartContract{
		Config:  config,
		Gas:     10000,
		CacheDB: nil,
	}
	engine, _ := sc.NewExecuteEngine(code, types.InvokeCntm)
	_, err := engine.Invoke()

	assert.NotNil(t, err)
}

func TestOpReadMemAttack(t *testing.T) {
	log.InitLog(4)
	defer func() {
		os.RemoveAll("Log")
	}()

	config := &Config{
		Time:   10,
		Height: 10,
		Tx:     &types.Transaction{},
	}

	bf := new(bytes.Buffer)
	builder := cntmvm.NewParamsBuilder(bf)
	builder.Emit(cntmvm.SYSCALL)
	sink := common.NewZeroCopySink(builder.ToArray())
	builder.EmitPushByteArray([]byte(cntmvm2.NATIVE_INVOKE_NAME))
	l := 0x7fffffc7 - 1
	sink.WriteVarUint(uint64(l))
	b := make([]byte, 4)
	sink.WriteBytes(b)

	sc := SmartContract{
		Config:  config,
		Gas:     100000,
		CacheDB: nil,
	}
	engine, _ := sc.NewExecuteEngine(sink.Bytes(), types.InvokeCntm)
	_, err := engine.Invoke()

	assert.NotNil(t, err)

}
