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
	"fmt"
	"github.com/Ontology/core/states"
	scommon "github.com/Ontology/core/store/common"
	"github.com/Ontology/core/store/statestore"
	"github.com/Ontology/core/payload"
	"github.com/Ontology/crypto"
	vmtypes"github.com/Ontology/vm/types"
	"testing"
)

func init() {
	crypto.SetAlg("")
}

func TestCcntmractState(t *testing.T) {
	batch, err := getStateBatch()
	if err != nil {
		t.Errorf("NewStateBatch error %s", err)
		return
	}
	testCode := []byte("testcode")

	deploy := &payload.DeployCode{
		Code:        testCode,
		VmType:      vmtypes.NEOVM,
		NeedStorage: false,
		Name:        "testsm",
		Version:     "v1.0",
		Author:     "",
		Email:       "",
		Description: "",
	}
	code := &vmtypes.VmCode{
		Code:   testCode,
		VmType: vmtypes.NEOVM,
	}
	codeHash := code.AddressFromVmCode()
	err = batch.TryGetOrAdd(
		scommon.ST_Ccntmract,
		codeHash[:],
		deploy,
		false)
	if err != nil {
		 t.Errorf("TryGetOrAdd ccntmract error %s", err)
		 return
	}

	err = batch.CommitTo()
	if err != nil {
		t.Errorf("batch.CommitTo error %s", err)
		return
	}
	err = testStateStore.CommitTo()
	if err != nil {
		t.Errorf("testStateStore.CommitTo error %s", err)
		return
	}
	ccntmractState1, err := testStateStore.GetCcntmractState(codeHash)
	if err != nil {
		t.Errorf("GetCcntmractState error %s", err)
		return
	}
	if ccntmractState1.Name != deploy.Name ||
		ccntmractState1.Version != deploy.Version ||
		ccntmractState1.Author != deploy.Author ||
		ccntmractState1.Description != deploy.Description ||
		ccntmractState1.Email != deploy.Email {
		t.Errorf("TestCcntmractState failed %+v != %+v", ccntmractState1, deploy)
		return
	}
}

func TestBookkeeperState(t *testing.T) {
	batch, err := getStateBatch()
	if err != nil {
		t.Errorf("NewStateBatch error %s", err)
		return
	}

	_, pubKey1, _ := crypto.GenKeyPair()
	_, pubKey2, _ := crypto.GenKeyPair()
	currBookkeepers := make([]*crypto.PubKey, 0)
	currBookkeepers = append(currBookkeepers, &pubKey1)
	currBookkeepers = append(currBookkeepers, &pubKey2)
	nextBookkeepers := make([]*crypto.PubKey, 0)
	nextBookkeepers = append(nextBookkeepers, &pubKey1)
	nextBookkeepers = append(nextBookkeepers, &pubKey2)

	bookkeeperState := &states.BookkeeperState{
		CurrBookkeeper: currBookkeepers,
		NextBookkeeper: nextBookkeepers,
	}
	batch.TryAdd(scommon.ST_Bookkeeper, BookerKeeper, bookkeeperState, false)
	err = batch.CommitTo()
	if err != nil {
		t.Errorf("batch.CommitTo error %s", err)
		return
	}
	err = testStateStore.CommitTo()
	if err != nil {
		t.Errorf("testStateStore.CommitTo error %s", err)
		return
	}
	bookState, err := testStateStore.GetBookkeeperState()
	if err != nil {
		t.Errorf("GetBookkeeperState error %s", err)
		return
	}
	currBookkeepers1 := bookState.CurrBookkeeper
	nextBookkeepers1 := bookState.NextBookkeeper
	for index, pk := range currBookkeepers {
		pk1 := currBookkeepers1[index]
		if pk.X.Cmp(pk1.X) != 0 || pk.Y.Cmp(pk1.Y) != 0 {
			t.Errorf("TestBookkeeperState currentBookkeeper failed")
			return
		}
	}
	for index, pk := range nextBookkeepers {
		pk1 := nextBookkeepers1[index]
		if pk.X.Cmp(pk1.X) != 0 || pk.Y.Cmp(pk1.Y) != 0 {
			t.Errorf("TestBookkeeperState currentBookkeeper failed")
			return
		}
	}
}

func getStateBatch() (*statestore.StateBatch, error) {
	err := testStateStore.NewBatch()
	if err != nil {
		return nil, fmt.Errorf("testStateStore.NewBatch error %s", err)
	}
	batch := testStateStore.NewStateBatch()
	return batch, nil
}
