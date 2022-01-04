package ccntmext

import (
	"github.com/Ontology/common"
	"github.com/Ontology/smartccntmract/event"
	vmtypes "github.com/Ontology/vm/types"
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
	Code            vmtypes.VmCode
}
