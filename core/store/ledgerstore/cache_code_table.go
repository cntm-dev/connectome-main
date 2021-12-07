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

package ledgerstore

import (
	"fmt"
	scom "github.com/Ontology/core/store/common"
	"github.com/Ontology/core/payload"
)

type CacheCodeTable struct {
	store scom.IStateStore
}

func (table *CacheCodeTable) GetCode(codeHash []byte) ([]byte, error) {
	value, _ := table.store.TryGet(scom.ST_Ccntmract, codeHash)
	if value == nil {
		return nil, fmt.Errorf("[GetCode] TryGet ccntmract error! codeHash:%x", codeHash)
	}

	return value.Value.(*payload.DeployCode).Code.Code, nil
}
