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

import "time"

type MessageProducer interface {
	// Tell sends a messages asynchronously to the PID
	Tell(pid *PID, message interface{})

	// Request sends a messages asynchronously to the PID. The actor may send a response back via respondTo, which is
	// available to the receiving actor via Ccntmext.Sender
	Request(pid *PID, message interface{}, respondTo *PID)

	// RequestFuture sends a message to a given PID and returns a Future
	RequestFuture(pid *PID, message interface{}, timeout time.Duration) *Future
}

type rootMessageProducer struct {
}

var (
	EmptyCcntmext MessageProducer = &rootMessageProducer{}
)

// Tell sends a messages asynchronously to the PID
func (*rootMessageProducer) Tell(pid *PID, message interface{}) {
	pid.Tell(message)
}

// Request sends a messages asynchronously to the PID. The actor may send a response back via respondTo, which is
// available to the receiving actor via Ccntmext.Sender
func (*rootMessageProducer) Request(pid *PID, message interface{}, respondTo *PID) {
	pid.Request(message, respondTo)
}

// RequestFuture sends a message to a given PID and returns a Future
func (*rootMessageProducer) RequestFuture(pid *PID, message interface{}, timeout time.Duration) *Future {
	return pid.RequestFuture(message, timeout)
}
