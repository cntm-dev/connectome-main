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
	"github.com/cntmio/cntmology-eventbus/actor"
	"github.com/cntmio/cntmology/events"
	"github.com/cntmio/cntmology/events/message"
)

type EventActor struct {
	blockPersistCompleted func(v interface{})
	smartCodeEvt          func(v interface{})
}

//receive from subscribed actor
func (t *EventActor) Receive(c actor.Ccntmext) {
	switch msg := c.Message().(type) {
	case *message.SaveBlockCompleteMsg:
		t.blockPersistCompleted(*msg.Block)
	case *message.SmartCodeEventMsg:
		t.smartCodeEvt(*msg.Event)
	default:
	}
}

//Subscribe save block complete and smartccntmract Event
func SubscribeEvent(topic string, handler func(v interface{})) {
	var props = actor.FromProducer(func() actor.Actor {
		if topic == message.TOPIC_SAVE_BLOCK_COMPLETE {
			return &EventActor{blockPersistCompleted: handler}
		} else if topic == message.TOPIC_SMART_CODE_EVENT {
			return &EventActor{smartCodeEvt: handler}
		} else {
			return &EventActor{}
		}
	})
	var pid = actor.Spawn(props)
	var sub = events.NewActorSubscriber(pid)
	sub.Subscribe(topic)
}
