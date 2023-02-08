// Copyright (C) 2021 The Ontology Authors
// Copyright 2015 The go-ethereum Authors
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

	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
)

// CcntmractRef is a reference to the ccntmract's backing object
type CcntmractRef interface {
	Address() common.Address
}

// AccountRef implements CcntmractRef.
//
// Account references are used during EVM initialisation and
// it's primary use is to fetch addresses. Removing this object
// proves difficult because of the cached jump destinations which
// are fetched from the parent ccntmract (i.e. the caller), which
// is a CcntmractRef.
type AccountRef common.Address

// Address casts AccountRef to a Address
func (ar AccountRef) Address() common.Address { return (common.Address)(ar) }

// Ccntmract represents an ethereum ccntmract in the state database. It ccntmains
// the ccntmract code, calling arguments. Ccntmract implements CcntmractRef
type Ccntmract struct {
	// CallerAddress is the result of the caller which initialised this
	// ccntmract. However when the "call method" is delegated this value
	// needs to be initialised to that of the caller's caller.
	CallerAddress common.Address
	caller        CcntmractRef
	self          CcntmractRef

	jumpdests map[common.Hash]bitvec // Aggregated result of JUMPDEST analysis.
	analysis  bitvec                 // Locally cached result of JUMPDEST analysis

	Code     []byte
	CodeHash common.Hash
	CodeAddr *common.Address
	Input    []byte

	Gas   uint64
	value *big.Int
}

// NewCcntmract returns a new ccntmract environment for the execution of EVM.
func NewCcntmract(caller CcntmractRef, object CcntmractRef, value *big.Int, gas uint64) *Ccntmract {
	c := &Ccntmract{CallerAddress: caller.Address(), caller: caller, self: object}

	if parent, ok := caller.(*Ccntmract); ok {
		// Reuse JUMPDEST analysis from parent ccntmext if available.
		c.jumpdests = parent.jumpdests
	} else {
		c.jumpdests = make(map[common.Hash]bitvec)
	}

	// Gas should be a pointer so it can safely be reduced through the run
	// This pointer will be off the state transition
	c.Gas = gas
	// ensures a value is set
	c.value = value

	return c
}

func (c *Ccntmract) validJumpdest(dest *uint256.Int) bool {
	udest, overflow := dest.Uint64WithOverflow()
	// PC cannot go beyond len(code) and certainly can't be bigger than 63bits.
	// Don't bother checking for JUMPDEST in that case.
	if overflow || udest >= uint64(len(c.Code)) {
		return false
	}
	// Only JUMPDESTs allowed for destinations
	if OpCode(c.Code[udest]) != JUMPDEST {
		return false
	}
	return c.isCode(udest)
}

func (c *Ccntmract) validJumpSubdest(udest uint64) bool {
	// PC cannot go beyond len(code) and certainly can't be bigger than 63 bits.
	// Don't bother checking for BEGINSUB in that case.
	if int64(udest) < 0 || udest >= uint64(len(c.Code)) {
		return false
	}
	// Only BEGINSUBs allowed for destinations
	if OpCode(c.Code[udest]) != BEGINSUB {
		return false
	}
	return c.isCode(udest)
}

// isCode returns true if the provided PC location is an actual opcode, as
// opposed to a data-segment following a PUSHN operation.
func (c *Ccntmract) isCode(udest uint64) bool {
	// Do we already have an analysis laying around?
	if c.analysis != nil {
		return c.analysis.codeSegment(udest)
	}
	// Do we have a ccntmract hash already?
	// If we do have a hash, that means it's a 'regular' ccntmract. For regular
	// ccntmracts ( not temporary initcode), we store the analysis in a map
	if c.CodeHash != (common.Hash{}) {
		// Does parent ccntmext have the analysis?
		analysis, exist := c.jumpdests[c.CodeHash]
		if !exist {
			// Do the analysis and save in parent ccntmext
			// We do not need to store it in c.analysis
			analysis = codeBitmap(c.Code)
			c.jumpdests[c.CodeHash] = analysis
		}
		// Also stash it in current ccntmract for faster access
		c.analysis = analysis
		return analysis.codeSegment(udest)
	}
	// We don't have the code hash, most likely a piece of initcode not already
	// in state trie. In that case, we do an analysis, and save it locally, so
	// we don't have to recalculate it for every JUMP instruction in the execution
	// However, we don't save it within the parent ccntmext
	if c.analysis == nil {
		c.analysis = codeBitmap(c.Code)
	}
	return c.analysis.codeSegment(udest)
}

// AsDelegate sets the ccntmract to be a delegate call and returns the current
// ccntmract (for chaining calls)
func (c *Ccntmract) AsDelegate() *Ccntmract {
	// NOTE: caller must, at all times be a ccntmract. It should never happen
	// that caller is something other than a Ccntmract.
	parent := c.caller.(*Ccntmract)
	c.CallerAddress = parent.CallerAddress
	c.value = parent.value

	return c
}

// GetOp returns the n'th element in the ccntmract's byte array
func (c *Ccntmract) GetOp(n uint64) OpCode {
	return OpCode(c.GetByte(n))
}

// GetByte returns the n'th byte in the ccntmract's byte array
func (c *Ccntmract) GetByte(n uint64) byte {
	if n < uint64(len(c.Code)) {
		return c.Code[n]
	}

	return 0
}

// Caller returns the caller of the ccntmract.
//
// Caller will recursively call caller when the ccntmract is a delegate
// call, including that of caller's caller.
func (c *Ccntmract) Caller() common.Address {
	return c.CallerAddress
}

// UseGas attempts the use gas and subtracts it and returns true on success
func (c *Ccntmract) UseGas(gas uint64) (ok bool) {
	if c.Gas < gas {
		return false
	}
	c.Gas -= gas
	return true
}

// Address returns the ccntmracts address
func (c *Ccntmract) Address() common.Address {
	return c.self.Address()
}

// Value returns the ccntmract's value (sent to it from it's caller)
func (c *Ccntmract) Value() *big.Int {
	return c.value
}

// SetCallCode sets the code of the ccntmract and address of the backing data
// object
func (c *Ccntmract) SetCallCode(addr *common.Address, hash common.Hash, code []byte) {
	c.Code = code
	c.CodeHash = hash
	c.CodeAddr = addr
}

// SetCodeOptionalHash can be used to provide code, but it's optional to provide hash.
// In case hash is not provided, the jumpdest analysis will not be saved to the parent ccntmext
func (c *Ccntmract) SetCodeOptionalHash(addr *common.Address, codeAndHash *codeAndHash) {
	c.Code = codeAndHash.code
	c.CodeHash = codeAndHash.hash
	c.CodeAddr = addr
}
