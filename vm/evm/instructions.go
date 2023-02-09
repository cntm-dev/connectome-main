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
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
	"github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/vm/evm/errors"
	"github.com/cntmio/cntmology/vm/evm/params"
	"golang.org/x/crypto/sha3"
)

func opAdd(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	x, y := callCcntmext.stack.pop(), callCcntmext.stack.peek()
	y.Add(&x, y)
	return nil, nil
}

func opSub(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	x, y := callCcntmext.stack.pop(), callCcntmext.stack.peek()
	y.Sub(&x, y)
	return nil, nil
}

func opMul(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	x, y := callCcntmext.stack.pop(), callCcntmext.stack.peek()
	y.Mul(&x, y)
	return nil, nil
}

func opDiv(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	x, y := callCcntmext.stack.pop(), callCcntmext.stack.peek()
	y.Div(&x, y)
	return nil, nil
}

func opSdiv(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	x, y := callCcntmext.stack.pop(), callCcntmext.stack.peek()
	y.SDiv(&x, y)
	return nil, nil
}

func opMod(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	x, y := callCcntmext.stack.pop(), callCcntmext.stack.peek()
	y.Mod(&x, y)
	return nil, nil
}

func opSmod(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	x, y := callCcntmext.stack.pop(), callCcntmext.stack.peek()
	y.SMod(&x, y)
	return nil, nil
}

func opExp(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	base, exponent := callCcntmext.stack.pop(), callCcntmext.stack.peek()
	exponent.Exp(&base, exponent)
	return nil, nil
}

func opSignExtend(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	back, num := callCcntmext.stack.pop(), callCcntmext.stack.peek()
	num.ExtendSign(num, &back)
	return nil, nil
}

func opNot(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	x := callCcntmext.stack.peek()
	x.Not(x)
	return nil, nil
}

func opLt(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	x, y := callCcntmext.stack.pop(), callCcntmext.stack.peek()
	if x.Lt(y) {
		y.SetOne()
	} else {
		y.Clear()
	}
	return nil, nil
}

func opGt(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	x, y := callCcntmext.stack.pop(), callCcntmext.stack.peek()
	if x.Gt(y) {
		y.SetOne()
	} else {
		y.Clear()
	}
	return nil, nil
}

func opSlt(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	x, y := callCcntmext.stack.pop(), callCcntmext.stack.peek()
	if x.Slt(y) {
		y.SetOne()
	} else {
		y.Clear()
	}
	return nil, nil
}

func opSgt(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	x, y := callCcntmext.stack.pop(), callCcntmext.stack.peek()
	if x.Sgt(y) {
		y.SetOne()
	} else {
		y.Clear()
	}
	return nil, nil
}

func opEq(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	x, y := callCcntmext.stack.pop(), callCcntmext.stack.peek()
	if x.Eq(y) {
		y.SetOne()
	} else {
		y.Clear()
	}
	return nil, nil
}

func opIszero(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	x := callCcntmext.stack.peek()
	if x.IsZero() {
		x.SetOne()
	} else {
		x.Clear()
	}
	return nil, nil
}

func opAnd(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	x, y := callCcntmext.stack.pop(), callCcntmext.stack.peek()
	y.And(&x, y)
	return nil, nil
}

func opOr(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	x, y := callCcntmext.stack.pop(), callCcntmext.stack.peek()
	y.Or(&x, y)
	return nil, nil
}

func opXor(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	x, y := callCcntmext.stack.pop(), callCcntmext.stack.peek()
	y.Xor(&x, y)
	return nil, nil
}

func opByte(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	th, val := callCcntmext.stack.pop(), callCcntmext.stack.peek()
	val.Byte(&th)
	return nil, nil
}

func opAddmod(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	x, y, z := callCcntmext.stack.pop(), callCcntmext.stack.pop(), callCcntmext.stack.peek()
	if z.IsZero() {
		z.Clear()
	} else {
		z.AddMod(&x, &y, z)
	}
	return nil, nil
}

