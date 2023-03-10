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
package common

import (
	"bytes"
	"encoding/json"

	utils2 "github.com/cntmio/cntmology/cmd/utils"
	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/core/payload"
	"github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/core/utils"
	common2 "github.com/cntmio/cntmology/http/base/common"
	"github.com/cntmio/cntmology/smartccntmract/states"
	"github.com/cntmio/cntmology/vm/neovm"
)

type TestEnv struct {
	Witness []common.Address `json:"witness"`
}

func (self TestEnv) MarshalJSON() ([]byte, error) {
	var witness []string
	for _, addr := range self.Witness {
		witness = append(witness, addr.ToBase58())
	}
	env := struct {
		Witness []string `json:"witness"`
	}{Witness: witness}

	return json.Marshal(env)
}

func (self *TestEnv) UnmarshalJSON(buf []byte) error {
	env := struct {
		Witness []string `json:"witness"`
	}{}
	err := json.Unmarshal(buf, &env)
	if err != nil {
		return err
	}
	var witness []common.Address
	for _, addr := range env.Witness {
		wit, err := common.AddressFromBase58(addr)
		if err != nil {
			return err
		}

		witness = append(witness, wit)
	}

	self.Witness = witness
	return nil
}

type TestCase struct {
	Env         TestEnv `json:"env"`
	NeedCcntmext bool    `json:"needccntmext"`
	Method      string  `json:"method"`
	Param       string  `json:"param"`
	Expect      string  `json:"expected"`
	Notify      string  `json:"notify"`
}

type TestCcntmext struct {
	Admin   common.Address
	AddrMap map[string]common.Address
}

func GenWasmTransaction(testCase TestCase, ccntmract common.Address, testConext *TestCcntmext) (*types.Transaction, error) {
	params, err := utils2.ParseParams(testCase.Param)
	if err != nil {
		return nil, err
	}
	allParam := append([]interface{}{}, testCase.Method)
	allParam = append(allParam, params...)
	tx, err := utils.NewWasmVMInvokeTransaction(0, 100000000, ccntmract, allParam)
	if err != nil {
		return nil, err
	}

	if testCase.NeedCcntmext {
		source := common.NewZeroCopySource(tx.Payload.(*payload.InvokeCode).Code)
		ccntmract := &states.WasmCcntmractParam{}
		err := ccntmract.Deserialization(source)
		if err != nil {
			return nil, err
		}
		ccntmextParam := buildTestConext(testConext)
		ccntmract.Args = append(ccntmract.Args, ccntmextParam...)

		tx.Payload.(*payload.InvokeCode).Code = common.SerializeToBytes(ccntmract)
	}

	imt, err := tx.IntoImmutable()
	if err != nil {
		return nil, err
	}

	imt.SignedAddr = append(imt.SignedAddr, testCase.Env.Witness...)
	imt.SignedAddr = append(imt.SignedAddr, testConext.Admin)

	return imt, nil
}

// when need pass testConext to neovm ccntmract, must write ccntmract as def Main(operation, args) api. and args need be a list.
func buildTestConextForNeo(testConext *TestCcntmext) []byte {
	addrMap := testConext.AddrMap
	builder := neovm.NewParamsBuilder(new(bytes.Buffer))

	// [args, operation]
	builder.Emit(neovm.SWAP)
	// [operation, args]
	builder.Emit(neovm.TOALTSTACK)
	// [operation]

	// construct [admin, map] array
	builder.EmitPushByteArray(testConext.Admin[:])
	builder.Emit(neovm.NEWMAP)
	for file, addr := range addrMap {
		builder.Emit(neovm.DUP)
		builder.EmitPushByteArray(addr[:])
		builder.Emit(neovm.SWAP)
		builder.EmitPushByteArray([]byte(file))
		builder.Emit(neovm.ROT)
		builder.Emit(neovm.SETITEM)
	}
	builder.Emit(neovm.PUSH2)
	builder.Emit(neovm.PACK)
	// end [addmin, map] array construct

	// [operation, [admin, map]]
	builder.Emit(neovm.FROMALTSTACK)
	// [operation, [admin, map], args]
	builder.Emit(neovm.UNPACK)
	builder.Emit(neovm.PUSH1)
	builder.Emit(neovm.ADD)
	builder.Emit(neovm.PACK)
	// [operation, [args,[admin, map]]]
	builder.Emit(neovm.SWAP)
	// the second list of last elt is the testConext
	// [[args,[admin, map]], operation] ==> topof the stack.
	return builder.ToArray()
}

func GenNeoVMTransaction(testCase TestCase, ccntmract common.Address, testConext *TestCcntmext) (*types.Transaction, error) {
	params, err := utils2.ParseParams(testCase.Param)
	if err != nil {
		return nil, err
	}
	allParam := append([]interface{}{}, testCase.Method)
	allParam = append(allParam, params...)
	tx, err := common2.NewNeovmInvokeTransaction(0, 100000000, ccntmract, allParam)
	if err != nil {
		return nil, err
	}

	if testCase.NeedCcntmext {
		args := buildTestConextForNeo(testConext)
		codelen := uint32(len(tx.Payload.(*payload.InvokeCode).Code))
		tx.Payload.(*payload.InvokeCode).Code = append(tx.Payload.(*payload.InvokeCode).Code[:codelen-(common.ADDR_LEN+1)], args...)
		tx.Payload.(*payload.InvokeCode).Code = append(tx.Payload.(*payload.InvokeCode).Code, 0x67)
		tx.Payload.(*payload.InvokeCode).Code = append(tx.Payload.(*payload.InvokeCode).Code, ccntmract[:]...)
		//neovms.Dumpcode(tx.Payload.(*payload.InvokeCode).Code[:], "")
	}

	imt, err := tx.IntoImmutable()
	if err != nil {
		return nil, err
	}

	imt.SignedAddr = append(imt.SignedAddr, testCase.Env.Witness...)
	imt.SignedAddr = append(imt.SignedAddr, testConext.Admin)

	return imt, nil
}

func buildTestConext(testConext *TestCcntmext) []byte {
	bf := common.NewZeroCopySink(nil)
	addrMap := testConext.AddrMap

	bf.WriteAddress(testConext.Admin)
	bf.WriteVarUint(uint64(len(addrMap)))
	for file, addr := range addrMap {
		bf.WriteString(file)
		bf.WriteAddress(addr)
	}

	return bf.Bytes()
}
