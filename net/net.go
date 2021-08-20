package net

import (
	. "github.com/Ontology/common"
	"github.com/Ontology/core/ledger"
	"github.com/Ontology/core/transaction"
	"github.com/Ontology/crypto"
	. "github.com/Ontology/errors"
	"github.com/Ontology/events"
	"github.com/Ontology/net/node"
	"github.com/Ontology/net/protocol"
)

type Neter interface {
	GetMemoryPool() map[common.Uint256]*transaction.Transaction
	SynchronizeMemoryPool()
	Xmit(common.Inventory) error // The transmit interface
	GetEvent(eventName string) *events.Event
}

func StartProtocol() (Neter, protocol.Noder) {
	seedNodes := config.Parameters.SeedList

	net := node.InitNode()
	for _, nodeAddr := range seedNodes {
		net.Connect(nodeAddr)
	}
	return net, net
}
