package ccntmract

import (
	"GoOnchain/crypto"
	"GoOnchain/vm"
	. "GoOnchain/common"
	pg "GoOnchain/core/ccntmract/program"
	"math/big"
	_ "sort"
)

//create a Single Singature ccntmract for owner  。
func CreateSignatureCcntmract(ownerPubKey *crypto.PubKey) (*Ccntmract,error){

	return &Ccntmract{
		Code: CreateSignatureRedeemScript(ownerPubKey),
		Parameters: []CcntmractParameterType{Signature},
		OwnerPubkeyHash: ToCodeHash(ownerPubKey.EncodePoint(true)),
	},nil
}

func CreateSignatureRedeemScript(pubkey *crypto.PubKey) []byte{
	sb := pg.NewProgramBuilder()
	sb.PushData(pubkey.EncodePoint(true))
	sb.AddOp(vm.OP_CHECKSIG)
	return sb.ToArray()
}

//create a Multi Singature ccntmract for owner  。
func CreateMultiSigCcntmract(publicKeyHash Uint160,m int, publicKeys []*crypto.PubKey) (*Ccntmract,error){

	params := make([]CcntmractParameterType,m)
	for i,_ := range params{
		params[i] = Signature
	}

	return &Ccntmract{
		Code: CreateMultiSigRedeemScript(m,publicKeys),
		Parameters: params,
		OwnerPubkeyHash: publicKeyHash,
	},nil
}

func CreateMultiSigRedeemScript(m int,pubkeys []*crypto.PubKey) []byte{
	if ! (m >= 1 && m <= len(pubkeys) && len(pubkeys) <= 24) {
		return nil //TODO: add panic
	}

	sb := pg.NewProgramBuilder()
	sb.PushNumber(big.NewInt(int64(m)))

	//TODO: sort pubkey
	//var keys *crypto.PubKeys = pubkeys
	//sort.Sort(keys)
	for _,pubkey := range pubkeys{
		sb.PushData(pubkey.EncodePoint(true))
	}



	sb.PushNumber(big.NewInt(int64(len(pubkeys))))
	sb.AddOp(vm.OP_CHECKMULTISIG)
	return sb.ToArray()
}
