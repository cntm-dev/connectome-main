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
package testsuite

import (
	"github.com/cntmio/cntmology/core/store/leveldbstore"
	"github.com/cntmio/cntmology/core/store/overlaydb"
)

func NewOverlayDB() *overlaydb.OverlayDB {
	store := leveldbstore.NewMemLevelDBStore()
	return overlaydb.NewOverlayDB(store)
}
