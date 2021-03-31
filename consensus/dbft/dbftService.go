package dbft

import (
	"time"
	"sync"
	. "GoOnchain/errors"
	. "GoOnchain/common"
	"errors"
	"GoOnchain/net"
	pl "GoOnchain/net/payload"
	inv "GoOnchain/net/inventory"
	tx "GoOnchain/core/transaction"
	va "GoOnchain/core/validation"
	sig "GoOnchain/core/signature"
	ct "GoOnchain/core/ccntmract"
	_ "GoOnchain/core/signature"
	"GoOnchain/core/ledger"
	"GoOnchain/consensus"
	cl "GoOnchain/client"
	"GoOnchain/events"
)

const TimePerBlock = 15
const SecondsPerBlock = 15

type DbftService struct {
	ccntmext ConsensusCcntmext
	mu           sync.Mutex
	Client *cl.Client
	timer *time.Timer
	timerHeight uint32
	timeView byte
	blockReceivedTime time.Time
	logDictionary string
	started bool
	localNode *net.Node

	newInventorySubscriber events.Subscriber
	blockPersistCompletedSubscriber events.Subscriber
}

func NewDbftService(localNode *net.Node,client *cl.Client,logDictionary string) *DbftService {
	return &DbftService{
		localNode: localNode,
		Client: client,
		timer: time.NewTimer(time.Second*15),
		started: false,
		logDictionary: logDictionary,
	}
}

func (ds *DbftService) AddTransaction(TX *tx.Transaction) error{

	hasTx := ledger.DefaultLedger.Blockchain.CcntmainsTransaction(TX.Hash())
	verifyTx := va.VerifyTransaction(TX,ledger.DefaultLedger,ds.ccntmext.GetTransactionList())
	checkPolicy :=  ds.CheckPolicy(TX)
	if hasTx || (verifyTx != nil) || (checkPolicy != nil) {
		//ADD Log: "reject tx"
		ds.RequestChangeView()
		return errors.New("Transcation is invalid.")
	}

	ds.ccntmext.Transactions[TX.Hash()] = TX
	if len(ds.ccntmext.TransactionHashes) == len(ds.ccntmext.Transactions) {

		//Get Miner list
		txlist := ds.ccntmext.GetTransactionList()
		minerAddress := ledger.GetMinerAddress(ledger.DefaultLedger.Blockchain.GetMinersByTXs(txlist))

		if minerAddress == ds.ccntmext.NextMiner{
			//TODO: add log "send prepare response"
			ds.ccntmext.State |= SignatureSent
			sig.SignBySigner(ds.ccntmext.MakeHeader(),ds.Client.GetAccount(ds.ccntmext.Miners[ds.ccntmext.MinerIndex]))
			ds.SignAndRelay(ds.ccntmext.MakePerpareResponse(ds.ccntmext.Signatures[ds.ccntmext.MinerIndex]))
			ds.CheckSignatures()
		} else {
			ds.RequestChangeView()
			return errors.New("No valid Next Miner.")
		}
	}
	return nil
}

func (ds *DbftService) BlockPersistCompleted(v interface{}){
	ds.blockReceivedTime = time.Now()
	ds.InitializeConsensus(0)
}

func (ds *DbftService) ChangeViewReceived(payload *pl.ConsensusPayload,message *ChangeView){
	//TODO: add log

	if message.NewViewNumber <= ds.ccntmext.ExpectedView[payload.MinerIndex] {
		return
	}

	ds.ccntmext.ExpectedView[payload.MinerIndex] = message.NewViewNumber
	ds.CheckExpectedView(message.NewViewNumber)
}

func (ds *DbftService) CheckExpectedView(viewNumber byte){
	if ds.ccntmext.ViewNumber == viewNumber {
		return
	}

	if len(ds.ccntmext.ExpectedView) >= ds.ccntmext.M(){
		ds.InitializeConsensus(viewNumber)
	}
}

func (ds *DbftService) CheckPolicy(transaction *tx.Transaction) error{
	//TODO: CheckPolicy

	return nil
}

