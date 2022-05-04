package states

import (
	"bytes"
	"testing"

	"github.com/cntmio/cntmology/smartccntmract/types"
)

func TestCcntmract_Serialize_Deserialize(t *testing.T) {
	vmcode := types.VmCode{
		VmType: types.Native,
		Code:   []byte{1},
	}

	addr := vmcode.AddressFromVmCode()

	c := &Ccntmract{
		Version: 0,
		Code:    []byte{1},
		Address: addr,
		Method:  "init",
		Args:    []byte{2},
	}
	bf := new(bytes.Buffer)
	if err := c.Serialize(bf); err != nil {
		t.Fatalf("Ccntmract serialize error: %v", err)
	}

	v := new(Ccntmract)
	if err := v.Deserialize(bf); err != nil {
		t.Fatalf("Ccntmract deserialize error: %v", err)
	}
}
