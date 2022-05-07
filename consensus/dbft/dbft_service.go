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
	"bytes"
	"fmt"
	"reflect"
	"time"

	"github.com/cntmio/cntmology-eventbus/actor"
	"github.com/cntmio/cntmology/account"
	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/config"
	"github.com/cntmio/cntmology/common/log"
	actorTypes "github.com/cntmio/cntmology/consensus/actor"
	"github.com/cntmio/cntmology/core/genesis"
	"github.com/cntmio/cntmology/core/ledger"
	"github.com/cntmio/cntmology/core/signature"
	"github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/core/vote"
	"github.com/cntmio/cntmology/events"
	"github.com/cntmio/cntmology/events/message"
	p2pmsg "github.com/cntmio/cntmology/p2pserver/message/types"
	"github.com/cntmio/cntmology/validator/increment"
)

type DbftService struct {
	ccntmext           ConsensusCcntmext
	Account           *account.Account
	timer             *time.Timer
	timerHeight       uint32
	timeView          byte
	blockReceivedTime time.Time
	started           bool
	ledger            *ledger.Ledger
	incrValidator     *increment.IncrementValidator
	poolActor         *actorTypes.TxPoolActor
	p2p               *actorTypes.P2PActor

	pid *actor.PID
	sub *events.ActorSubscriber
}

func NewDbftService(bkAccount *account.Account, txpool, p2p *actor.PID) (*DbftService, error) {

	service := &DbftService{
		Account:       bkAccount,
		timer:         time.NewTimer(time.Second * 15),
		started:       false,
		ledger:        ledger.DefLedger,
		incrValidator: increment.NewIncrementValidator(10),
		poolActor:     &actorTypes.TxPoolActor{Pool: txpool},
		p2p:           &actorTypes.P2PActor{P2P: p2p},
	}

	if !service.timer.Stop() {
		<-service.timer.C
	}

	go func() {
		for {
			select {
			case <-service.timer.C:
				log.Debug("******Get a timeout notice")
				service.pid.Tell(&actorTypes.TimeOut{})
			}
		}
	}()

	props := actor.FromProducer(func() actor.Actor {
		return service
	})

	pid, err := actor.SpawnNamed(props, "consensus_dbft")
	service.pid = pid

	service.sub = events.NewActorSubscriber(pid)
	return service, err
}

func (this *DbftService) Receive(ccntmext actor.Ccntmext) {
	if _, ok := ccntmext.Message().(*actorTypes.StartConsensus); this.started == false && ok == false {
		return
	}

	switch msg := ccntmext.Message().(type) {
	case *actor.Restarting:
		log.Warn("dbft actor restarting")
	case *actor.Stopping:
		log.Warn("dbft actor stopping")
	case *actor.Stopped:
		log.Warn("dbft actor stopped")
	case *actor.Started:
		log.Warn("dbft actor started")
	case *actor.Restart:
		log.Warn("dbft actor restart")
	case *actorTypes.StartConsensus:
		this.start()
	case *actorTypes.StopConsensus:
		this.incrValidator.Clean()
		this.halt()
	case *actorTypes.TimeOut:
		log.Info("dbft receive timeout")
		this.Timeout()
	case *message.SaveBlockCompleteMsg:
		log.Infof("dbft actor receives block complete event. block height=%d, numtx=%d",
			msg.Block.Header.Height, len(msg.Block.Transactions))
		this.incrValidator.AddBlock(msg.Block)
		this.handleBlockPersistCompleted(msg.Block)
	case *p2pmsg.ConsensusPayload:
		this.NewConsensusPayload(msg)

	default:
		log.Info("dbft actor: Unknown msg ", msg, "type", reflect.TypeOf(msg))
	}
}

func (this *DbftService) GetPID() *actor.PID {
	return this.pid
}
func (this *DbftService) Start() error {
	this.pid.Tell(&actorTypes.StartConsensus{})
	return nil
}

func (this *DbftService) Halt() error {
	this.pid.Tell(&actorTypes.StopConsensus{})
	return nil
}

