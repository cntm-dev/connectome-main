package net

import (
	"github.com/DNAProject/DNA/common"
	"github.com/DNAProject/DNA/config"
	"github.com/DNAProject/DNA/core/transaction"
	"github.com/DNAProject/DNA/crypto"
	"github.com/DNAProject/DNA/events"
	"github.com/DNAProject/DNA/net/node"
	"github.com/DNAProject/DNA/net/protocol"
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
