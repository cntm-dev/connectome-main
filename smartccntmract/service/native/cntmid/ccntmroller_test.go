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
	"bytes"
	"testing"

	"github.com/cntmio/cntmology-crypto/keypair"
	"github.com/cntmio/cntmology/account"
	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/smartccntmract/service/native"
	"github.com/cntmio/cntmology/smartccntmract/service/native/utils"
)

func TestCaseCcntmroller(t *testing.T) {
	testcase(t, CaseCcntmroller)
}

func TestGroupCcntmroller(t *testing.T) {
	testcase(t, CaseGroupCcntmroller)
}

// Test case: register an ID ccntmrolled by another ID
func CaseCcntmroller(t *testing.T, n *native.NativeService) {
	a0 := account.NewAccount("")
	id0, _ := account.GenerateID()
	id1, _ := account.GenerateID()

	// 1. unregistered ccntmroller, should fail
	if err := regCcntmrolledID(n, id1, id0, 1, a0.Address); err == nil {
		t.Error("registered ccntmrolled id with unregistered ccntmroller")
	}

	// 2. register the ccntmroller
	if err := regID(n, id0, a0); err != nil {
		t.Fatal(err)
	}

	// 3. register without valid signature, should fail
	if err := regCcntmrolledID(n, id1, id0, 1, common.ADDRESS_EMPTY); err == nil {
		t.Error("registered without valid signature")
	}

	// 4. register with invalid key index, should fail
	if err := regCcntmrolledID(n, id1, id0, 2, a0.Address); err == nil {
		t.Error("registered with invalid key index")
	}

	// 5. register with invalid id, should fail
	if err := regCcntmrolledID(n, "did:cntm::123", id0, 1, a0.Address); err == nil {
		t.Error("invalid id registered")
	}

	// 6. register the ccntmrolled ID
	if err := regCcntmrolledID(n, id1, id0, 1, a0.Address); err != nil {
		t.Fatal(err)
	}

	// 7. register again, should fail
	if err := regCcntmrolledID(n, id1, id0, 1, a0.Address); err == nil {
		t.Fatal("register twice")
	}

	// 8. verify ccntmroller
	if ok, err := verifyCtrl(n, id1, 1, a0.Address); !ok || err != nil {
		t.Fatal("verify ccntmroller error", err)
	}

	// 9. verify invalid ccntmroller, should fail
	if ok, err := verifyCtrl(n, id1, 2, a0.Address); ok && err == nil {
		t.Error("invalid ccntmroller key index passed verification")
	}

	// 10. verify ccntmroller without valid signature, should fail
	if ok, err := verifyCtrl(n, id1, 1, common.ADDRESS_EMPTY); ok && err == nil {
		t.Error("ccntmroller passed verification without valid signature")
	}

	// 11. add attribute by invalid ccntmroller, should fail
	attr := attribute{
		[]byte("test key"),
		[]byte("test value"),
		[]byte("test type"),
	}
	if err := ctrlAddAttr(n, id1, attr, 1, common.Address{}); err == nil {
		t.Error("attribute added by invalid ccntmroller")
	}

	// 12. add attribute
	if err := ctrlAddAttr(n, id1, attr, 1, a0.Address); err != nil {
		t.Fatal(err)
	}

	// 13. check attribute
	if err := checkAttribute(n, id1, []attribute{attr}); err != nil {
		t.Error("check attribute error", err)
	}

	// 14. remove attribute by invalid ccntmroller, should fail
	if err := ctrlRmAttr(n, id1, attr.key, 1, common.Address{}); err == nil {
		t.Error("attribute removed by invalid ccntmroller")
	}

	// 15. remove nonexistent attribute, should fail
	if err := ctrlRmAttr(n, id1, []byte("unknown key"), 1, a0.Address); err == nil {
		t.Error("removed nonexistent attribute")
	}

	// 16. remove attribute by ccntmroller
	if err := ctrlRmAttr(n, id1, attr.key, 1, a0.Address); err != nil {
		t.Fatal(err)
	}

	// 17. add invalid key, should fail
	if err := ctrlAddKey(n, id1, []byte("test invalid key"), 1, a0.Address); err == nil {
		t.Error("invalid key added by ccntmroller")
	}

	// 18. add key by invalid ccntmroller, should fail
	a1 := account.NewAccount("")
	pk := keypair.SerializePublicKey(a1.PubKey())
	if err := ctrlAddKey(n, id1, pk, 1, common.Address{}); err == nil {
		t.Error("key added by invalid ccntmroller")
	}

	// 19. add key
	if err := ctrlAddKey(n, id1, pk, 1, a0.Address); err != nil {
		t.Fatal(err)
	}

	// 20. remove key by invalid ccntmroller, should fail
	if err := ctrlRmKey(n, id1, 1, 1, common.ADDRESS_EMPTY); err == nil {
		t.Error("key removed by invalid ccntmroller")
	}

	// 21. remove invalid key, should fail
	if err := ctrlRmKey(n, id1, 2, 1, a0.Address); err == nil {
		t.Error("invlid key removed")
	}

	// 22. remove key
	if err := ctrlRmKey(n, id1, 1, 1, a0.Address); err != nil {
		t.Fatal(err)
	}

	// 23. add the removed key again, should fail
	if err := ctrlAddKey(n, id1, pk, 1, a0.Address); err == nil {
		t.Error("removed key added again")
	}

	// 24. add a new key
	a2 := account.NewAccount("")
	pk = keypair.SerializePublicKey(a2.PubKey())
	if err := ctrlAddKey(n, id1, pk, 1, a0.Address); err != nil {
		t.Fatal(err)
	}

	// 25, remove ccntmroller by invalid key, should fail
	if err := rmCtrl(n, id1, 1, a1.Address); err == nil {
		t.Error("ccntmroller removed by invalid key")
	}

	// 26. remove ccntmroller without valid signature, should fail
	if err := rmCtrl(n, id1, 2, common.Address{}); err == nil {
		t.Error("ccntmroller removed without valid signature")
	}

	// 27. remove ccntmoller
	if err := rmCtrl(n, id1, 2, a2.Address); err != nil {
		t.Fatal(err)
	}

	// 28. use removed ccntmroller, should all fail
	if ok, err := verifyCtrl(n, id1, 1, a0.Address); ok && err == nil {
		t.Error("removed ccntmroller passed verification")
	}
	if err := ctrlAddAttr(n, id1, attr, 1, a0.Address); err == nil {
		t.Error("attribute added by removed ccntmroller")
	}
	a3 := account.NewAccount("")
	pk = keypair.SerializePublicKey(a3.PubKey())
	if err := ctrlAddKey(n, id1, pk, 1, a0.Address); err == nil {
		t.Error("key added by removed ccntmroller")
	}
}

