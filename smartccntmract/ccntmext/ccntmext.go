package ccntmext

import (
	"github.com/Ontology/common"
	vmtypes "github.com/Ontology/vm/types"
)

type CcntmextRef interface {
	PushCcntmext(ccntmext *Ccntmext)
	CurrentCcntmext() *Ccntmext
	CallingCcntmext() *Ccntmext
	EntryCcntmext() *Ccntmext
	PopCcntmext()
	Execute() error
}


type Ccntmext struct {
	CcntmractAddress common.Address
	Code vmtypes.VmCode
}
