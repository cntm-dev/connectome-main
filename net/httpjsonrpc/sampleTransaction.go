package httpjsonrpc

import (
	. "github.com/Ontology/account"
	. "github.com/Ontology/common"
	"github.com/Ontology/common/log"
	. "github.com/Ontology/core/asset"
	"github.com/Ontology/core/ccntmract"
	"github.com/Ontology/core/signature"
	"github.com/Ontology/core/transaction"
	"strconv"
)

const (
	ASSETPREFIX = "Ontology"
)

func NewRegTx(rand string, index int, admin, issuer *Account) *transaction.Transaction {
	name := ASSETPREFIX + "-" + strconv.Itoa(index) + "-" + rand
	description := "description"
	asset := &Asset{name, description, byte(MaxPrecision), AssetType(Share), UTXO}
	amount := Fixed64(1000)
	ccntmroller, _ := ccntmract.CreateSignatureCcntmract(admin.PubKey())
	tx, _ := transaction.NewRegisterAssetTransaction(asset, amount, issuer.PubKey(), ccntmroller.ProgramHash)
	return tx
}

func SignTx(admin *Account, tx *transaction.Transaction) {
	signdate, err := signature.SignBySigner(tx, admin)
	if err != nil {
		log.Error(err, "signdate SignBySigner failed")
	}
	transactionCcntmract, _ := ccntmract.CreateSignatureCcntmract(admin.PublicKey)
	transactionCcntmractCcntmext := ccntmract.NewCcntmractCcntmext(tx)
	transactionCcntmractCcntmext.AddCcntmract(transactionCcntmract, admin.PublicKey, signdate)
	tx.SetPrograms(transactionCcntmractCcntmext.GetPrograms())
}