func (self *DbftService) handleBlockPersistCompleted(block *types.Block) {
	log.Infof("persist block: %x", block.Hash())
	self.p2p.Broadcast(block.Hash())

	self.blockReceivedTime = time.Now()

	self.InitializeConsensus(0)
}

func (ds *DbftService) BlockPersistCompleted(v interface{}) {
	if block, ok := v.(*types.Block); ok {
		log.Infof("persist block: %x", block.Hash())

		ds.p2p.Broadcast(block.Hash())
	}

}

func (ds *DbftService) CheckExpectedView(viewNumber byte) {
	log.Debug()
	if ds.ccntmext.State.HasFlag(BlockGenerated) {
		return
	}
	if ds.ccntmext.ViewNumber == viewNumber {
		return
	}

	//check the count for same view number
	count := 0
	for i, expectedViewNumber := range ds.ccntmext.ExpectedView {
		log.Debug(fmt.Sprintf("[CheckExpectedView] ExpectedView #%d: %d",i,expectedViewNumber))
		if expectedViewNumber == viewNumber {
			count++
		}
	}

	log.Debug("[CheckExpectedView] same view number count: ",count)
	log.Debug("[CheckExpectedView] ds.ccntmext.M(): ",ds.ccntmext.M())

	M := ds.ccntmext.M()
	if count >= M {

		log.Debug("[CheckExpectedView] Begin InitializeConsensus.")
		go ds.InitializeConsensus(viewNumber)
		//ds.InitializeConsensus(viewNumber)
	}
}

func (ds *DbftService) CheckPolicy(transaction *types.Transaction) error {
	//TODO: CheckPolicy

	return nil
}

func (ds *DbftService) CheckSignatures() error {
	log.Debug()

	//check if get enough signatures
	if ds.ccntmext.GetSignaturesCount() >= ds.ccntmext.M() {
		//build block
		block := ds.ccntmext.MakeHeader()
		sigs := make([]SignaturesData, ds.ccntmext.M())
		for i, j := 0, 0; i < len(ds.ccntmext.Bookkeepers) && j < ds.ccntmext.M(); i++ {
			if ds.ccntmext.Signatures[i] != nil {
				sig := ds.ccntmext.Signatures[i]
				sigs[j].Index = uint16(i)
				sigs[j].Signature = sig

				block.Header.SigData = append(block.Header.SigData, sig)
				j++
			}
		}

		block.Header.Bookkeepers = ds.ccntmext.Bookkeepers

		//fill transactions
		block.Transactions = ds.ccntmext.Transactions

		hash := block.Hash()
		isExist, err := ds.ledger.IsCcntmainBlock(hash)
		if err != nil {
			log.Errorf("DefLedger.IsCcntmainBlock Hash:%x error:%s", hash, err)
			return err
		}
		if !isExist {
			// save block
			err := ds.ledger.AddBlock(block)
			if err != nil {
				return fmt.Errorf("CheckSignatures DefLedgerPid.RequestFuture Height:%d error:%s", block.Header.Height, err)
			}

			ds.ccntmext.State |= BlockGenerated
			payload := ds.ccntmext.MakeBlockSignatures(sigs)
			ds.SignAndRelay(payload)
		}
	}
	return nil
}

func (ds *DbftService) ChangeViewReceived(payload *p2pmsg.ConsensusPayload, message *ChangeView) {
	log.Debug()
	log.Info(fmt.Sprintf("Change View Received: height=%d View=%d index=%d nv=%d", payload.Height, message.ViewNumber(), payload.BookkeeperIndex, message.NewViewNumber))

	if message.NewViewNumber <= ds.ccntmext.ExpectedView[payload.BookkeeperIndex] {
		return
	}

	ds.ccntmext.ExpectedView[payload.BookkeeperIndex] = message.NewViewNumber

	ds.CheckExpectedView(message.NewViewNumber)
}

func (ds *DbftService) halt() error {
	log.Info("DBFT Stop")
	if ds.timer != nil {
		ds.timer.Stop()
	}

	if ds.started {
		ds.sub.Unsubscribe(message.TOPIC_SAVE_BLOCK_COMPLETE)
	}
	return nil
}