func CaseGroupCcntmroller(t *testing.T, n *native.NativeService) {
	id, _ := account.GenerateID()
	id0, _ := account.GenerateID()
	id1, _ := account.GenerateID()
	id2, _ := account.GenerateID()
	id3, _ := account.GenerateID()
	a0 := account.NewAccount("")
	a1 := account.NewAccount("")
	a2 := account.NewAccount("")
	a3 := account.NewAccount("")

	// ccntmroller group
	g := &Group{
		Threshold: 2,
		Members: []interface{}{
			[]byte(id0),
			&Group{
				Threshold: 1,
				Members: []interface{}{
					[]byte(id1),
					[]byte(id2),
				},
			},
		},
	}
	// signers
	signers := []Signer{
		Signer{[]byte(id0), 1},
		Signer{[]byte(id1), 1},
		Signer{[]byte(id2), 1},
	}
	// signed addresses
	addr := []common.Address{a0.Address, a1.Address, a2.Address}

	// 1. register id by unregistered ccntmrollers, should fail
	if err := regGroupCcntmrolledID(n, id, g, signers, addr); err == nil {
		t.Error("ccntmrolled id registered with unregistered ccntmrollers")
	}

	// 2. register ccntmrollers
	if err := regID(n, id0, a0); err != nil {
		t.Fatal("register id0 error")
	}
	if err := regID(n, id1, a1); err != nil {
		t.Fatal("register id1 error")
	}
	if err := regID(n, id2, a2); err != nil {
		t.Fatal("register id2 error")
	}
	if err := regID(n, id3, a3); err != nil {
		t.Fatal("register id3 error")
	}

	// 3. register without valid signature, should fail
	if err := regGroupCcntmrolledID(n, id, g, signers, addr[1:]); err == nil {
		t.Error("registered without valid signatures")
	}

	// 4. register without enough signers, should fail
	if err := regGroupCcntmrolledID(n, id, g, signers[1:], addr[1:]); err == nil {
		t.Error("registered without enough signers")
	}

	// 5. register with invalid signers, should fail
	signers[0].id = []byte(id3)
	addr[0] = a3.Address
	if err := regGroupCcntmrolledID(n, id, g, signers, addr); err == nil {
		t.Error("registered invalid ccntmroller")
	}

	// 5. register ccntmrolled id
	signers[0].id = []byte(id0)
	addr[0] = a0.Address
	if err := regGroupCcntmrolledID(n, id, g, signers, addr); err != nil {
		t.Fatal(err)
	}

	// 6. verify ccntmroller
	if ok, err := verifyGroupCtrl(n, id, signers, addr); !ok || err != nil {
		t.Error("verify group ccntmroller failed", err)
	}

	// 7. verify invalid ccntmroller, should fail
	if ok, err := verifyGroupCtrl(n, id, signers[1:], addr[1:]); ok && err == nil {
		t.Error("invalid group ccntmroller passed verification")
	}

	// 8. revoke id by invalid ccntmroller, should fail
	if err := revokeByCtrl(n, id, signers[1:], addr[1:]); err == nil {
		t.Error("id revoked by invalid ccntmroller")
	}

	// 9. revoke id by ccntmroller
	if err := revokeByCtrl(n, id, signers, addr); err != nil {
		t.Fatal(err)
	}

	// 10. check id state
	enc, _ := encodeID([]byte(id))
	if checkIDState(n, enc) != flag_revoke {
		t.Fatal("id state is not revoked")
	}

	// 11. verify ccntmroller, should fail
	if ok, err := verifyGroupCtrl(n, id, signers, addr); ok && err == nil {
		t.Error("revoked id passed verification")
	}

	// 12. register again, should fail
	if err := regGroupCcntmrolledID(n, id, g, signers, addr); err == nil {
		t.Error("revoked id should not be registered again")
	}
}

