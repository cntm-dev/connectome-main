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
	"os"
	"testing"

	"github.com/cntmio/cntmology-crypto/keypair"
	"github.com/cntmio/cntmology/account"
	"github.com/cntmio/cntmology/common/config"
	"github.com/cntmio/cntmology/common/log"
	vconfig "github.com/cntmio/cntmology/consensus/vbft/config"
	"github.com/cntmio/cntmology/core/genesis"
	"github.com/cntmio/cntmology/core/ledger"
)

var testBookkeeperAccounts []*account.Account

func newTestChainStore(t *testing.T) *ChainStore {
	log.InitLog(log.InfoLog, log.Stdout)
	var err error
	acct := account.NewAccount("SHA256withECDSA")
	if acct == nil {
		t.Fatalf("GetDefaultAccount error: acc is nil")
	}
	os.RemoveAll(config.DEFAULT_DATA_DIR)
	var bookkeepers []keypair.PublicKey
	if len(testBookkeeperAccounts) == 0 {
		for i := 0; i < 7; i++ {
			acc := account.NewAccount("")
			testBookkeeperAccounts = append(testBookkeeperAccounts, acc)
			bookkeepers = append(bookkeepers, acc.PublicKey)
		}
	}

	genesisConfig := config.DefConfig.Genesis

	// update peers in genesis
	for i, p := range genesisConfig.VBFT.Peers {
		if i > 0 && i <= len(testBookkeeperAccounts) {
			p.PeerPubkey = vconfig.PubkeyID(testBookkeeperAccounts[i-1].PublicKey)
		}
	}
	block, err := genesis.BuildGenesisBlock(bookkeepers, genesisConfig)
	if err != nil {
		t.Fatalf("BuildGenesisBlock error %s", err)
	}

	db, err := ledger.InitLedger(config.DEFAULT_DATA_DIR, 0, bookkeepers, block)
	if err != nil {
		t.Fatalf("NewLedger error %s", err)
	}

	chainstore, err := OpenBlockStore(db, nil)
	if err != nil {
		t.Fatalf("openblockstore failed: %v\n", err)
	}
	return chainstore
}

func cleanTestChainStore() {
	os.RemoveAll(config.DEFAULT_DATA_DIR)
	testBookkeeperAccounts = make([]*account.Account, 0)
}

func TestGetChainedBlockNum(t *testing.T) {
	chainstore := newTestChainStore(t)
	if chainstore == nil {
		t.Error("newChainStrore error")
		return
	}
	defer cleanTestChainStore()

	blocknum := chainstore.GetChainedBlockNum()
	t.Logf("TestGetChainedBlockNum :%d", blocknum)
}
