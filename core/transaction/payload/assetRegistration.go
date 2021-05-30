package payload

import (
	"github.com/DNAProject/DNA/common"
	"github.com/DNAProject/DNA/core/asset"
	"github.com/DNAProject/DNA/crypto"
	"io"
	. "github.com/DNAProject/DNA/errors"

)

type AssetRegistration struct {
	Asset      *asset.Asset
	Amount     *common.Fixed64
	//Precision  byte
	Issuer     *crypto.PubKey
	Ccntmroller *common.Uint160
}

func (a *AssetRegistration) Data() []byte {
	//TODO: implement AssetRegistration.Data()
	return []byte{0}

}

func (a *AssetRegistration) Serialize(w io.Writer) {
	a.Asset.Serialize(w)
	a.Amount.Serialize(w)
	//w.Write([]byte{a.Precision})
	a.Issuer.Serialize(w)
	a.Ccntmroller.Serialize(w)
}

func (a *AssetRegistration) Deserialize(r io.Reader) error {

	//asset
	a.Asset = new(asset.Asset)
	err := a.Asset.Deserialize(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[AssetRegistration], Asset Deserialize failed.")
	}

	//Amount
	a.Amount = new(common.Fixed64)
	err = a.Amount.Deserialize(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[AssetRegistration], Ammount Deserialize failed.")
	}

	//Precision  byte 02/10 comment out by wjj
	//p := make([]byte, 1)
	//n, err := r.Read(p)
	//if n > 0 {
	//	a.Precision = p[0]
	//} else {
	//	return NewDetailErr(err, ErrNoCode, "[AssetRegistration], Precision Deserialize failed.")
	//}

	//Issuer     *crypto.PubKey
	a.Issuer = new(crypto.PubKey)
	err = a.Issuer.DeSerialize(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[AssetRegistration], Ammount Deserialize failed.")
	}

	//Ccntmroller *common.Uint160
	a.Ccntmroller = new(common.Uint160)
	err = a.Ccntmroller.Deserialize(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[AssetRegistration], Ammount Deserialize failed.")
	}
	return nil
}