func (ds *DbftService) InitializeConsensus(viewNum byte) error {
	log.Debug("[InitializeConsensus] Start InitializeConsensus.")
	ds.ccntmext.ccntmextMu.Lock()
	defer ds.ccntmext.ccntmextMu.Unlock()

	log.Debug("[InitializeConsensus] viewNum: ",viewNum)

	if viewNum == 0 {
		ds.ccntmext.Reset(ds.Client)
	} else {
		ds.ccntmext.ChangeView(viewNum)
	}

	if ds.ccntmext.BookkeeperIndex < 0 {
		log.Info("You aren't bookkeeper")
		return nil
	}

	if ds.ccntmext.BookkeeperIndex == int(ds.ccntmext.PrimaryIndex) {

		//primary peer
		ds.ccntmext.State |= Primary
		ds.timerHeight = ds.ccntmext.Height
		ds.timeView = viewNum
		span := time.Now().Sub(ds.blockReceivedTime)
		if span > genesis.GenBlockTime {
			//TODO: double check the is the stop necessary
			ds.timer.Stop()
			ds.timer.Reset(0)
			//go ds.Timeout()
		} else {
			ds.timer.Stop()
			ds.timer.Reset(genesis.GenBlockTime - span)
		}
	} else {

		//backup peer
		ds.ccntmext.State = Backup
		ds.timerHeight = ds.ccntmext.Height
		ds.timeView = viewNum

		ds.timer.Stop()
		ds.timer.Reset(genesis.GenBlockTime << (viewNum + 1))
	}
	return nil
}

func (ds *DbftService) LocalNodeNewInventory(v interface{}) {
	log.Debug()
	if inventory, ok := v.(common.Inventory); ok {
		if inventory.Type() == common.CONSENSUS {
			payload, ret := inventory.(*p2pmsg.ConsensusPayload)
			if ret == true {
				ds.NewConsensusPayload(payload)
			}
		}
	}
}

func (ds *DbftService) NewConsensusPayload(payload *p2pmsg.ConsensusPayload) {
	//if payload from current peer, ignore it
	if int(payload.BookkeeperIndex) == ds.ccntmext.BookkeeperIndex {
		return
	}

	//if payload is not same height with current ccntmex, ignore it
	if payload.Version != CcntmextVersion || payload.PrevHash != ds.ccntmext.PrevHash || payload.Height != ds.ccntmext.Height {
		log.Debug("unmatched height")
		return
	}

	if ds.ccntmext.State.HasFlag(BlockGenerated) {
		log.Debug("has flag 'BlockGenerated'")
		return
	}

	if int(payload.BookkeeperIndex) >= len(ds.ccntmext.Bookkeepers) {
		log.Debug("bookkeeper index out of range")
		return
	}

	message, err := DeserializeMessage(payload.Data)
	if err != nil {
		log.Error(fmt.Sprintf("DeserializeMessage failed: %s\n", err))
		return
	}

	if message.ViewNumber() != ds.ccntmext.ViewNumber && message.Type() != ChangeViewMsg {
		return
	}

	err = payload.Verify()
	if err != nil {
		log.Warn(err.Error())
		return
	}

	switch message.Type() {
	case ChangeViewMsg:
		if cv, ok := message.(*ChangeView); ok {
			ds.ChangeViewReceived(payload, cv)
		}
		break
	case PrepareRequestMsg:
		if pr, ok := message.(*PrepareRequest); ok {
			ds.PrepareRequestReceived(payload, pr)
		}
		break
	case PrepareResponseMsg:
		if pres, ok := message.(*PrepareResponse); ok {
			ds.PrepareResponseReceived(payload, pres)
		}
		break
	case BlockSignaturesMsg:
		if blockSigs, ok := message.(*BlockSignatures); ok {
			ds.BlockSignaturesReceived(payload, blockSigs)
		}
		break
	default:
		log.Warn("unknown consensus message type")
	}
}

