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

package event

import (
	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/events"
	"github.com/cntmio/cntmology/events/message"
)

const (
	EVENT_LOG    = "Log"
	EVENT_NOTIFY = "Notify"
)

// PushSmartCodeEvent push event ccntment to socket.io
func PushSmartCodeEvent(txHash common.Uint256, errcode int64, action string, result interface{}) {
	if events.DefActorPublisher == nil {
		return
	}
	smartCodeEvt := &types.SmartCodeEvent{
		TxHash: txHash,
		Action: action,
		Result: result,
		Error:  errcode,
	}
	events.DefActorPublisher.Publish(message.TOPIC_SMART_CODE_EVENT, &message.SmartCodeEventMsg{Event: smartCodeEvt})
}
