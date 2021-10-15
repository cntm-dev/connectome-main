package httpjsonrpc

/*
import (
	. "github.com/Ontology/account"
	. "github.com/Ontology/common"
	"github.com/Ontology/common/log"
	. "github.com/Ontology/core/asset"
	"github.com/Ontology/core/ccntmract"
	"github.com/Ontology/core/signature"
	"github.com/Ontology/core/types"
	"strconv"
)

const (
	ASSETPREFIX = "Ontology"
)

func SignTx(admin *Account, tx *types.Transaction) {
	signdate, err := signature.SignBySigner(tx, admin)
	if err != nil {
		log.Error(err, "signdate SignBySigner failed")
	}
	transactionCcntmract, _ := ccntmract.CreateSignatureCcntmract(admin.PublicKey)
	transactionCcntmractCcntmext := ccntmract.NewCcntmractCcntmext(tx)
	transactionCcntmractCcntmext.AddCcntmract(transactionCcntmract, admin.PublicKey, signdate)
	tx.SetPrograms(transactionCcntmractCcntmext.GetPrograms())
}
*/
