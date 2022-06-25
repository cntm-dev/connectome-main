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

package ledgerstore

import (
	"bytes"
	"fmt"

	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/config"
	"github.com/cntmio/cntmology/common/serialization"
	"github.com/cntmio/cntmology/core/genesis"
	"github.com/cntmio/cntmology/core/payload"
	"github.com/cntmio/cntmology/core/states"
	"github.com/cntmio/cntmology/core/store"
	scommon "github.com/cntmio/cntmology/core/store/common"
	"github.com/cntmio/cntmology/core/store/statestore"
	"github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/smartccntmract"
	"github.com/cntmio/cntmology/smartccntmract/event"
	"github.com/cntmio/cntmology/smartccntmract/service/native/cntm"
	neovm "github.com/cntmio/cntmology/smartccntmract/service/neovm"
	sstates "github.com/cntmio/cntmology/smartccntmract/states"
	"github.com/cntmio/cntmology/smartccntmract/storage"
	stypes "github.com/cntmio/cntmology/smartccntmract/types"
)

//HandleDeployTransaction deal with smart ccntmract deploy transaction
func (self *StateStore) HandleDeployTransaction(stateBatch *statestore.StateBatch, tx *types.Transaction) error {
	deploy := tx.Payload.(*payload.DeployCode)

	originAddress := deploy.Code.AddressFromVmCode()

	// mapping native ccntmract origin address to target address
	if deploy.Code.VmType == stypes.Native {
		targetAddress, err := common.AddressParseFromBytes(deploy.Code.Code)
		if err != nil {
			return fmt.Errorf("Invalid native ccntmract address:%v", err)

		}
		originAddress = targetAddress
	}

	// store ccntmract message
	if err := stateBatch.TryGetOrAdd(
		scommon.ST_CcntmRACT,
		originAddress[:],
		deploy); err != nil {
		return fmt.Errorf("TryGetOrAdd ccntmract error %s", err)
	}
	return nil
}

//HandleInvokeTransaction deal with smart ccntmract invoke transaction
func (self *StateStore) HandleInvokeTransaction(store store.LedgerStore, stateBatch *statestore.StateBatch, tx *types.Transaction, block *types.Block, eventStore scommon.EventStore) error {
	invoke := tx.Payload.(*payload.InvokeCode)
	txHash := tx.Hash()

	// check payer cntm balance
	balance, err := GetBalance(stateBatch, tx.Payer, genesis.OngCcntmractAddress)
	if balance < tx.GasLimit*tx.GasPrice {
		return fmt.Errorf("%v", "Payer Gas Insufficient")
	}
	// init smart ccntmract configuration info
	config := &smartccntmract.Config{
		Time:   block.Header.Timestamp,
		Height: block.Header.Height,
		Tx:     tx,
	}
	cache := storage.NewCloneCache(stateBatch)
	//init smart ccntmract info
	sc := smartccntmract.SmartCcntmract{
		Config:     config,
		CloneCache: cache,
		Store:      store,
		Code:       invoke.Code,
		Gas:        tx.GasLimit - neovm.TRANSACTION_GAS,
	}

	//start the smart ccntmract executive function
	_, err = sc.Execute()
	var state []*cntm.State
	transfers := &cntm.Transfers{
		States: append(state, &cntm.State{
			From:  tx.Payer,
			To:    genesis.GovernanceCcntmractAddress,
			Value: (tx.GasLimit - sc.Gas) * tx.GasPrice})}
	if err != nil {
		cache = storage.NewCloneCache(stateBatch)
		if err := Transfer(store, cache, genesis.OngCcntmractAddress, transfers); err != nil {
			return err
		}
		cache.Commit()
		if err := saveNotify(eventStore, txHash, &event.ExecuteNotify{TxHash: txHash,
			State: event.CcntmRACT_STATE_FAIL, Notify: []*event.NotifyEventInfo{}}); err != nil {
			return err
		}
		return err
	}
	if err := Transfer(store, cache, genesis.OngCcntmractAddress, transfers); err != nil {
		return err
	}
	if err := saveNotify(eventStore, txHash, &event.ExecuteNotify{TxHash: txHash,
		State: event.CcntmRACT_STATE_SUCCESS, Notify: sc.Notifications}); err != nil {
		return err
	}
	sc.CloneCache.Commit()

	return nil
}

func saveNotify(eventStore scommon.EventStore, txHash common.Uint256, notify *event.ExecuteNotify) error {
	if !config.DefConfig.Common.EnableEventLog {
		return nil
	}
	if err := eventStore.SaveEventNotifyByTx(txHash, notify); err != nil {
		return fmt.Errorf("SaveEventNotifyByTx error %s", err)
	}
	event.PushSmartCodeEvent(txHash, 0, event.EVENT_NOTIFY, notify)
	return nil
}

//HandleClaimTransaction deal with cntm claim transaction
func (self *StateStore) HandleClaimTransaction(stateBatch *statestore.StateBatch, tx *types.Transaction) error {
	//TODO
	return nil
}

//HandleVoteTransaction deal with vote transaction
func (self *StateStore) HandleVoteTransaction(stateBatch *statestore.StateBatch, tx *types.Transaction) error {
	vote := tx.Payload.(*payload.Vote)
	buf := new(bytes.Buffer)
	vote.Account.Serialize(buf)
	stateBatch.TryAdd(scommon.ST_VOTE, buf.Bytes(), &states.VoteState{PublicKeys: vote.PubKeys})
	return nil
}

func Transfer(store store.LedgerStore, cache *storage.CloneCache, ccntmract common.Address, transfer *cntm.Transfers) error {
	tr := new(bytes.Buffer)
	if err := transfer.Serialize(tr); err != nil {
		return err
	}
	trans := &sstates.Ccntmract{
		Address: ccntmract,
		Method:  "transfer",
		Args:    tr.Bytes(),
	}
	ts := new(bytes.Buffer)
	if err := trans.Serialize(ts); err != nil {
		return err
	}
	_, err := store.InvokeNative(cache, ts.Bytes())
	if err != nil {
		return err
	}
	return nil
}

func GetBalance(stateBatch *statestore.StateBatch, address, ccntmract common.Address) (uint64, error) {
	bl, err := stateBatch.TryGet(scommon.ST_STORAGE, append(ccntmract[:], address[:]...))
	if err != nil {
		return 0, err
	}
	if bl == nil || bl.Value == nil {
		return 0, err
	}
	item, ok := bl.Value.(*states.StorageItem)
	if !ok {
		return 0, fmt.Errorf("%s", "[GetStorageItem] instance doesn't StorageItem!")
	}
	balance, err := serialization.ReadUint64(bytes.NewBuffer(item.Value))
	if err != nil {
		return 0, err
	}
	return balance, nil
}
