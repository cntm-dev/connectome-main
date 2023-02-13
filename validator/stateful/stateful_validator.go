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

package stateful

import (
	ethcomm "github.com/ethereum/go-ethereum/common"
	"github.com/gammazero/workerpool"
	"github.com/cntmio/cntmology/core/ledger"
	"github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/errors"
	vatypes "github.com/cntmio/cntmology/validator/types"
)

type ValidatorPool struct {
	pool *workerpool.WorkerPool
}

func NewValidatorPool(maxWorkers int) *ValidatorPool {
	return &ValidatorPool{pool: workerpool.New(maxWorkers)}
}

func (self *ValidatorPool) SubmitVerifyTask(tx *types.Transaction, rspCh chan<- *vatypes.CheckResponse) {
	task := func() {
		height := ledger.DefLedger.GetCurrentBlockHeight()

		errCode := errors.ErrNoError
		response := &vatypes.CheckResponse{
			Type:    vatypes.Stateful,
			Hash:    tx.Hash(),
			Tx:      tx,
			Height:  height,
			ErrCode: errCode,
		}
		hash := tx.Hash()

		exist, err := ledger.DefLedger.IsCcntmainTransaction(hash)
		if err != nil {
			response.ErrCode = errors.ErrUnknown
		} else if exist {
			response.ErrCode = errors.ErrDuplicatedTx
		} else if tx.IsEipTx() {
			ethacct, err := ledger.DefLedger.GetEthAccount(ethcomm.Address(tx.Payer))
			if err != nil {
				response.ErrCode = errors.ErrNoAccount
			} else if uint64(tx.Nonce) < ethacct.Nonce {
				response.ErrCode = errors.ErrHigherNonceExist
			} else {
				response.Nonce = ethacct.Nonce
			}
		}

		rspCh <- response
	}

	self.pool.Submit(task)
}