func (ds *DbftService) CheckSignatures() error{

	if ds.ccntmext.GetSignaturesCount() >= ds.ccntmext.M() && ds.ccntmext.CheckTxHashesExist() {
		ccntmract,err := ct.CreateMultiSigCcntmract(ToCodeHash(ds.ccntmext.Miners[ds.ccntmext.MinerIndex].EncodePoint(true)),ds.ccntmext.M(),ds.ccntmext.Miners)
		if err != nil{
			return err
		}

		block := ds.ccntmext.MakeHeader()
		cxt := ct.NewCcntmractCcntmext(block)

		for i,j :=0,0; i < len(ds.ccntmext.Miners) && j < ds.ccntmext.M() ; i++ {
			if ds.ccntmext.Signatures[i] != nil{
				cxt.AddCcntmract(ccntmract,ds.ccntmext.Miners[i],ds.ccntmext.Signatures[i])
				j++
			}
		}

		cxt.Data.SetPrograms(cxt.GetPrograms())
		block.Transcations = ds.ccntmext.GetTXByHashes()

		//TODO: add log "relay block"

		if err := ds.localNode.Relay(block); err != nil{
			//TODO: add log "reject block"
		}

		ds.ccntmext.State |= BlockSent

	}
	return nil
}

func (ds *DbftService) Halt() error  {
	if ds.timer != nil {
		ds.timer.Stop()
	}

	if ds.started {
		ledger.DefaultLedger.Blockchain.BCEvents.UnSubscribe(ledger.EventBlockPersistCompleted,ds.blockPersistCompletedSubscriber)
		ds.localNode.NodeEvent.UnSubscribe(net.EventNewInventory,ds.newInventorySubscriber)
	}
	return nil
}

func (ds *DbftService) InitializeConsensus(viewNum byte) error  {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if viewNum == 0 {
		ds.ccntmext.Reset(ds.Client)
	} else {
		ds.ccntmext.ChangeView(viewNum)
	}

	if ds.ccntmext.MinerIndex < 0 {
		return NewDetailErr(errors.New("Miner Index incorrect"),ErrNoCode,"")
	}

	if ds.ccntmext.MinerIndex == int(ds.ccntmext.PrimaryIndex) {
		ds.ccntmext.State |= Primary
		ds.timerHeight = ds.ccntmext.Height
		ds.timeView = viewNum

		span := time.Now().Sub(ds.blockReceivedTime)

		if span > TimePerBlock {
			ds.Timeout() //TODO: double check timer check
		} else {
			time.AfterFunc(TimePerBlock-span,ds.Timeout)//TODO: double check time usage
		}
	} else {
		ds.ccntmext.State = Backup
		ds.timerHeight = ds.ccntmext.Height
		ds.timeView = viewNum
		//ds.timer.Reset()
	}
	return nil
}

func (ds *DbftService) LocalNodeNewInventory(v interface{}){
	if inventory,ok := v.(inv.Inventory);ok {
		if inventory.InvertoryType() == inv.Consensus {
			payload, isConsensusPayload := inventory.(*pl.ConsensusPayload)
			if isConsensusPayload {
				ds.NewConsensusPayload(payload)
			}
		} else if inventory.InvertoryType() == inv.Transaction  {
			transaction, isTransaction := inventory.(*tx.Transaction)
			if isTransaction{
				ds.NewTransactionPayload(transaction)
			}
		}
	}
}

