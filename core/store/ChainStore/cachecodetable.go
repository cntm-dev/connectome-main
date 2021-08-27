package ChainStore

import (
	"github.com/Ontology/core/states"
	"github.com/Ontology/core/store"
	"github.com/Ontology/errors"

	"fmt"
)

type CacheCodeTable struct {
	store store.IStateStore
}

func (table *CacheCodeTable) GetCode(codeHash []byte) ([]byte, error) {
	value, _ := table.store.TryGet(store.ST_Ccntmract, codeHash)
	if value == nil {
		return nil, errors.NewErr(fmt.Sprintf("[GetCode] TryGet ccntmract error! codeHash:%x", codeHash))
	}

	return value.Value.(*states.CcntmractState).Code.Code, nil
}
