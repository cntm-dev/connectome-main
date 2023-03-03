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

package Cbft

import (
	"fmt"
	"testing"

	"github.com/conntectome/cntm/account"
)

func constructMsg() *blockProposalMsg {
	acc := account.NewAccount("SHA256withECDSA")
	if acc == nil {
		fmt.Println("GetDefaultAccount error: acc is nil")
		return nil
	}
	msg := constructProposalMsgTest(acc)
	return msg
}
func TestSerializeCbftMsg(t *testing.T) {
	msg := constructMsg()
	_, err := SerializeCbftMsg(msg)
	if err != nil {
		t.Errorf("TestSerializeCbftMsg failed :%v", err)
		return
	}
	t.Logf("TestSerializeCbftMsg succ")
}

func TestDeserializeCbftMsg(t *testing.T) {
	msg := constructMsg()
	data, err := SerializeCbftMsg(msg)
	if err != nil {
		t.Errorf("TestSerializeCbftMsg failed :%v", err)
		return
	}
	_, err = DeserializeCbftMsg(data)
	if err != nil {
		t.Errorf("DeserializeCbftMsg failed :%v", err)
		return
	}
	t.Logf("TestDeserializeCbftMsg succ")
}
