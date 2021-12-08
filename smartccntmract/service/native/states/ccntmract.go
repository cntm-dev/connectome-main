package states

import (
	"github.com/Ontology/common"
	"io"
	"github.com/Ontology/common/serialization"
	"github.com/Ontology/errors"
)

type Ccntmract struct {
	Version byte
	Address common.Address
	Method string
	Args []byte
}

func (this *Ccntmract) Serialize(w io.Writer) error {
	if err := serialization.WriteByte(w, this.Version); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Ccntmract] Version serialize error!")
	}
	if err := this.Address.Serialize(w); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Ccntmract] Address serialize error!")
	}
	if err := serialization.WriteVarBytes(w, []byte(this.Method)); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Ccntmract] Method serialize error!")
	}
	if err := serialization.WriteVarBytes(w, this.Args); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Ccntmract] Args serialize error!")
	}
	return nil
}

func (this *Ccntmract) Deserialize(r io.Reader) error {
	var err error
	this.Version, err = serialization.ReadByte(r); if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Ccntmract] Version deserialize error!")
	}

	if err := this.Address.Deserialize(r); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Ccntmract] Address deserialize error!")
	}

	method, err := serialization.ReadVarBytes(r); if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Ccntmract] Method deserialize error!")
	}
	this.Method = string(method)

	this.Args, err = serialization.ReadVarBytes(r); if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Ccntmract] Args deserialize error!")
	}
	return nil
}
