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

package neovm

import (
	"bytes"
	"fmt"
	"github.com/Ontology/common"
	"github.com/Ontology/core/states"
	scommon "github.com/Ontology/core/store/common"
	"github.com/Ontology/core/store"
	"github.com/Ontology/errors"
	"github.com/Ontology/smartccntmract/storage"
	stypes "github.com/Ontology/smartccntmract/types"
	vm "github.com/Ontology/vm/neovm"
	"github.com/Ontology/core/payload"
	vmtypes "github.com/Ontology/vm/types"
)

type StateMachine struct {
	*StateReader
	ldgerStore store.ILedgerStore
	CloneCache *storage.CloneCache
	trigger    stypes.TriggerType
	time       uint32
}

func NewStateMachine(ldgerStore store.ILedgerStore, dbCache scommon.IStateStore, trigger stypes.TriggerType, time uint32) *StateMachine {
	var stateMachine StateMachine
	stateMachine.ldgerStore = ldgerStore
	stateMachine.CloneCache = storage.NewCloneCache(dbCache)
	stateMachine.StateReader = NewStateReader(ldgerStore, trigger)
	stateMachine.trigger = trigger
	stateMachine.time = time

	stateMachine.StateReader.Register("Neo.Runtime.GetTrigger", stateMachine.RuntimeGetTrigger)
	stateMachine.StateReader.Register("Neo.Runtime.GetTime", stateMachine.RuntimeGetTime)

	stateMachine.StateReader.Register("Neo.Ccntmract.Create", stateMachine.CcntmractCreate)
	stateMachine.StateReader.Register("Neo.Ccntmract.Migrate", stateMachine.CcntmractMigrate)
	stateMachine.StateReader.Register("Neo.Ccntmract.GetStorageCcntmext", stateMachine.GetStorageCcntmext)
	stateMachine.StateReader.Register("Neo.Ccntmract.GetScript", stateMachine.CcntmractGetCode)
	stateMachine.StateReader.Register("Neo.Ccntmract.Destroy", stateMachine.CcntmractDestory)

	stateMachine.StateReader.Register("Neo.Storage.Get", stateMachine.StorageGet)
	stateMachine.StateReader.Register("Neo.Storage.Put", stateMachine.StoragePut)
	stateMachine.StateReader.Register("Neo.Storage.Delete", stateMachine.StorageDelete)
	return &stateMachine
}

func (s *StateMachine) RuntimeGetTrigger(engine *vm.ExecutionEngine) (bool, error) {
	vm.PushData(engine, int(s.trigger))
	return true, nil
}

func (s *StateMachine) RuntimeGetTime(engine *vm.ExecutionEngine) (bool, error) {
	vm.PushData(engine, s.time)
	return true, nil
}

func (s *StateMachine) CcntmractCreate(engine *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(engine) < 7 {
		return false, errors.NewErr("[CcntmractCreate] Too few input parameters")
	}
	code := vm.PopByteArray(engine); if len(code) > 1024 * 1024 {
		return false, errors.NewErr("[CcntmractCreate] Code too lcntm!")
	}
	needStorage := vm.PopBoolean(engine)
	name := vm.PopByteArray(engine); if len(name) > 252 {
		return false, errors.NewErr("[CcntmractCreate] Name too lcntm!")
	}
	version := vm.PopByteArray(engine); if len(version) > 252 {
		return false, errors.NewErr("[CcntmractCreate] Version too lcntm!")
	}
	author := vm.PopByteArray(engine); if len(author) > 252 {
		return false, errors.NewErr("[CcntmractCreate] Author too lcntm!")
	}
	email := vm.PopByteArray(engine); if len(email) > 252 {
		return false, errors.NewErr("[CcntmractCreate] Email too lcntm!")
	}
	desc := vm.PopByteArray(engine); if len(desc) > 65536 {
		return false, errors.NewErr("[CcntmractCreate] Desc too lcntm!")
	}
	vmCode := &vmtypes.VmCode{VmType:vmtypes.NEOVM, Code: code}
	ccntmractState := &payload.DeployCode{
		Code:        vmCode,
		NeedStorage: needStorage,
		Name:        string(name),
		Version:     string(version),
		Author:      string(author),
		Email:       string(email),
		Description: string(desc),
	}
	ccntmractAddress := vmCode.AddressFromVmCode()
	state, err := s.CloneCache.GetOrAdd(scommon.ST_Ccntmract, ccntmractAddress[:], ccntmractState)
	if err != nil {
		return false, errors.NewDetailErr(err, errors.ErrNoCode, "[CcntmractCreate] GetOrAdd error!")
	}
	vm.PushData(engine, state)
	return true, nil
}

