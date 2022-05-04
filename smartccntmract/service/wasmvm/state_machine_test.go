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
