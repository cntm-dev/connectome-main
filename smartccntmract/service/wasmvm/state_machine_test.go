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

import "testing"

func TestNewWasmStateMachine(t *testing.T) {
	sm := NewWasmStateMachine()
	if sm == nil {
		t.Fatal("NewWasmStateMachine should return a non nil state machine")
	}

	if sm.WasmStateReader == nil {
		t.Fatal("NewWasmStateMachine should return a non nil state reader")
	}

	if !sm.Exists("CcntmractLogDebug") {
		t.Error("NewWasmStateMachine should has CcntmractLogDebug service")
	}

	if !sm.Exists("CcntmractLogInfo") {
		t.Error("NewWasmStateMachine should has CcntmractLogInfo service")
	}

	if !sm.Exists("CcntmractLogError") {
		t.Error("NewWasmStateMachine should has CcntmractLogError service")
	}
}
