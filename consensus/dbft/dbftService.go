package dbft

import (
	cl "DNA/client"
	. "DNA/common"
	"DNA/common/log"
	"DNA/config"
	con "DNA/consensus"
	ct "DNA/core/ccntmract"
	"DNA/core/ccntmract/program"
	"DNA/core/ledger"
	_ "DNA/core/signature"
	sig "DNA/core/signature"
	tx "DNA/core/transaction"
	"DNA/core/transaction/payload"
	va "DNA/core/validation"
	. "DNA/errors"
	"DNA/events"
	"DNA/net"
	msg "DNA/net/message"
	"errors"
	"fmt"
	"time"
)

var GenBlockTime = (2 * time.Second)

type DbftService struct {
	ccntmext           ConsensusCcntmext
	Client            cl.Client
	timer             *time.Timer
	timerHeight       uint32
	timeView          byte
	blockReceivedTime time.Time
	logDictionary     string
	started           bool
	localNet          net.Neter

	newInventorySubscriber          events.Subscriber
	blockPersistCompletedSubscriber events.Subscriber
}

func NewDbftService(client cl.Client, logDictionary string, localNet net.Neter) *DbftService {
	Trace()

	ds := &DbftService{
		Client:        client,
		timer:         time.NewTimer(time.Second * 15),
		started:       false,
		localNet:      localNet,
		logDictionary: logDictionary,
	}

	if !ds.timer.Stop() {
		<-ds.timer.C
	}
	Trace()
	go ds.timerRoutine()
	return ds
}

func (ds *DbftService) AddTransaction(TX *tx.Transaction, needVerify bool) error {
	Trace()

	//check whether the new TX already exist in ledger
	if ledger.DefaultLedger.Blockchain.CcntmainsTransaction(TX.Hash()) {
		log.Warn(fmt.Sprintf("[AddTransaction] TX already Exist: %v", TX.Hash()))
		ds.RequestChangeView()
		return errors.New("TX already Exist.")
	}

	//verify the TX
	if needVerify {
		err := va.VerifyTransaction(TX, ledger.DefaultLedger, ds.ccntmext.GetTransactionList())
		if err != nil {
			log.Warn(fmt.Sprintf("[AddTransaction] TX Verfiy failed: %v", TX.Hash()))
			ds.RequestChangeView()
			return errors.New("TX Verfiy failed.")
		}
	}

	//check the TX policy
	//checkPolicy :=  ds.CheckPolicy(TX)

	//set TX to current ccntmext
	ds.ccntmext.Transactions[TX.Hash()] = TX

	//if enough TXs already added to ccntmext, build block and sign/relay
	if len(ds.ccntmext.TransactionHashes) == len(ds.ccntmext.Transactions) {

		minerAddress, err := ledger.GetMinerAddress(ds.ccntmext.Miners)
		if err != nil {
			return NewDetailErr(err, ErrNoCode, "[DbftService] ,GetMinerAddress failed")
		}

		if minerAddress == ds.ccntmext.NextMiner {
			log.Info("send prepare response")
			ds.ccntmext.State |= SignatureSent
			miner, err := ds.Client.GetAccount(ds.ccntmext.Miners[ds.ccntmext.MinerIndex])
			if err != nil {
				return NewDetailErr(err, ErrNoCode, "[DbftService] ,GetAccount failed.")
			}
			//sig.SignBySigner(ds.ccntmext.MakeHeader(), miner)
			ds.ccntmext.Signatures[ds.ccntmext.MinerIndex], err = sig.SignBySigner(ds.ccntmext.MakeHeader(), miner)
			if err != nil {
				log.Error("[DbftService], SignBySigner failed.")
				return NewDetailErr(err, ErrNoCode, "[DbftService], SignBySigner failed.")
			}
			payload := ds.ccntmext.MakePrepareResponse(ds.ccntmext.Signatures[ds.ccntmext.MinerIndex])
			ds.SignAndRelay(payload)
			err = ds.CheckSignatures()
			if err != nil {
				return NewDetailErr(err, ErrNoCode, "[DbftService] ,CheckSignatures failed.")
			}
		} else {
			ds.RequestChangeView()
			return errors.New("No valid Next Miner.")

		}
	}
	return nil
}

