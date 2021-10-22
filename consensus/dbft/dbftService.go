package dbft

import (
	"bytes"
	"fmt"
	"time"

	. "github.com/Ontology/common"
	"github.com/Ontology/common/config"
	"github.com/Ontology/common/log"
	"github.com/Ontology/core"
	"github.com/Ontology/core/ccntmract"
	"github.com/Ontology/core/ccntmract/program"
	"github.com/Ontology/core/genesis"
	"github.com/Ontology/core/ledger"
	"github.com/Ontology/core/payload"
	"github.com/Ontology/core/signature"
	"github.com/Ontology/core/transaction/utxo"
	"github.com/Ontology/core/types"
	"github.com/Ontology/core/vote"
	"github.com/Ontology/crypto"
	"github.com/Ontology/events"
	"github.com/Ontology/net"
	msg "github.com/Ontology/net/message"
	"github.com/Ontology/core/ledger/ledgerevent"
	"github.com/Ontology/account"
	clientActor "github.com/Ontology/consensus/actor"
	cntmErrors "github.com/Ontology/errors"
	"github.com/Ontology/eventbus/actor"
)

type DbftService struct {
	ccntmext           ConsensusCcntmext
	Account           *account.Account
	timer             *time.Timer
	timerHeight       uint32
	timeView          byte
	blockReceivedTime time.Time
	logDictionary     string
	started           bool
	localNet          net.Neter
	poolActor         *clientActor.TxPoolActor

	newInventorySubscriber          events.Subscriber
	blockPersistCompletedSubscriber events.Subscriber
}

func NewDbftService(bkAccount *account.Account, logDictionary string, txpool *actor.PID) *DbftService {

	ds := &DbftService{
		Account:       bkAccount,
		timer:         time.NewTimer(time.Second * 15),
		started:       false,
		poolActor:     &clientActor.TxPoolActor{Pool: txpool},
		logDictionary: logDictionary,
	}

	if !ds.timer.Stop() {
		<-ds.timer.C
	}
	go ds.timerRoutine()
	return ds
}

func (ds *DbftService) BlockPersistCompleted(v interface{}) {
	if block, ok := v.(*types.Block); ok {
		log.Infof("persist block: %x", block.Hash())

		ds.localNet.Xmit(block.Hash())
	}

	ds.blockReceivedTime = time.Now()

	go ds.InitializeConsensus(0)
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

		//get current index's hash
		ep, err := ds.ccntmext.BookKeepers[ds.ccntmext.BookKeeperIndex].EncodePoint(true)
		if err != nil {
			return cntmErrors.NewDetailErr(err, cntmErrors.ErrNoCode, "[DbftService] ,EncodePoint failed")
		}
		codehash := ToCodeHash(ep)

		//create multi-sig ccntmract with all bookKeepers
		ct, err := ccntmract.CreateMultiSigCcntmract(codehash, ds.ccntmext.M(), ds.ccntmext.BookKeepers)
		if err != nil {
			log.Error("CheckSignatures CreateMultiSigCcntmract error: ", err)
			return err
		}

		//build block
		block := ds.ccntmext.MakeHeader()
		//sign the block with all bookKeepers and add signed ccntmract to ccntmext
		sb := program.NewProgramBuilder()

		sigs := make([]SignaturesData, ds.ccntmext.M())
		for i, j := 0, 0; i < len(ds.ccntmext.BookKeepers) && j < ds.ccntmext.M(); i++ {
			if ds.ccntmext.Signatures[i] != nil {
				sigs[j].Index = uint16(i)
				sigs[j].Signature = ds.ccntmext.Signatures[i]

				sb.PushData(ds.ccntmext.Signatures[i])
				j++
			}
		}
		//set signed program to the block
		block.Header.Program = &program.Program{
			Code:      ct.Code,
			Parameter: sb.ToArray(),
		}
		//fill transactions
		block.Transactions = ds.ccntmext.Transactions

		hash := block.Hash()
		isExist, err := ledger.DefLedger.IsCcntmainBlock(&hash)
		if err != nil {
			log.Errorf("DefLedger.IsCcntmainBlock Hash:%x error:%s", hash, err)
			return err
		}
		if !isExist {
			// save block
			if err := ledger.DefLedger.AddBlock(block); err != nil {
				log.Error(fmt.Sprintf("[CheckSignatures] Xmit block Error: %s, blockHash: %d", err.Error(), block.Hash()))
				return cntmErrors.NewDetailErr(err, cntmErrors.ErrNoCode, "[DbftService], CheckSignatures AddCcntmract failed.")
			}

			ds.ccntmext.State |= BlockGenerated
			payload := ds.ccntmext.MakeBlockSignatures(sigs)
			ds.SignAndRelay(payload)
		}
	}
	return nil
}

