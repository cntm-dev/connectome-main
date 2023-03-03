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
package utils

import (
	"fmt"
	"testing"
)

func TestParseCntmvmFunc(t *testing.T) {
	var testCntmvmAbi = `{
  "hash": "0xe827bf96529b5780ad0702757b8bad315e2bb8ce",
  "entrypoint": "Main",
  "functions": [
    {
      "name": "Main",
      "parameters": [
        {
          "name": "operation",
          "type": "String"
        },
        {
          "name": "args",
          "type": "Array"
        }
      ],
      "returntype": "Any"
    },
    {
      "name": "Add",
      "parameters": [
        {
          "name": "a",
          "type": "Integer"
        },
        {
          "name": "b",
          "type": "Integer"
        }
      ],
      "returntype": "Integer"
    }
  ],
  "events": []
}`
	contractAbi, err := NewCntmvmContractAbi([]byte(testCntmvmAbi))
	if err != nil {
		t.Errorf("TestParseCntmvmFunc NewCntmvmContractAbi error:%s", err)
		return
	}
	funcAbi := contractAbi.GetFunc("Add")
	if funcAbi == nil {
		t.Error("TestParseCntmvmFunc cannot find func abi")
		return
	}

	params, err := ParseCntmvmFunc([]string{"12", "34"}, funcAbi)
	if err != nil {
		t.Errorf("TestParseCntmvmFunc ParseCntmvmFunc error:%s", err)
		return
	}
	fmt.Printf("TestParseCntmvmFunc %v\n", params)
}
