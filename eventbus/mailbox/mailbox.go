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

package mailbox

import (
	"runtime"
	"sync/atomic"

	"github.com/Ontology/eventbus/internal/queue/mpsc"
	"github.com/Ontology/common/log"
)

type Statistics interface {
	MailboxStarted()
	MessagePosted(message interface{})
	MessageReceived(message interface{})
	MailboxEmpty()
}

// MessageInvoker is the interface used by a mailbox to forward messages for processing
type MessageInvoker interface {
	InvokeSystemMessage(interface{})
	InvokeUserMessage(interface{})
	EscalateFailure(reason interface{}, message interface{})
}

// The Inbound interface is used to enqueue messages to the mailbox
type Inbound interface {
	PostUserMessage(message interface{})
	PostSystemMessage(message interface{})
	Start()
}

// Producer is a function which creates a new mailbox
type Producer func(invoker MessageInvoker, dispatcher Dispatcher) Inbound

const (
	idle int32 = iota
	running
)

type defaultMailbox struct {
	userMailbox     queue
	systemMailbox   *mpsc.Queue
	schedulerStatus int32
	userMessages    int32
	sysMessages     int32
	invoker         MessageInvoker
	dispatcher      Dispatcher
	suspended       bool
	mailboxStats    []Statistics
}

func (m *defaultMailbox) PostUserMessage(message interface{}) {
	for _, ms := range m.mailboxStats {
		ms.MessagePosted(message)
	}
	m.userMailbox.Push(message)
	atomic.AddInt32(&m.userMessages, 1)
	m.schedule()
}

func (m *defaultMailbox) PostSystemMessage(message interface{}) {
	for _, ms := range m.mailboxStats {
		ms.MessagePosted(message)
	}
	m.systemMailbox.Push(message)
	atomic.AddInt32(&m.sysMessages, 1)
	m.schedule()
}

func (m *defaultMailbox) schedule() {
	if atomic.CompareAndSwapInt32(&m.schedulerStatus, idle, running) {
		m.dispatcher.Schedule(m.processMessages)
	}
}

func (m *defaultMailbox) processMessages() {

process:
	m.run()

	// set mailbox to idle
	atomic.StoreInt32(&m.schedulerStatus, idle)
	sys := atomic.LoadInt32(&m.sysMessages)
	user := atomic.LoadInt32(&m.userMessages)
	// check if there are still messages to process (sent after the message loop ended)
	if sys > 0 || (!m.suspended && user > 0) {
		// try setting the mailbox back to running
		if atomic.CompareAndSwapInt32(&m.schedulerStatus, idle, running) {
			//	fmt.Printf("looping %v %v %v\n", sys, user, m.suspended)
			goto process
		}
	}

	for _, ms := range m.mailboxStats {
		ms.MailboxEmpty()
	}
}

func (m *defaultMailbox) run() {
	var msg interface{}

	defer func() {
		if r := recover(); r != nil {
			log.Debug("[ACTOR] Recovering")
			m.invoker.EscalateFailure(r, msg)
		}
	}()

	i, t := 0, m.dispatcher.Throughput()
	for {
		if i > t {
			i = 0
			runtime.Gosched()
		}

		i++

		// keep processing system messages until queue is empty
		if msg = m.systemMailbox.Pop(); msg != nil {
			atomic.AddInt32(&m.sysMessages, -1)
			switch msg.(type) {
			case *SuspendMailbox:
				m.suspended = true
			case *ResumeMailbox:
				m.suspended = false
			default:
				m.invoker.InvokeSystemMessage(msg)
			}
			for _, ms := range m.mailboxStats {
				ms.MessageReceived(msg)
			}
			ccntminue
		}

		// didn't process a system message, so break until we are resumed
		if m.suspended {
			return
		}

		if msg = m.userMailbox.Pop(); msg != nil {
			atomic.AddInt32(&m.userMessages, -1)
			m.invoker.InvokeUserMessage(msg)
			for _, ms := range m.mailboxStats {
				ms.MessageReceived(msg)
			}
		} else {
			return
		}
	}

}

func (m *defaultMailbox) Start() {
	for _, ms := range m.mailboxStats {
		ms.MailboxStarted()
	}
}