func (ds *DbftService) CreateBookkeepingTransaction(nonce uint64, fee Fixed64) *types.Transaction {
	log.Debug()
	//TODO: sysfee
	bookKeepingPayload := &payload.BookKeeping{
		Nonce: uint64(time.Now().UnixNano()),
	}
	signatureRedeemScript, err := ccntmract.CreateSignatureRedeemScript(ds.ccntmext.Owner)
	if err != nil {
		return nil
	}
	signatureRedeemScriptHashToCodeHash := ToCodeHash(signatureRedeemScript)
	if err != nil {
		return nil
	}
	outputs := []*utxo.TxOutput{}
	if fee > 0 {
		feeOutput := &utxo.TxOutput{
			AssetID:     genesis.cntmTokenID,
			Value:       fee,
			ProgramHash: signatureRedeemScriptHashToCodeHash,
		}
		outputs = append(outputs, feeOutput)
	}
	return &types.Transaction{
		TxType: types.BookKeeping,
		//PayloadVersion: payload.BookKeepingPayloadVersion,
		Payload:    bookKeepingPayload,
		Attributes: []*types.TxAttribute{},
	}
}

func (ds *DbftService) ChangeViewReceived(payload *msg.ConsensusPayload, message *ChangeView) {
	log.Debug()
	log.Info(fmt.Sprintf("Change View Received: height=%d View=%d index=%d nv=%d", payload.Height, message.ViewNumber(), payload.BookKeeperIndex, message.NewViewNumber))

	if message.NewViewNumber <= ds.ccntmext.ExpectedView[payload.BookKeeperIndex] {
		return
	}

	ds.ccntmext.ExpectedView[payload.BookKeeperIndex] = message.NewViewNumber

	ds.CheckExpectedView(message.NewViewNumber)
}