func (ds *DbftService) PrepareRequestReceived(payload *p2pmsg.ConsensusPayload, message *PrepareRequest) {
	log.Info(fmt.Sprintf("Prepare Request Received: height=%d View=%d index=%d tx=%d", payload.Height, message.ViewNumber(), payload.BookkeeperIndex, len(message.Transactions)))

	if !ds.ccntmext.State.HasFlag(Backup) || ds.ccntmext.State.HasFlag(RequestReceived) {
		return
	}

	if uint32(payload.BookkeeperIndex) != ds.ccntmext.PrimaryIndex {
		return
	}

	header, err := ds.ledger.GetHeaderByHash(ds.ccntmext.PrevHash)
	if err != nil {
		log.Errorf("PrepareRequestReceived GetHeader failed with ds.ccntmext.PrevHash:%x", ds.ccntmext.PrevHash)
		return
	}
	if header == nil {
		log.Errorf("PrepareRequestReceived cannot GetHeaderByHash by PrevHash:%x", ds.ccntmext.PrevHash)
		return
	}

	//TODO Add Error Catch
	prevBlockTimestamp := header.Timestamp
	if payload.Timestamp <= prevBlockTimestamp || payload.Timestamp > uint32(time.Now().Add(time.Minute*10).Unix()) {
		log.Info(fmt.Sprintf("Prepare Reques tReceived: Timestamp incorrect: %d", payload.Timestamp))
		return
	}

	//ds.ccntmext copy
	ds.ccntmext.State |= RequestReceived
	ds.ccntmext.Timestamp = payload.Timestamp
	ds.ccntmext.Nonce = message.Nonce
	ds.ccntmext.NextMiner = message.NextMiner
	ds.ccntmext.TransactionHashes = message.TransactionHashes
	ds.ccntmext.Transactions = make(map[Uint256]*tx.Transaction)

	//block header verification
	_, err = va.VerifySignature(ds.ccntmext.MakeHeader(), ds.ccntmext.Miners[payload.MinerIndex], message.Signature)
	if err != nil {
		log.Warn("PrepareRequestReceived VerifySignature failed.", err)
		return
	}

	ds.ccntmext.Signatures = make([][]byte, len(ds.ccntmext.Miners))
	ds.ccntmext.Signatures[payload.MinerIndex] = message.Signature

	mempool := ds.localNet.GetMemoryPool()
	for _, hash := range ds.ccntmext.TransactionHashes[1:] {
		if transaction, ok := mempool[hash]; ok {
			if err := ds.AddTransaction(transaction, false); err != nil {
				log.Info("PrepareRequestReceived AddTransaction failed.")
				return
			}
		}
	}

	if err := ds.AddTransaction(message.BookkeepingTransaction, true); err != nil {
		log.Warn("PrepareRequestReceived AddTransaction failed", err)
		return
	}

	//TODO: LocalNode allow hashes (add Except method)
	//AllowHashes(ds.ccntmext.TransactionHashes)
	log.Info("Prepare Requst finished")
	if len(ds.ccntmext.Transactions) < len(ds.ccntmext.TransactionHashes) {
		ds.localNet.SynchronizeMemoryPool()
	}
}

func (ds *DbftService) PrepareResponseReceived(payload *msg.ConsensusPayload, message *PrepareResponse) {
	log.Debug()

	log.Info(fmt.Sprintf("Prepare Response Received: height=%d View=%d index=%d", payload.Height, message.ViewNumber(), payload.MinerIndex))

	if ds.ccntmext.State.HasFlag(BlockSent) {
		return
	}

	//if the signature already exist, needn't handle again
	if ds.ccntmext.Signatures[payload.MinerIndex] != nil {
		return
	}

	header := ds.ccntmext.MakeHeader()
	if header == nil {
		return
	}
	blockHash := header.Hash()
	err := signature.Verify(ds.ccntmext.Bookkeepers[payload.BookkeeperIndex], blockHash[:], message.Signature)
	if err != nil {
		return
	}

	ds.ccntmext.Signatures[payload.BookkeeperIndex] = message.Signature
	err = ds.CheckSignatures()
	if err != nil {
		log.Error("CheckSignatures failed", err)
		return
	}
	log.Info("Prepare Response finished")
}

