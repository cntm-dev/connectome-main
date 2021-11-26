/*
 * Copyright (C) 2018 The cntmology Authors
 * This file is part of The cntmology library.
 *
 * The cntmology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The cntmology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * alcntm with The cntmology.  If not, see <http://www.gnu.org/licenses/>.
 */

package common

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