func (ds *DbftService) NewConsensusPayload(payload *pl.ConsensusPayload){
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if int(payload.MinerIndex) == ds.ccntmext.MinerIndex {return }

	if payload.Version != CcntmextVersion || payload.PrevHash != ds.ccntmext.PrevHash || payload.Height != ds.ccntmext.Height {
		return
	}

	if int(payload.MinerIndex) >= len(ds.ccntmext.Miners) {return }

	message,_ := DeserializeMessage(payload.Data)

	if message.ViewNumber() != ds.ccntmext.ViewNumber && message.Type() != ChangeViewMsg {
		return
	}

	switch message.Type() {
	case ChangeViewMsg:
		if cv, ok := message.(*ChangeView); ok {
			ds.ChangeViewReceived(payload,cv)
		}
		break
	case PrepareRequestMsg:
		if pr, ok := message.(*PrepareRequest); ok {
			ds.PrepareRequestReceived(payload,pr)
		}
		break
	case PrepareResponseMsg:
		if pres, ok := message.(*PrepareResponse); ok {
			ds.PrepareResponseReceived(payload,pres)
		}
		break
	}
}

func (ds *DbftService) NewTransactionPayload(transaction *tx.Transaction) error{
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if ds.ccntmext.State.HasFlag(Backup) || !ds.ccntmext.State.HasFlag(RequestReceived) || ds.ccntmext.State.HasFlag(SignatureSent) {
		return NewDetailErr(errors.New("Consensus State is incorrect."),ErrNoCode,"")
	}

	if _, hasTx := ds.ccntmext.Transactions[transaction.Hash()]; hasTx {
		return NewDetailErr(errors.New("The transaction already exist."),ErrNoCode,"")
	}

	if !ds.ccntmext.HasTxHash(transaction.Hash()) {
		return NewDetailErr(errors.New("The transaction hash is not exist."),ErrNoCode,"")
	}
	return ds.AddTransaction(transaction)
}

func (ds *DbftService) PrepareRequestReceived(payload *pl.ConsensusPayload,message *PrepareRequest) {
	//TODO: add log

	if ds.ccntmext.State.HasFlag(Backup) || ds.ccntmext.State.HasFlag(RequestReceived) {
		return
	}

	if uint32(payload.MinerIndex) != ds.ccntmext.PrimaryIndex {return }

	prevBlockTimestamp := ledger.DefaultLedger.Blockchain.GetHeader(ds.ccntmext.PrevHash).Blockdata.Timestamp
	if payload.Timestamp <= prevBlockTimestamp || payload.Timestamp > uint32(time.Now().Add(time.Minute*10).Unix()){
		//TODO: add log "Timestamp incorrect"
		return
	}

	ds.ccntmext.State |= RequestReceived
	ds.ccntmext.Timestamp = payload.Timestamp
	ds.ccntmext.Nonce = message.Nonce
	ds.ccntmext.NextMiner = message.NextMiner
	ds.ccntmext.TransactionHashes = message.TransactionHashes
	ds.ccntmext.Transactions = make(map[Uint256]*tx.Transaction)

	if err := va.VerifySignature(ds.ccntmext.MakeHeader(),ds.ccntmext.Miners[payload.MinerIndex],message.Signature); err != nil {
		return
	}

	minerLen := len(ds.ccntmext.Miners)
	ds.ccntmext.Signatures = make([][]byte,minerLen)
	ds.ccntmext.Signatures[payload.MinerIndex] = message.Signature

	if err := ds.AddTransaction(message.MinerTransaction); err != nil {return }

	mempool := ds.localNode.GetMemoryPool()
	for _, hash := range ds.ccntmext.TransactionHashes[1:] {
		if transaction,ok := mempool[hash]; ok{
			if err := ds.AddTransaction(transaction); err != nil {
				return
			}
		}
	}

	//TODO: LocalNode allow hashes (add Except method)
	//AllowHashes(ds.ccntmext.TransactionHashes)

	if len(ds.ccntmext.Transactions) < len(ds.ccntmext.TransactionHashes){
		ds.localNode.SynchronizeMemoryPool()
	}
}

func (ds *DbftService) PrepareResponseReceived(payload *pl.ConsensusPayload,message *PrepareResponse){
	//TODO: add log

	if ds.ccntmext.State.HasFlag(BlockSent)  {return}
	if ds.ccntmext.Signatures[payload.MinerIndex] != nil {return }

	header := ds.ccntmext.MakeHeader()
	if  header == nil {return }
	if err := va.VerifySignature(header,ds.ccntmext.Miners[payload.MinerIndex],message.Signature); err != nil {
		return
	}

	ds.ccntmext.Signatures[payload.MinerIndex] = message.Signature
	ds.CheckSignatures()
}

