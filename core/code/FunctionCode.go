package code

import (
	"fmt"
	. "github.com/Ontology/common"
	"github.com/Ontology/common/log"
	"github.com/Ontology/common/serialization"
	. "github.com/Ontology/core/ccntmract"
	. "github.com/Ontology/errors"
	"io"
)

type FunctionCode struct {
	// Ccntmract Code
	Code []byte

	// Ccntmract parameter type list
	ParameterTypes []CcntmractParameterType

	// Ccntmract return type
	ReturnType CcntmractParameterType

	codeHash Uint160
}

// method of SerializableData
func (fc *FunctionCode) Serialize(w io.Writer) error {
	var err error
	err = serialization.WriteVarBytes(w, fc.Code)
	if err != nil {
		return err
	}

	err = serialization.WriteVarBytes(w, CcntmractParameterTypeToByte(fc.ParameterTypes))
	if err != nil {
		return err
	}

	err = serialization.WriteByte(w, byte(fc.ReturnType))
	if err != nil {
		return err
	}

	return nil
}

// method of SerializableData
func (fc *FunctionCode) Deserialize(r io.Reader) error {
	var err error

	fc.Code, err = serialization.ReadVarBytes(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Transaction FunctionCode Code Deserialize failed.")
	}

	p, err := serialization.ReadVarBytes(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Transaction FunctionCode ParameterTypes Deserialize failed.")
	}
	fc.ParameterTypes = ByteToCcntmractParameterType(p)

	returnType, err := serialization.ReadByte(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Transaction FunctionCode returnType Deserialize failed.")
	}
	fc.ReturnType = CcntmractParameterType(returnType)
	return nil
}

// method of ICode
// Get the hash of the smart ccntmract
func (fc *FunctionCode) CodeHash() Uint160 {
	u160 := Uint160{}
	if fc.codeHash == u160 {
		u160, err := ToCodeHash(fc.Code)
		if err != nil {
			log.Debug( fmt.Sprintf("[FunctionCode] ToCodeHash err=%s",err) )
			return u160
		}
		fc.codeHash = u160
	}
	return fc.codeHash
}
