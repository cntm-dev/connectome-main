package ccntmract

import (
	"GoOnchain/crypto"
	. "GoOnchain/common"
)

//create a Single Singature ccntmract for owner  。
func CreateSignatureCcntmract(ownerPubKey crypto.PubKey) (*Ccntmract,error){
	//TODO: implement func CreateSignatureCcntmract
	return nil,nil
}

//create a Multi Singature ccntmract for owner  。
func CreateMultiSigCcntmract(publicKeyHash Uint160,m int, publicKeys ...[]*crypto.PubKey) (*Ccntmract,error){
	//TODO: implement func CreateSignatureCcntmract
	return nil,nil
}
