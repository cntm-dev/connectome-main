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
	"github.com/cntmio/cntmology/core/payload"
	"github.com/cntmio/cntmology/core/states"
	"github.com/cntmio/cntmology/core/store"
	scommon "github.com/cntmio/cntmology/core/store/common"
	"github.com/cntmio/cntmology/core/store/statestore"
	"github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/smartccntmract"
	"github.com/cntmio/cntmology/smartccntmract/event"
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

	// init smart ccntmract configuration info
	config := &smartccntmract.Config{
		Time:   block.Header.Timestamp,
		Height: block.Header.Height,
		Tx:     tx,
	}

	//init smart ccntmract info
	sc := smartccntmract.SmartCcntmract{
		Config:     config,
		CloneCache: storage.NewCloneCache(stateBatch),
		Store:      store,
		Code:       invoke.Code,
		Gas:        tx.GasLimit,
	}

	//start the smart ccntmract executive function
	if _, err := sc.Execute(); err != nil {
		return err
	}

	sc.CloneCache.Commit()

	if len(sc.Notifications) > 0 {
		if err := eventStore.SaveEventNotifyByTx(txHash, sc.Notifications); err != nil {
			return fmt.Errorf("SaveEventNotifyByTx error %s", err)
		}
		event.PushSmartCodeEvent(txHash, 0, event.EVENT_NOTIFY, sc.Notifications)
	}
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
