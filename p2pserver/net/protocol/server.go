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

// Package p2p provides an network interface
package p2p

import (
	"github.com/cntmio/cntmology/p2pserver/common"
	"github.com/cntmio/cntmology/p2pserver/message/types"
	"github.com/cntmio/cntmology/p2pserver/peer"
)

//P2P represent the net interface of p2p package
type P2P interface {
	Start()
	Halt()
	Connect(addr string, isConsensus bool) error
	GetID() uint64
	GetVersion() uint32
	GetSyncPort() uint16
	GetConsPort() uint16
	GetHttpInfoPort() uint16
	GetRelay() bool
	GetHeight() uint64
	GetTime() int64
	GetServices() uint64
	GetNeighbors() []*peer.Peer
	GetNeighborAddrs() []common.PeerAddr
	GetConnectionCnt() uint32
	GetNp() *peer.NbrPeers
	GetPeer(uint64) *peer.Peer
	SetHeight(uint64)
	IsPeerEstablished(p *peer.Peer) bool
	Send(p *peer.Peer, msg types.Message, isConsensus bool) error
	GetMsgChan(isConsensus bool) chan *types.MsgPayload
	GetPeerFromAddr(addr string) *peer.Peer
	AddOutConnectingList(addr string) (added bool)
	GetOutConnRecordLen() int
	RemoveFromConnectingList(addr string)
	RemoveFromOutConnRecord(addr string)
	RemoveFromInConnRecord(addr string)
	AddPeerSyncAddress(addr string, p *peer.Peer)
	AddPeerConsAddress(addr string, p *peer.Peer)
	GetOutConnectingListLen() (count uint)
	RemovePeerSyncAddress(addr string)
	RemovePeerConsAddress(addr string)
	AddNbrNode(*peer.Peer)
	DelNbrNode(id uint64) (*peer.Peer, bool)
	NodeEstablished(uint64) bool
	Xmit(msg types.Message, isCons bool)
	SetOwnAddress(addr string)
	IsAddrFromConnecting(addr string) bool
}