func (s *StateMachine) CcntmractMigrate(engine *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(engine) < 7 {
		return false, errors.NewErr("[CcntmractMigrate] Too few input parameters ")
	}
	code := vm.PopByteArray(engine); if len(code) > 1024 * 1024 {
		return false, errors.NewErr("[CcntmractMigrate] Code too lcntm!")
	}
	vmCode := &vmtypes.VmCode{
		Code: code,
		VmType: vmtypes.NEOVM,
	}
	ccntmractAddress := vmCode.AddressFromVmCode()
	item, err := s.CloneCache.Get(scommon.ST_Ccntmract, ccntmractAddress[:]); if err != nil {
		return false, errors.NewErr("[CcntmractMigrate] Get Ccntmract error!")
	}
	if item != nil {
		return false, errors.NewErr("[CcntmractMigrate] Migrate Ccntmract has exist!")
	}

	nameByte := vm.PopByteArray(engine); if len(nameByte) > 252 {
		return false, errors.NewErr("[CcntmractMigrate] Name too lcntm!")
	}
	versionByte := vm.PopByteArray(engine); if len(versionByte) > 252 {
		return false, errors.NewErr("[CcntmractMigrate] Version too lcntm!")
	}
	authorByte := vm.PopByteArray(engine); if len(authorByte) > 252 {
		return false, errors.NewErr("[CcntmractMigrate] Author too lcntm!")
	}
	emailByte := vm.PopByteArray(engine); if len(emailByte) > 252 {
		return false, errors.NewErr("[CcntmractMigrate] Email too lcntm!")
	}
	descByte := vm.PopByteArray(engine); if len(descByte) > 65536 {
		return false, errors.NewErr("[CcntmractMigrate] Desc too lcntm!")
	}
	ccntmractState := &payload.DeployCode{
		Code:        vmCode,
		Name:        string(nameByte),
		Version:     string(versionByte),
		Author:      string(authorByte),
		Email:       string(emailByte),
		Description: string(descByte),
	}
	s.CloneCache.Add(scommon.ST_Ccntmract, ccntmractAddress[:], ccntmractState)
	stateValues, err := s.CloneCache.Store.Find(scommon.ST_Ccntmract, ccntmractAddress[:]); if err != nil {
		return false, errors.NewDetailErr(err, errors.ErrNoCode, "[CcntmractMigrate] Find error!")
	}
	for _, v := range stateValues {
		key := new(states.StorageKey)
		bf := bytes.NewBuffer([]byte(v.Key))
		if err := key.Deserialize(bf); err != nil {
			return false, errors.NewErr("[CcntmractMigrate] Key deserialize error!")
		}
		key = &states.StorageKey{CodeHash: ccntmractAddress, Key: key.Key}
		b := new(bytes.Buffer)
		if _, err := key.Serialize(b); err != nil {
			return false, errors.NewErr("[CcntmractMigrate] Key Serialize error!")
		}
		s.CloneCache.Add(scommon.ST_Storage, key.ToArray(), v.Value)
	}
	vm.PushData(engine, ccntmractState)
	return s.CcntmractDestory(engine)
}

func (s *StateMachine) CcntmractDestory(engine *vm.ExecutionEngine) (bool, error) {
	ccntmext, err := engine.CurrentCcntmext(); if err != nil {
		return false, err
	}
	hash, err := ccntmext.GetCodeHash(); if err != nil {
		return false, nil
	}
	item, err := s.CloneCache.Store.TryGet(scommon.ST_Ccntmract, hash[:]); if err != nil {
		return false, err
	}
	if item == nil {
		return false, nil
	}
	s.CloneCache.Delete(scommon.ST_Ccntmract, hash[:])
	stateValues, err := s.CloneCache.Store.Find(scommon.ST_Ccntmract, hash[:]); if err != nil {
		return false, errors.NewDetailErr(err, errors.ErrNoCode, "[CcntmractDestory] Find error!")
	}
	for _, v := range stateValues {
		s.CloneCache.Delete(scommon.ST_Storage, []byte(v.Key))
	}
	return true, nil
}

func (s *StateMachine) CheckStorageCcntmext(ccntmext *StorageCcntmext) (bool, error) {
	item, err := s.CloneCache.Get(scommon.ST_Ccntmract, ccntmext.codeHash[:])
	if err != nil {
		return false, err
	}
	if item == nil {
		return false, errors.NewErr(fmt.Sprintf("get ccntmract by codehash=%v nil", ccntmext.codeHash))
	}
	return true, nil
}

