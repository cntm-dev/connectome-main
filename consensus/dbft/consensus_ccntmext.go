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

package dbft

import (
	"fmt"
	"github.com/Ontology/account"
	. "github.com/Ontology/common"
	"github.com/Ontology/common/log"
	ser "github.com/Ontology/common/serialization"
	"github.com/Ontology/core/ledger"
	"github.com/Ontology/core/types"
	"github.com/Ontology/core/vote"
	"github.com/Ontology/crypto"
	msg "github.com/Ontology/net/message"
)

const CcntmextVersion uint32 = 0

type ConsensusCcntmext struct {
	State           ConsensusState
	PrevHash        Uint256
	Height          uint32
	ViewNumber      byte
	Bookkeepers     []*crypto.PubKey
	NextBookkeepers []*crypto.PubKey
	Owner           *crypto.PubKey
	BookkeeperIndex int
	PrimaryIndex    uint32
	Timestamp       uint32
	Nonce           uint64
	NextBookkeeper  Address
	Transactions    []*types.Transaction
	Signatures      [][]byte
	ExpectedView    []byte

	header *types.Block

	isBookkeeperChanged bool
	nmChangedblkHeight  uint32
}

func (cxt *ConsensusCcntmext) M() int {
	log.Debug()
	return len(cxt.Bookkeepers) - (len(cxt.Bookkeepers)-1)/3
}

func NewConsensusCcntmext() *ConsensusCcntmext {
	log.Debug()
	return &ConsensusCcntmext{}
}

func (cxt *ConsensusCcntmext) ChangeView(viewNum byte) {
	log.Debug()
	p := (cxt.Height - uint32(viewNum)) % uint32(len(cxt.Bookkeepers))
	cxt.State &= SignatureSent
	cxt.ViewNumber = viewNum
	if p >= 0 {
		cxt.PrimaryIndex = uint32(p)
	} else {
		cxt.PrimaryIndex = uint32(p) + uint32(len(cxt.Bookkeepers))
	}

	if cxt.State == Initial {
		cxt.Transactions = nil
		cxt.Signatures = make([][]byte, len(cxt.Bookkeepers))
		cxt.header = nil
	}
}

func (cxt *ConsensusCcntmext) MakeChangeView() *msg.ConsensusPayload {
	log.Debug()
	cv := &ChangeView{
		NewViewNumber: cxt.ExpectedView[cxt.BookkeeperIndex],
	}
	cv.msgData.Type = ChangeViewMsg
	return cxt.MakePayload(cv)
}

func (cxt *ConsensusCcntmext) MakeHeader() *types.Block {
	log.Debug()
	if cxt.Transactions == nil {
		return nil
	}
	if cxt.header == nil {
		txHash := []Uint256{}
		for _, t := range cxt.Transactions {
			txHash = append(txHash, t.Hash())
		}
		txRoot, err := crypto.ComputeRoot(txHash)
		if err != nil {
			return nil
		}
		blockRoot := ledger.DefLedger.GetBlockRootWithNewTxRoot(txRoot)
		header := &types.Header{
			Version:          CcntmextVersion,
			PrevBlockHash:    cxt.PrevHash,
			TransactionsRoot: txRoot,
			BlockRoot:        blockRoot,
			Timestamp:        cxt.Timestamp,
			Height:           cxt.Height,
			ConsensusData:    cxt.Nonce,
			NextBookkeeper:   cxt.NextBookkeeper,
		}
		cxt.header = &types.Block{
			Header:       header,
			Transactions: []*types.Transaction{},
		}
	}
	return cxt.header
}

func (cxt *ConsensusCcntmext) MakePayload(message ConsensusMessage) *msg.ConsensusPayload {
	log.Debug()
	message.ConsensusMessageData().ViewNumber = cxt.ViewNumber
	return &msg.ConsensusPayload{
		Version:         CcntmextVersion,
		PrevHash:        cxt.PrevHash,
		Height:          cxt.Height,
		BookkeeperIndex: uint16(cxt.BookkeeperIndex),
		Timestamp:       cxt.Timestamp,
		Data:            ser.ToArray(message),
		Owner:           cxt.Owner,
	}
}

