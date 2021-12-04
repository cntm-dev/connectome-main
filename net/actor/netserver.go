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

package actor

import (
	"github.com/Ontology/common/log"
	"github.com/Ontology/net/protocol"
	"github.com/cntmio/cntmology-eventbus/actor"
	"reflect"
)

var netServerPid *actor.PID

var node protocol.Noder

type NetServer struct{}

type GetNodeVersionReq struct {
}
type GetNodeVersionRsp struct {
	Version uint32
}

type GetConnectionCntReq struct {
}
type GetConnectionCntRsp struct {
	Cnt uint
}

type GetNodeIdReq struct {
}
type GetNodeIdRsp struct {
	Id uint64
}

type GetNodePortReq struct {
}
type GetNodePortRsp struct {
	Port uint16
}

type GetConsensusPortReq struct {
}
type GetConsensusPortRsp struct {
	Port uint16
}

type GetConnectionStateReq struct {
}
type GetConnectionStateRsp struct {
	State uint32
}

type GetNodeTimeReq struct {
}
type GetNodeTimeRsp struct {
	Time int64
}

type GetNodeTypeReq struct {
}
type GetNodeTypeRsp struct {
	NodeType uint64
}

type GetRelayStateReq struct {
}
type GetRelayStateRsp struct {
	Relay bool
}

type GetNeighborAddrsReq struct {
}
type GetNeighborAddrsRsp struct {
	Addrs []protocol.NodeAddr
	Count uint64
}

func (state *NetServer) Receive(ccntmext actor.Ccntmext) {
	switch ccntmext.Message().(type) {
	case *actor.Restarting:
		log.Warn("p2p actor restarting")
	case *actor.Stopping:
		log.Warn("p2p actor stopping")
	case *actor.Stopped:
		log.Warn("p2p actor stopped")
	case *actor.Started:
		log.Warn("p2p actor started")
	case *actor.Restart:
		log.Warn("p2p actor restart")
	case *GetNodeVersionReq:
		version := node.Version()
		ccntmext.Sender().Request(&GetNodeVersionRsp{Version: version}, ccntmext.Self())
	case *GetConnectionCntReq:
		connectionCnt := node.GetConnectionCnt()
		ccntmext.Sender().Request(&GetConnectionCntRsp{Cnt: connectionCnt}, ccntmext.Self())
	case *GetNodePortReq:
		nodePort := node.GetPort()
		ccntmext.Sender().Request(&GetNodePortRsp{Port: nodePort}, ccntmext.Self())
	case *GetConsensusPortReq:
		conPort := node.GetPort()
		ccntmext.Sender().Request(&GetConsensusPortRsp{Port: conPort}, ccntmext.Self())
	case *GetNodeIdReq:
		id := node.GetID()
		ccntmext.Sender().Request(&GetNodeIdRsp{Id: id}, ccntmext.Self())
	case *GetConnectionStateReq:
		state := node.GetState()
		ccntmext.Sender().Request(&GetConnectionStateRsp{State: state}, ccntmext.Self())
	case *GetNodeTimeReq:
		time := node.GetTime()
		ccntmext.Sender().Request(&GetNodeTimeRsp{Time: time}, ccntmext.Self())
	case *GetNodeTypeReq:
		nodeType := node.Services()
		ccntmext.Sender().Request(&GetNodeTypeRsp{NodeType: nodeType}, ccntmext.Self())
	case *GetRelayStateReq:
		relay := node.GetRelay()
		ccntmext.Sender().Request(&GetRelayStateRsp{Relay: relay}, ccntmext.Self())
	case *GetNeighborAddrsReq:
		addrs, count := node.GetNeighborAddrs()
		ccntmext.Sender().Request(&GetNeighborAddrsRsp{Addrs: addrs, Count: count}, ccntmext.Self())
	default:
		err := node.Xmit(ccntmext.Message())
		if nil != err {
			log.Error("Error Xmit message ", err.Error(), reflect.TypeOf(ccntmext.Message()))
		}
	}
}

func InitNetServer(netNode protocol.Noder) (*actor.PID, error) {
	props := actor.FromProducer(func() actor.Actor { return &NetServer{} })
	netServerPid, err := actor.SpawnNamed(props, "net_server")
	if err != nil {
		return nil, err
	}
	node = netNode
	return netServerPid, err
}
