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

package vbft

import (
	"fmt"
	"os"
	"sort"
	"testing"

	"github.com/cntmio/cntmology-crypto/keypair"
	"github.com/cntmio/cntmology/account"
	"github.com/cntmio/cntmology/common/config"
	"github.com/cntmio/cntmology/common/log"
	actorTypes "github.com/cntmio/cntmology/consensus/actor"
	"github.com/cntmio/cntmology/core/ledger"
	ldgactor "github.com/cntmio/cntmology/core/ledger/actor"
	"github.com/cntmio/cntmology/core/signature"
)

func newChainStore() *ChainStore {
	log.Init(log.PATH, log.Stdout)
	var err error
	err = signature.SetDefaultScheme(config.Parameters.SignatureScheme)
	if err != nil {
		log.Warn("Config error: ", err)
	}
	passwd := string("passwordtest")
	acct := account.Open(account.WALLET_FILENAME, []byte(passwd))

	defBookkeepers, err := acct.GetBookkeepers()
	sort.Sort(keypair.NewPublicList(defBookkeepers))
	if err != nil {
		log.Fatalf("GetBookkeepers error:%s", err)
		os.Exit(1)
	}
	ledger.DefLedger, err = ledger.NewLedger()
	if err != nil {
		log.Fatalf("NewLedger error %s", err)
		os.Exit(1)
	}
	ldgerActor := ldgactor.NewLedgerActor()
	ledgerPID := ldgerActor.Start()
	var ledger *actorTypes.LedgerActor
	ledger = &actorTypes.LedgerActor{Ledger: ledgerPID}
	store, err := OpenBlockStore(ledger)
	if err != nil {
		fmt.Printf("openblockstore failed: %v\n", err)
		return nil
	}
	return store
}

func TestGetChainedBlockNum(t *testing.T) {
	chainstore := newChainStore()
	if chainstore == nil {
		t.Error("newChainStrore error")
		return
	}
	blocknum := chainstore.GetChainedBlockNum()
	t.Logf("TestGetChainedBlockNum :%d", blocknum)
}

func TestGetBlock(t *testing.T) {
	chainstore := newChainStore()
	if chainstore == nil {
		t.Error("newChainStrore error")
		return
	}
	blk, err := constructBlock()
	if err != nil {
		t.Errorf("constructBlock failed: %v", err)
		return
	}
	chainstore.pendingBlocks[1] = blk
	_, err = chainstore.GetBlock(uint64(1))
	if err != nil {
		t.Errorf("TestGetBlock failed :%v", err)
		return
	}
	t.Log("TestGetBlock succ")
}
