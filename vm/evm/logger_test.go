// Copyright (C) 2021 The Ontology Authors
// Copyright 2016 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// alcntm with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package evm

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
	"github.com/cntmio/cntmology/smartccntmract/storage"
	"github.com/cntmio/cntmology/vm/evm/params"
)

type dummyCcntmractRef struct {
	calledForEach bool
}

func (dummyCcntmractRef) ReturnGas(*big.Int)          {}
func (dummyCcntmractRef) Address() common.Address     { return common.Address{} }
func (dummyCcntmractRef) Value() *big.Int             { return new(big.Int) }
func (dummyCcntmractRef) SetCode(common.Hash, []byte) {}
func (d *dummyCcntmractRef) ForEachStorage(callback func(key, value common.Hash) bool) {
	d.calledForEach = true
}
func (d *dummyCcntmractRef) SubBalance(amount *big.Int) {}
func (d *dummyCcntmractRef) AddBalance(amount *big.Int) {}
func (d *dummyCcntmractRef) SetBalance(*big.Int)        {}
func (d *dummyCcntmractRef) SetNonce(uint64)            {}
func (d *dummyCcntmractRef) Balance() *big.Int          { return new(big.Int) }

type dummyStatedb struct {
	storage.StateDB
}

func (*dummyStatedb) GetRefund() uint64 { return 1337 }

func TestStoreCapture(t *testing.T) {
	var (
		env      = NewEVM(BlockCcntmext{}, TxCcntmext{}, &dummyStatedb{}, params.TestChainConfig, Config{})
		logger   = NewStructLogger(nil)
		mem      = NewMemory()
		stack    = newstack()
		rstack   = newReturnStack()
		ccntmract = NewCcntmract(&dummyCcntmractRef{}, &dummyCcntmractRef{}, new(big.Int), 0)
	)
	stack.push(uint256.NewInt().SetUint64(1))
	stack.push(uint256.NewInt())
	var index common.Hash
	logger.CaptureState(env, 0, SSTORE, 0, 0, mem, stack, rstack, nil, ccntmract, 0, nil)
	if len(logger.storage[ccntmract.Address()]) == 0 {
		t.Fatalf("expected exactly 1 changed value on address %x, got %d", ccntmract.Address(), len(logger.storage[ccntmract.Address()]))
	}
	exp := common.BigToHash(big.NewInt(1))
	if logger.storage[ccntmract.Address()][index] != exp {
		t.Errorf("expected %x, got %x", exp, logger.storage[ccntmract.Address()][index])
	}
}