func (ds *DbftService) BlockSignaturesReceived(payload *p2pmsg.ConsensusPayload, message *BlockSignatures) {
	log.Info(fmt.Sprintf("BlockSignatures Received: height=%d View=%d index=%d", payload.Height, message.ViewNumber(), payload.BookkeeperIndex))

	if ds.ccntmext.State.HasFlag(BlockGenerated) {
		return
	}

	//if the signature already exist, needn't handle again
	if ds.ccntmext.Signatures[payload.BookkeeperIndex] != nil {
		return
	}

	header := ds.ccntmext.MakeHeader()
	if header == nil {
		return
	}

	blockHash := header.Hash()

	for i := 0; i < len(message.Signatures); i++ {
		sigdata := message.Signatures[i]

		if ds.ccntmext.Signatures[sigdata.Index] != nil {
			ccntminue
		}

		err := signature.Verify(ds.ccntmext.Bookkeepers[sigdata.Index], blockHash[:], sigdata.Signature)
		if err != nil {
			ccntminue
		}

		ds.ccntmext.Signatures[sigdata.Index] = sigdata.Signature
		if ds.ccntmext.GetSignaturesCount() >= ds.ccntmext.M() {
			log.Info("BlockSignatures got enough signatures")
			break
		}
	}

	err := ds.CheckSignatures()
	if err != nil {
		log.Error("CheckSignatures failed")
		return
	}
	log.Info("BlockSignatures finished")
}

func (ds *DbftService) RefreshPolicy() {
}

func (ds *DbftService) RequestChangeView() {
	if ds.ccntmext.State.HasFlag(BlockGenerated) {
		return
	}
	// FIXME if there is no save block notifcation, when the timeout call this function it will crash
	if ds.ccntmext.ViewNumber > ds.ccntmext.ExpectedView[ds.ccntmext.BookkeeperIndex] {
		ds.ccntmext.ExpectedView[ds.ccntmext.BookkeeperIndex] = ds.ccntmext.ViewNumber + 1
	} else {
		ds.ccntmext.ExpectedView[ds.ccntmext.BookkeeperIndex] += 1
	}
	log.Info(fmt.Sprintf("Request change view: height=%d View=%d nv=%d state=%s", ds.ccntmext.Height,
		ds.ccntmext.ViewNumber, ds.ccntmext.ExpectedView[ds.ccntmext.BookkeeperIndex], ds.ccntmext.GetStateDetail()))

	ds.timer.Stop()
	ds.timer.Reset(genesis.GenBlockTime << (ds.ccntmext.ExpectedView[ds.ccntmext.BookkeeperIndex] + 1))

	ds.SignAndRelay(ds.ccntmext.MakeChangeView())
	ds.CheckExpectedView(ds.ccntmext.ExpectedView[ds.ccntmext.BookkeeperIndex])
}

func (ds *DbftService) SignAndRelay(payload *p2pmsg.ConsensusPayload) {
	log.Debug()

	prohash, err := payload.GetProgramHashes()
	if err != nil {
		log.Debug("[SignAndRelay] payload.GetProgramHashes failed: ", err.Error())
		return
	}
	log.Debug("[SignAndRelay] ConsensusPayload Program Hashes: ", prohash)

	ccntmext := ccntmract.NewCcntmractCcntmext(payload)

	if prohash[0] != ds.Account.ProgramHash {
		log.Error("[SignAndRelay] wrcntm program hash")
	}

	sig, _ := signature.SignBySigner(ccntmext.Data, ds.Account)
	ct, _ := ccntmract.CreateSignatureCcntmract(ds.Account.PublicKey)
	ccntmext.AddCcntmract(ct, ds.Account.PublicKey, sig)

	prog := ccntmext.GetPrograms()
	if prog == nil {
		log.Warn("[SignAndRelay] Get program failure")
	}
	payload.SetPrograms(prog)
	ds.localNet.Xmit(payload)
}

