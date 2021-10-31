package ledgerstore

import (
	"fmt"
	"github.com/Ontology/common"
	"github.com/Ontology/core/code"
	"github.com/Ontology/core/states"
	scommon "github.com/Ontology/core/store/common"
	"github.com/Ontology/core/store/statestore"
	"github.com/Ontology/crypto"
	"github.com/Ontology/vm/types"
	"testing"
)

func init() {
	crypto.SetAlg("")
}

func TestCcntmractState(t *testing.T) {
	batch, err := getStateBatch(&common.Uint256{})
	if err != nil {
		t.Errorf("NewStateBatch error %s", err)
		return
	}
	testCode := []byte("testcode")
	codeHash := common.ToCodeHash(testCode)
	ccntmactState := &states.CcntmractState{
		Code:        &code.FunctionCode{Code: testCode},
		VmType:      types.NEOVM,
		NeedStorage: false,
		Name:        "test",
		Version:     "1.0",
		Author:      "test",
		Email:       "test",
		Description: "test",
	}
	batch.TryAdd(scommon.ST_Ccntmract, codeHash.ToArray(), ccntmactState, false)
	_, err = batch.CommitTo()
	if err != nil {
		t.Errorf("batch.CommitTo error %s", err)
		return
	}
	err = testStateStore.CommitTo()
	if err != nil {
		t.Errorf("testStateStore.CommitTo error %s", err)
		return
	}
	ccntmractState1, err := testStateStore.GetCcntmractState(&codeHash)
	if err != nil {
		t.Errorf("GetCcntmractState error %s", err)
		return
	}
	if ccntmractState1.Name != ccntmactState.Name ||
		ccntmractState1.Version != ccntmactState.Version ||
		ccntmractState1.Author != ccntmactState.Author ||
		ccntmractState1.Description != ccntmactState.Description ||
		ccntmractState1.Email != ccntmactState.Email {
		t.Errorf("TestCcntmractState failed %+v != %+v", ccntmractState1, ccntmactState)
		return
	}
}

func TestBookKeeperState(t *testing.T) {
	batch, err := getStateBatch(&common.Uint256{})
	if err != nil {
		t.Errorf("NewStateBatch error %s", err)
		return
	}

	_, pubKey1, _ := crypto.GenKeyPair()
	_, pubKey2, _ := crypto.GenKeyPair()
	currBookKeepers := make([]*crypto.PubKey, 0)
	currBookKeepers = append(currBookKeepers, &pubKey1)
	currBookKeepers = append(currBookKeepers, &pubKey2)
	nextBookKeepers := make([]*crypto.PubKey, 0)
	nextBookKeepers = append(nextBookKeepers, &pubKey1)
	nextBookKeepers = append(nextBookKeepers, &pubKey2)

	bookKeeperState := &states.BookKeeperState{
		CurrBookKeeper: currBookKeepers,
		NextBookKeeper: nextBookKeepers,
	}
	batch.TryAdd(scommon.ST_BookKeeper, BookerKeeper, bookKeeperState, false)
	_, err = batch.CommitTo()
	if err != nil {
		t.Errorf("batch.CommitTo error %s", err)
		return
	}
	err = testStateStore.CommitTo()
	if err != nil {
		t.Errorf("testStateStore.CommitTo error %s", err)
		return
	}
	bookState, err := testStateStore.GetBookKeeperState()
	if err != nil {
		t.Errorf("GetBookKeeperState error %s", err)
		return
	}
	currBookKeepers1 := bookState.CurrBookKeeper
	nextBookKeepers1 := bookState.NextBookKeeper
	for index, pk := range currBookKeepers {
		pk1 := currBookKeepers1[index]
		if pk.X.Cmp(pk1.X) != 0 || pk.Y.Cmp(pk1.Y) != 0 {
			t.Errorf("TestBookKeeperState currentBookKeeper failed")
			return
		}
	}
	for index, pk := range nextBookKeepers {
		pk1 := nextBookKeepers1[index]
		if pk.X.Cmp(pk1.X) != 0 || pk.Y.Cmp(pk1.Y) != 0 {
			t.Errorf("TestBookKeeperState currentBookKeeper failed")
			return
		}
	}
}

func getStateBatch(stateRoot *common.Uint256) (*statestore.StateBatch, error) {
	err := testStateStore.NewBatch()
	if err != nil {
		return nil, fmt.Errorf("testStateStore.NewBatch error %s", err)
	}
	batch, err := testStateStore.NewStateBatch(&common.Uint256{})
	if err != nil {
		return nil, fmt.Errorf("NewStateBatch error %s", err)
	}
	return batch, nil
}
