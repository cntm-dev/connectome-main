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

package proc

import (
	"os"
	"testing"
	"time"

	"github.com/cntmio/cntmology/common/config"
	"github.com/cntmio/cntmology/common/log"
	"github.com/cntmio/cntmology/core/genesis"
	"github.com/cntmio/cntmology/core/ledger"

	"github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/events/message"
	tc "github.com/cntmio/cntmology/txnpool/common"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	log.InitLog(log.InfoLog, log.Stdout)
	bookKeepers, err := config.DefConfig.GetBookkeepers()
	if err != nil {
		return
	}
	genesisConfig := config.DefConfig.Genesis
	genesisBlock, err := genesis.BuildGenesisBlock(bookKeepers, genesisConfig)
	if err != nil {
		return
	}
	ledger.DefLedger, err = ledger.InitLedger(config.DEFAULT_DATA_DIR, 0, bookKeepers, genesisBlock)
	if err != nil {
		return
	}

	m.Run()

	ledger.DefLedger.Close()
	os.RemoveAll(config.DEFAULT_DATA_DIR)
}

func TestTxPoolActor(t *testing.T) {
	t.Log("Starting tx pool actor test")
	s := NewTxPoolServer(true, false)
	if s == nil {
		t.Error("Test case: new tx pool server failed")
		return
	}

	txPoolActor := NewTxPoolActor(s)
	txPoolPid := startActor(txPoolActor)
	if txPoolPid == nil {
		t.Error("Test case: start tx actor failed")
		s.Stop()
		return
	}

	txEntry := &tc.VerifiedTx{
		Tx:             txn,
		VerifiedHeight: 10,
	}
	e := s.txPool.AddTxList(txEntry)
	assert.True(t, e.Success())

	future := txPoolPid.RequestFuture(&tc.GetTxnPoolReq{ByCount: false}, 2*time.Second)
	result, err := future.Result()
	assert.Nil(t, err)
	rsp := (result).(*tc.GetTxnPoolRsp)
	assert.NotNil(t, rsp.TxnPool)

	bk := &tc.VerifyBlockReq{
		Height: 0,
		Txs:    []*types.Transaction{txn},
	}
	future = txPoolPid.RequestFuture(bk, 10*time.Second)
	result, err = future.Result()
	assert.Nil(t, err)

	sbc := &message.SaveBlockCompleteMsg{}
	txPoolPid.Tell(sbc)

	s.Stop()
	t.Log("Ending tx pool actor test")
}