func opMulmod(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	x, y, z := callCcntmext.stack.pop(), callCcntmext.stack.pop(), callCcntmext.stack.peek()
	z.MulMod(&x, &y, z)
	return nil, nil
}

// opSHL implements Shift Left
// The SHL instruction (shift left) pops 2 values from the stack, first arg1 and then arg2,
// and pushes on the stack arg2 shifted to the left by arg1 number of bits.
func opSHL(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	// Note, second operand is left in the stack; accumulate result into it, and no need to push it afterwards
	shift, value := callCcntmext.stack.pop(), callCcntmext.stack.peek()
	if shift.LtUint64(256) {
		value.Lsh(value, uint(shift.Uint64()))
	} else {
		value.Clear()
	}
	return nil, nil
}

// opSHR implements Logical Shift Right
// The SHR instruction (logical shift right) pops 2 values from the stack, first arg1 and then arg2,
// and pushes on the stack arg2 shifted to the right by arg1 number of bits with zero fill.
func opSHR(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	// Note, second operand is left in the stack; accumulate result into it, and no need to push it afterwards
	shift, value := callCcntmext.stack.pop(), callCcntmext.stack.peek()
	if shift.LtUint64(256) {
		value.Rsh(value, uint(shift.Uint64()))
	} else {
		value.Clear()
	}
	return nil, nil
}

// opSAR implements Arithmetic Shift Right
// The SAR instruction (arithmetic shift right) pops 2 values from the stack, first arg1 and then arg2,
// and pushes on the stack arg2 shifted to the right by arg1 number of bits with sign extension.
func opSAR(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	shift, value := callCcntmext.stack.pop(), callCcntmext.stack.peek()
	if shift.GtUint64(256) {
		if value.Sign() >= 0 {
			value.Clear()
		} else {
			// Max negative shift: all bits set
			value.SetAllOne()
		}
		return nil, nil
	}
	n := uint(shift.Uint64())
	value.SRsh(value, n)
	return nil, nil
}

func opSha3(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	offset, size := callCcntmext.stack.pop(), callCcntmext.stack.peek()
	data := callCcntmext.memory.GetPtr(int64(offset.Uint64()), int64(size.Uint64()))

	if interpreter.hasher == nil {
		interpreter.hasher = sha3.NewLegacyKeccak256().(keccakState)
	} else {
		interpreter.hasher.Reset()
	}
	interpreter.hasher.Write(data)
	interpreter.hasher.Read(interpreter.hasherBuf[:])

	evm := interpreter.evm
	if evm.vmConfig.EnablePreimageRecording {
		evm.StateDB.AddPreimage(interpreter.hasherBuf, data)
	}

	size.SetBytes(interpreter.hasherBuf[:])
	return nil, nil
}
func opAddress(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	callCcntmext.stack.push(new(uint256.Int).SetBytes(callCcntmext.ccntmract.Address().Bytes()))
	return nil, nil
}

func opBalance(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	slot := callCcntmext.stack.peek()
	address := common.Address(slot.Bytes20())
	slot.SetFromBig(interpreter.evm.StateDB.GetBalance(address))
	return nil, nil
}

func opOrigin(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	callCcntmext.stack.push(new(uint256.Int).SetBytes(interpreter.evm.Origin.Bytes()))
	return nil, nil
}
func opCaller(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	callCcntmext.stack.push(new(uint256.Int).SetBytes(callCcntmext.ccntmract.Caller().Bytes()))
	return nil, nil
}

func opCallValue(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	v, _ := uint256.FromBig(callCcntmext.ccntmract.value)
	callCcntmext.stack.push(v)
	return nil, nil
}

func opCallDataLoad(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	x := callCcntmext.stack.peek()
	if offset, overflow := x.Uint64WithOverflow(); !overflow {
		data := getData(callCcntmext.ccntmract.Input, offset, 32)
		x.SetBytes(data)
	} else {
		x.Clear()
	}
	return nil, nil
}

func opCallDataSize(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	callCcntmext.stack.push(new(uint256.Int).SetUint64(uint64(len(callCcntmext.ccntmract.Input))))
	return nil, nil
}

