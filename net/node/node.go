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
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"runtime"
	"time"
)

// The node capability flag
const (
	RELAY        = 0x01
	SERVER       = 0x02
	NODESERVICES = 0x01
)

type node struct {
	state    uint   // node status
	id       uint64 // The nodes's id
	cap      uint32 // The node capability set
	version  uint32 // The network protocol the node used
	services uint64 // The services the node supplied
	relay    bool   // The relay capability of the node (merge into capbility flag)
	height   uint64 // The node latest block height
	// TODO does this channel should be a buffer channel
	chF        chan func() error // Channel used to operate the node without lock
	link                         // The link status and infomation
	local      *node             // The pointer to local node
	nbrNodes                     // The neighbor node connect with currently node except itself
	eventQueue                   // The event queue to notice notice other modules
	TXNPool                      // Unconfirmed transaction pool
	idCache                      // The buffer to store the id of the items which already be processed
	ledger     *ledger.Ledger    // The Local ledger
}

func (node node) DumpInfo() {
	log.Trace("Node info:\n")
	fmt.Printf("\t state = %d\n", node.state)
	fmt.Printf("\t id = 0x%x\n", node.id)
	fmt.Printf("\t addr = %s\n", node.addr)
	fmt.Printf("\t conn = %v\n", node.conn)
	fmt.Printf("\t cap = %d\n", node.cap)
	fmt.Printf("\t version = %d\n", node.version)
	fmt.Printf("\t services = %d\n", node.services)
	fmt.Printf("\t port = %d\n", node.port)
	fmt.Printf("\t relay = %v\n", node.relay)
	fmt.Printf("\t height = %v\n", node.height)

	fmt.Printf("\t conn cnt = %v\n", node.link.connCnt)
}

func (node *node) UpdateInfo(t time.Time, version uint32, services uint64,
	port uint16, nonce uint64, relay uint8, height uint64) {
	// TODO need lock
	node.UpdateTime(t)
	node.id = nonce
	node.version = version
	node.services = services
	node.port = port
	if relay == 0 {
		node.relay = false
	} else {
		node.relay = true
	}
	node.height = uint64(height)
}

func NewNode() *node {
	n := node{
		state: protocol.INIT,
		chF:   make(chan func() error),
	}
	runtime.SetFinalizer(&n, rmNode)
	go n.backend()
	return &n
}

func InitNode() Noder {
	var err error
	n := NewNode()

	n.version = PROTOCOLVERSION
	n.services = NODESERVICES
	n.link.port = uint16(Parameters.NodePort)
	n.relay = true
	rand.Seed(time.Now().UTC().UnixNano())
	// Fixme replace with the real random number
	n.id = uint64(rand.Uint32())<<32 + uint64(rand.Uint32())
	fmt.Printf("Init node ID to 0x%0x \n", n.id)
	n.nbrNodes.init()
	n.local = n
	n.TXNPool.init()
	n.eventQueue.init()
	n.ledger, err = ledger.GetDefaultLedger()
	if err != nil {
		fmt.Printf("Get Default Ledger error\n")
		errors.New("Get Default Ledger error")
	}

	go n.initConnection()
	go n.updateNodeInfo()

	return n
}

func rmNode(node *node) {
	log.Debug(fmt.Sprintf("Remove unused/deuplicate node: 0x%0x", node.id))
}

func (node *node) backend() {
	for f := range node.chF {
		f()
	}
}

func (node node) GetID() uint64 {
	return node.id
}

func (node node) GetState() uint {
	return node.state
}

func (node node) getConn() net.Conn {
	return node.conn
}

func (node node) GetPort() uint16 {
	return node.port
}

func (node node) GetRelay() bool {
	return node.relay
}

func (node node) Version() uint32 {
	return node.version
}

func (node node) Services() uint64 {
	return node.services
}

func (node *node) SetState(state uint) {
	node.state = state
}

func (node *node) LocalNode() Noder {
	return node.local
}

func (node *node) GetHeight() uint64 {
	return node.height
}

func (node node) GetLedger() *ledger.Ledger {
	return node.ledger
}

func (node *node) UpdateTime(t time.Time) {
	node.time = t
}

func (node node) GetMemoryPool() map[common.Uint256]*transaction.Transaction {
	return node.GetTxnPool()
	// TODO refresh the pending transaction pool
}

func (node node) SynchronizeMemoryPool() {
	// Fixme need lock
	for _, n := range node.nbrNodes.List {
		if n.state == ESTABLISH {
			ReqMemoryPool(n)
		}
	}
}

func (node node) Xmit(inv common.Inventory) error {
	common.Trace()
	var buffer []byte
	var err error

	if inv.Type() == common.TRANSACTION {
		log.Info("****TX transaction message*****\n")
		transaction, ret := inv.(*transaction.Transaction)
		if ret {
			//transaction.Serialize(tmpBuffer)
			buffer, err = NewTxn(transaction)
			if err != nil {
				log.Warn("Error New Tx message ", err.Error())
				return err
			}
		}
		node.txnCnt++
	} else if inv.Type() == common.BLOCK {
		log.Info("****TX block message****\n")
		block, isBlock := inv.(*ledger.Block)
		// FiXME, should be moved to higher layer
		if isBlock == false {
			log.Warn("Wrcntm block be Xmit")
			return errors.New("Wrcntm block be Xmit")
		}

		err := ledger.DefaultLedger.Blockchain.AddBlock(block)
		if err != nil {
			log.Error("Add block error before Xmit")
			return errors.New("Add block error before Xmit")
		}
		buffer, err = NewBlock(block)
		if err != nil {
			log.Warn("Error New Block message ", err.Error())
			return err
		}
	} else if inv.Type() == common.CONSENSUS {
		log.Info("*****TX consensus message****\n")
		payload, ret := inv.(*ConsensusPayload)
		if ret {
			buffer, err = NewConsensus(payload)
			if err != nil {
				log.Warn("Error New consensus message ", err.Error())
				return err
			}
		}
	} else {
		log.Info("Unknown Xmit message type")
		return errors.New("Unknow Xmit message type\n")
	}

	node.nbrNodes.Broadcast(buffer)

	return nil
}

func (node *node) GetAddr() string {
	return node.addr
}

func (node *node) GetAddr16() ([16]byte, error) {
	var result [16]byte
	ip := net.ParseIP(node.addr).To16()
	if ip == nil {
		log.Error("Parse IP address error\n")
		return result, errors.New("Parse IP address error")
	}

	copy(result[:], ip[:16])
	return result, nil
}

func (node node) GetTime() int64 {
	t := time.Now()
	return t.UnixNano()
}

func (node node) GetNeighborAddrs() ([]NodeAddr, uint64) {
	var i uint64
	var addrs []NodeAddr
	// TODO read lock
	for _, n := range node.nbrNodes.List {
		if n.GetState() != ESTABLISH {
			ccntminue
		}
		var addr NodeAddr
		addr.IpAddr, _ = n.GetAddr16()
		addr.Time = n.GetTime()
		addr.Services = n.Services()
		addr.Port = n.GetPort()
		addr.ID = n.GetID()
		addrs = append(addrs, addr)

		i++
	}

	return addrs, i
}
