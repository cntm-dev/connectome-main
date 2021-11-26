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

package message

import (
	"github.com/Ontology/core/types"
	"github.com/Ontology/common"
	"github.com/Ontology/net/protocol"
)

const (
	TopicSaveBlockComplete       = "svblkcmp"
	TopicNewInventory            = "newinv"
	TopicNodeDisconnect          = "noddis"
	TopicNodeConsensusDisconnect = "nodcnsdis"
	TopicSmartCodeEvent          = "scevt"
)

type SaveBlockCompleteMsg struct {
	Block *types.Block
}

type NewInventoryMsg struct {
	Inventory *common.Inventory
}

type NodeDisconnectMsg struct {
	Node protocol.Noder
}

type NodeConsensusDisconnectMsg struct {
	Node protocol.Noder
}

type SmartCodeEventMsg struct {
	Event *types.SmartCodeEvent
}