package states

import (
	"github.com/Ontology/common"
	"io"
	. "github.com/Ontology/errors"
)

type CcntmractMapping struct {
	OriginAddress common.Address
	TargetAddress common.Address
}

func (this *CcntmractMapping) Serialize(w io.Writer) error {
	if err := this.OriginAddress.Serialize(w); err != nil {
		return NewDetailErr(err, ErrNoCode, "[CcntmractMapping] OriginAddress serialize failed.")
	}
	if err := this.TargetAddress.Serialize(w); err != nil {
		return NewDetailErr(err, ErrNoCode, "[CcntmractMapping] TargetAddress serialize failed.")
	}
	return nil
}

func (this *CcntmractMapping) Deserialize(r io.Reader) error {
	origin := new(common.Address)
	if err := origin.Deserialize(r); err != nil {
		return NewDetailErr(err, ErrNoCode, "[CcntmractMapping] OriginAddress deserialize failed.")
	}

	target := new(common.Address)
	if err := target.Deserialize(r); err != nil {
		return NewDetailErr(err, ErrNoCode, "[CcntmractMapping] TargetAddress deserialize failed.")
	}
	return nil
}