func opCallDataCopy(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	var (
		memOffset  = callCcntmext.stack.pop()
		dataOffset = callCcntmext.stack.pop()
		length     = callCcntmext.stack.pop()
	)
	dataOffset64, overflow := dataOffset.Uint64WithOverflow()
	if overflow {
		dataOffset64 = 0xffffffffffffffff
	}
	// These values are checked for overflow during gas cost calculation
	memOffset64 := memOffset.Uint64()
	length64 := length.Uint64()
	callCcntmext.memory.Set(memOffset64, length64, getData(callCcntmext.ccntmract.Input, dataOffset64, length64))

	return nil, nil
}

func opReturnDataSize(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	callCcntmext.stack.push(new(uint256.Int).SetUint64(uint64(len(interpreter.returnData))))
	return nil, nil
}

func opReturnDataCopy(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	var (
		memOffset  = callCcntmext.stack.pop()
		dataOffset = callCcntmext.stack.pop()
		length     = callCcntmext.stack.pop()
	)

	offset64, overflow := dataOffset.Uint64WithOverflow()
	if overflow {
		return nil, errors.ErrReturnDataOutOfBounds
	}
	// we can reuse dataOffset now (aliasing it for clarity)
	var end = dataOffset
	end.Add(&dataOffset, &length)
	end64, overflow := end.Uint64WithOverflow()
	if overflow || uint64(len(interpreter.returnData)) < end64 {
		return nil, errors.ErrReturnDataOutOfBounds
	}
	callCcntmext.memory.Set(memOffset.Uint64(), length.Uint64(), interpreter.returnData[offset64:end64])
	return nil, nil
}

func opExtCodeSize(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	slot := callCcntmext.stack.peek()
	slot.SetUint64(uint64(interpreter.evm.StateDB.GetCodeSize(slot.Bytes20())))
	return nil, nil
}

func opCodeSize(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	l := new(uint256.Int)
	l.SetUint64(uint64(len(callCcntmext.ccntmract.Code)))
	callCcntmext.stack.push(l)
	return nil, nil
}

func opCodeCopy(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	var (
		memOffset  = callCcntmext.stack.pop()
		codeOffset = callCcntmext.stack.pop()
		length     = callCcntmext.stack.pop()
	)
	uint64CodeOffset, overflow := codeOffset.Uint64WithOverflow()
	if overflow {
		uint64CodeOffset = 0xffffffffffffffff
	}
	codeCopy := getData(callCcntmext.ccntmract.Code, uint64CodeOffset, length.Uint64())
	callCcntmext.memory.Set(memOffset.Uint64(), length.Uint64(), codeCopy)

	return nil, nil
}

func opExtCodeCopy(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	var (
		stack      = callCcntmext.stack
		a          = stack.pop()
		memOffset  = stack.pop()
		codeOffset = stack.pop()
		length     = stack.pop()
	)
	uint64CodeOffset, overflow := codeOffset.Uint64WithOverflow()
	if overflow {
		uint64CodeOffset = 0xffffffffffffffff
	}
	addr := common.Address(a.Bytes20())
	codeCopy := getData(interpreter.evm.StateDB.GetCode(addr), uint64CodeOffset, length.Uint64())
	callCcntmext.memory.Set(memOffset.Uint64(), length.Uint64(), codeCopy)

	return nil, nil
}

// opExtCodeHash returns the code hash of a specified account.
// There are several cases when the function is called, while we can relay everything
// to `state.GetCodeHash` function to ensure the correctness.
//   (1) Caller tries to get the code hash of a normal ccntmract account, state
// should return the relative code hash and set it as the result.
//
//   (2) Caller tries to get the code hash of a non-existent account, state should
// return common.Hash{} and zero will be set as the result.
//
//   (3) Caller tries to get the code hash for an account without ccntmract code,
// state should return emptyCodeHash(0xc5d246...) as the result.
//
//   (4) Caller tries to get the code hash of a precompiled account, the result
// should be zero or emptyCodeHash.
//
// It is worth noting that in order to avoid unnecessary create and clean,
// all precompile accounts on mainnet have been transferred 1 wei, so the return
// here should be emptyCodeHash.
// If the precompile account is not transferred any amount on a private or
// customized chain, the return value will be zero.
//
//   (5) Caller tries to get the code hash for an account which is marked as suicided
// in the current transaction, the code hash of this account should be returned.
//
//   (6) Caller tries to get the code hash for an account which is marked as deleted,
// this account should be regarded as a non-existent account and zero should be returned.
func opExtCodeHash(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	slot := callCcntmext.stack.peek()
	address := common.Address(slot.Bytes20())
	if interpreter.evm.StateDB.Empty(address) {
		slot.Clear()
	} else {
		slot.SetBytes(interpreter.evm.StateDB.GetCodeHash(address).Bytes())
	}
	return nil, nil
}

