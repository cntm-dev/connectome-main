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
	. "GoOnchain/net/message"
	. "GoOnchain/net/protocol"
	"time"
)

func keepAlive(from *Noder, dst *Noder) {
	// Need move to node function or keep here?
}

func (node node) GetBlkHdrs() {
	for _, n := range node.local.List {
		h1 := n.GetHeight()
		h2:= node.local.GetLedger().GetLocalBlockChainHeight()
		if (node.GetState() == ESTABLISH) && (h1 > uint64(h2)) {
			buf, _ := NewMsg("getheaders", node.local)
			go node.Tx(buf)
		}
	}
}

func (node *node) SyncBlk() {
	headerHeight := ledger.DefaultLedger.Store.GetHeaderHeight()
	currentBlkHeight := ledger.DefaultLedger.Blockchain.BlockHeight
	if currentBlkHeight >= headerHeight {
		return
	}
	var dValue int32
	var reqCnt uint32
	var i uint32
	noders := node.local.GetNeighborNoder()

	for _, n := range noders {
		if uint32(n.GetHeight()) <= currentBlkHeight {
			ccntminue
		}
		n.RemoveFlightHeightLessThan(currentBlkHeight)
		count := protocol.MAX_REQ_BLK_ONCE - uint32(n.GetFlightHeightCnt())
		dValue = int32(headerHeight - currentBlkHeight - reqCnt)
		flights := n.GetFlightHeights()
		if count == 0 {
			for _, f := range flights {
				hash := ledger.DefaultLedger.Store.GetHeaderHashByHeight(f)
				if ledger.DefaultLedger.Store.BlockInCache(hash) == false {
					ReqBlkData(n, hash)
				}
			}

		}
		for i = 1; i <= count && dValue >= 0; i++ {
			hash := ledger.DefaultLedger.Store.GetHeaderHashByHeight(currentBlkHeight + reqCnt)

			if ledger.DefaultLedger.Store.BlockInCache(hash) == false {
				ReqBlkData(n, hash)
				n.StoreFlightHeight(currentBlkHeight + reqCnt)
			}
			reqCnt++
			dValue--
		}
	}
}

func (node *node) SendPingToNbr() {
	noders := node.local.GetNeighborNoder()
	for _, n := range noders {
		if n.GetState() == protocol.ESTABLISH {
			buf, err := message.NewPingMsg()
			if err != nil {
				log.Error("failed build a new ping message")
			} else {
				go n.Tx(buf)
			}
		}
	}
}

func (node *node) HeartBeatMonitor() {
	noders := node.local.GetNeighborNoder()
	var periodUpdateTime uint
	if config.Parameters.GenBlockTime > config.MIN_GEN_BLOCK_TIME {
		periodUpdateTime = config.Parameters.GenBlockTime / protocol.UPDATE_RATE_PER_BLOCK
	} else {
		periodUpdateTime = config.DEFAULT_GEN_BLOCK_TIME / protocol.UPDATE_RATE_PER_BLOCK
	}
	for _, n := range noders {
		if n.GetState() == protocol.ESTABLISH {
			t := n.GetLastRXTime()
			if t.Before(time.Now().Add(-1 * time.Second * time.Duration(periodUpdateTime) * protocol.KEEPALIVE_TIMEOUT)) {
				log.Warn("keepalive timeout!!!")
				n.SetState(protocol.INACTIVITY)
				n.CloseConn()
			}
		}
	}
}

func (node *node) ReqNeighborList() {
	buf, _ := message.NewMsg("getaddr", node.local)
	go node.Tx(buf)
}

func (node *node) ConnectSeeds() {
	if node.IsUptoMinNodeCount() {
		return
	}
	seedNodes := config.Parameters.SeedList
	for _, nodeAddr := range seedNodes {
		found := false
		var n protocol.Noder
		var ip net.IP
		node.nbrNodes.Lock()
		for _, tn := range node.nbrNodes.List {
			addr := getNodeAddr(tn)
			ip = addr.IpAddr[:]
			addrString := ip.To16().String() + ":" + strconv.Itoa(int(addr.Port))
			if nodeAddr == addrString {
				n = tn
				found = true
				break
			}
		}
		node.nbrNodes.Unlock()
		if found {
			if n.GetState() == protocol.ESTABLISH {
				n.ReqNeighborList()
			}
		} else {
			go node.Connect(nodeAddr)
		}
	}
}

func getNodeAddr(n *node) protocol.NodeAddr {
	var addr protocol.NodeAddr
	addr.IpAddr, _ = n.GetAddr16()
	addr.Time = n.GetTime()
	addr.Services = n.Services()
	addr.Port = n.GetPort()
	addr.ID = n.GetID()
	return addr
}

func (node *node) reconnect() {
	node.RetryConnAddrs.Lock()
	lst := make(map[string]int)
	addrs := make([]string, 0, len(node.RetryAddrs))
	for addr, v := range node.RetryAddrs {
		v += 1
		addrs = append(addrs, addr)
		if v < protocol.MAX_RETRY_COUNT {
			lst[addr] = v
		}
	}
	node.RetryAddrs = lst
	node.RetryConnAddrs.Unlock()

	for _, addr := range addrs {
		rand.Seed(time.Now().UnixNano())
		log.Info("Try to reconnect peer, peer addr is ", addr)
		<-time.After(time.Duration(rand.Intn(protocol.CONN_MAX_BACK)) * time.Millisecond)
		log.Info("Back off time`s up, start connect node")
		node.Connect(addr)
	}
}

func (n *node) TryConnect() {
	if n.fetchRetryNodeFromNeighborList() > 0 {
		n.reconnect()
	}
}

func (n *node) fetchRetryNodeFromNeighborList() int {
	n.nbrNodes.Lock()
	defer n.nbrNodes.Unlock()
	var ip net.IP
	neighborNodes := make(map[uint64]*node)
	for _, tn := range n.nbrNodes.List {
		addr := getNodeAddr(tn)
		ip = addr.IpAddr[:]
		nodeAddr := ip.To16().String() + ":" + strconv.Itoa(int(addr.Port))
		if tn.GetState() == protocol.INACTIVITY {
			//add addr to retry list
			n.AddInRetryList(nodeAddr)
			//close legacy node
			if tn.conn != nil {
				tn.CloseConn()
			}
		} else {
			//add others to tmp node map
			n.RemoveFromRetryList(nodeAddr)
			neighborNodes[tn.GetID()] = tn
		}
	}
	n.nbrNodes.List = neighborNodes
	return len(n.RetryAddrs)
}

func (node *node) updateNodeInfo() {
	var periodUpdateTime uint
	if config.Parameters.GenBlockTime > config.MIN_GEN_BLOCK_TIME {
		periodUpdateTime = config.Parameters.GenBlockTime / protocol.UPDATE_RATE_PER_BLOCK
	} else {
		periodUpdateTime = config.DEFAULT_GEN_BLOCK_TIME / protocol.UPDATE_RATE_PER_BLOCK
	}
	ticker := time.NewTicker(time.Second * (time.Duration(periodUpdateTime)))
	quit := make(chan struct{})
	for {
		select {
		case <-ticker.C:
			node.SendPingToNbr()
			node.HeartBeatMonitor()
		case <-quit:
			ticker.Stop()
			return
		}
	}
}

func (node *node) updateConnection() {
	t := time.NewTimer(time.Second * protocol.CONN_MONITOR)
	for {
		select {
		case <-t.C:
			node.ConnectSeeds()
			node.TryConnect()
			t.Stop()
			t.Reset(time.Second * protocol.CONN_MONITOR)
		}
	}
}
