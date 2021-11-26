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

package signtest

import (
	"github.com/Ontology/eventbus/actor"
	"fmt"
	"github.com/Ontology/crypto"
)

type SignActor struct{
	PrivateKey []byte

}

func (s *SignActor) Receive(ccntmext actor.Ccntmext) {
	switch msg := ccntmext.Message().(type) {
	case *actor.Started:
		fmt.Println("Started, initialize actor here")
	case *actor.Stopping:
		fmt.Println("Stopping, actor is about shut down")
	case *actor.Restarting:
		fmt.Println("Restarting, actor is about restart")

	case *SetPrivKey:
		//fmt.Println(ccntmext.Self().Id," set Privkey")
		s.PrivateKey = msg.PrivKey

	case *SignRequest:
		//fmt.Println(ccntmext.Self().Id," is signing")
		signature,_:=crypto.Sign(s.PrivateKey, msg.Data)
		response := &SignResponse{Signature:signature,Seq:msg.Seq}
		//fmt.Println(ccntmext.Self().Id," done signing")
		ccntmext.Sender().Request(response,ccntmext.Self())

	default:
		//fmt.Println("unknown message")
	}
}