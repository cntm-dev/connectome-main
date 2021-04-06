package dbft

import (
	. "GoOnchain/common"
	"GoOnchain/crypto"
	tx "GoOnchain/core/transaction"
	 "GoOnchain/core/ledger"
	msg "GoOnchain/net/message"
	ser "GoOnchain/common/serialization"
	cl "GoOnchain/client"
)

const CcntmextVersion uint32 = 0

type ConsensusCcntmext struct {

	State ConsensusState
	PrevHash Uint256
	Height uint32
	ViewNumber byte
	Miners []*crypto.PubKey
	MinerIndex int
	PrimaryIndex uint32
	Timestamp uint32
	Nonce uint64
	NextMiner Uint160
	TransactionHashes []Uint256
	Transactions map[Uint256]*tx.Transaction
	Signatures [][]byte
	ExpectedView []byte

	txlist []*tx.Transaction

	header *ledger.Block
}

func (cxt *ConsensusCcntmext)  M() int {
	return len(cxt.Miners) - (len(cxt.Miners) - 1) / 3
}

func NewConsensusCcntmext() *ConsensusCcntmext {
	return  &ConsensusCcntmext{
	}
}

func (cxt *ConsensusCcntmext)  ChangeView(viewNum byte)  {
	p := (cxt.Height - uint32(viewNum)) % uint32(len(cxt.Miners))
	cxt.State &= SignatureSent
	cxt.ViewNumber = viewNum
	if p >= 0 {
		cxt.PrimaryIndex = uint32(p)
	} else {
		cxt.PrimaryIndex = uint32(p) + uint32(len(cxt.Miners))
	}

	if cxt.State == Initial{
		cxt.TransactionHashes = nil
		cxt.Signatures = make([][]byte,len(cxt.Miners))
	}
	cxt.header = nil
}

func (cxt *ConsensusCcntmext)  HasTxHash(txHash Uint256) bool {
	for _, hash :=  range cxt.TransactionHashes{
		if hash == txHash {
			return true
		}
	}
	return false
}

func (cxt *ConsensusCcntmext)  MakeChangeView() *msg.ConsensusPayload {
	cv := &ChangeView{
		msgData: &ConsensusMessageData{
			Type: ChangeViewMsg,
		},
		NewViewNumber: cxt.ExpectedView[cxt.MinerIndex],
	}
	return cxt.MakePayload(cv)
}

func (cxt *ConsensusCcntmext)  MakeHeader() *ledger.Block {
	if cxt.TransactionHashes == nil {
		return nil
	}

	txRoot,_ := crypto.ComputeRoot(cxt.TransactionHashes)


	if cxt.header == nil{
		blockData := &ledger.Blockdata{
			Version: CcntmextVersion,
			PrevBlockHash: cxt.PrevHash,
			TransactionsRoot: txRoot,
			Timestamp: cxt.Timestamp,
			Height: cxt.Height,
			ConsensusData: cxt.Nonce,
			NextMiner: cxt.NextMiner,
		}
		cxt.header = &ledger.Block{
			Blockdata: blockData,
			Transcations: []*tx.Transaction{},
		}
	}
	return cxt.header
}

func (cxt *ConsensusCcntmext)  MakePayload(message ConsensusMessage) *msg.ConsensusPayload{
	message.ConsensusMessageData().ViewNumber = cxt.ViewNumber
	return &msg.ConsensusPayload{
		Version: CcntmextVersion,
		PrevHash: cxt.PrevHash,
		Height: cxt.Height,
		MinerIndex: uint16(cxt.MinerIndex),
		Timestamp: cxt.Timestamp,
		Data: ser.ToArray(message),
	}
}

func (cxt *ConsensusCcntmext)  MakePrepareRequest() *msg.ConsensusPayload{
	preReq := &PrepareRequest{
		msgData: &ConsensusMessageData{
			Type: PrepareRequestMsg,
		},
		Nonce: cxt.Nonce,
		NextMiner: cxt.NextMiner,
		TransactionHashes: cxt.TransactionHashes,
		BookkeepingTransaction: cxt.Transactions[cxt.TransactionHashes[0]],
		Signature: cxt.Signatures[cxt.MinerIndex],
	}
	return cxt.MakePayload(preReq)
}

func (cxt *ConsensusCcntmext)  MakePerpareResponse(signature []byte) *msg.ConsensusPayload{
	preRes := &PrepareResponse{
		msgData: &ConsensusMessageData{
			Type: PrepareResponseMsg,
		},
		Signature: signature,
	}
	return cxt.MakePayload(preRes)
}

func (cxt *ConsensusCcntmext)  GetSignaturesCount() (count int){
	count = 0
	for _,sig := range cxt.Signatures {
		if sig != nil {
			count += 1
		}
	}
	return count
}

func (cxt *ConsensusCcntmext)  GetTransactionList()  []*tx.Transaction{
	if cxt.txlist == nil{
		cxt.txlist = []*tx.Transaction{}
		for _,TX := range cxt.Transactions {
			cxt.txlist = append(cxt.txlist,TX)
		}
	}
	return cxt.txlist
}

func (cxt *ConsensusCcntmext)  GetTXByHashes()  []*tx.Transaction{
	TXs := []*tx.Transaction{}
	for _,hash := range cxt.TransactionHashes {
		if TX,ok:=cxt.Transactions[hash]; ok{
			TXs = append(TXs,TX)
		}
	}
	return TXs
}

func (cxt *ConsensusCcntmext)  CheckTxHashesExist() bool {
	for _,hash := range cxt.TransactionHashes {
		if _,ok:=cxt.Transactions[hash]; !ok{
			return false
		}
	}
	return true
}

func (cxt *ConsensusCcntmext) Reset(client *cl.Client){
	cxt.State = Initial
	cxt.PrevHash = ledger.DefaultLedger.Blockchain.CurrentBlockHash()
	cxt.Height = ledger.DefaultLedger.Blockchain.BlockHeight + 1
	cxt.ViewNumber = 0
	cxt.Miners = ledger.DefaultLedger.Blockchain.GetMiners()
	cxt.MinerIndex = -1

	minerLen := len(cxt.Miners)
	cxt.PrimaryIndex = cxt.Height % uint32(minerLen)
	cxt.TransactionHashes = nil
	cxt.Signatures = make([][]byte,minerLen)
	cxt.ExpectedView = make([]byte,minerLen)

	for i:=0;i<minerLen ;i++  {
		if client.CcntmainsAccount(cxt.Miners[i]){
			cxt.MinerIndex = i
			break
		}
	}
	cxt.header = nil
}