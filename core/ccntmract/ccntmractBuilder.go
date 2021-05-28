package ccntmract

import (
	"DNA/crypto"
	"DNA/vm"
	. "DNA/common"
	pg "DNA/core/ccntmract/program"
	"math/big"
	. "DNA/errors"
	"sort"
)

//create a Single Singature ccntmract for owner  。
func CreateSignatureCcntmract(ownerPubKey *crypto.PubKey) (*Ccntmract,error){

	temp,err := ownerPubKey.EncodePoint(true)
	if err !=nil{
		return nil,NewDetailErr(err, ErrNoCode, "[Ccntmract],CreateSignatureCcntmract failed.")
	}
	signatureRedeemScript,err := CreateSignatureRedeemScript(ownerPubKey)
	if err !=nil{
		return nil,NewDetailErr(err, ErrNoCode, "[Ccntmract],CreateSignatureCcntmract failed.")
	}
	hash,err:= ToCodeHash(temp)
	if err !=nil{
		return nil,NewDetailErr(err, ErrNoCode, "[Ccntmract],CreateSignatureCcntmract failed.")
	}
	signatureRedeemScriptHashToCodeHash,err:= ToCodeHash(signatureRedeemScript)
	if err !=nil{
		return nil,NewDetailErr(err, ErrNoCode, "[Ccntmract],CreateSignatureCcntmract failed.")
	}
	return &Ccntmract{
		Code: signatureRedeemScript,
		Parameters: []CcntmractParameterType{Signature},
		ProgramHash:signatureRedeemScriptHashToCodeHash,
		OwnerPubkeyHash: hash,
	},nil
}

func CreateSignatureRedeemScript(pubkey *crypto.PubKey) ([]byte,error){
	temp,err := pubkey.EncodePoint(true)
	if err !=nil{
		return nil,NewDetailErr(err, ErrNoCode, "[Ccntmract],CreateSignatureRedeemScript failed.")
	}
	sb := pg.NewProgramBuilder()
	sb.PushData(temp)
	sb.AddOp(vm.OP_CHECKSIG)
	return sb.ToArray(),nil
}

//create a Multi Singature ccntmract for owner  。
func CreateMultiSigCcntmract(publicKeyHash Uint160,m int, publicKeys []*crypto.PubKey) (*Ccntmract,error){

	params := make([]CcntmractParameterType,m)
	for i,_ := range params{
		params[i] = Signature
	}
	MultiSigRedeemScript, err := CreateMultiSigRedeemScript(m,publicKeys)
	if err !=nil{
		return nil,NewDetailErr(err, ErrNoCode, "[Ccntmract],CreateSignatureRedeemScript failed.")
	}
	signatureRedeemScriptHashToCodeHash,err:= ToCodeHash(MultiSigRedeemScript)
	if err !=nil{
		return nil,NewDetailErr(err, ErrNoCode, "[Ccntmract],CreateSignatureCcntmract failed.")
	}
	return &Ccntmract{
		Code: MultiSigRedeemScript,
		Parameters: params,
		ProgramHash:signatureRedeemScriptHashToCodeHash,
		OwnerPubkeyHash: publicKeyHash,
	},nil
}

func CreateMultiSigRedeemScript(m int,pubkeys []*crypto.PubKey) ([]byte,error){
	if ! (m >= 1 && m <= len(pubkeys) && len(pubkeys) <= 24) {
		return nil,nil //TODO: add panic
	}

	sb := pg.NewProgramBuilder()
	sb.PushNumber(big.NewInt(int64(m)))

	//sort pubkey
	sort.Sort(crypto.PubKeySlice(pubkeys))

	for _,pubkey := range pubkeys{
		temp,err := pubkey.EncodePoint(true)
		if err !=nil{
			return nil,NewDetailErr(err, ErrNoCode, "[Ccntmract],CreateSignatureCcntmract failed.")
		}
		sb.PushData(temp)
	}

	sb.PushNumber(big.NewInt(int64(len(pubkeys))))
	sb.AddOp(vm.OP_CHECKMULTISIG)
	return sb.ToArray(),nil
}