func (ds *DbftService) BlockPersistCompleted(v interface{}) {
	Trace()
	if block, ok := v.(*ledger.Block); ok {
		log.Info(fmt.Sprintf("persist block: %d", block.Hash()))
	}

	ds.blockReceivedTime = time.Now()

	go ds.InitializeConsensus(0)
}

func (ds *DbftService) CheckExpectedView(viewNumber byte) {
	Trace()
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

func (ds *DbftService) CheckPolicy(transaction *tx.Transaction) error {
	//TODO: CheckPolicy

	return nil
}

func (ds *DbftService) CheckSignatures() error {
	Trace()

	//check have enought signatures and all required TXs already in ccntmext
	if ds.ccntmext.GetSignaturesCount() >= ds.ccntmext.M() && ds.ccntmext.CheckTxHashesExist() {

		//get current index's hash
		ep, err := ds.ccntmext.Miners[ds.ccntmext.MinerIndex].EncodePoint(true)
		if err != nil {
			return NewDetailErr(err, ErrNoCode, "[DbftService] ,EncodePoint failed")
		}
		codehash, err := ToCodeHash(ep)
		if err != nil {
			return NewDetailErr(err, ErrNoCode, "[DbftService] ,ToCodeHash failed")
		}

		//create multi-sig ccntmract with all miners
		ccntmract, err := ct.CreateMultiSigCcntmract(codehash, ds.ccntmext.M(), ds.ccntmext.Miners)
		if err != nil {
			return err
		}

		//build block
		block := ds.ccntmext.MakeHeader()

		//sign the block with all miners and add signed ccntmract to ccntmext
		cxt := ct.NewCcntmractCcntmext(block)
		for i, j := 0, 0; i < len(ds.ccntmext.Miners) && j < ds.ccntmext.M(); i++ {
			if ds.ccntmext.Signatures[i] != nil {
				err := cxt.AddCcntmract(ccntmract, ds.ccntmext.Miners[i], ds.ccntmext.Signatures[i])
				if err != nil {
					log.Error("[CheckSignatures] Multi-sign add ccntmract error:", err.Error())
					return NewDetailErr(err, ErrNoCode, "[DbftService], CheckSignatures AddCcntmract failed.")
				}
				j++
			}
		}
		//set signed program to the block
		cxt.Data.SetPrograms(cxt.GetPrograms())

		block.Transcations = ds.ccntmext.GetTXByHashes()

		if err := ds.localNet.Xmit(block); err != nil {
			log.Info(fmt.Sprintf("[CheckSignatures] Xmit block Error: %s, blockHash: %d", err.Error(), block.Hash()))
		}
		ds.ccntmext.State |= BlockSent
	}
	return nil
}

func (ds *DbftService) CreateBookkeepingTransaction(nonce uint64) *tx.Transaction {
	Trace()

	//TODO: sysfee

	return &tx.Transaction{
		TxType:         tx.BookKeeping,
		PayloadVersion: 0x2,
		Payload:        &payload.MinerPayload{},
		Nonce:          nonce, //TODO: update the nonce
		Attributes:     []*tx.TxAttribute{},
		UTXOInputs:     []*tx.UTXOTxInput{},
		BalanceInputs:  []*tx.BalanceTxInput{},
		Outputs:        []*tx.TxOutput{},
		Programs:       []*program.Program{},
	}
}

func (ds *DbftService) ChangeViewReceived(payload *msg.ConsensusPayload, message *ChangeView) {
	Trace()
	log.Info(fmt.Sprintf("Change View Received: height=%d View=%d index=%d nv=%d", payload.Height, message.ViewNumber(), payload.MinerIndex, message.NewViewNumber))

	if message.NewViewNumber <= ds.ccntmext.ExpectedView[payload.MinerIndex] {
		return
	}

	ds.ccntmext.ExpectedView[payload.MinerIndex] = message.NewViewNumber

	ds.CheckExpectedView(message.NewViewNumber)
}

func (ds *DbftService) Halt() error {
	Trace()
	log.Info("DBFT Stop")
	if ds.timer != nil {
		ds.timer.Stop()
	}

	if ds.started {
		ledger.DefaultLedger.Blockchain.BCEvents.UnSubscribe(events.EventBlockPersistCompleted, ds.blockPersistCompletedSubscriber)
		ds.localNet.GetEvent("consensus").UnSubscribe(events.EventNewInventory, ds.newInventorySubscriber)
	}
	return nil
}

func (ds *DbftService) InitializeConsensus(viewNum byte) error {
	log.Debug("[InitializeConsensus] Start InitializeConsensus.")
	Trace()
	ds.ccntmext.ccntmextMu.Lock()
	defer ds.ccntmext.ccntmextMu.Unlock()

	log.Debug("[InitializeConsensus] viewNum: ",viewNum)

	if viewNum == 0 {
		ds.ccntmext.Reset(ds.Client)
	} else {
		ds.ccntmext.ChangeView(viewNum)
	}

	if ds.ccntmext.MinerIndex < 0 {
		log.Error("Miner Index incorrect ", ds.ccntmext.MinerIndex)
		return NewDetailErr(errors.New("Miner Index incorrect"), ErrNoCode, "")
	}

	if ds.ccntmext.MinerIndex == int(ds.ccntmext.PrimaryIndex) {

		//primary peer
		Trace()
		ds.ccntmext.State |= Primary
		ds.timerHeight = ds.ccntmext.Height
		ds.timeView = viewNum
		span := time.Now().Sub(ds.blockReceivedTime)
		if span > GenBlockTime {
			//TODO: double check the is the stop necessary
			ds.timer.Stop()
			ds.timer.Reset(0)
			//go ds.Timeout()
		} else {
			ds.timer.Stop()
			ds.timer.Reset(GenBlockTime - span)
		}
	} else {

		//backup peer
		ds.ccntmext.State = Backup
		ds.timerHeight = ds.ccntmext.Height
		ds.timeView = viewNum

		ds.timer.Stop()
		ds.timer.Reset(GenBlockTime << (viewNum + 1))
	}
	return nil
}

func (ds *DbftService) LocalNodeNewInventory(v interface{}) {
	Trace()
	if inventory, ok := v.(Inventory); ok {
		if inventory.Type() == CONSENSUS {
			payload, ret := inventory.(*msg.ConsensusPayload)
			if ret == true {
				ds.NewConsensusPayload(payload)
			}
		} else if inventory.Type() == TRANSACTION {
			transaction, isTransaction := inventory.(*tx.Transaction)
			if isTransaction {
				ds.NewTransactionPayload(transaction)
			}
		}
	}
}

//TODO: add invenory receiving

func (ds *DbftService) NewConsensusPayload(payload *msg.ConsensusPayload) {
	Trace()
	ds.ccntmext.ccntmextMu.Lock()
	defer ds.ccntmext.ccntmextMu.Unlock()

	//if payload from current peer, ignore it
	if int(payload.MinerIndex) == ds.ccntmext.MinerIndex {
		return
	}

	//if payload is not same height with current ccntmex, ignore it
	if payload.Version != CcntmextVersion || payload.PrevHash != ds.ccntmext.PrevHash || payload.Height != ds.ccntmext.Height {
		return
	}

	if int(payload.MinerIndex) >= len(ds.ccntmext.Miners) {
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
	}
}

func (ds *DbftService) NewTransactionPayload(transaction *tx.Transaction) error {
	Trace()
	ds.ccntmext.ccntmextMu.Lock()
	defer ds.ccntmext.ccntmextMu.Unlock()

	if !ds.ccntmext.State.HasFlag(Backup) || !ds.ccntmext.State.HasFlag(RequestReceived) || ds.ccntmext.State.HasFlag(SignatureSent) {
		return NewDetailErr(errors.New("Consensus State is incorrect."), ErrNoCode, "")
	}

	if _, hasTx := ds.ccntmext.Transactions[transaction.Hash()]; hasTx {
		return NewDetailErr(errors.New("The transaction already exist."), ErrNoCode, "")
	}

	if !ds.ccntmext.HasTxHash(transaction.Hash()) {
		return NewDetailErr(errors.New("The transaction hash is not exist."), ErrNoCode, "")
	}
	return ds.AddTransaction(transaction, true)
}

func (ds *DbftService) PrepareRequestReceived(payload *msg.ConsensusPayload, message *PrepareRequest) {
	Trace()
	log.Info(fmt.Sprintf("Prepare Request Received: height=%d View=%d index=%d tx=%d", payload.Height, message.ViewNumber(), payload.MinerIndex, len(message.TransactionHashes)))

	if !ds.ccntmext.State.HasFlag(Backup) || ds.ccntmext.State.HasFlag(RequestReceived) {
		return
	}

	if uint32(payload.MinerIndex) != ds.ccntmext.PrimaryIndex {
		return
	}
	header, err := ledger.DefaultLedger.Blockchain.GetHeader(ds.ccntmext.PrevHash)
	if err != nil {
		log.Info("PrepareRequestReceived GetHeader failed with ds.ccntmext.PrevHash", ds.ccntmext.PrevHash)
	}

	Trace()
	//TODO Add Error Catch
	prevBlockTimestamp := header.Blockdata.Timestamp
	if payload.Timestamp <= prevBlockTimestamp || payload.Timestamp > uint32(time.Now().Add(time.Minute*10).Unix()) {
		log.Info(fmt.Sprintf("Prepare Reques tReceived: Timestamp incorrect: %d", payload.Timestamp))
		return
	}

	ds.ccntmext.State |= RequestReceived
	ds.ccntmext.Timestamp = payload.Timestamp
	ds.ccntmext.Nonce = message.Nonce
	ds.ccntmext.NextMiner = message.NextMiner
	ds.ccntmext.TransactionHashes = message.TransactionHashes
	ds.ccntmext.Transactions = make(map[Uint256]*tx.Transaction)

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
	Trace()

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
	if _, err := va.VerifySignature(header, ds.ccntmext.Miners[payload.MinerIndex], message.Signature); err != nil {
		return
	}

	ds.ccntmext.Signatures[payload.MinerIndex] = message.Signature
	ds.CheckSignatures()
	log.Info("Prepare Response finished")
}

func (ds *DbftService) RefreshPolicy() {
	Trace()
	con.DefaultPolicy.Refresh()
}

func (ds *DbftService) RequestChangeView() {
	Trace()
	// FIXME if there is no save block notifcation, when the timeout call this function it will crash
	ds.ccntmext.ExpectedView[ds.ccntmext.MinerIndex] = ds.ccntmext.ExpectedView[ds.ccntmext.MinerIndex] + 1
	log.Info(fmt.Sprintf("Request change view: height=%d View=%d nv=%d state=%s", ds.ccntmext.Height, ds.ccntmext.ViewNumber, ds.ccntmext.ExpectedView[ds.ccntmext.MinerIndex], ds.ccntmext.GetStateDetail()))

	ds.timer.Stop()
	ds.timer.Reset(GenBlockTime << (ds.ccntmext.ExpectedView[ds.ccntmext.MinerIndex] + 1))

	ds.SignAndRelay(ds.ccntmext.MakeChangeView())
	ds.CheckExpectedView(ds.ccntmext.ExpectedView[ds.ccntmext.MinerIndex])
}

func (ds *DbftService) SignAndRelay(payload *msg.ConsensusPayload) {
	Trace()

	prohash, err := payload.GetProgramHashes()
	if err != nil {
		log.Debug("[SignAndRelay] payload.GetProgramHashes failed: ", err.Error())
		return
	}
	log.Debug("[SignAndRelay] ConsensusPayload Program Hashes: ", prohash)

	ctCxt := ct.NewCcntmractCcntmext(payload)

	ret := ds.Client.Sign(ctCxt)
	if ret == false {
		log.Warn("[SignAndRelay] Sign ccntmract failure")
	}
	prog := ctCxt.GetPrograms()
	if prog == nil {
		log.Warn("[SignAndRelay] Get programe failure")
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
	Trace()
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
		ds.timer.Reset(GenBlockTime << (ds.timeView + 1))
	} else if (ds.ccntmext.State.HasFlag(Primary) && ds.ccntmext.State.HasFlag(RequestSent)) || ds.ccntmext.State.HasFlag(Backup) {
		ds.RequestChangeView()
	}


}

func (ds *DbftService) timerRoutine() {
	Trace()
	for {
		select {
		case <-ds.timer.C:
			log.Debug("******Get a timeout notice")
			go ds.Timeout()
		}
	}
}
