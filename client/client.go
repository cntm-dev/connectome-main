package client

import (
	"GoOnchain/crypto"
	ct "GoOnchain/core/ccntmract"
	. "GoOnchain/common"
	"sync"
	sig "GoOnchain/core/signature"
	. "GoOnchain/errors"
	"GoOnchain/core/ledger"
	"time"
)


type Client struct {
	mu           sync.Mutex
	path string
	iv []byte
	masterKey []byte

	accounts map[Uint160]*Account
	ccntmracts map[Uint160]*ct.Ccntmract

	watchOnly []Uint160
	currentHeight uint32

	store ClientStore
	isrunning bool

}

//TODO: adjust ccntmract folder structure


func NewClient(path string,passwordKey []byte,store ClientStore,create bool) *Client {
	newClient := &Client{
		path: path,
		isrunning: true,
	}

	if create {
		//create new client
		newClient.iv = make([]byte,16)
		newClient.masterKey = make([]byte,32)
		newClient.watchOnly = []Uint160{}
		if ledger.DefaultLedger.Blockchain == nil{
			newClient.currentHeight = 0
		} else { newClient.currentHeight = 1}

		//TODO: generate random number for iv/masterkey

		//TODO: new client store (build DB)

		newClient.store.SaveStoredData("PasswordHash",crypto.Sha256(passwordKey))
		newClient.store.SaveStoredData("IV",newClient.iv)
		newClient.store.SaveStoredData("MasterKey",newClient.masterKey) //TODO: AES Encrypt
		newClient.store.SaveStoredData("Version",[]byte{1}) //TODO: Version setting
		newClient.store.SaveStoredData("Height",IntToBytes(int(newClient.currentHeight)))
	} else {
		//load client
		passwordHash := newClient.store.LoadStoredData("PasswordHash")
		if passwordHash != nil && !IsEqualBytes(passwordHash,crypto.Sha256(passwordKey)){
			return nil //TODO: add panic
		}

		newClient.iv = newClient.store.LoadStoredData("IV")
		newClient.masterKey = newClient.store.LoadStoredData("MasterKey") //TODO: AES Dncrypt
		newClient.accounts = newClient.store.LoadAccount()
		newClient.ccntmracts = newClient.store.LoadCcntmracts()

		//TODO: watch only
		ClearBytes(passwordKey)

		go newClient.ProcessBlocks()

	}

	return newClient
}

func (cl *Client) GetAccount(pubKey *crypto.PubKey) (*Account,error){
	temp,err := pubKey.EncodePoint(true)
	if err !=nil{
		return nil,NewDetailErr(err, ErrNoCode, "[Ccntmract],CreateSignatureCcntmract failed.")
	}
	hash,err :=ToCodeHash(temp)
	if err !=nil{
		return nil,NewDetailErr(err, ErrNoCode, "[Ccntmract],CreateSignatureCcntmract failed.")
	}
	return cl.GetAccountByKeyHash(hash),nil
}

func (cl *Client) GetAccountByKeyHash(publicKeyHash Uint160) *Account{
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if account,ok := cl.accounts[publicKeyHash]; ok{
		return account
	}
	return nil
}

func (cl *Client) GetAccountByProgramHash(programHash Uint160) *Account{
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if ccntmract,ok := cl.ccntmracts[programHash]; ok{
		return cl.accounts[ccntmract.OwnerPubkeyHash]
	}
	return nil
}

func (cl *Client) GetCcntmract(codeHash Uint160) *ct.Ccntmract{
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if ccntmract,ok := cl.ccntmracts[codeHash]; ok{
		return ccntmract
	}
	return nil
}

func (cl *Client) ChangePassword(oldPassword string,newPassword string) bool{
	if !cl.VerifyPassword(oldPassword) {
		return  false
	}

	//TODO: ChangePassword

	return false
}

func (cl *Client) CcntmainsAccount(pubKey *crypto.PubKey) bool{
	//TODO: CcntmainsAccount
	return false
}

func (cl *Client) CreateAccount() *Account{
	privateKey := make([]byte,32)

	//TODO: Generate Private Key

	account := cl.CreateAccountByPrivateKey(privateKey)
	ClearBytes(privateKey)

	return account
}

func (cl *Client) CreateAccountByPrivateKey(privateKey []byte) *Account {
	account,_ := NewAccount(privateKey)
	cl.mu.Lock()
	defer cl.mu.Unlock()
	cl.accounts[account.PublicKeyHash] = account

	return account
}

func (cl *Client) ProcessBlocks() {
	for  {
		if !cl.isrunning { break}

		for{
			if ledger.DefaultLedger.Blockchain == nil {break}
			if cl.currentHeight > ledger.DefaultLedger.Blockchain.BlockHeight {break}

			cl.mu.Lock()

			block ,_:= ledger.DefaultLedger.GetBlockWithHeight(cl.currentHeight)
			if block != nil{
				cl.ProcessNewBlock(block)
			}

			cl.mu.Unlock()
		}

		for i:=0;i < 20 ;i++ {
			time.Sleep(time.Millisecond * 100)
		}
	}
}

func (cl *Client) ProcessNewBlock(block *ledger.Block) {
	//TODO: ProcessNewBlock

}

func (cl *Client) Sign(ccntmext *ct.CcntmractCcntmext) bool{
	fSuccess := false
	for i,hash := range ccntmext.ProgramHashes{
		ccntmract := cl.GetCcntmract(hash)
		if ccntmract == nil {ccntminue}

		account := cl.GetAccountByProgramHash(hash)
		if account == nil {ccntminue}

		signature,errx:= sig.SignBySigner(ccntmext.Data,account)
		if errx != nil{
			return false
		}
		err := ccntmext.AddCcntmract(ccntmract,account.PublicKey,signature)

		if err != nil {
			fSuccess = false
		} else {
			if i == 0 {
				fSuccess = true
			}
		}
	}
	return fSuccess
}

func ClearBytes(bytes []byte){
	for i:=0; i<len(bytes) ;i++  {
		bytes[i] = 0
	}
}

func (cl *Client) VerifyPassword(password string) bool{
	//TODO: VerifyPassword
	return true
}