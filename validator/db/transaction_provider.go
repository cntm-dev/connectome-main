/*
 * Copyright (C) 2018 The cntm Authors
 * This file is part of The cntm library.
 *
 * The cntm is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The cntm is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The cntm.  If not, see <http://www.gnu.org/licenses/>.
 */

package db

import (
	"github.com/conntectome/cntm/common"
	"github.com/conntectome/cntm/core/types"
)

type TransactionProvider interface {
	BestStateProvider
	ContainTransaction(hash common.Uint256) bool
	GetTransactionBytes(hash common.Uint256) ([]byte, error)
	GetTransaction(hash common.Uint256) (*types.Transaction, error)
	PersistBlock(block *types.Block) error
}