func (ds *DbftService) Start() error {
	Trace()
	ds.started = true

	ds.newInventorySubscriber = ledger.DefaultLedger.Blockchain.BCEvents.Subscribe(events.EventBlockPersistCompleted, ds.BlockPersistCompleted)
	ds.blockPersistCompletedSubscriber = ds.localNet.GetEvent("consensus").Subscribe(events.EventNewInventory, ds.LocalNodeNewInventory)

	go ds.InitializeConsensus(0)
	//ds.InitializeConsensus(0)
	return nil
}

func (ds *DbftService) Timeout() {
	if ds.timerHeight != ds.ccntmext.Height || ds.timeView != ds.ccntmext.ViewNumber {
		return
	}

	log.Info("Timeout: height: ", ds.timerHeight, " View: ", ds.timeView, " State: ", ds.ccntmext.GetStateDetail())

	////temp change view number test
	//if ledger.DefaultLedger.Blockchain.BlockHeight > 2 {
	//	ds.RequestChangeView()
	//	return
	//}

	if (ds.ccntmext.State.HasFlag(Primary) && !ds.ccntmext.State.HasFlag(RequestSent)) {

		//parimary peer send the prepare request
		log.Info("Send prepare request: height: ", ds.timerHeight, " View: ", ds.timeView, " State: ", ds.ccntmext.GetStateDetail())
		ds.ccntmext.State |= RequestSent
		if !ds.ccntmext.State.HasFlag(SignatureSent) {

			//do signature
			now := uint32(time.Now().Unix())
			header, _ := ledger.DefaultLedger.Blockchain.GetHeader(ds.ccntmext.PrevHash)

			//set ccntmext Timestamp
			blockTime := header.Blockdata.Timestamp + 1
			if blockTime > now {
				ds.ccntmext.Timestamp = blockTime
			} else {
				ds.ccntmext.Timestamp = now
			}

			ds.ccntmext.Nonce = GetNonce()
			transactionsPool := ds.localNet.GetMemoryPool() //TODO: add policy

			//TODO: add max TX limitation

			//convert txPool to tx list
			transactions := []*tx.Transaction{}

			//add new book keeping TX first
			txBookkeeping := ds.CreateBookkeepingTransaction(ds.ccntmext.Nonce)
			transactions = append(transactions, txBookkeeping)

			//add TXs from mem pool
			for _, tx := range transactionsPool {
				transactions = append(transactions, tx)
			}

			//add Transaction hashes
			trxhashes := []Uint256{}
			txMap := make(map[Uint256]*tx.Transaction)
			for _, tx := range transactions {
				txHash := tx.Hash()
				trxhashes = append(trxhashes, txHash)
				txMap[txHash] = tx
			}

			ds.ccntmext.TransactionHashes = trxhashes
			ds.ccntmext.Transactions = txMap

			//build block and sign
			ds.ccntmext.NextMiner, _ = ledger.GetMinerAddress(ds.ccntmext.Miners)
			block := ds.ccntmext.MakeHeader()
			account, _ := ds.Client.GetAccount(ds.ccntmext.Miners[ds.ccntmext.MinerIndex]) //TODO: handle error
			ds.ccntmext.Signatures[ds.ccntmext.MinerIndex], _ = sig.SignBySigner(block, account)

		}
		payload := ds.ccntmext.MakePrepareRequest()
		ds.SignAndRelay(payload)
		ds.timer.Stop()
		ds.timer.Reset(ledger.GenBlockTime << (ds.timeView + 1))
	} else if (ds.ccntmext.State.HasFlag(Primary) && ds.ccntmext.State.HasFlag(RequestSent)) || ds.ccntmext.State.HasFlag(Backup) {
		ds.RequestChangeView()
	}


}

func (ds *DbftService) timerRoutine() {
	log.Debug()
	for {
		select {
		case <-ds.timer.C:
			log.Debug("******Get a timeout notice")
			go ds.Timeout()
		}
	}
}