func  (ds *DbftService)  RefreshPolicy(){
	consensus.DefaultPolicy.Refresh()
}

func  (ds *DbftService)  RequestChangeView() {
	ds.ccntmext.ExpectedView[ds.ccntmext.MinerIndex]++
	//TODO: add log request change view

	time.AfterFunc(SecondsPerBlock << (ds.ccntmext.ExpectedView[ds.ccntmext.MinerIndex]+1),ds.Timeout) //TODO: double check timer
	ds.SignAndRelay(ds.ccntmext.MakeChangeView())
	ds.CheckExpectedView(ds.ccntmext.ExpectedView[ds.ccntmext.MinerIndex])
}

func (ds *DbftService) SignAndRelay(payload *pl.ConsensusPayload){

	ctCxt := ct.NewCcntmractCcntmext(payload)

	ds.Client.Sign(ctCxt)
	ctCxt.Data.SetPrograms(ctCxt.GetPrograms())
	ds.localNode.Relay(payload)
}

func (ds *DbftService) Start() error  {

	ds.started = true

	ds.newInventorySubscriber = ledger.DefaultLedger.Blockchain.BCEvents.Subscribe(ledger.EventBlockPersistCompleted,ds.BlockPersistCompleted)
	ds.blockPersistCompletedSubscriber = ds.localNode.NodeEvent.Subscribe(net.EventNewInventory,ds.LocalNodeNewInventory)

	ds.InitializeConsensus(0)
	return nil
}

func (ds *DbftService) Timeout() {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if ds.timerHeight != ds.ccntmext.Height || ds.timeView != ds.ccntmext.ViewNumber {
		return
	}

	//TODO: add log "timeout”

	if ds.ccntmext.State.HasFlag(Primary) && !ds.ccntmext.State.HasFlag(RequestSent) {
		//TODO: add log “send prepare request”

		ds.ccntmext.State |= RequestSent
		if !ds.ccntmext.State.HasFlag(SignatureSent) {

			//set ccntmext Timestamp
			now := uint32(time.Now().Unix())
			blockTime := ledger.DefaultLedger.Blockchain.GetHeader(ds.ccntmext.PrevHash).Blockdata.Timestamp

			if blockTime > now {
				ds.ccntmext.Timestamp = blockTime
			} else {
				ds.ccntmext.Timestamp = now
			}

			ds.ccntmext.Nonce = GetNonce()
			transactions := ds.localNode.GetMemoryPool() //TODO: add policy
			//Insert miner transaction
			if ds.ccntmext.TransactionHashes == nil {
				ds.ccntmext.TransactionHashes = []Uint256{}
			}

			for _,TX := range transactions {
				ds.ccntmext.TransactionHashes = append(ds.ccntmext.TransactionHashes,TX.Hash())
			}
			ds.ccntmext.Transactions = transactions

			txlist := ds.ccntmext.GetTransactionList()
			ds.ccntmext.NextMiner = ledger.GetMinerAddress(ledger.DefaultLedger.Blockchain.GetMinersByTXs(txlist))

			block := ds.ccntmext.MakeHeader()
			account := ds.Client.GetAccount(ds.ccntmext.Miners[ds.ccntmext.MinerIndex])
			ds.ccntmext.Signatures[ds.ccntmext.MinerIndex] = sig.SignBySigner(block,account)
		}
		ds.SignAndRelay(ds.ccntmext.MakePerpareRequest())
		time.AfterFunc(SecondsPerBlock << (ds.timeView + 1), ds.Timeout) //TODO: double check change timer

	} else if ds.ccntmext.State.HasFlag(Primary) && ds.ccntmext.State.HasFlag(RequestSent) || ds.ccntmext.State.HasFlag(Backup){
		ds.RequestChangeView()
	}
}
