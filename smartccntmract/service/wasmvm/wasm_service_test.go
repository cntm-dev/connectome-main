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
