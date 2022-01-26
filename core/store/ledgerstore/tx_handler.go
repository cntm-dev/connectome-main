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
	"github.com/cntmio/cntmology/smartccntmract/ccntmext"
	"github.com/cntmio/cntmology/smartccntmract/event"
	vmtypes "github.com/cntmio/cntmology/vm/types"
)

const (
	INVOKE_TRANSACTION = "InvokeTransaction"
)

func (self *StateStore) HandleDeployTransaction(stateBatch *statestore.StateBatch, tx *types.Transaction) error {
	deploy := tx.Payload.(*payload.DeployCode)

	originAddress := deploy.Code.AddressFromVmCode()

	// mapping native ccntmract origin address to target address
	if deploy.Code.VmType == vmtypes.Native {
		targetAddress, err := common.AddressParseFromBytes(deploy.Code.Code)
		if err != nil {
			return fmt.Errorf("Invalid native ccntmract address:%v", err)
		}
		if err := stateBatch.TryGetOrAdd(
			scommon.ST_CcntmRACT,
			targetAddress[:],
			&states.CcntmractMapping{
				OriginAddress: originAddress,
				TargetAddress: targetAddress,
			},
			false); err != nil {
			return fmt.Errorf("TryGetOrAdd ccntmract error %s", err)
		}
	}

	// store ccntmract message
	if err := stateBatch.TryGetOrAdd(
		scommon.ST_CcntmRACT,
		originAddress[:],
		deploy,
		false); err != nil {
		return fmt.Errorf("TryGetOrAdd ccntmract error %s", err)
	}
	return nil
}

func (self *StateStore) HandleInvokeTransaction(store store.LedgerStore, stateBatch *statestore.StateBatch, tx *types.Transaction, block *types.Block, eventStore scommon.EventStore) error {
	invoke := tx.Payload.(*payload.InvokeCode)
	txHash := tx.Hash()

	// init smart ccntmract configuration info
	config := &smartccntmract.Config{
		Time:    block.Header.Timestamp,
		Height:  block.Header.Height,
		Tx:      tx,
		Table:   &CacheCodeTable{stateBatch},
		DBCache: stateBatch,
		Store:   store,
	}

	//init smart ccntmract ccntmext info
	ctx := &ccntmext.Ccntmext{
		Code:            invoke.Code,
		CcntmractAddress: invoke.Code.AddressFromVmCode(),
	}

	//init smart ccntmract info
	sc := smartccntmract.SmartCcntmract{
		Config: config,
	}

	//load current ccntmext to smart ccntmract
	sc.PushCcntmext(ctx)

	//start the smart ccntmract executive function
	if err := sc.Execute(); err != nil {
		return err
	}

	if len(sc.Notifications) > 0 {
		if err := eventStore.SaveEventNotifyByTx(txHash, sc.Notifications); err != nil {
			return fmt.Errorf("SaveEventNotifyByTx error %s", err)
		}
		event.PushSmartCodeEvent(txHash, 0, INVOKE_TRANSACTION, sc.Notifications)
	}
	return nil
}

func (self *StateStore) HandleClaimTransaction(stateBatch *statestore.StateBatch, tx *types.Transaction) error {
	//TODO
	return nil
}

func (self *StateStore) HandleVoteTransaction(stateBatch *statestore.StateBatch, tx *types.Transaction) error {
	vote := tx.Payload.(*payload.Vote)
	buf := new(bytes.Buffer)
	vote.Account.Serialize(buf)
	stateBatch.TryAdd(scommon.ST_VOTE, buf.Bytes(), &states.VoteState{PublicKeys: vote.PubKeys}, false)
	return nil
}
