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
	"github.com/Ontology/core/payload"
	"github.com/Ontology/core/states"
	scommon "github.com/Ontology/core/store/common"
	"github.com/Ontology/common"
	"github.com/Ontology/core/store/statestore"
	"github.com/Ontology/core/types"
	vmtypes "github.com/Ontology/vm/types"
	"github.com/Ontology/smartccntmract"
	"github.com/Ontology/core/store"
	"github.com/Ontology/smartccntmract/ccntmext"
)

func (this *StateStore) HandleDeployTransaction(stateBatch *statestore.StateBatch, tx *types.Transaction) error {
	deploy := tx.Payload.(*payload.DeployCode)

	originAddress := deploy.Code.AddressFromVmCode()
	targetAddress, err := common.AddressParseFromBytes(deploy.Code.Code)
	if err != nil {
		return fmt.Errorf("Invalid native ccntmract address:%v", err)
	}

	// mapping native ccntmract origin address to target address
	if deploy.Code.VmType == vmtypes.Native {
		if err := stateBatch.TryGetOrAdd(
			scommon.ST_Ccntmract,
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
		scommon.ST_Ccntmract,
		originAddress[:],
		deploy,
		false); err != nil {
		return fmt.Errorf("TryGetOrAdd ccntmract error %s", err)
	}
	return nil
}

func (this *StateStore) HandleInvokeTransaction(store store.ILedgerStore, stateBatch *statestore.StateBatch, tx *types.Transaction, block *types.Block, eventStore scommon.IEventStore) error {
	invoke := tx.Payload.(*payload.InvokeCode)
	//txHash := tx.Hash()

	// init smart ccntmract configuration info
	config := &smartccntmract.Config{
		Time: block.Header.Timestamp,
		Height: block.Header.Height,
		Tx: tx,
		Table: &CacheCodeTable{stateBatch},
		DBCache: stateBatch,
		Store: store,
	}

	//init smart ccntmract ccntmext info
	ctx := &ccntmext.Ccntmext{
		Code: invoke.Code,
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

	//if len(sc.Notifications) > 0 {
	//	if err := eventStore.SaveEventNotifyByTx(txHash, sc.Notifications); err != nil {
	//		return fmt.Errorf("SaveEventNotifyByTx error %s", err)
	//	}
	//}

	return nil
}

func (this *StateStore) HandleClaimTransaction(stateBatch *statestore.StateBatch, tx *types.Transaction) error {
	//TODO
	//p := tx.Payload.(*payload.Claim)
	//for _, c := range p.Claims {
	//	state, err := this.TryGetAndChange(ST_SpentCoin, c.ReferTxID.ToArray(), false)
	//	if err != nil {
	//		return fmt.Errorf("TryGetAndChange error %s", err)
	//	}
	//	spentcoins := state.(*SpentCoinState)
	//
	//	newItems := make([]*Item, 0, len(spentcoins.Items)-1)
	//	for index, item := range spentcoins.Items {
	//		if uint16(index) != c.ReferTxOutputIndex {
	//			newItems = append(newItems, item)
	//		}
	//	}
	//	spentcoins.Items = newItems
	//}
	return nil
}

func (this *StateStore) HandleEnrollmentTransaction(stateBatch *statestore.StateBatch, tx *types.Transaction) error {
	en := tx.Payload.(*payload.Enrollment)
	bf := new(bytes.Buffer)
	if err := en.PublicKey.Serialize(bf); err != nil {
		return err
	}
	stateBatch.TryAdd(scommon.ST_Validator, bf.Bytes(), &states.ValidatorState{PublicKey: en.PublicKey}, false)
	return nil
}

func (this *StateStore) HandleVoteTransaction(stateBatch *statestore.StateBatch, tx *types.Transaction) error {
	vote := tx.Payload.(*payload.Vote)
	buf := new(bytes.Buffer)
	vote.Account.Serialize(buf)
	stateBatch.TryAdd(scommon.ST_Vote, buf.Bytes(), &states.VoteState{PublicKeys: vote.PubKeys}, false)
	return nil
}




