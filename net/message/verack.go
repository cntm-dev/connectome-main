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

package message

import (
	"GoOnchain/common"
	"GoOnchain/common/log"
	"GoOnchain/core/ledger"
	. "GoOnchain/net/protocol"
	"encoding/hex"
	"time"
)

type verACK struct {
	msgHdr
}

func NewVerack() ([]byte, error) {
	var msg verACK
	var sum []byte
	sum = []byte{0x5d, 0xf6, 0xe0, 0xe2}
	msg.msgHdr.init("verack", sum, 0)

	buf, err := msg.Serialization()
	if err != nil {
		return nil, err
	}

	str := hex.EncodeToString(buf)
	log.Debug("The message tx verack length is ", len(buf), ", ", str)

	return buf, err
}

/*
 * The node state switch table after rx message, there is time limitation for each action
 * The Hanshake status will switch to INIT after TIMEOUT if not received the VerACK
 * in this time window
 *  _______________________________________________________________________
 * |          |    INIT         | HANDSHAKE |  ESTABLISH | INACTIVITY      |
 * |-----------------------------------------------------------------------|
 * | version  | HANDSHAKE(timer)|           |            | HANDSHAKE(timer)|
 * |          | if helloTime > 3| Tx verack | Depend on  | if helloTime > 3|
 * |          | Tx version      |           | node update| Tx version      |
 * |          | then Tx verack  |           |            | then Tx verack  |
 * |-----------------------------------------------------------------------|
 * | verack   |                 | ESTABLISH |            |                 |
 * |          |   No Action     |           | No Action  | No Action       |
 * |------------------------------------------------------------------------
 *
 */
// TODO The process should be adjusted based on above table
func (msg verACK) Handle(node Noder) error {
	common.Trace()

	t := time.Now()
	// TODO we loading the state&time without consider race case
	s := node.GetState()
	if s == HANDSHAKE {
		node.SetState(ESTABLISH)
	} else {
		log.Error("Unkown status when get the verack")
	}
	// TODO update other node info
	node.UpdateTime(t)
	node.DumpInfo()
	if node.GetState() == ESTABLISH {
		node.ReqNeighborList()

		if uint64(ledger.DefaultLedger.Blockchain.BlockHeight) < node.GetHeight() {
			buf, err := NewHeadersReq(node)
			if err != nil {
				log.Error("failed build a new headersReq")
			} else {
				node.Tx(buf)
			}
		}
	}
	return nil
}
