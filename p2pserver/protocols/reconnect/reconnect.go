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

package reconnect

import (
	"math/rand"
	"sync"
	"time"

	"github.com/cntmio/cntmology/common/config"
	"github.com/cntmio/cntmology/common/log"
	"github.com/cntmio/cntmology/p2pserver/common"
	p2p "github.com/cntmio/cntmology/p2pserver/net/protocol"
	"github.com/cntmio/cntmology/p2pserver/peer"
)

type ReconnectPeerInfo struct {
	count int // current retry count
	id    common.PeerId
}

//ReconnectService ccntmain addr need to reconnect
type ReconnectService struct {
	sync.RWMutex
	MaxRetryCount uint
	RetryAddrs    map[string]*ReconnectPeerInfo
	net           p2p.P2P
	quit          chan bool
}

func NewReconectService(net p2p.P2P) *ReconnectService {
	return &ReconnectService{
		net:           net,
		MaxRetryCount: common.MAX_RETRY_COUNT,
		quit:          make(chan bool),
		RetryAddrs:    make(map[string]*ReconnectPeerInfo),
	}
}

func (self *ReconnectService) Start() {
	go self.keepOnlineService()
}

func (self *ReconnectService) Stop() {
	close(self.quit)
}

func (this *ReconnectService) keepOnlineService() {
	tick := time.NewTicker(time.Second * common.CONN_MONITOR)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			this.retryInactivePeer()
		case <-this.quit:
			return
		}
	}
}

func (self *ReconnectService) OnAddPeer(p *peer.PeerInfo) {
	listenAddr := p.RemoteListenAddress()
	self.Lock()
	delete(self.RetryAddrs, listenAddr)
	self.Unlock()
}

func (self *ReconnectService) OnDelPeer(p *peer.PeerInfo) {
	nodeAddr := p.RemoteListenAddress()
	self.Lock()
	self.RetryAddrs[nodeAddr] = &ReconnectPeerInfo{count: 0, id: p.Id}
	self.Unlock()
}

func (this *ReconnectService) retryInactivePeer() {
	net := this.net
	connCount := net.GetOutConnRecordLen()
	if connCount >= config.DefConfig.P2PNode.MaxConnOutBound {
		log.Warnf("[p2p]Connect: out connections(%d) reach max limit(%d)", connCount,
			config.DefConfig.P2PNode.MaxConnOutBound)
		return
	}

	//try connect
	if len(this.RetryAddrs) > 0 {
		this.Lock()

		list := make(map[string]*ReconnectPeerInfo)
		addrs := make([]string, 0, len(this.RetryAddrs))
		for addr, v := range this.RetryAddrs {
			v.count += 1
			if v.count <= common.MAX_RETRY_COUNT && net.GetPeer(v.id) == nil {
				addrs = append(addrs, addr)
				list[addr] = v
			}
		}

		this.RetryAddrs = list
		this.Unlock()
		for _, addr := range addrs {
			rand.Seed(time.Now().UnixNano())
			log.Debug("[p2p]Try to reconnect peer, peer addr is ", addr)
			<-time.After(time.Duration(rand.Intn(common.CONN_MAX_BACK)) * time.Millisecond)
			log.Debug("[p2p]Back off time`s up, start connect node")
			net.Connect(addr)
		}
	}
}

func (self *ReconnectService) ReconnectCount() int {
	self.RLock()
	defer self.RUnlock()
	return len(self.RetryAddrs)
}
