package dbft

import (
	"time"
	"sync"
	. "GoOnchain/errors"
	. "GoOnchain/common"
	"errors"
	"GoOnchain/net"
	msg "GoOnchain/net/message"
	tx "GoOnchain/core/transaction"
	va "GoOnchain/core/validation"
	sig "GoOnchain/core/signature"
	ct "GoOnchain/core/ccntmract"
	_ "GoOnchain/core/signature"
	"GoOnchain/core/ledger"
	con "GoOnchain/consensus"
	cl "GoOnchain/client"
	"GoOnchain/events"
	"fmt"
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
	localNet net.Neter

	newInventorySubscriber events.Subscriber
	blockPersistCompletedSubscriber events.Subscriber
}

func NewDbftService(client *cl.Client,logDictionary string,localNet net.Neter) *DbftService {
	Trace()
	return &DbftService{
		//localNode: localNode,
		Client: client,
		timer: time.NewTimer(time.Second*15),
		started: false,
		localNet:localNet,
		logDictionary: logDictionary,
	}
}

func (ds *DbftService) AddTransaction(TX *tx.Transaction) error{
	Trace()

	hasTx := ledger.DefaultLedger.Blockchain.CcntmainsTransaction(TX.Hash())
	verifyTx := va.VerifyTransaction(TX,ledger.DefaultLedger,ds.ccntmext.GetTransactionList())
	checkPolicy :=  ds.CheckPolicy(TX)
	if hasTx || (verifyTx != nil) || (checkPolicy != nil) {

		con.Log(fmt.Sprintf("reject tx: %s",TX.Hash()))
		ds.RequestChangeView()
		return errors.New("Transcation is invalid.")
	}

	ds.ccntmext.Transactions[TX.Hash()] = TX
	if len(ds.ccntmext.TransactionHashes) == len(ds.ccntmext.Transactions) {

		//Get Miner list
		txlist := ds.ccntmext.GetTransactionList()
		minerAddress,err := ledger.GetMinerAddress(ledger.DefaultLedger.Blockchain.GetMinersByTXs(txlist))
		if err != nil {
			return NewDetailErr(err,ErrNoCode,"[DbftService] ,GetMinerAddress failed")
		}

		if minerAddress == ds.ccntmext.NextMiner{
			con.Log("send perpare response")
			ds.ccntmext.State |= SignatureSent
			miner,err:=ds.Client.GetAccount(ds.ccntmext.Miners[ds.ccntmext.MinerIndex])
			if err != nil {
				return NewDetailErr(err,ErrNoCode,"[DbftService] ,GetAccount failed.")
			}
			sig.SignBySigner(ds.ccntmext.MakeHeader(),miner)
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
	Trace()
	ds.blockReceivedTime = time.Now()
	ds.InitializeConsensus(0)
}

func (ds *DbftService) CheckSignatures() error{
	Trace()
	if ds.ccntmext.GetSignaturesCount() >= ds.ccntmext.M() && ds.ccntmext.CheckTxHashesExist() {
		ep , err := ds.ccntmext.Miners[ds.ccntmext.MinerIndex].EncodePoint(true)
		if err != nil {
			return NewDetailErr(err,ErrNoCode,"[DbftService] ,EncodePoint failed")
		}
		codehash ,err := ToCodeHash(ep)
		if err != nil {
			return NewDetailErr(err,ErrNoCode,"[DbftService] ,ToCodeHash failed")
		}
		ccntmract,err := ct.CreateMultiSigCcntmract(codehash,ds.ccntmext.M(),ds.ccntmext.Miners)
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

		con.Log(fmt.Sprintf("relay block: %s", block.Hash()))

		if err := ds.localNet.Xmit(block); err != nil{
			con.Log(fmt.Sprintf("reject block: %s", block.Hash()))
		}

		ds.ccntmext.State |= BlockSent

	}
	return nil
}

func (ds *DbftService) CreateBookkeepingTransaction(txs map[Uint256]*tx.Transaction,nonce uint64) *tx.Transaction {
	Trace()
	return &tx.Transaction{
		TxType: tx.BookKeeping,
	}
}

func (ds *DbftService) ChangeViewReceived(payload *msg.ConsensusPayload,message *ChangeView){
	Trace()
	con.Log(fmt.Sprintf("Change View Received: height=%d View=%d index=%d nv=%d",payload.Height,message.ViewNumber(),payload.MinerIndex,message.NewViewNumber))

	if message.NewViewNumber <= ds.ccntmext.ExpectedView[payload.MinerIndex] {
		return
	}

	ds.ccntmext.ExpectedView[payload.MinerIndex] = message.NewViewNumber
	ds.CheckExpectedView(message.NewViewNumber)
}

func (ds *DbftService) CheckExpectedView(viewNumber byte){
	Trace()
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

func (ds *DbftService) Halt() error  {
	Trace()
	if ds.timer != nil {
		ds.timer.Stop()
	}

	if ds.started {
		ledger.DefaultLedger.Blockchain.BCEvents.UnSubscribe(events.EventBlockPersistCompleted,ds.blockPersistCompletedSubscriber)
		ds.localNet.GetEvent("consensus").UnSubscribe(events.EventNewInventory,ds.newInventorySubscriber)
	}
	return nil
}

func (ds *DbftService) InitializeConsensus(viewNum byte) error  {
	Trace()
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if viewNum == 0 {
		ds.ccntmext.Reset(ds.Client)
	} else {
		ds.ccntmext.ChangeView(viewNum)
	}
	fmt.Println("ds.ccntmext.MinerIndex= ",ds.ccntmext.MinerIndex)
	if ds.ccntmext.MinerIndex < 0 {
		return NewDetailErr(errors.New("Miner Index incorrect"),ErrNoCode,"")
	}
	fmt.Println("ds.ccntmext.MinerIndex",ds.ccntmext.MinerIndex)
	fmt.Println("ds.ccntmext.PrimaryIndex",ds.ccntmext.PrimaryIndex)
	if ds.ccntmext.MinerIndex == int(ds.ccntmext.PrimaryIndex) {
		Trace()
		ds.ccntmext.State |= Primary
		ds.timerHeight = ds.ccntmext.Height
		ds.timeView = viewNum
		span := time.Now().Sub(ds.blockReceivedTime)
		Trace()
		if span > TimePerBlock {
			Trace()
			ds.Timeout()
		} else {
			time.AfterFunc(TimePerBlock-span,ds.Timeout)
		}
	} else {
		ds.ccntmext.State = Backup
		ds.timerHeight = ds.ccntmext.Height
		ds.timeView = viewNum
	}
	return nil
}

func (ds *DbftService) LocalNodeNewInventory(v interface{}){
	Trace()
	if inventory,ok := v.(Inventory);ok {
		if inventory.Type() == CONSENSUS {
			payload, isConsensusPayload := inventory.(*msg.ConsensusPayload)
			if isConsensusPayload {
				ds.NewConsensusPayload(payload)
			}
		} else if inventory.Type() == TRANSACTION  {
			transaction, isTransaction := inventory.(*tx.Transaction)
			if isTransaction{
				ds.NewTransactionPayload(transaction)
			}
		}
	}
}

func (ds *DbftService) NewConsensusPayload(payload *msg.ConsensusPayload){
	Trace()
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
	Trace()
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

func (ds *DbftService) PrepareRequestReceived(payload *msg.ConsensusPayload,message *PrepareRequest) {
	Trace()
	con.Log(fmt.Sprintf("Prepare Request Received: height=%d View=%d index=%d tx=%d",payload.Height,message.ViewNumber(),payload.MinerIndex,len(message.TransactionHashes)))

	if ds.ccntmext.State.HasFlag(Backup) || ds.ccntmext.State.HasFlag(RequestReceived) {
		return
	}

	if uint32(payload.MinerIndex) != ds.ccntmext.PrimaryIndex {return }
	header,_ := ledger.DefaultLedger.Blockchain.GetHeader(ds.ccntmext.PrevHash)
	/*
	* TODO Add Error Catch
	* 2017/2/27 luodanwg
	* */
	prevBlockTimestamp := header.Blockdata.Timestamp
	if payload.Timestamp <= prevBlockTimestamp || payload.Timestamp > uint32(time.Now().Add(time.Minute*10).Unix()){
		con.Log(fmt.Sprintf("Timestamp incorrect: %d",payload.Timestamp))
		return
	}

	ds.ccntmext.State |= RequestReceived
	ds.ccntmext.Timestamp = payload.Timestamp
	ds.ccntmext.Nonce = message.Nonce
	ds.ccntmext.NextMiner = message.NextMiner
	ds.ccntmext.TransactionHashes = message.TransactionHashes
	ds.ccntmext.Transactions = make(map[Uint256]*tx.Transaction)

	if _,err := va.VerifySignature(ds.ccntmext.MakeHeader(),ds.ccntmext.Miners[payload.MinerIndex],message.Signature); err != nil {
		return
	}

	minerLen := len(ds.ccntmext.Miners)
	ds.ccntmext.Signatures = make([][]byte,minerLen)
	ds.ccntmext.Signatures[payload.MinerIndex] = message.Signature

	if err := ds.AddTransaction(message.BookkeepingTransaction); err != nil {return }

	mempool :=  ds.localNet.GetMemoryPool()
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
		ds.localNet.SynchronizeMemoryPool()
	}
}

func (ds *DbftService) PrepareResponseReceived(payload *msg.ConsensusPayload,message *PrepareResponse){
	Trace()

	con.Log(fmt.Sprintf("Prepare Response Received: height=%d View=%d index=%d",payload.Height,message.ViewNumber(),payload.MinerIndex))

	if ds.ccntmext.State.HasFlag(BlockSent)  {return}
	if ds.ccntmext.Signatures[payload.MinerIndex] != nil {return }

	header := ds.ccntmext.MakeHeader()
	if  header == nil {return }
	if _,err := va.VerifySignature(header,ds.ccntmext.Miners[payload.MinerIndex],message.Signature); err != nil {
		return
	}

	ds.ccntmext.Signatures[payload.MinerIndex] = message.Signature
	ds.CheckSignatures()
}

func  (ds *DbftService)  RefreshPolicy(){
	Trace()
	con.DefaultPolicy.Refresh()
}

func  (ds *DbftService)  RequestChangeView() {
	Trace()
	ds.ccntmext.ExpectedView[ds.ccntmext.MinerIndex]++
	con.Log(fmt.Sprintf("Request change view: height=%d View=%d nv=%d state=%d",ds.ccntmext.Height,ds.ccntmext.ViewNumber,ds.ccntmext.MinerIndex,ds.ccntmext.State))

	time.AfterFunc(SecondsPerBlock << (ds.ccntmext.ExpectedView[ds.ccntmext.MinerIndex]+1),ds.Timeout)
	ds.SignAndRelay(ds.ccntmext.MakeChangeView())
	ds.CheckExpectedView(ds.ccntmext.ExpectedView[ds.ccntmext.MinerIndex])
}

func (ds *DbftService) SignAndRelay(payload *msg.ConsensusPayload){
	Trace()
	ctCxt := ct.NewCcntmractCcntmext(payload)

	ds.Client.Sign(ctCxt)
	ctCxt.Data.SetPrograms(ctCxt.GetPrograms())
	ds.localNet.Xmit(payload)
}

func (ds *DbftService) Start() error  {
	Trace()
	ds.started = true

	ds.newInventorySubscriber = ledger.DefaultLedger.Blockchain.BCEvents.Subscribe(events.EventBlockPersistCompleted,ds.BlockPersistCompleted)
	ds.blockPersistCompletedSubscriber = ds.localNet.GetEvent("consensus").Subscribe(events.EventNewInventory,ds.LocalNodeNewInventory)

	ds.InitializeConsensus(0)
	return nil
}

func (ds *DbftService) Timeout() {
	Trace()
	//ds.mu.Lock()
	//defer ds.mu.Unlock()
	if ds.timerHeight != ds.ccntmext.Height || ds.timeView != ds.ccntmext.ViewNumber {
		return
	}
	fmt.Sprintf("Timeout: height=%d View=%d state=%d",ds.timerHeight,ds.timeView,ds.ccntmext.State)
	con.Log(fmt.Sprintf("Timeout: height=%d View=%d state=%d",ds.timerHeight,ds.timeView,ds.ccntmext.State))
	fmt.Println("ds.ccntmext.State.HasFlag(Primary)=",ds.ccntmext.State.HasFlag(Primary))
	fmt.Println("ds.ccntmext.State.HasFlag(RequestSent)=",ds.ccntmext.State.HasFlag(RequestSent))
	fmt.Println("ds.ccntmext.State.HasFlag(Backup)=",ds.ccntmext.State.HasFlag(Backup))

	if ds.ccntmext.State.HasFlag(Primary) && !ds.ccntmext.State.HasFlag(RequestSent) {
		con.Log(fmt.Sprintf("Send prepare request: height=%d View=%d",ds.timerHeight,ds.timeView,ds.ccntmext.State))
		ds.ccntmext.State |= RequestSent
		if !ds.ccntmext.State.HasFlag(SignatureSent) {

			//set ccntmext Timestamp
			now := uint32(time.Now().Unix())
			fmt.Println("ds.ccntmext.PrevHash",ds.ccntmext.PrevHash)
			header,_:= ledger.DefaultLedger.Blockchain.GetHeader(ds.ccntmext.PrevHash)
			fmt.Println(" ledger.DefaultLedger.Blockchain.GetHeader(ds.ccntmext.PrevHash)",header)
			/*
			* TODO Error Catch
			* 2017/2/27 luodanwg
			* */
			blockTime := header.Blockdata.Timestamp

			if blockTime > now {
				ds.ccntmext.Timestamp = blockTime
			} else {
				ds.ccntmext.Timestamp = now
			}

			ds.ccntmext.Nonce = GetNonce()
			transactions := ds.localNet.GetMemoryPool() //TODO: add policy

			txBookkeeping := ds.CreateBookkeepingTransaction(transactions,ds.ccntmext.Nonce)
			transactions[txBookkeeping.Hash()] = txBookkeeping

			if ds.ccntmext.TransactionHashes == nil {
				ds.ccntmext.TransactionHashes = []Uint256{}
			}

			ds.ccntmext.TransactionHashes = append(ds.ccntmext.TransactionHashes,txBookkeeping.Hash())
			for _,TX := range transactions {
				ds.ccntmext.TransactionHashes = append(ds.ccntmext.TransactionHashes,TX.Hash())
			}
			ds.ccntmext.Transactions = transactions

			txlist := ds.ccntmext.GetTransactionList()
			ds.ccntmext.NextMiner,_= ledger.GetMinerAddress(ledger.DefaultLedger.Blockchain.GetMinersByTXs(txlist))
			/*
			* TODO  add error catch
			* 2017/2/27 luodanwg
			* */
			block := ds.ccntmext.MakeHeader()
			account,_:= ds.Client.GetAccount(ds.ccntmext.Miners[ds.ccntmext.MinerIndex])
			/*
			* TODO add error catch
			* 2017/2/27 luodanwg
			* */
			ds.ccntmext.Signatures[ds.ccntmext.MinerIndex],_ = sig.SignBySigner(block,account)
			/*
			* TODO add error catch
			* 2017/2/27 luodanwg
			* */
		}
		ds.SignAndRelay(ds.ccntmext.MakePrepareRequest())
		time.AfterFunc(SecondsPerBlock << (ds.timeView + 1), ds.Timeout)

	} else if ds.ccntmext.State.HasFlag(Primary) && ds.ccntmext.State.HasFlag(RequestSent) || ds.ccntmext.State.HasFlag(Backup){
		ds.RequestChangeView()
	}
}
