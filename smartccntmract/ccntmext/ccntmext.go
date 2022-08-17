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

package ccntmext

import (
	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/smartccntmract/event"
)

// CcntmextRef is a interface of smart ccntmext
// when need call a ccntmract, push current ccntmext to smart ccntmract ccntmexts
// when execute smart ccntmract finish, pop current ccntmext from smart ccntmract ccntmexts
// when need to check authorization, use CheckWitness
// when smart ccntmract execute trigger event, use PushNotifications push it to smart ccntmract notifications
// when need to invoke a smart ccntmract, use AppCall to invoke it
type CcntmextRef interface {
	PushCcntmext(ccntmext *Ccntmext)
	CurrentCcntmext() *Ccntmext
	CallingCcntmext() *Ccntmext
	EntryCcntmext() *Ccntmext
	PopCcntmext()
	CheckWitness(address common.Address) bool
	PushNotifications(notifications []*event.NotifyEventInfo)
	NewExecuteEngine(code []byte) (Engine, error)
	CheckUseGas(gas uint64) bool
	CheckExecStep() bool
}

type Engine interface {
	Invoke() (interface{}, error)
}

// Ccntmext describe smart ccntmract execute ccntmext struct
type Ccntmext struct {
	CcntmractAddress common.Address
	Code            []byte
}