func opGasprice(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	v, _ := uint256.FromBig(interpreter.evm.GasPrice)
	callCcntmext.stack.push(v)
	return nil, nil
}

func opBlockhash(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	num := callCcntmext.stack.peek()
	num64, overflow := num.Uint64WithOverflow()
	if overflow {
		num.Clear()
		return nil, nil
	}
	var upper, lower uint64
	upper = interpreter.evm.Ccntmext.BlockNumber.Uint64()
	if upper < 257 {
		lower = 0
	} else {
		lower = upper - 256
	}
	if num64 >= lower && num64 < upper {
		num.SetBytes(interpreter.evm.Ccntmext.GetHash(num64).Bytes())
	} else {
		num.Clear()
	}
	return nil, nil
}

func opCoinbase(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	callCcntmext.stack.push(new(uint256.Int).SetBytes(interpreter.evm.Ccntmext.Coinbase.Bytes()))
	return nil, nil
}

func opTimestamp(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	v, _ := uint256.FromBig(interpreter.evm.Ccntmext.Time)
	callCcntmext.stack.push(v)
	return nil, nil
}

func opNumber(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	v, _ := uint256.FromBig(interpreter.evm.Ccntmext.BlockNumber)
	callCcntmext.stack.push(v)
	return nil, nil
}

func opDifficulty(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	v, _ := uint256.FromBig(interpreter.evm.Ccntmext.Difficulty)
	callCcntmext.stack.push(v)
	return nil, nil
}

func opGasLimit(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	callCcntmext.stack.push(new(uint256.Int).SetUint64(interpreter.evm.Ccntmext.GasLimit))
	return nil, nil
}

func opPop(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	callCcntmext.stack.pop()
	return nil, nil
}

func opMload(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	v := callCcntmext.stack.peek()
	offset := int64(v.Uint64())
	v.SetBytes(callCcntmext.memory.GetPtr(offset, 32))
	return nil, nil
}

func opMstore(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	// pop value of the stack
	mStart, val := callCcntmext.stack.pop(), callCcntmext.stack.pop()
	callCcntmext.memory.Set32(mStart.Uint64(), &val)
	return nil, nil
}

func opMstore8(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	off, val := callCcntmext.stack.pop(), callCcntmext.stack.pop()
	callCcntmext.memory.store[off.Uint64()] = byte(val.Uint64())
	return nil, nil
}

func opSload(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	loc := callCcntmext.stack.peek()
	hash := common.Hash(loc.Bytes32())
	val := interpreter.evm.StateDB.GetState(callCcntmext.ccntmract.Address(), hash)
	loc.SetBytes(val.Bytes())
	return nil, nil
}

func opSstore(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	loc := callCcntmext.stack.pop()
	val := callCcntmext.stack.pop()
	interpreter.evm.StateDB.SetState(callCcntmext.ccntmract.Address(),
		loc.Bytes32(), val.Bytes32())
	return nil, nil
}

func opJump(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	pos := callCcntmext.stack.pop()
	if !callCcntmext.ccntmract.validJumpdest(&pos) {
		return nil, errors.ErrInvalidJump
	}
	*pc = pos.Uint64()
	return nil, nil
}

