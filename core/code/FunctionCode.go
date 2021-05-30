package code

import (
	"github.com/DNAProject/DNA/common/log"
	."github.com/DNAProject/DNA/common"
	."github.com/DNAProject/DNA/core/ccntmract"
	"github.com/DNAProject/DNA/common/serialization"
	"fmt"
	"io"
)

type FunctionCode struct {
	// Ccntmract Code
	Code []byte

	// Ccntmract parameter type list
	ParameterTypes []CcntmractParameterType

	// Ccntmract return type list
	ReturnTypes []CcntmractParameterType
}

// method of SerializableData
func (fc *FunctionCode) Serialize(w io.Writer) error {
	err := serialization.WriteVarBytes(w,CcntmractParameterTypeToByte(fc.ParameterTypes))
	if err != nil {
		return err
	}

	err = serialization.WriteVarBytes(w,fc.Code)
	if err != nil {
		return err
	}

	return nil
}

// method of SerializableData
func (fc *FunctionCode) Deserialize(r io.Reader) error {
	p,err := serialization.ReadVarBytes(r)
	if err != nil {
		return err
	}
	fc.ParameterTypes = ByteToCcntmractParameterType(p)

	fc.Code,err = serialization.ReadVarBytes(r)
	if err != nil {
		return err
	}

	return nil
}

// method of ICode
// Get code
func (fc *FunctionCode) GetCode() []byte {
	return fc.Code
}

// method of ICode
// Get the list of parameter value
func (fc *FunctionCode) GetParameterTypes() []CcntmractParameterType {
	return fc.ParameterTypes
}

// method of ICode
// Get the list of return value
func (fc *FunctionCode) GetReturnTypes() []CcntmractParameterType {
	return fc.ReturnTypes
}

// method of ICode
// Get the hash of the smart ccntmract
func (fc *FunctionCode) CodeHash() Uint160 {
	hash,err := ToCodeHash(fc.Code)
	if err != nil {
		log.Debug( fmt.Sprintf("[FunctionCode] ToCodeHash err=%s",err) )
		return Uint160{0}
	}

	return hash
}