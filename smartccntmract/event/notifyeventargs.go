package event

import (
	"github.com/Ontology/common"
	"github.com/Ontology/vm/neovm/types"
)

type NotifyEventArgs struct {
	Ccntmainer common.Uint256
	CodeHash common.Address
	States types.StackItemInterface
}

type NotifyEventInfo struct {
	Ccntmainer common.Uint256
	CodeHash common.Address
	States interface{}
}

