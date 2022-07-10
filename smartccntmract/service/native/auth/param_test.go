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
package auth

import (
	"bytes"
	"crypto/rand"
	"testing"
)

var (
	admin    []byte
	newAdmin []byte
	p1       []byte
	p2       []byte
)
var (
	funcs           = []string{"foo1", "foo2"}
	role            = "role"
	OntCcntmractAddr = []byte{0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}
)

func init() {
	admin = make([]byte, 32)
	newAdmin = make([]byte, 32)
	p1 = make([]byte, 20)
	p2 = make([]byte, 20)
	rand.Read(admin)
	rand.Read(newAdmin)
	rand.Read(p1)
	rand.Read(p2)
}
func TestSerialization_Init(t *testing.T) {
	param := &InitCcntmractAdminParam{
		AdminOntID: admin,
	}
	bf := new(bytes.Buffer)
	if err := param.Serialize(bf); err != nil {
		t.Fatal(err)
	}
	rd := bytes.NewReader(bf.Bytes())

	param2 := new(InitCcntmractAdminParam)
	if err := param2.Deserialize(rd); err != nil {
		t.Fatal(err)
	}

	if bytes.Compare(param.AdminOntID, param2.AdminOntID) != 0 {
		t.Fatalf("failed")
	}
}

func TestSerialization_Transfer(t *testing.T) {
	param := &TransferParam{
		CcntmractAddr:  OntCcntmractAddr,
		NewAdminOntID: newAdmin,
	}
	bf := new(bytes.Buffer)
	if err := param.Serialize(bf); err != nil {
		t.Fatal(err)
	}
	rd := bytes.NewReader(bf.Bytes())

	param2 := new(TransferParam)
	if err := param2.Deserialize(rd); err != nil {
		t.Fatal(err)
	}

	if bytes.Compare(param.CcntmractAddr, param2.CcntmractAddr) != 0 ||
		bytes.Compare(param.NewAdminOntID, param2.NewAdminOntID) != 0 {
		t.Fatalf("failed")
	}
}

func TestSerialization_AssignFuncs(t *testing.T) {
	param := &FuncsToRoleParam{
		CcntmractAddr: OntCcntmractAddr,
		AdminOntID:   admin,
		Role:         []byte("role"),
		FuncNames:    funcs,
	}
	bf := new(bytes.Buffer)
	if err := param.Serialize(bf); err != nil {
		t.Fatal(err)
	}
	rd := bytes.NewReader(bf.Bytes())

	param2 := new(FuncsToRoleParam)
	if err := param2.Deserialize(rd); err != nil {
		t.Fatal(err)
	}

	if bytes.Compare(param.CcntmractAddr, param2.CcntmractAddr) != 0 ||
		bytes.Compare(param.AdminOntID, param2.AdminOntID) != 0 ||
		bytes.Compare(param.Role, param2.Role) != 0 {
		t.Fatalf("failed")
	}
}

func TestSerialization_AssignOntIDs(t *testing.T) {
	param := &OntIDsToRoleParam{
		CcntmractAddr: OntCcntmractAddr,
		AdminOntID:   admin,
		Role:         []byte(role),
		Persons:      [][]byte{[]byte{0x03, 0x04, 0x05, 0x06}, []byte{0x07, 0x08, 0x09, 0x0a}},
	}
	bf := new(bytes.Buffer)
	if err := param.Serialize(bf); err != nil {
		t.Fatal(err)
	}
	rd := bytes.NewReader(bf.Bytes())
	param2 := new(OntIDsToRoleParam)
	if err := param2.Deserialize(rd); err != nil {
		t.Fatal(err)
	}

	if bytes.Compare(param.CcntmractAddr, param2.CcntmractAddr) != 0 ||
		bytes.Compare(param.AdminOntID, param2.AdminOntID) != 0 ||
		bytes.Compare(param.Role, param2.Role) != 0 {
		t.Fatalf("failed")
	}
}

func TestSerialization_Delegate(t *testing.T) {
	param := &DelegateParam{
		CcntmractAddr: OntCcntmractAddr,
		From:         p1,
		To:           p2,
		Role:         []byte(role),
		Period:       60 * 60 * 24,
		Level:        3,
	}
	bf := new(bytes.Buffer)
	if err := param.Serialize(bf); err != nil {
		t.Fatal(err)
	}
	rd := bytes.NewReader(bf.Bytes())
	param2 := new(DelegateParam)
	if err := param2.Deserialize(rd); err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(param.CcntmractAddr, param2.CcntmractAddr) != 0 ||
		bytes.Compare(param.From, param2.From) != 0 ||
		bytes.Compare(param.To, param2.To) != 0 ||
		bytes.Compare(param.Role, param2.Role) != 0 ||
		param.Period != param2.Period || param.Level != param2.Level {
		t.Fatalf("failed")
	}
}

func TestSerialization_Withdraw(t *testing.T) {
	param := &WithdrawParam{
		CcntmractAddr: OntCcntmractAddr,
		Initiator:    p1,
		Delegate:     p2,
		Role:         []byte(role),
	}
	bf := new(bytes.Buffer)
	if err := param.Serialize(bf); err != nil {
		t.Fatal(err)
	}
	rd := bytes.NewReader(bf.Bytes())
	param2 := new(WithdrawParam)
	if err := param2.Deserialize(rd); err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(param.CcntmractAddr, param2.CcntmractAddr) != 0 ||
		bytes.Compare(param.Initiator, param2.Initiator) != 0 ||
		bytes.Compare(param.Delegate, param2.Delegate) != 0 ||
		bytes.Compare(param.Role, param2.Role) != 0 {
		t.Fatalf("failed")
	}
}

func TestSerialization_VerifyToken(t *testing.T) {
	param := &VerifyTokenParam{
		CcntmractAddr: OntCcntmractAddr,
		Caller:       p1,
		Fn:           []byte("foo1"),
	}
	bf := new(bytes.Buffer)
	if err := param.Serialize(bf); err != nil {
		t.Fatal(err)
	}
	rd := bytes.NewReader(bf.Bytes())
	param2 := new(VerifyTokenParam)
	if err := param2.Deserialize(rd); err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(param.CcntmractAddr, param2.CcntmractAddr) != 0 ||
		bytes.Compare(param.Caller, param2.Caller) != 0 ||
		bytes.Compare(param.Fn, param2.Fn) != 0 {
		t.Fatalf("failed")
	}
}
