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
package wasmvm

import (
	"github.com/cntmio/cntmology/smartccntmract/types"
	"strings"
	"testing"
)

func TestWasmVmService_GetCcntmractCodeFromAddress(t *testing.T) {
	code := "0061736d01000000010c0260027f7f017f60017f017f03030200010710020361646400000673717561726500010a11020700200120006a0b0700200020006c0b"
	address := GetCcntmractAddress(code, types.WASMVM)

	if len(address[:]) != 20 {
		t.Error("TestWasmVmService_GetCcntmractCodeFromAddress get address failed")
	}

	hexaddr := address.ToHexString()

	if strings.Index(hexaddr, "9") != 0 {
		t.Error("TestWasmVmService_GetCcntmractCodeFromAddress is not a wasm ccntmract address")
	}
}
