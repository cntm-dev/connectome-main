package ledgerstore

import (
	"fmt"
	."github.com/Ontology/core/store/common"
	"github.com/Ontology/core/payload"
)

type CacheCodeTable struct {
	store IStateStore
}

func (table *CacheCodeTable) GetCode(codeHash []byte) ([]byte, error) {
	value, _ := table.store.TryGet(ST_Ccntmract, codeHash)
	if value == nil {
		return nil, fmt.Errorf("[GetCode] TryGet ccntmract error! codeHash:%x", codeHash)
	}

	return value.Value.(*payload.DeployCode).Code, nil
}
