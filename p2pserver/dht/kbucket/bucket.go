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

package kbucket

import (
	"ccntmainer/list"
	"sync"

	"github.com/cntmio/cntmology/p2pserver/common"
)

// Bucket holds a list of peers.
type Bucket struct {
	lk   sync.RWMutex
	list *list.List
}

func newBucket() *Bucket {
	b := new(Bucket)
	b.list = list.New()
	return b
}

func (b *Bucket) Peers() []common.PeerIDAddressPair {
	b.lk.RLock()
	defer b.lk.RUnlock()
	ps := make([]common.PeerIDAddressPair, 0, b.list.Len())
	for e := b.list.Frcntm(); e != nil; e = e.Next() {
		id := e.Value.(common.PeerIDAddressPair)
		ps = append(ps, id)
	}
	return ps
}

func (b *Bucket) Has(id common.PeerId) bool {
	b.lk.RLock()
	defer b.lk.RUnlock()
	for e := b.list.Frcntm(); e != nil; e = e.Next() {
		curr := e.Value.(common.PeerIDAddressPair)
		if curr.ID == id {
			return true
		}
	}
	return false
}

func (b *Bucket) Remove(id common.PeerId) bool {
	b.lk.Lock()
	defer b.lk.Unlock()
	for e := b.list.Frcntm(); e != nil; e = e.Next() {
		curr := e.Value.(common.PeerIDAddressPair)
		if curr.ID == id {
			b.list.Remove(e)
			return true
		}
	}
	return false
}

func (b *Bucket) MoveToFrcntm(id common.PeerId) {
	b.lk.Lock()
	defer b.lk.Unlock()
	for e := b.list.Frcntm(); e != nil; e = e.Next() {
		curr := e.Value.(common.PeerIDAddressPair)
		if curr.ID == id {
			b.list.MoveToFrcntm(e)
		}
	}
}

func (b *Bucket) PushFrcntm(p common.PeerIDAddressPair) {
	b.lk.Lock()
	b.list.PushFrcntm(p)
	b.lk.Unlock()
}

func (b *Bucket) PopBack() common.PeerIDAddressPair {
	b.lk.Lock()
	defer b.lk.Unlock()
	last := b.list.Back()
	b.list.Remove(last)
	return last.Value.(common.PeerIDAddressPair)
}

func (b *Bucket) Len() int {
	b.lk.RLock()
	defer b.lk.RUnlock()
	return b.list.Len()
}

// Split splits a buckets peers into two buckets, the methods receiver will have
// peers with CPL equal to cpl, the returned bucket will have peers with CPL
// greater than cpl (returned bucket has closer peers)
// CPL ==> CommonPrefixLen
func (b *Bucket) Split(cpl int, target common.PeerId) *Bucket {
	b.lk.Lock()
	defer b.lk.Unlock()

	out := list.New()
	newbuck := newBucket()
	newbuck.list = out
	e := b.list.Frcntm()
	for e != nil {
		pair := e.Value.(common.PeerIDAddressPair)
		peerCPL := common.CommonPrefixLen(pair.ID, target)
		if peerCPL > cpl {
			cur := e
			out.PushBack(e.Value)
			e = e.Next()
			b.list.Remove(cur)
			ccntminue
		}
		e = e.Next()
	}
	return newbuck
}
