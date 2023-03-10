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

package statestore

import (
	"bytes"
	"fmt"

	"github.com/cntmio/cntmology/core/payload"
	"github.com/cntmio/cntmology/core/states"
	"github.com/cntmio/cntmology/core/store/common"
	"github.com/syndtr/goleveldb/leveldb"
	"strings"
)

type StateBatch struct {
	store       common.PersistStore
	memoryStore common.MemoryCacheStore
	dbErr       error
}

func NewStateStoreBatch(memoryStore common.MemoryCacheStore, store common.PersistStore) *StateBatch {
	return &StateBatch{
		store:       store,
		memoryStore: memoryStore,
	}
}

func (self *StateBatch) Find(prefix common.DataEntryPrefix, key []byte) ([]*common.StateItem, error) {
	var sts []*common.StateItem
	bp := []byte{byte(prefix)}
	iter := self.store.NewIterator(append(bp, key...))
	defer iter.Release()
	for iter.Next() {
		k := iter.Key()
		kv := k[1:]
		if self.memoryStore.Get(byte(prefix), kv) == nil {
			value := iter.Value()
			state, err := getStateObject(prefix, value)
			if err != nil {
				return nil, err
			}
			sts = append(sts, &common.StateItem{Key: string(kv), Value: state})
		}
	}
	if err := iter.Error(); err != nil {
		return nil, err
	}
	keyP := string(append(bp, key...))
	for _, v := range self.memoryStore.Find() {
		if v.State != common.Deleted && strings.HasPrefix(v.Key, keyP) {
			sts = append(sts, v.Copy())
		}
	}
	return sts, nil
}

func (self *StateBatch) TryAdd(prefix common.DataEntryPrefix, key []byte, value states.StateValue) {
	self.setStateObject(byte(prefix), key, value, common.Changed)
}

func (self *StateBatch) TryGetOrAdd(prefix common.DataEntryPrefix, key []byte, value states.StateValue) error {
	state := self.memoryStore.Get(byte(prefix), key)
	if state != nil {
		if state.State == common.Deleted {
			self.setStateObject(byte(prefix), key, value, common.Changed)
			return nil
		}
		return nil
	}
	item, err := self.store.Get(append([]byte{byte(prefix)}, key...))
	if err != nil && err != leveldb.ErrNotFound {
		return err
	}
	if item != nil {
		return nil
	}
	self.setStateObject(byte(prefix), key, value, common.Changed)
	return nil
}

func (self *StateBatch) TryGet(prefix common.DataEntryPrefix, key []byte) (*common.StateItem, error) {
	state := self.memoryStore.Get(byte(prefix), key)
	if state != nil {
		if state.State == common.Deleted {
			return nil, nil
		}
		return state, nil
	}
	enc, err := self.store.Get(append([]byte{byte(prefix)}, key...))
	if err != nil && err != leveldb.ErrNotFound {
		return nil, err
	}

	if enc == nil {
		return nil, nil
	}
	stateVal, err := getStateObject(prefix, enc)
	if err != nil {
		return nil, err
	}
	self.setStateObject(byte(prefix), key, stateVal, common.None)
	return &common.StateItem{Key: string(append([]byte{byte(prefix)}, key...)), Value: stateVal, State: common.None}, nil
}

func (self *StateBatch) TryGetAndChange(prefix common.DataEntryPrefix, key []byte) (states.StateValue, error) {
	state := self.memoryStore.Get(byte(prefix), key)
	if state != nil {
		if state.State == common.Deleted {
			return nil, nil
		} else if state.State == common.None {
			state.State = common.Changed
		}
		return state.Value, nil
	}
	k := append([]byte{byte(prefix)}, key...)
	enc, err := self.store.Get(k)
	if err != nil && err != leveldb.ErrNotFound {
		return nil, err
	}

	if enc == nil {
		return nil, nil
	}

	val, err := getStateObject(prefix, enc)
	if err != nil {
		return nil, err
	}
	self.setStateObject(byte(prefix), key, val, common.Changed)
	return val, nil
}

func (self *StateBatch) TryDelete(prefix common.DataEntryPrefix, key []byte) {
	self.memoryStore.Delete(byte(prefix), key)
}

func (self *StateBatch) CommitTo() error {
	for k, v := range self.memoryStore.GetChangeSet() {
		if v.State == common.Deleted {
			self.store.BatchDelete([]byte(k))
		} else {
			data := new(bytes.Buffer)
			err := v.Value.Serialize(data)
			if err != nil {
				return fmt.Errorf("error: key %v, value:%v", k, v.Value)
			}
			self.store.BatchPut([]byte(k), data.Bytes())
		}
	}
	return nil
}

func (self *StateBatch) setStateObject(prefix byte, key []byte, value states.StateValue, state common.ItemState) {
	self.memoryStore.Put(prefix, key, value, state)
}

func (self *StateBatch) SetError(err error) {
	if self.dbErr == nil {
		self.dbErr = err
	}
}

func (self *StateBatch) Error() error {
	return self.dbErr
}

func getStateObject(prefix common.DataEntryPrefix, enc []byte) (states.StateValue, error) {
	reader := bytes.NewBuffer(enc)
	switch prefix {
	case common.ST_BOOKKEEPER:
		bookkeeper := new(payload.Bookkeeper)
		if err := bookkeeper.Deserialize(reader); err != nil {
			return nil, err
		}
		return bookkeeper, nil
	case common.ST_CcntmRACT:
		ccntmract := new(payload.DeployCode)
		if err := ccntmract.Deserialize(reader); err != nil {
			return nil, err
		}
		return ccntmract, nil
	case common.ST_STORAGE:
		storage := new(states.StorageItem)
		if err := storage.Deserialize(reader); err != nil {
			return nil, err
		}
		return storage, nil
	default:
		panic("[getStateObject] invalid state type!")
	}
}
