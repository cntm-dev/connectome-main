package states

import (
	"io"
	"bytes"
	. "github.com/Ontology/common/serialization"
	"github.com/Ontology/core/code"
	"github.com/Ontology/smartccntmract/types"
	. "github.com/Ontology/errors"
	"github.com/Ontology/vm/neovm/interfaces"
)

type CcntmractState struct {
	StateBase
	Code        *code.FunctionCode
	VmType      types.VmType
	NeedStorage bool
	Name        string
	Version     string
	Author      string
	Email       string
	Description string
}

func (this *CcntmractState) Serialize(w io.Writer) error {
	this.StateBase.Serialize(w)
	err := this.Code.Serialize(w)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "CcntmractState Code Serialize failed.")
	}
	err = WriteByte(w, byte(this.VmType))
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "CcntmractState VmType Serialize failed.")
	}

	err = WriteBool(w, this.NeedStorage)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "CcntmractState NeedStorage Serialize failed.")
	}

	err = WriteVarString(w, this.Name)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "CcntmractState Name Serialize failed.")
	}
	err = WriteVarString(w, this.Version)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "CcntmractState Version Serialize failed.")
	}
	err = WriteVarString(w, this.Author)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "CcntmractState Author Serialize failed.")
	}
	err = WriteVarString(w, this.Email)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "CcntmractState Email Serialize failed.")
	}
	err = WriteVarString(w, this.Description)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "CcntmractState Description Serialize failed.")
	}
	return nil
}

func (this *CcntmractState) Deserialize(r io.Reader) error {
	if this == nil {
		this = new(CcntmractState)
	}
	f := new(code.FunctionCode)

	err := this.StateBase.Deserialize(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "CcntmractState StateBase Deserialize failed.")
	}
	err = f.Deserialize(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "CcntmractState Code Deserialize failed.")
	}
	this.Code = f

	vmType, err := ReadByte(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "CcntmractState VmType Deserialize failed.")
	}
	this.VmType = types.VmType(vmType)

	this.NeedStorage, err = ReadBool(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "CcntmractState NeedStorage Deserialize failed.")
	}

	this.Name, err = ReadVarString(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "CcntmractState Name Deserialize failed.")
	}
	this.Version, err = ReadVarString(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "CcntmractState Version Deserialize failed.")
	}
	this.Author, err = ReadVarString(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "CcntmractState Author Deserialize failed.")
	}
	this.Email, err = ReadVarString(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "CcntmractState Email Deserialize failed.")
	}
	this.Description, err = ReadVarString(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "CcntmractState Description Deserialize failed.")
	}
	return nil
}

func (ccntmractState *CcntmractState) ToArray() []byte {
	b := new(bytes.Buffer)
	ccntmractState.Serialize(b)
	return b.Bytes()
}

func (ccntmractState *CcntmractState) Clone() interfaces.IInteropInterface {
	cs := *ccntmractState
	return &cs
}