func (s *StateMachine) StoragePut(engine *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(engine) < 3 {
		return false, errors.NewErr("[StoragePut] Too few input parameters ")
	}
	opInterface := vm.PopInteropInterface(engine); if opInterface == nil {
		return false, errors.NewErr("[StoragePut] Get StorageCcntmext nil")
	}
	ccntmext := opInterface.(*StorageCcntmext)
	key := vm.PopByteArray(engine)
	if len(key) > 1024 {
		return false, errors.NewErr("[StoragePut] Get Storage key to lcntm")
	}
	value := vm.PopByteArray(engine)
	k, err := serializeStorageKey(ccntmext.codeHash, key); if err != nil {
		return false, err
	}
	s.CloneCache.Add(scommon.ST_Storage, k, &states.StorageItem{Value: value})
	return true, nil
}

func (s *StateMachine) StorageDelete(engine *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(engine) < 2 {
		return false, errors.NewErr("[StorageDelete] Too few input parameters ")
	}
	opInterface := vm.PopInteropInterface(engine)
	if opInterface == nil {
		return false, errors.NewErr("[StorageDelete] Get StorageCcntmext nil")
	}
	ccntmext := opInterface.(*StorageCcntmext)
	key := vm.PopByteArray(engine)
	k, err := serializeStorageKey(ccntmext.codeHash, key); if err != nil {
		return false, err
	}
	s.CloneCache.Delete(scommon.ST_Storage, k)
	return true, nil
}

func (s *StateMachine) StorageGet(engine *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(engine) < 2 {
		return false, errors.NewErr("[StorageGet] Too few input parameters ")
	}
	opInterface := vm.PopInteropInterface(engine)
	if opInterface == nil {
		return false, errors.NewErr("[StorageGet] Get StorageCcntmext error!")
	}
	ccntmext := opInterface.(*StorageCcntmext)
	if exist, err := s.CheckStorageCcntmext(ccntmext); !exist {
		return false, err
	}
	key := vm.PopByteArray(engine)
	k, err := serializeStorageKey(ccntmext.codeHash, key); if err != nil {
		return false, err
	}
	item, err := s.CloneCache.Get(scommon.ST_Storage, k); if err != nil {
		return false, err
	}
	if item == nil {
		vm.PushData(engine, []byte{})
	} else {
		vm.PushData(engine, item.(*states.StorageItem).Value)
	}
	return true, nil
}

func (s *StateMachine) GetStorageCcntmext(engine *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(engine) < 1 {
		return false, errors.NewErr("[GetStorageCcntmext] Too few input parameters ")
	}
	opInterface := vm.PopInteropInterface(engine)
	if opInterface == nil {
		return false, errors.NewErr("[GetStorageCcntmext] Get StorageCcntmext nil")
	}
	ccntmractState := opInterface.(*payload.DeployCode)
	codeHash := ccntmractState.Code.AddressFromVmCode()
	item, err := s.CloneCache.Store.TryGet(scommon.ST_Ccntmract, codeHash[:])
	if err != nil {
		return false, errors.NewDetailErr(err, errors.ErrNoCode, "[GetStorageCcntmext] Get StorageCcntmext nil")
	}
	ccntmext, err := engine.CurrentCcntmext(); if err != nil {
		return false, err
	}
	if item == nil {
		return false, errors.NewErr(fmt.Sprintf("[GetStorageCcntmext] Get ccntmract by codehash:%v nil", codeHash))
	}
	currentHash, err := ccntmext.GetCodeHash(); if err != nil {
		return false, err
	}
	if codeHash != currentHash {
		return false, errors.NewErr("[GetStorageCcntmext] CodeHash not equal!")
	}
	vm.PushData(engine, &StorageCcntmext{codeHash: codeHash})
	return true, nil
}

func ccntmains(programHashes []common.Address, programHash common.Address) bool {
	for _, v := range programHashes {
		if v == programHash {
			return true
		}
	}
	return false
}

func serializeStorageKey(codeHash common.Address, key []byte) ([]byte, error) {
	bf := new(bytes.Buffer)
	storageKey := &states.StorageKey{CodeHash: codeHash, Key: key}
	if _, err := storageKey.Serialize(bf); err != nil {
		return []byte{}, errors.NewErr("[serializeStorageKey] StorageKey serialize error!")
	}
	return bf.Bytes(), nil
}
