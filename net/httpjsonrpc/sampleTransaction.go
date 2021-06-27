package httpjsonrpc

import (
	"DNA/client"
	. "DNA/common"
	"DNA/common/log"
	. "DNA/core/asset"
	"DNA/core/ccntmract"
	"DNA/core/signature"
	"DNA/core/transaction"
	"strconv"
)

const (
	ASSETPREFIX = "DNA"
)

func NewRegTx(rand string, index int, admin, issuer *client.Account) *transaction.Transaction {
	name := ASSETPREFIX + "-" + strconv.Itoa(index) + "-" + rand
	asset := &Asset{name, byte(0x00), AssetType(Share), UTXO}
	amount := Fixed64(1000)
	ccntmroller, _ := ccntmract.CreateSignatureCcntmract(admin.PubKey())
	tx, _ := transaction.NewRegisterAssetTransaction(asset, amount, issuer.PubKey(), ccntmroller.ProgramHash)
	return tx
}

func SignTx(admin *client.Account, tx *transaction.Transaction) {
	signdate, err := signature.SignBySigner(tx, admin)
	if err != nil {
		log.Error(err, "signdate SignBySigner failed")
	}
	transactionCcntmract, _ := ccntmract.CreateSignatureCcntmract(admin.PublicKey)
	transactionCcntmractCcntmext := ccntmract.NewCcntmractCcntmext(tx)
	transactionCcntmractCcntmext.AddCcntmract(transactionCcntmract, admin.PublicKey, signdate)
	tx.SetPrograms(transactionCcntmractCcntmext.GetPrograms())
}
