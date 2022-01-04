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
	"testing"

	"github.com/Ontology/core/payload"
	"github.com/Ontology/core/states"
	scommon "github.com/Ontology/core/store/common"
	"github.com/Ontology/core/store/statestore"
	vmtypes "github.com/Ontology/vm/types"
	"github.com/cntmio/cntmology-crypto/keypair"
)

func TestCcntmractState(t *testing.T) {
	batch, err := getStateBatch()
	if err != nil {
		t.Errorf("NewStateBatch error %s", err)
		return
	}
	testCode := []byte("testcode")

	vmCode := &vmtypes.VmCode{
		VmType: vmtypes.NEOVM,
		Code:   testCode,
	}
	deploy := &payload.DeployCode{
		Code:        vmCode,
		NeedStorage: false,
		Name:        "testsm",
		Version:     "v1.0",
		Author:      "",
		Email:       "",
		Description: "",
	}
	code := &vmtypes.VmCode{
		Code:   testCode,
		VmType: vmtypes.NEOVM,
	}
	codeHash := code.AddressFromVmCode()
	err = batch.TryGetOrAdd(
		scommon.ST_CcntmRACT,
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

	_, pubKey1, _ := keypair.GenerateKeyPair(keypair.PK_ECDSA, keypair.P256)
	_, pubKey2, _ := keypair.GenerateKeyPair(keypair.PK_ECDSA, keypair.P256)
	currBookkeepers := make([]keypair.PublicKey, 0)
	currBookkeepers = append(currBookkeepers, &pubKey1)
	currBookkeepers = append(currBookkeepers, &pubKey2)
	nextBookkeepers := make([]keypair.PublicKey, 0)
	nextBookkeepers = append(nextBookkeepers, &pubKey1)
	nextBookkeepers = append(nextBookkeepers, &pubKey2)

	bookkeeperState := &states.BookkeeperState{
		CurrBookkeeper: currBookkeepers,
		NextBookkeeper: nextBookkeepers,
	}
	batch.TryAdd(scommon.ST_BOOK_KEEPER, BookerKeeper, bookkeeperState, false)
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
		if !keypair.ComparePublicKey(pk, pk1) {
			t.Errorf("TestBookkeeperState currentBookkeeper failed")
			return
		}
	}
	for index, pk := range nextBookkeepers {
		pk1 := nextBookkeepers1[index]
		if !keypair.ComparePublicKey(pk, pk1) {
			t.Errorf("TestBookkeeperState nextBookkeeper failed")
			return
		}
	}
}

func getStateBatch() (*statestore.StateBatch, error) {
	testStateStore.NewBatch()
	batch := testStateStore.NewStateBatch()
	return batch, nil
}
