package event

import (
	"github.com/Ontology/common"
	"github.com/Ontology/vm/neovm/types"
)

type NotifyEventArgs struct {
	Ccntmainer common.Uint256
	CodeHash common.Uint160
	States types.StackItemInterface
}

type NotifyEventInfo struct {
	Ccntmainer common.Uint256
	CodeHash common.Uint160
	States interface{}
}

