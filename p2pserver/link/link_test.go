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

package link

import (
	"math/rand"
	"testing"
	"time"

	"github.com/cntmio/cntmology-crypto/keypair"
	"github.com/cntmio/cntmology/account"
	comm "github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/log"
	ct "github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/p2pserver/common"
	mt "github.com/cntmio/cntmology/p2pserver/message/types"
)

var (
	cliLink    *Link
	serverLink *Link
	cliChan    chan *mt.MsgPayload
	serverChan chan *mt.MsgPayload
	cliAddr    string
	serAddr    string
)

func init() {
	log.InitLog(log.InfoLog, log.Stdout)

	cliLink = NewLink()
	serverLink = NewLink()
	id := common.PseudoPeerIdFromUint64(0x733936)
	cliLink.SetID(id)
	id2 := common.PseudoPeerIdFromUint64(0x8274950)
	serverLink.SetID(id2)

	cliChan = make(chan *mt.MsgPayload, 100)
	serverChan = make(chan *mt.MsgPayload, 100)
	//listen ip addr
	cliAddr = "127.0.0.1:50338"
	serAddr = "127.0.0.1:50339"
}

func TestNewLink(t *testing.T) {
	id := 0x74936295

	if cliLink.GetID().ToUint64() != 0x733936 {
		t.Fatal("link GetID failed")
	}
	i := common.PseudoPeerIdFromUint64(uint64(id))
	cliLink.SetID(i)
	if cliLink.GetID().ToUint64() != uint64(id) {
		t.Fatal("link SetID failed")
	}

	cliLink.SetChan(cliChan)
	serverLink.SetChan(serverChan)

	cliLink.UpdateRXTime(time.Now())

	msg := &mt.MsgPayload{
		Id:      cliLink.GetID(),
		Addr:    cliLink.GetAddr(),
		Payload: &mt.NotFound{comm.UINT256_EMPTY},
	}
	go func() {
		time.Sleep(5000000)
		cliChan <- msg
	}()

	timeout := time.NewTimer(time.Second)
	select {
	case <-cliChan:
		t.Log("read data from channel")
	case <-timeout.C:
		timeout.Stop()
		t.Fatal("can`t read data from link channel")
	}

}

func TestUnpackBufNode(t *testing.T) {
	cliLink.SetChan(cliChan)

	msgType := "block"

	var msg mt.Message

	switch msgType {
	case "addr":
		var newaddrs []common.PeerAddr
		for i := 0; i < 10000000; i++ {
			idd := common.PseudoPeerIdFromUint64(uint64(i))
			newaddrs = append(newaddrs, common.PeerAddr{
				Time: time.Now().Unix(),
				ID:   idd,
			})
		}
		var addr mt.Addr
		addr.NodeAddrs = newaddrs
		msg = &addr
	case "consensuspayload":
		acct := account.NewAccount("SHA256withECDSA")
		key := acct.PubKey()
		payload := mt.ConsensusPayload{
			Owner: key,
		}
		for i := 0; uint32(i) < 200000000; i++ {
			byteInt := rand.Intn(256)
			payload.Data = append(payload.Data, byte(byteInt))
		}

		msg = &mt.Consensus{payload}
	case "consensus":
		acct := account.NewAccount("SHA256withECDSA")
		key := acct.PubKey()
		payload := &mt.ConsensusPayload{
			Owner: key,
		}
		for i := 0; uint32(i) < 200000000; i++ {
			byteInt := rand.Intn(256)
			payload.Data = append(payload.Data, byte(byteInt))
		}
		consensus := mt.Consensus{
			Cons: *payload,
		}
		msg = &consensus
	case "blkheader":
		var headers []*ct.Header
		blkHeader := &mt.BlkHeader{}
		for i := 0; uint32(i) < 100000000; i++ {
			header := &ct.Header{}
			header.Height = uint32(i)
			header.Bookkeepers = make([]keypair.PublicKey, 0)
			header.SigData = make([][]byte, 0)
			headers = append(headers, header)
		}
		blkHeader.BlkHdr = headers
		msg = blkHeader
	case "tx":
		trn := &mt.Trn{}
		rawTXBytes, _ := comm.HexToBytes("00d1af758596f401000000000000204e000000000000b09ba6a4fe99eb2b2dc1d86a6d453423a6be03f02e0101011552c1126765744469736b506c61796572734c697374676a6f1082c6cec3a1bcbb5a3892cf770061e4b98200014241015d434467639fd8e7b4331d2f3fc0d4168e2d68a203593c6399f5746d2324217aeeb3db8ff31ba0fdb1b13aa6f4c3cd25f7b3d0d26c144bbd75e2963d0a443629232103fdcae8110c9a60d1fc47f8111a12c1941e1f3584b0b0028157736ed1eecd101eac")
		tx, _ := ct.TransactionFromRawBytes(rawTXBytes)
		trn.Txn = tx
		msg = trn
	case "block":
		var blk ct.Block
		mBlk := &mt.Block{}
		var txs []*ct.Transaction
		header := ct.Header{}
		header.Height = uint32(1)
		header.Bookkeepers = make([]keypair.PublicKey, 0)
		header.SigData = make([][]byte, 0)
		blk.Header = &header

		rawTXBytes, _ := comm.HexToBytes("00d1af758596f401000000000000204e000000000000b09ba6a4fe99eb2b2dc1d86a6d453423a6be03f02e0101011552c1126765744469736b506c61796572734c697374676a6f1082c6cec3a1bcbb5a3892cf770061e4b98200014241015d434467639fd8e7b4331d2f3fc0d4168e2d68a203593c6399f5746d2324217aeeb3db8ff31ba0fdb1b13aa6f4c3cd25f7b3d0d26c144bbd75e2963d0a443629232103fdcae8110c9a60d1fc47f8111a12c1941e1f3584b0b0028157736ed1eecd101eac")
		tx, _ := ct.TransactionFromRawBytes(rawTXBytes)
		txs = append(txs, tx)

		blk.Transactions = txs
		mBlk.Blk = &blk

		msg = mBlk
	}

	sink := comm.NewZeroCopySink(nil)
	mt.WriteMessage(sink, msg)
}
