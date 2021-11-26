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

package remote

import (
	"time"

	"github.com/Ontology/eventbus/actor"
	"github.com/Ontology/eventbus/eventstream"
	"github.com/Ontology/common/log"
	"golang.org/x/net/ccntmext"
	"google.golang.org/grpc"
)

func newEndpointWriter(address string, config *remoteConfig) actor.Producer {
	return func() actor.Actor {
		return &endpointWriter{
			address: address,
			config:  config,
		}
	}
}

type endpointWriter struct {
	config              *remoteConfig
	address             string
	conn                *grpc.ClientConn
	stream              Remoting_ReceiveClient
	defaultSerializerId int32
}

func (state *endpointWriter) initialize() {
	err := state.initializeInternal()
	if err != nil {
		log.Error("EndpointWriter failed to connect"+err.Error())
		//Wait 2 seconds to restart and retry
		//Replace with Exponential Backoff
		time.Sleep(2 * time.Second)
		panic(err)
	}
}

func (state *endpointWriter) initializeInternal() error {
	log.Info("Started EndpointWriter", string(state.address))
	log.Info("EndpointWriter connecting", string(state.address))
	conn, err := grpc.Dial(state.address, state.config.dialOptions...)
	if err != nil {
		return err
	}
	state.conn = conn
	c := NewRemotingClient(conn)
	resp, err := c.Connect(ccntmext.Background(), &ConnectRequest{})
	if err != nil {
		return err
	}
	state.defaultSerializerId = resp.DefaultSerializerId

	//	log.Printf("Getting stream from address %v", state.address)
	stream, err := c.Receive(ccntmext.Background(), state.config.callOptions...)
	if err != nil {
		return err
	}
	go func() {
		_, err := stream.Recv()
		if err != nil {
			log.Info("EndpointWriter lost connection to address", string(state.address))

			//notify that the endpoint terminated
			terminated := &EndpointTerminatedEvent{
				Address: state.address,
			}
			eventstream.Publish(terminated)
		}
	}()

	log.Info("EndpointWriter connected", string(state.address))
	connected := &EndpointConnectedEvent{Address: state.address}
	eventstream.Publish(connected)
	state.stream = stream
	return nil
}

func (state *endpointWriter) sendEnvelopes(msg []interface{}, ctx actor.Ccntmext) {
	envelopes := make([]*MessageEnvelope, len(msg))

	//type name uniqueness map name string to type index
	typeNames := make(map[string]int32)
	typeNamesArr := make([]string, 0)
	targetNames := make(map[string]int32)
	targetNamesArr := make([]string, 0)
	var header *MessageHeader
	var typeID int32
	var targetID int32
	var serializerID int32
	for i, tmp := range msg {
		rd := tmp.(*remoteDeliver)

		if rd.serializerID == -1 {
			serializerID = state.defaultSerializerId
		} else {
			serializerID = rd.serializerID
		}

		if rd.header == nil || rd.header.Length() == 0 {
			header = nil
		} else {
			header = &MessageHeader{rd.header.ToMap()}
		}

		bytes, typeName, err := Serialize(rd.message, serializerID)
		if err != nil {
			panic(err)
		}
		typeID, typeNamesArr = addToLookup(typeNames, typeName, typeNamesArr)
		targetID, targetNamesArr = addToLookup(targetNames, rd.target.Id, targetNamesArr)

		envelopes[i] = &MessageEnvelope{
			MessageHeader: header,
			MessageData:   bytes,
			Sender:        rd.sender,
			Target:        targetID,
			TypeId:        typeID,
			SerializerId:  serializerID,
		}
	}

	batch := &MessageBatch{
		TypeNames:   typeNamesArr,
		TargetNames: targetNamesArr,
		Envelopes:   envelopes,
	}
	err := state.stream.Send(batch)

	if err != nil {
		ctx.Stash()
		log.Debug("gRPC Failed to send", string(state.address))
		panic("restart it")
	}
}

func addToLookup(m map[string]int32, name string, a []string) (int32, []string) {
	max := int32(len(m))
	id, ok := m[name]
	if !ok {
		m[name] = max
		id = max
		a = append(a, name)
	}
	return id, a
}

func (state *endpointWriter) Receive(ctx actor.Ccntmext) {
	switch msg := ctx.Message().(type) {
	case *actor.Started:
		state.initialize()
	case *actor.Stopped:
		state.conn.Close()
	case *actor.Restarting:
		state.conn.Close()
	case []interface{}:
		state.sendEnvelopes(msg, ctx)
	case actor.SystemMessage, actor.AutoReceiveMessage:
		//ignore
	default:
		log.Error("EndpointWriter received unknown message")
	}
}
