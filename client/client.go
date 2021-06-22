package client

import (
	. "DNA/common"
	"DNA/common/log"
	"DNA/common/serialization"
	"DNA/core/ccntmract"
	ct "DNA/core/ccntmract"
	"DNA/core/ledger"
	sig "DNA/core/signature"
	"DNA/crypto"
	. "DNA/errors"
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type ClientVersion struct {
	Major    uint32
	Minor    uint32
	Build    uint32
	Revision uint32
}

func (v *ClientVersion) ToArray() []byte {
	vbuf := bytes.NewBuffer(nil)
	serialization.WriteUint32(vbuf, v.Major)
	serialization.WriteUint32(vbuf, v.Minor)
	serialization.WriteUint32(vbuf, v.Build)
	serialization.WriteUint32(vbuf, v.Revision)

	return vbuf.Bytes()
}

type Client interface {
	Sign(ccntmext *ct.CcntmractCcntmext) bool
	CcntmainsAccount(pubKey *crypto.PubKey) bool
	GetAccount(pubKey *crypto.PubKey) (*Account, error)
	GetDefaultAccount() (*Account, error)
}

type ClientImpl struct {
	mu sync.Mutex

	path      string
	iv        []byte
	masterKey []byte

	accounts  map[Uint160]*Account
	ccntmracts map[Uint160]*ct.Ccntmract

	watchOnly     []Uint160
	currentHeight uint32

	store     FileStore
	isrunning bool
}

//TODO: adjust ccntmract folder structure

func CreateClient(path string, passwordKey []byte) *ClientImpl {
	cl := NewClient(path, passwordKey, true)

	_, err := cl.CreateAccount()
	if err != nil {
		fmt.Println(err)
	}

	return cl
}

func OpenClient(path string, passwordKey []byte) *ClientImpl {
	cl := NewClient(path, passwordKey, false)

	if cl != nil {
		cl.accounts = cl.LoadAccount()
		cl.ccntmracts = cl.LoadCcntmracts()

		return cl
	}

	return nil
}

func NewClient(path string, passwordKey []byte, create bool) *ClientImpl {
	newClient := &ClientImpl{
		path:      path,
		accounts:  map[Uint160]*Account{},
		ccntmracts: map[Uint160]*ct.Ccntmract{},
		store:     FileStore{path: path},
		isrunning: true,
	}

	passwordKey = crypto.ToAesKey(passwordKey)

	if create {
		//create new client
		newClient.iv = make([]byte, 16)
		newClient.masterKey = make([]byte, 32)
		newClient.watchOnly = []Uint160{}
		newClient.currentHeight = 0

		//generate random number for iv/masterkey
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		for i := 0; i < 16; i++ {
			newClient.iv[i] = byte(r.Intn(256))
		}
		for i := 0; i < 32; i++ {
			newClient.masterKey[i] = byte(r.Intn(256))
		}

		//new client store (build DB)
		newClient.store.BuildDatabase(path)

		// SaveStoredData
		pwdhash := sha256.Sum256(passwordKey)
		err := newClient.store.SaveStoredData("PasswordHash", pwdhash[:])
		if err != nil {
			fmt.Println(err)
			return nil
		}
		err = newClient.store.SaveStoredData("IV", newClient.iv[:])
		if err != nil {
			fmt.Println(err)
			return nil
		}

		aesmk, err := crypto.AesEncrypt(newClient.masterKey[:], passwordKey, newClient.iv)
		if err == nil {
			err = newClient.store.SaveStoredData("MasterKey", aesmk)
			if err != nil {
				log.Error(err)
				return nil
			}
		} else {
			log.Error(err)
			return nil
		}
	} else {
		//load client
		passwordHash, err := newClient.store.LoadStoredData("PasswordHash")
		if err != nil {
			log.Error(err)
			return nil
		}

		pwdhash := sha256.Sum256(passwordKey)
		if passwordHash == nil {
			log.Error("ERROR: passwordHash = nil")
			return nil
		}

		if !IsEqualBytes(passwordHash, pwdhash[:]) {
			log.Error("ERROR: password wrcntm!")
			return nil
		}

		log.Info("[OpenClient] Password Verify.")

		newClient.iv, err = newClient.store.LoadStoredData("IV")
		if err != nil {
			log.Error(err)
			return nil
		}

		masterKey, err := newClient.store.LoadStoredData("MasterKey")
		if err != nil {
			log.Error(err)
			return nil
		}

		newClient.masterKey, err = crypto.AesDecrypt(masterKey, passwordKey, newClient.iv)
		if err != nil {
			log.Error(err)
			return nil
		}
	}

	ClearBytes(passwordKey, len(passwordKey))

	return newClient
}

func (cl *ClientImpl) GetDefaultAccount() (*Account, error) {
	for k, _ := range cl.accounts {
		return cl.GetAccountByKeyHash(k), nil
	}

	return nil, NewDetailErr(errors.New("Can't load default account."), ErrNoCode, "")
}

func (cl *ClientImpl) GetAccount(pubKey *crypto.PubKey) (*Account, error) {
	temp, err := pubKey.EncodePoint(true)
	if err != nil {
		return nil, NewDetailErr(err, ErrNoCode, "[Ccntmract],CreateSignatureCcntmract failed.")
	}
	hash, err := ToCodeHash(temp)
	if err != nil {
		return nil, NewDetailErr(err, ErrNoCode, "[Ccntmract],CreateSignatureCcntmract failed.")
	}
	return cl.GetAccountByKeyHash(hash), nil
}

func (cl *ClientImpl) GetAccountByKeyHash(publicKeyHash Uint160) *Account {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if account, ok := cl.accounts[publicKeyHash]; ok {
		return account
	}
	return nil
}

func (cl *ClientImpl) GetAccountByProgramHash(programHash Uint160) *Account {
	log.Debug()
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if ccntmract, ok := cl.ccntmracts[programHash]; ok {
		return cl.accounts[ccntmract.OwnerPubkeyHash]
	}
	return nil
}

func (cl *ClientImpl) GetCcntmract(codeHash Uint160) *ct.Ccntmract {
	log.Debug()
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if ccntmract, ok := cl.ccntmracts[codeHash]; ok {
		return ccntmract
	}
	return nil
}

func (cl *ClientImpl) ChangePassword(oldPassword string, newPassword string) bool {
	if !cl.VerifyPassword(oldPassword) {
		return false
	}

	//TODO: ChangePassword

	return false
}

func (cl *ClientImpl) CcntmainsAccount(pubKey *crypto.PubKey) bool {
	acpubkey, err := pubKey.EncodePoint(true)
	if err == nil {
		Publickey, err := ToCodeHash(acpubkey)
		if err == nil {
			if cl.GetAccountByKeyHash(Publickey) != nil {
				return true
			} else {
				return false
			}

		} else {
			log.Error(err)
			return false
		}
	} else {
		log.Error(err)
		return false
	}
}

func (cl *ClientImpl) CreateAccount() (*Account, error) {
	ac, err := NewAccount()
	if err != nil {
		return nil, err
	}

	cl.mu.Lock()
	cl.accounts[ac.PublicKeyHash] = ac
	cl.mu.Unlock()

	err = cl.SaveAccount(ac)
	if err != nil {
		return nil, err
	}

	ct, err := ccntmract.CreateSignatureCcntmract(ac.PublicKey)
	if err == nil {
		cl.AddCcntmract(ct)
		log.Info("[CreateCcntmract] Address: %s\n", ct.ProgramHash.ToAddress())
	}

	log.Info("Create account Success")
	return ac, nil
}

func (cl *ClientImpl) CreateAccountByPrivateKey(privateKey []byte) (*Account, error) {
	ac, err := NewAccountWithPrivatekey(privateKey)
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if err != nil {
		return nil, err
	}

	cl.accounts[ac.PublicKeyHash] = ac
	err = cl.SaveAccount(ac)
	if err != nil {
		return nil, err
	}
	return ac, nil
}

func (cl *ClientImpl) ProcessBlocks() {
	for {
		if !cl.isrunning {
			break
		}

		for {
			if ledger.DefaultLedger.Blockchain == nil {
				break
			}
			if cl.currentHeight > ledger.DefaultLedger.Blockchain.BlockHeight {
				break
			}

			cl.mu.Lock()

			block, _ := ledger.DefaultLedger.GetBlockWithHeight(cl.currentHeight)
			if block != nil {
				cl.ProcessNewBlock(block)
			}

			cl.mu.Unlock()
		}

		for i := 0; i < 20; i++ {
			time.Sleep(time.Millisecond * 100)
		}
	}
}

func (cl *ClientImpl) ProcessNewBlock(block *ledger.Block) {
	//TODO: ProcessNewBlock

}

func (cl *ClientImpl) Sign(ccntmext *ct.CcntmractCcntmext) bool {
	log.Debug()
	fSuccess := false
	for i, hash := range ccntmext.ProgramHashes {
		ccntmract := cl.GetCcntmract(hash)
		if ccntmract == nil {
			ccntminue
		}
		account := cl.GetAccountByProgramHash(hash)
		if account == nil {
			ccntminue
		}

		signature, err := sig.SignBySigner(ccntmext.Data, account)
		if err != nil {
			return fSuccess
		}
		err = ccntmext.AddCcntmract(ccntmract, account.PublicKey, signature)

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

func (cl *ClientImpl) VerifyPassword(password string) bool {
	//TODO: VerifyPassword
	return true
}

func (cl *ClientImpl) EncryptPrivateKey(prikey []byte) ([]byte, error) {
	enc, err := crypto.AesEncrypt(prikey, cl.masterKey, cl.iv)
	if err != nil {
		return nil, err
	}

	return enc, nil
}

func (cl *ClientImpl) DecryptPrivateKey(prikey []byte) ([]byte, error) {
	if prikey == nil {
		return nil, NewDetailErr(errors.New("The PriKey is nil"), ErrNoCode, "")
	}
	if len(prikey) != 96 {
		return nil, NewDetailErr(errors.New("The len of PriKeyEnc is not 96bytes"), ErrNoCode, "")
	}

	dec, err := crypto.AesDecrypt(prikey, cl.masterKey, cl.iv)
	if err != nil {
		return nil, err
	}

	return dec, nil
}

func (cl *ClientImpl) SaveAccount(ac *Account) error {

	decryptedPrivateKey := make([]byte, 96)
	temp, err := ac.PublicKey.EncodePoint(false)
	if err != nil {
		return err
	}
	for i := 1; i <= 64; i++ {
		decryptedPrivateKey[i-1] = temp[i]
	}

	for i := 0; i < 32; i++ {
		decryptedPrivateKey[64+i] = ac.PrivateKey[i]
	}

	encryptedPrivateKey, err := cl.EncryptPrivateKey(decryptedPrivateKey)
	if err != nil {
		return err
	}

	ClearBytes(decryptedPrivateKey, 96)

	err = cl.store.SaveAccountData(ac.PublicKeyHash.ToArray(), encryptedPrivateKey)
	if err != nil {
		return err
	}

	return nil
}

func (cl *ClientImpl) LoadAccount() map[Uint160]*Account {

	i := 0
	accounts := map[Uint160]*Account{}
	for true {
		pubkeyhash, prikeyenc, err := cl.store.LoadAccountData(i)
		if err != nil {
			// TODO: report the error
		}

		decryptedPrivateKey, err := cl.DecryptPrivateKey(prikeyenc)
		if err != nil {
			log.Error(err)
		}

		prikey := decryptedPrivateKey[64:96]
		ac, err := NewAccountWithPrivatekey(prikey)

		pk, _ := ac.PublicKey.EncodePoint(true)
		log.Debug("[LoadAccount] PublicKey: %x\n", pk)
		pkhash, _ := Uint160ParseFromBytes(pubkeyhash)
		accounts[pkhash] = ac
		i++
		break
	}

	return accounts
}

func (cl *ClientImpl) LoadCcntmracts() map[Uint160]*ct.Ccntmract {

	i := 0
	ccntmracts := map[Uint160]*ct.Ccntmract{}

	for true {
		ph, _, rd, err := cl.store.LoadCcntmractData(i)
		if err != nil {
			//fmt.Println( err )
			break
		}

		rdreader := bytes.NewReader(rd)
		ct := new(ct.Ccntmract)
		ct.Deserialize(rdreader)

		programhash, err := Uint160ParseFromBytes(ph)
		ct.ProgramHash = programhash

		ccntmracts[ct.ProgramHash] = ct

		log.Info("[LoadCcntmract] Address: %s\n", ct.ProgramHash.ToAddress())
		i++
		break
	}

	return ccntmracts
}
func (cl *ClientImpl) AddCcntmract(ct *ccntmract.Ccntmract) error {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if cl.accounts[ct.OwnerPubkeyHash] == nil {
		return NewDetailErr(errors.New("AddCcntmract(): ccntmract.OwnerPubkeyHash not in []accounts"), ErrNoCode, "")
	}

	cl.ccntmracts[ct.ProgramHash] = ct

	err := cl.store.SaveCcntmractData(ct)
	return err
}
