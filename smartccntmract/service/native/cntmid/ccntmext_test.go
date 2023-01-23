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
package cntmid

import (
	"fmt"
	"testing"

	"github.com/cntmio/cntmology/account"
	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/smartccntmract/service/native"
)

func TestCcntmext(t *testing.T) {
	testcase(t, CaseCcntmext)
}

func CaseCcntmext(t *testing.T, n *native.NativeService) {
	id, err := account.GenerateID()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Print(id)
	acc := account.NewAccount("")
	if regID(n, id, acc) != nil {
		t.Fatal("register id error")
	}
	var ccntmexts = [][]byte{[]byte("https://www.w3.org/ns0/did/v1"), []byte("https://cntmid.cntm.io0/did/v1"), []byte("https://cntmid.cntm.io0/did/v1")}
	ccntmext := &Ccntmext{
		OntId:    []byte(id),
		Ccntmexts: ccntmexts,
		Index:    1,
	}
	sink := common.NewZeroCopySink(nil)
	ccntmext.Serialization(sink)
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{acc.Address}
	_, err = addCcntmext(n)
	if err != nil {
		t.Fatal()
	}
	encId, err := encodeID([]byte(id))
	if err != nil {
		t.Fatal()
	}
	key := append(encId, FIELD_CcntmEXT)
	res, err := getCcntmexts(n, key)
	if err != nil {
		t.Fatal()
	}
	for i := 0; i < len(res); i++ {
		fmt.Println(common.ToHexString(res[i]))
	}

	ccntmextsJson, err := getCcntmextsWithDefault(n, encId)
	if err != nil {
		t.Fatal()
	}
	fmt.Println(ccntmextsJson)

	ccntmexts = [][]byte{[]byte("https://www.w3.org/ns0/did/v1")}
	ccntmext = &Ccntmext{
		OntId:    []byte(id),
		Ccntmexts: ccntmexts,
		Index:    1,
	}
	sink = common.NewZeroCopySink(nil)
	ccntmext.Serialization(sink)
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{acc.Address}
	_, err = removeCcntmext(n)
	if err != nil {
		t.Fatal()
	}
	res, err = getCcntmexts(n, key)
	if err != nil {
		t.Fatal()
	}
	for i := 0; i < len(res); i++ {
		fmt.Println(common.ToHexString(res[i]))
	}
}