func opJumpi(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	pos, cond := callCcntmext.stack.pop(), callCcntmext.stack.pop()
	if !cond.IsZero() {
		if !callCcntmext.ccntmract.validJumpdest(&pos) {
			return nil, errors.ErrInvalidJump
		}
		*pc = pos.Uint64()
	} else {
		*pc++
	}
	return nil, nil
}

func opJumpdest(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	return nil, nil
}

func opBeginSub(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	return nil, errors.ErrInvalidSubroutineEntry
}

func opJumpSub(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	if len(callCcntmext.rstack.data) >= 1023 {
		return nil, errors.ErrReturnStackExceeded
	}
	pos := callCcntmext.stack.pop()
	if !pos.IsUint64() {
		return nil, errors.ErrInvalidJump
	}
	posU64 := pos.Uint64()
	if !callCcntmext.ccntmract.validJumpSubdest(posU64) {
		return nil, errors.ErrInvalidJump
	}
	callCcntmext.rstack.push(uint32(*pc))
	*pc = posU64 + 1
	return nil, nil
}

func opReturnSub(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	if len(callCcntmext.rstack.data) == 0 {
		return nil, errors.ErrInvalidRetsub
	}
	// Other than the check that the return stack is not empty, there is no
	// need to validate the pc from 'returns', since we only ever push valid
	//values cntmo it via jumpsub.
	*pc = uint64(callCcntmext.rstack.pop()) + 1
	return nil, nil
}

func opPc(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	callCcntmext.stack.push(new(uint256.Int).SetUint64(*pc))
	return nil, nil
}

func opMsize(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	callCcntmext.stack.push(new(uint256.Int).SetUint64(uint64(callCcntmext.memory.Len())))
	return nil, nil
}

func opGas(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	callCcntmext.stack.push(new(uint256.Int).SetUint64(callCcntmext.ccntmract.Gas))
	return nil, nil
}

func opCreate(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	var (
		value        = callCcntmext.stack.pop()
		offset, size = callCcntmext.stack.pop(), callCcntmext.stack.pop()
		input        = callCcntmext.memory.GetCopy(int64(offset.Uint64()), int64(size.Uint64()))
		gas          = callCcntmext.ccntmract.Gas
	)
	if interpreter.evm.chainRules.IsEIP150 {
		gas -= gas / 64
	}
	// reuse size int for stackvalue
	stackvalue := size

	callCcntmext.ccntmract.UseGas(gas)
	//TODO: use uint256.Int instead of converting with toBig()
	var bigVal = big0
	if !value.IsZero() {
		bigVal = value.ToBig()
	}

	res, addr, returnGas, suberr := interpreter.evm.Create(callCcntmext.ccntmract, input, gas, bigVal)
	// Push item on the stack based on the returned error. If the ruleset is
	// homestead we must check for CodeStoreOutOfGasError (homestead only
	// rule) and treat as an error, if the ruleset is frcntmier we must
	// ignore this error and pretend the operation was successful.
	if interpreter.evm.chainRules.IsHomestead && suberr == errors.ErrCodeStoreOutOfGas {
		stackvalue.Clear()
	} else if suberr != nil && suberr != errors.ErrCodeStoreOutOfGas {
		stackvalue.Clear()
	} else {
		stackvalue.SetBytes(addr.Bytes())
	}
	callCcntmext.stack.push(&stackvalue)
	callCcntmext.ccntmract.Gas += returnGas

	if suberr == errors.ErrExecutionReverted {
		return res, nil
	}
	return nil, nil
}

func opCreate2(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	var (
		endowment    = callCcntmext.stack.pop()
		offset, size = callCcntmext.stack.pop(), callCcntmext.stack.pop()
		salt         = callCcntmext.stack.pop()
		input        = callCcntmext.memory.GetCopy(int64(offset.Uint64()), int64(size.Uint64()))
		gas          = callCcntmext.ccntmract.Gas
	)

	// Apply EIP150
	gas -= gas / 64
	callCcntmext.ccntmract.UseGas(gas)
	// reuse size int for stackvalue
	stackvalue := size
	//TODO: use uint256.Int instead of converting with toBig()
	bigEndowment := big0
	if !endowment.IsZero() {
		bigEndowment = endowment.ToBig()
	}
	res, addr, returnGas, suberr := interpreter.evm.Create2(callCcntmext.ccntmract, input, gas,
		bigEndowment, &salt)
	// Push item on the stack based on the returned error.
	if suberr != nil {
		stackvalue.Clear()
	} else {
		stackvalue.SetBytes(addr.Bytes())
	}
	callCcntmext.stack.push(&stackvalue)
	callCcntmext.ccntmract.Gas += returnGas

	if suberr == errors.ErrExecutionReverted {
		return res, nil
	}
	return nil, nil
}

