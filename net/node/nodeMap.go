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

package node

import (
	"fmt"
	"github.com/Ontology/common/config"
	//	"github.com/Ontology/common/log"
	. "github.com/Ontology/net/protocol"
	"strings"
	"sync"
)

// The neigbor node list
type nbrNodes struct {
	Lock sync.RWMutex
	List map[uint64]*node
}

func (nm *nbrNodes) Broadcast(buf []byte) {
	// TODO lock the map
	// TODO Check whether the node existed or not
	for _, node := range nm.List {
		if node.state == ESTABLISH && node.relay == true {
			go node.Tx(buf)
		}
	}
}

func (nm *nbrNodes) NodeExisted(uid uint64) bool {
	_, ok := nm.List[uid]
	return ok
}

func (nm *nbrNodes) AddNbrNode(n Noder) {
	//TODO lock the node Map
	// TODO multi client from the same IP address issue
	if (nm.NodeExisted(n.GetID())) {
               fmt.Printf("Insert a existed node\n")
	} else {
		node, err := n.(*node)
		if (err == false) {
			fmt.Println("Convert the noder error when add node")
			return
		}
		nm.List[n.GetID()] = node
	}
}

func (nm *nbrNodes) DelNbrNode(id uint64) (Noder, bool) {
	//TODO lock the node Map
	n, ok := nm.List[id]
	if (ok == false) {
		return nil, false
	}
	delete(nm.List, id)
	return n, true
}

func (nm nbrNodes) GetConnectionCnt() uint {
	//TODO lock the node Map
	var cnt uint
	for _, node := range nm.List {
		if node.state == ESTABLISH {
			cnt++
		}
	}
	return cnt
}

func (nm *nbrNodes) init() {
	nm.List = make(map[uint64]*node)
}

func (nm nbrNodes) NodeEstablished(id uint64) bool {
	n, ok := nm.List[id]
	if (ok == false) {
		return false
	}

	if (n.state != ESTABLISH) {
		return false
	}

	return true
}