// Register id0 which is ccntmrolled by id1
func regCcntmrolledID(n *native.NativeService, id0, id1 string, index uint64, addr common.Address) error {
	// make arguments
	sink := common.NewZeroCopySink(nil)
	sink.WriteVarBytes([]byte(id0))
	sink.WriteVarBytes([]byte(id1))
	utils.EncodeVarUint(sink, index)
	n.Input = sink.Bytes()
	// set signing address
	n.Tx.SignedAddr = []common.Address{addr}
	// call
	_, err := regIdWithCcntmroller(n)
	return err
}

func verifyCtrl(n *native.NativeService, id string, index uint64, addr common.Address) (bool, error) {
	sink := common.NewZeroCopySink(nil)
	sink.WriteString(id)
	utils.EncodeVarUint(sink, index)
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{addr}
	res, err := verifyCcntmroller(n)
	return bytes.Equal(res, utils.BYTE_TRUE), err
}

func ctrlAddAttr(n *native.NativeService, id string, attr attribute, index uint64, addr common.Address) error {
	sink := common.NewZeroCopySink(nil)
	sink.WriteString(id)
	// attribute
	utils.EncodeVarUint(sink, 1)
	attr.Serialization(sink)
	// signer
	utils.EncodeVarUint(sink, index)
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{addr}
	_, err := addAttributesByCcntmroller(n)
	return err
}

func ctrlRmAttr(n *native.NativeService, id string, key []byte, index uint64, addr common.Address) error {
	sink := common.NewZeroCopySink(nil)
	sink.WriteString(id)
	sink.WriteVarBytes(key)
	utils.EncodeVarUint(sink, index)
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{addr}
	_, err := removeAttributeByCcntmroller(n)
	return err
}

func ctrlAddKey(n *native.NativeService, id string, key []byte, index uint64, addr common.Address) error {
	sink := common.NewZeroCopySink(nil)
	sink.WriteString(id)
	// key
	sink.WriteVarBytes(key)
	// signer
	utils.EncodeVarUint(sink, index)
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{addr}
	_, err := addKeyByCcntmroller(n)
	return err
}

func ctrlRmKey(n *native.NativeService, id string, keyIndex, signIndex uint64, addr common.Address) error {
	sink := common.NewZeroCopySink(nil)
	sink.WriteString(id)
	utils.EncodeVarUint(sink, keyIndex)
	utils.EncodeVarUint(sink, signIndex)
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{addr}
	_, err := removeKeyByCcntmroller(n)
	return err
}

func rmCtrl(n *native.NativeService, id string, index uint64, addr common.Address) error {
	sink := common.NewZeroCopySink(nil)
	sink.WriteString(id)
	// signing key index
	utils.EncodeVarUint(sink, index)
	n.Input = sink.Bytes()
	// set signing address
	n.Tx.SignedAddr = []common.Address{addr}
	_, err := removeCcntmroller(n)
	return err
}

func regGroupCcntmrolledID(n *native.NativeService, id string, g *Group, s []Signer, addr []common.Address) error {
	sink := common.NewZeroCopySink(nil)
	sink.WriteString(id)
	sink.WriteVarBytes(g.Serialize())
	sink.WriteVarBytes(SerializeSigners(s))
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = addr
	_, err := regIdWithCcntmroller(n)
	return err
}

func verifyGroupCtrl(n *native.NativeService, id string, s []Signer, addr []common.Address) (bool, error) {
	sink := common.NewZeroCopySink(nil)
	sink.WriteString(id)
	sink.WriteVarBytes(SerializeSigners(s))
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = addr
	res, err := verifyCcntmroller(n)
	return bytes.Equal(res, utils.BYTE_TRUE), err
}
func revokeByCtrl(n *native.NativeService, id string, s []Signer, addr []common.Address) error {
	sink := common.NewZeroCopySink(nil)
	sink.WriteString(id)
	sink.WriteVarBytes(SerializeSigners(s))
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = addr
	_, err := revokeIDByCcntmroller(n)
	return err
}