func opCall(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	stack := callCcntmext.stack
	// Pop gas. The actual gas in interpreter.evm.callGasTemp.
	// We can use this as a temporary value
	temp := stack.pop()
	gas := interpreter.evm.callGasTemp
	// Pop other call parameters.
	addr, value, inOffset, inSize, retOffset, retSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()
	toAddr := common.Address(addr.Bytes20())
	// Get the arguments from the memory.
	args := callCcntmext.memory.GetPtr(int64(inOffset.Uint64()), int64(inSize.Uint64()))

	var bigVal = big0
	//TODO: use uint256.Int instead of converting with toBig()
	// By using big0 here, we save an alloc for the most common case (non-ether-transferring ccntmract calls),
	// but it would make more sense to extend the usage of uint256.Int
	if !value.IsZero() {
		gas += params.CallStipend
		bigVal = value.ToBig()
	}

	ret, returnGas, err := interpreter.evm.Call(callCcntmext.ccntmract, toAddr, args, gas, bigVal)

	if err != nil {
		temp.Clear()
	} else {
		temp.SetOne()
	}
	stack.push(&temp)
	if err == nil || err == errors.ErrExecutionReverted {
		callCcntmext.memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
	}
	callCcntmext.ccntmract.Gas += returnGas

	return ret, nil
}

func opCallCode(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	// Pop gas. The actual gas is in interpreter.evm.callGasTemp.
	stack := callCcntmext.stack
	// We use it as a temporary value
	temp := stack.pop()
	gas := interpreter.evm.callGasTemp
	// Pop other call parameters.
	addr, value, inOffset, inSize, retOffset, retSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()
	toAddr := common.Address(addr.Bytes20())
	// Get arguments from the memory.
	args := callCcntmext.memory.GetPtr(int64(inOffset.Uint64()), int64(inSize.Uint64()))

	//TODO: use uint256.Int instead of converting with toBig()
	var bigVal = big0
	if !value.IsZero() {
		gas += params.CallStipend
		bigVal = value.ToBig()
	}

	ret, returnGas, err := interpreter.evm.CallCode(callCcntmext.ccntmract, toAddr, args, gas, bigVal)
	if err != nil {
		temp.Clear()
	} else {
		temp.SetOne()
	}
	stack.push(&temp)
	if err == nil || err == errors.ErrExecutionReverted {
		callCcntmext.memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
	}
	callCcntmext.ccntmract.Gas += returnGas

	return ret, nil
}

func opDelegateCall(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	stack := callCcntmext.stack
	// Pop gas. The actual gas is in interpreter.evm.callGasTemp.
	// We use it as a temporary value
	temp := stack.pop()
	gas := interpreter.evm.callGasTemp
	// Pop other call parameters.
	addr, inOffset, inSize, retOffset, retSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()
	toAddr := common.Address(addr.Bytes20())
	// Get arguments from the memory.
	args := callCcntmext.memory.GetPtr(int64(inOffset.Uint64()), int64(inSize.Uint64()))

	ret, returnGas, err := interpreter.evm.DelegateCall(callCcntmext.ccntmract, toAddr, args, gas)
	if err != nil {
		temp.Clear()
	} else {
		temp.SetOne()
	}
	stack.push(&temp)
	if err == nil || err == errors.ErrExecutionReverted {
		callCcntmext.memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
	}
	callCcntmext.ccntmract.Gas += returnGas

	return ret, nil
}

