/*
 * Copyright (C) 2018 The cntm Authors
 * This file is part of The cntm library.
 *
 * The cntm is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The cntm is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The cntm.  If not, see <http://www.gnu.org/licenses/>.
 */

package context

import (
	"github.com/conntectome/cntm/common"
	"github.com/conntectome/cntm/core/types"
	"github.com/conntectome/cntm/smartcontract/event"
)

// ContextRef is a interface of smart context
// when need call a contract, push current context to smart contract contexts
// when execute smart contract finish, pop current context from smart contract contexts
// when need to check authorization, use CheckWitness
// when smart contract execute trigger event, use PushNotifications push it to smart contract notifications
// when need to invoke a smart contract, use AppCall to invoke it
type ContextRef interface {
	PushContext(context *Context)
	CurrentContext() *Context
	CallingContext() *Context
	EntryContext() *Context
	PopContext()
	CheckWitness(address common.Address) bool
	PushNotifications(notifications []*event.NotifyEventInfo)
	NewExecuteEngine(code []byte, txtype types.TransactionType) (Engine, error)
	CheckUseGas(gas uint64) bool
	CheckExecStep() bool
	GetCallerAddress() []common.Address
	SetInternalErr()
	IsInternalErr() bool
	PutCrossStateHashes(hashes []common.Uint256)
}

type Engine interface {
	Invoke() (interface{}, error)
}

// Context describe smart contract execute context struct
type Context struct {
	ContractAddress common.Address
	Code            []byte
}
