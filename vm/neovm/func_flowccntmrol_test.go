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

package neovm

import (
	"testing"
)

func TestOpDcall(t *testing.T) {
	var e ExecutionEngine
	stack := NewRandAccessStack()
	e.EvaluationStack = stack
	ccntmext := NewExecutionCcntmext([]byte{byte(PUSH8), byte(PUSH2), byte(PUSH0), byte(DCALL)})
	e.PushCcntmext(ccntmext)

	PushData(&e, 1)
	opDCALL(&e)
	e.ExecuteCode()

	if len(e.Ccntmexts) != 2 {
		t.Fatalf("NeoVM opDCALL test failed")
	}

	if e.OpCode != PUSH2 {
		t.Fatalf("NeoVM opDCALL test failed, expect PUSH2 , get %s.", OpExecList[e.OpCode].Name)
	}
}
