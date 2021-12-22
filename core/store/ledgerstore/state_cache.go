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
	"github.com/hashicorp/golang-lru"
	"github.com/Ontology/core/states"
)

const(
	STATE_CACHE_SIZE = 100000
)

type StateCache struct {
	stateCache *lru.ARCCache
}

func NewStateCache() (*StateCache, error){
	stateCache, err := lru.NewARC(STATE_CACHE_SIZE)
	if err != nil {
		return nil, err
	}
	return &StateCache{
		stateCache:stateCache,
	}, nil
}

func (this *StateCache) GetState(key []byte)states.IStateValue{
	state, ok := this.stateCache.Get(string(key))
	if !ok {
		return nil
	}
	return state.(states.IStateValue)
}

func (this *StateCache) AddState(key []byte, state states.IStateValue){
	this.stateCache.Add(string(key), state)
}

func (this *StateCache) DeleteState(key []byte){
	this.stateCache.Remove(string(key))
}