func opStaticCall(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	// Pop gas. The actual gas is in interpreter.evm.callGasTemp.
	stack := callCcntmext.stack
	// We use it as a temporary value
	temp := stack.pop()
	gas := interpreter.evm.callGasTemp
	// Pop other call parameters.
	addr, inOffset, inSize, retOffset, retSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()
	toAddr := common.Address(addr.Bytes20())
	// Get arguments from the memory.
	args := callCcntmext.memory.GetPtr(int64(inOffset.Uint64()), int64(inSize.Uint64()))

	ret, returnGas, err := interpreter.evm.StaticCall(callCcntmext.ccntmract, toAddr, args, gas)
	if err != nil {
		temp.Clear()
	} else {
		temp.SetOne()
	}
	stack.push(&temp)
	if err == nil || err == errors.ErrExecutionReverted {
		callCcntmext.memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
	}
	callCcntmext.ccntmract.Gas += returnGas

	return ret, nil
}

func opReturn(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	offset, size := callCcntmext.stack.pop(), callCcntmext.stack.pop()
	ret := callCcntmext.memory.GetPtr(int64(offset.Uint64()), int64(size.Uint64()))

	return ret, nil
}

func opRevert(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	offset, size := callCcntmext.stack.pop(), callCcntmext.stack.pop()
	ret := callCcntmext.memory.GetPtr(int64(offset.Uint64()), int64(size.Uint64()))

	return ret, nil
}

func opStop(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	return nil, nil
}

func opSuicide(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	beneficiary := callCcntmext.stack.pop()
	balance := interpreter.evm.StateDB.GetBalance(callCcntmext.ccntmract.Address())
	interpreter.evm.StateDB.AddBalance(beneficiary.Bytes20(), balance)
	interpreter.evm.StateDB.Suicide(callCcntmext.ccntmract.Address())
	return nil, nil
}

// following functions are used by the instruction jump  table

// make log instruction function
func makeLog(size int) executionFunc {
	return func(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
		topics := make([]common.Hash, size)
		stack := callCcntmext.stack
		mStart, mSize := stack.pop(), stack.pop()
		for i := 0; i < size; i++ {
			addr := stack.pop()
			topics[i] = addr.Bytes32()
		}

		d := callCcntmext.memory.GetCopy(int64(mStart.Uint64()), int64(mSize.Uint64()))
		interpreter.evm.StateDB.AddLog(&types.StorageLog{
			Address: callCcntmext.ccntmract.Address(),
			Topics:  topics,
			Data:    d,
		})

		return nil, nil
	}
}

// opPush1 is a specialized version of pushN
func opPush1(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
	var (
		codeLen = uint64(len(callCcntmext.ccntmract.Code))
		integer = new(uint256.Int)
	)
	*pc += 1
	if *pc < codeLen {
		callCcntmext.stack.push(integer.SetUint64(uint64(callCcntmext.ccntmract.Code[*pc])))
	} else {
		callCcntmext.stack.push(integer.Clear())
	}
	return nil, nil
}

// make push instruction function
func makePush(size uint64, pushByteSize int) executionFunc {
	return func(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
		codeLen := len(callCcntmext.ccntmract.Code)

		startMin := codeLen
		if int(*pc+1) < startMin {
			startMin = int(*pc + 1)
		}

		endMin := codeLen
		if startMin+pushByteSize < endMin {
			endMin = startMin + pushByteSize
		}

		integer := new(uint256.Int)
		callCcntmext.stack.push(integer.SetBytes(common.RightPadBytes(
			callCcntmext.ccntmract.Code[startMin:endMin], pushByteSize)))

		*pc += size
		return nil, nil
	}
}

// make dup instruction function
func makeDup(size int64) executionFunc {
	return func(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
		callCcntmext.stack.dup(int(size))
		return nil, nil
	}
}

// make swap instruction function
func makeSwap(size int64) executionFunc {
	// switch n + 1 otherwise n would be swapped with n
	size++
	return func(pc *uint64, interpreter *EVMInterpreter, callCcntmext *callCtx) ([]byte, error) {
		callCcntmext.stack.swap(int(size))
		return nil, nil
	}
}
