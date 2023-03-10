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

package validation

import (
	"errors"
	"fmt"

	"github.com/Ontology/core/ledger"
	"github.com/Ontology/core/types"
	"github.com/Ontology/crypto"
	. "github.com/Ontology/errors"
)

func VerifyBlock(block *types.Block, ld *ledger.Ledger, completely bool) error {
	header := block.Header
	if header.Height == 0 {
		return nil
	}

	m := len(header.BookKeepers) - (len(header.BookKeepers)-1)/3
	hash := block.Hash()
	err := crypto.VerifyMultiSignature(hash[:], header.BookKeepers, m, header.SigData)
	if err != nil {
		return err
	}

	prevHeader, err := ld.GetHeaderByHash(block.Header.PrevBlockHash)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[BlockValidator], Cannnot find prevHeader..")
	}

	err = VerifyHeader(block.Header, prevHeader)
	if err != nil {
		return err
	}

	if block.Transactions == nil {
		return errors.New(fmt.Sprintf("No Transactions Exist in Block."))
	}

	if block.Transactions[0].TxType != types.BookKeeping {
		return errors.New(fmt.Sprintf("Header Verify failed first Transacion in block is not BookKeeping type."))
	}

	for index, v := range block.Transactions {
		if v.TxType == types.BookKeeping && index != 0 {
			return errors.New(fmt.Sprintf("This Block Has BookKeeping transaction after first transaction in block."))
		}
	}

	//verfiy block's transactions
	if completely {
		/*
			//TODO: NextBookKeeper Check.
			bookKeeperaddress, err := ledger.GetBookKeeperAddress(ld.Blockchain.GetBookKeepersByTXs(block.Transactions))
			if err != nil {
				return errors.New(fmt.Sprintf("GetBookKeeperAddress Failed."))
			}
			if block.Header.NextBookKeeper != bookKeeperaddress {
				return errors.New(fmt.Sprintf("BookKeeper is not validate."))
			}
		*/
		for _, txVerify := range block.Transactions {
			if errCode := VerifyTransaction(txVerify); errCode != ErrNoError {
				return errors.New(fmt.Sprintf("VerifyTransaction failed when verifiy block"))
			}

			if errCode := VerifyTransactionWithLedger(txVerify, ld); errCode != ErrNoError {
				return errors.New(fmt.Sprintf("VerifyTransaction failed when verifiy block"))
			}
		}
	}

	return nil
}

func VerifyHeader(header, prevHeader *types.Header) error {
	if header.Height == 0 {
		return nil
	}

	if prevHeader == nil {
		return NewDetailErr(errors.New("[BlockValidator] error"), ErrNoCode, "[BlockValidator], Cannnot find previous block.")
	}

	if prevHeader.Height+1 != header.Height {
		return NewDetailErr(errors.New("[BlockValidator] error"), ErrNoCode, "[BlockValidator], block height is incorrect.")
	}

	if prevHeader.Timestamp >= header.Timestamp {
		return NewDetailErr(errors.New("[BlockValidator] error"), ErrNoCode, "[BlockValidator], block timestamp is incorrect.")
	}

	address, err := types.AddressFromBookKeepers(header.BookKeepers)
	if err != nil {
		return err
	}

	if prevHeader.NextBookKeeper != address {
		return fmt.Errorf("bookkeeper address error")
	}

	return nil
}
