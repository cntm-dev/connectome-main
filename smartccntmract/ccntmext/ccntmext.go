package ccntmext

import (
	"github.com/Ontology/common"
	vmtypes "github.com/Ontology/vm/types"
	"github.com/Ontology/smartccntmract/event"
)

type CcntmextRef interface {
	PushCcntmext(ccntmext *Ccntmext)
	CurrentCcntmext() *Ccntmext
	CallingCcntmext() *Ccntmext
	EntryCcntmext() *Ccntmext
	PopCcntmext()
	CheckWitness(address common.Address) bool
	PushNotifications(notifications []*event.NotifyEventInfo)
	Execute() error
}


type Ccntmext struct {
	CcntmractAddress common.Address
	Code vmtypes.VmCode
}