func (cxt *ConsensusCcntmext) MakePrepareRequest() *msg.ConsensusPayload {
	log.Debug()
	preReq := &PrepareRequest{
		Nonce:          cxt.Nonce,
		NextBookkeeper: cxt.NextBookkeeper,
		Transactions:   cxt.Transactions,
		Signature:      cxt.Signatures[cxt.BookkeeperIndex],
	}
	preReq.msgData.Type = PrepareRequestMsg
	return cxt.MakePayload(preReq)
}

func (cxt *ConsensusCcntmext) MakePrepareResponse(signature []byte) *msg.ConsensusPayload {
	log.Debug()
	preRes := &PrepareResponse{
		Signature: signature,
	}
	preRes.msgData.Type = PrepareResponseMsg
	return cxt.MakePayload(preRes)
}

func (cxt *ConsensusCcntmext) MakeBlockSignatures(signatures []SignaturesData) *msg.ConsensusPayload {
	log.Debug()
	sigs := &BlockSignatures{
		Signatures: signatures,
	}
	sigs.msgData.Type = BlockSignaturesMsg
	return cxt.MakePayload(sigs)
}

func (cxt *ConsensusCcntmext) GetSignaturesCount() (count int) {
	log.Debug()
	count = 0
	for _, sig := range cxt.Signatures {
		if sig != nil {
			count += 1
		}
	}
	return count
}

func (cxt *ConsensusCcntmext) GetStateDetail() string {

	return fmt.Sprintf("Initial: %t, Primary: %t, Backup: %t, RequestSent: %t, RequestReceived: %t, SignatureSent: %t, BlockGenerated: %t, ",
		cxt.State.HasFlag(Initial),
		cxt.State.HasFlag(Primary),
		cxt.State.HasFlag(Backup),
		cxt.State.HasFlag(RequestSent),
		cxt.State.HasFlag(RequestReceived),
		cxt.State.HasFlag(SignatureSent),
		cxt.State.HasFlag(BlockGenerated))

}

func (cxt *ConsensusCcntmext) Reset(bkAccount *account.Account) {
	preHash := ledger.DefLedger.GetCurrentBlockHash()
	height := ledger.DefLedger.GetCurrentBlockHeight()
	header := cxt.MakeHeader()

	if height != cxt.Height || header == nil || header.Hash() != preHash || len(cxt.NextBookkeepers) == 0 {
		log.Info("[ConsensusCcntmext] Calculate Bookkeepers from db")
		var err error
		cxt.Bookkeepers, err = vote.GetValidators([]*types.Transaction{})
		if err != nil {
			log.Error("[ConsensusCcntmext] GetNextBookkeeper failed", err)
		}
	} else {
		cxt.Bookkeepers = cxt.NextBookkeepers
	}

	cxt.State = Initial
	cxt.PrevHash = preHash
	cxt.Height = height + 1
	cxt.ViewNumber = 0
	cxt.BookkeeperIndex = -1
	cxt.NextBookkeepers = nil
	bookkeeperLen := len(cxt.Bookkeepers)
	cxt.PrimaryIndex = cxt.Height % uint32(bookkeeperLen)
	cxt.Transactions = nil
	cxt.header = nil
	cxt.Signatures = make([][]byte, bookkeeperLen)
	cxt.ExpectedView = make([]byte, bookkeeperLen)

	for i := 0; i < bookkeeperLen; i++ {
		if bkAccount.PublicKey.X.Cmp(cxt.Bookkeepers[i].X) == 0 {
			cxt.BookkeeperIndex = i
			cxt.Owner = cxt.Bookkeepers[i]
			break
		}
	}

}