func (ds *DbftService) Halt() error {
	log.Debug()
	log.Info("DBFT Stop")
	if ds.timer != nil {
		ds.timer.Stop()
	}

	if ds.started {
		ledgerevent.DefLedgerEvt.UnSubscribe(events.EventBlockPersistCompleted, ds.blockPersistCompletedSubscriber)
		ds.localNet.GetEvent("consensus").UnSubscribe(events.EventNewInventory, ds.newInventorySubscriber)
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

	if ds.ccntmext.BookKeeperIndex < 0 {
		log.Info("You aren't bookkeeper")
		return nil
	}

	if ds.ccntmext.BookKeeperIndex == int(ds.ccntmext.PrimaryIndex) {

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
	if inventory, ok := v.(Inventory); ok {
		if inventory.Type() == CONSENSUS {
			payload, ret := inventory.(*msg.ConsensusPayload)
			if ret == true {
				ds.NewConsensusPayload(payload)
			}
		}
	}
}

//TODO: add invenory receiving

func (ds *DbftService) NewConsensusPayload(payload *msg.ConsensusPayload) {
	log.Debug()
	ds.ccntmext.ccntmextMu.Lock()
	defer ds.ccntmext.ccntmextMu.Unlock()

	//if payload from current peer, ignore it
	if int(payload.BookKeeperIndex) == ds.ccntmext.BookKeeperIndex {
		return
	}

	//if payload is not same height with current ccntmex, ignore it
	if payload.Version != CcntmextVersion || payload.PrevHash != ds.ccntmext.PrevHash || payload.Height != ds.ccntmext.Height {
		return
	}

	if ds.ccntmext.State.HasFlag(BlockGenerated) {
		return
	}

	if ds.ccntmext.State.HasFlag(BlockGenerated) {
		return
	}

	if int(payload.BookKeeperIndex) >= len(ds.ccntmext.BookKeepers) {
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
	}
}

func (ds *DbftService) PrepareRequestReceived(payload *msg.ConsensusPayload, message *PrepareRequest) {
	log.Info(fmt.Sprintf("Prepare Request Received: height=%d View=%d index=%d tx=%d", payload.Height, message.ViewNumber(), payload.BookKeeperIndex, len(message.Transactions)))

	if !ds.ccntmext.State.HasFlag(Backup) || ds.ccntmext.State.HasFlag(RequestReceived) {
		return
	}

	if uint32(payload.BookKeeperIndex) != ds.ccntmext.PrimaryIndex {
		return
	}

	header, err := ledger.DefLedger.GetHeaderByHash(&ds.ccntmext.PrevHash)
	if err != nil {
		log.Info("PrepareRequestReceived GetHeader failed with ds.ccntmext.PrevHash", ds.ccntmext.PrevHash)
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
	buf := new(bytes.Buffer)
	header.SerializeUnsigned(buf)
	err := crypto.Verify(*ds.ccntmext.BookKeepers[payload.BookKeeperIndex], buf.Bytes(), message.Signature)
	if err != nil {
		return
	}

	ds.ccntmext.Signatures[payload.BookKeeperIndex] = message.Signature
	err = ds.CheckSignatures()
	if err != nil {
		log.Error("CheckSignatures failed")
		return
	}
	log.Info("Prepare Response finished")
}

func (ds *DbftService) BlockSignaturesReceived(payload *msg.ConsensusPayload, message *BlockSignatures) {
	log.Info(fmt.Sprintf("BlockSignatures Received: height=%d View=%d index=%d", payload.Height, message.ViewNumber(), payload.BookKeeperIndex))

	if ds.ccntmext.State.HasFlag(BlockGenerated) {
		return
	}

	//if the signature already exist, needn't handle again
	if ds.ccntmext.Signatures[payload.BookKeeperIndex] != nil {
		return
	}

	header := ds.ccntmext.MakeHeader()
	if header == nil {
		return
	}

	buf := new(bytes.Buffer)
	header.SerializeUnsigned(buf)

	for i := 0; i < len(message.Signatures); i++ {
		sigdata := message.Signatures[i]

		if ds.ccntmext.Signatures[sigdata.Index] != nil {
			ccntminue
		}

		err := crypto.Verify(*ds.ccntmext.BookKeepers[sigdata.Index], buf.Bytes(), sigdata.Signature)
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
	log.Debug()
	//con.DefaultPolicy.Refresh()
}

func (ds *DbftService) RequestChangeView() {
	if ds.ccntmext.State.HasFlag(BlockGenerated) {
		return
	}
	// FIXME if there is no save block notifcation, when the timeout call this function it will crash
	if ds.ccntmext.ViewNumber > ds.ccntmext.ExpectedView[ds.ccntmext.BookKeeperIndex] {
		ds.ccntmext.ExpectedView[ds.ccntmext.BookKeeperIndex] = ds.ccntmext.ViewNumber + 1
	} else {
		ds.ccntmext.ExpectedView[ds.ccntmext.BookKeeperIndex] += 1
	}
	log.Info(fmt.Sprintf("Request change view: height=%d View=%d nv=%d state=%s", ds.ccntmext.Height,
		ds.ccntmext.ViewNumber, ds.ccntmext.ExpectedView[ds.ccntmext.BookKeeperIndex], ds.ccntmext.GetStateDetail()))

	ds.timer.Stop()
	ds.timer.Reset(genesis.GenBlockTime << (ds.ccntmext.ExpectedView[ds.ccntmext.BookKeeperIndex] + 1))

	ds.SignAndRelay(ds.ccntmext.MakeChangeView())
	ds.CheckExpectedView(ds.ccntmext.ExpectedView[ds.ccntmext.BookKeeperIndex])
}

func (ds *DbftService) SignAndRelay(payload *msg.ConsensusPayload) {
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
	log.Debug()
	ds.ccntmext.ccntmextMu.Lock()
	defer ds.ccntmext.ccntmextMu.Unlock()
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
