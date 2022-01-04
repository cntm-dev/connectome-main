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

package account

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Ontology/common"
	"github.com/Ontology/common/config"
	"github.com/Ontology/common/log"
	"github.com/Ontology/common/password"
	"github.com/Ontology/core/ccntmract"
	ct "github.com/Ontology/core/ccntmract"
	sig "github.com/Ontology/core/signature"
	cntmErrors "github.com/Ontology/errors"
	"github.com/Ontology/net/protocol"
	"github.com/cntmio/cntmology-crypto/aes"
	"github.com/cntmio/cntmology-crypto/keypair"
)

const (
	DEFAULT_BOOKKEEPER_COUNT = 4
	WALLET_FILENAME          = "wallet.dat"
)

type Client interface {
	Sign(ccntmext *ct.CcntmractCcntmext) bool
	CcntmainsAccount(pubKey keypair.PublicKey) bool
	GetAccount(pubKey keypair.PublicKey) (*Account, error)
	GetDefaultAccount() (*Account, error)
	GetBookkeepers() ([]keypair.PublicKey, error)
}

type ClientImpl struct {
	mu sync.Mutex

	path      string
	iv        []byte
	masterKey []byte

	accounts  map[common.Address]*Account
	ccntmracts map[common.Address]*ct.Ccntmract

	watchOnly     []common.Address
	currentHeight uint32

	FileStore
	isrunning bool
}

//TODO: adjust ccntmract folder structure
func Create(path string, passwordKey []byte) *ClientImpl {
	cl := NewClient(path, passwordKey, true)

	_, err := cl.CreateAccount()
	if err != nil {
		fmt.Println(err)
	}

	return cl
}

func Open(path string, passwordKey []byte) *ClientImpl {
	cl := NewClient(path, passwordKey, false)
	if cl == nil {
		log.Error("Alloc new client failure")
		return nil
	}

	cl.accounts = cl.LoadAccount()
	if cl.accounts == nil {
		log.Error("Load accounts failure")
	}
	cl.ccntmracts = cl.LoadCcntmracts()
	if cl.ccntmracts == nil {
		log.Error("Load ccntmracts failure")
	}
	return cl
}

func NewClient(path string, password []byte, create bool) *ClientImpl {
	newClient := &ClientImpl{
		path:      path,
		accounts:  map[common.Address]*Account{},
		ccntmracts: map[common.Address]*ct.Ccntmract{},
		FileStore: FileStore{path: path},
		isrunning: true,
	}

	passwordKey := DoubleHash(password)
	if create {
		//create new client
		newClient.iv = make([]byte, 16)
		newClient.masterKey = make([]byte, 32)
		newClient.watchOnly = []common.Address{}
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
		newClient.BuildDatabase(path)

		// SaveStoredData
		pwdhash := sha256.Sum256(passwordKey)
		err := newClient.SaveStoredData("PasswordHash", pwdhash[:])
		if err != nil {
			log.Error(err)
			return nil
		}
		err = newClient.SaveStoredData("IV", newClient.iv[:])
		if err != nil {
			log.Error(err)
			return nil
		}

		aesmk, err := aes.AesEncrypt(newClient.masterKey[:], passwordKey, newClient.iv)
		if err == nil {
			err = newClient.SaveStoredData("MasterKey", aesmk)
			if err != nil {
				log.Error(err)
				return nil
			}
		} else {
			log.Error(err)
			return nil
		}
	} else {
		if b := newClient.verifyPasswordKey(passwordKey); b == false {
			return nil
		}
		if err := newClient.loadClient(passwordKey); err != nil {
			return nil
		}
	}
	ClearBytes(passwordKey, len(passwordKey))
	return newClient
}

func (cl *ClientImpl) loadClient(passwordKey []byte) error {
	var err error
	cl.iv, err = cl.LoadStoredData("IV")
	if err != nil {
		fmt.Println("error: failed to load iv")
		return err
	}
	encryptedMasterKey, err := cl.LoadStoredData("MasterKey")
	if err != nil {
		fmt.Println("error: failed to load master key")
		return err
	}
	cl.masterKey, err = aes.AesDecrypt(encryptedMasterKey, passwordKey, cl.iv)
	if err != nil {
		fmt.Println("error: failed to decrypt master key")
		return err
	}
	return nil
}

func (cl *ClientImpl) GetDefaultAccount() (*Account, error) {
	for programHash, _ := range cl.accounts {
		return cl.GetAccountByProgramHash(programHash), nil
	}

	return nil, cntmErrors.NewDetailErr(errors.New("Can't load default account."), cntmErrors.ErrNoCode, "")
}

func (cl *ClientImpl) GetAccount(pubKey keypair.PublicKey) (*Account, error) {
	signatureRedeemScript, err := ccntmract.CreateSignatureRedeemScript(pubKey)
	if err != nil {
		return nil, cntmErrors.NewDetailErr(err, cntmErrors.ErrNoCode, "CreateSignatureRedeemScript failed")
	}
	programHash := common.ToCodeHash(signatureRedeemScript)
	return cl.GetAccountByProgramHash(programHash), nil
}

func (cl *ClientImpl) GetAccountByProgramHash(programHash common.Address) *Account {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if account, ok := cl.accounts[programHash]; ok {
		return account
	}
	return nil
}

func (cl *ClientImpl) GetCcntmract(programHash common.Address) *ct.Ccntmract {
	log.Debug()
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if ccntmract, ok := cl.ccntmracts[programHash]; ok {
		return ccntmract
	}
	return nil
}

func (cl *ClientImpl) ChangePassword(oldPassword []byte, newPassword []byte) bool {
	// check password
	oldPasswordKey := DoubleHash(oldPassword)
	if !cl.verifyPasswordKey(oldPasswordKey) {
		fmt.Println("error: password verification failed")
		return false
	}
	if err := cl.loadClient(oldPasswordKey); err != nil {
		fmt.Println("error: load wallet info failed")
		return false
	}

	// encrypt master key with new password
	newPasswordKey := DoubleHash(newPassword)
	newMasterKey, err := aes.AesEncrypt(cl.masterKey, newPasswordKey, cl.iv)
	if err != nil {
		fmt.Println("error: set new password failed")
		return false
	}

	// update wallet file
	newPasswordHash := sha256.Sum256(newPasswordKey)
	if err := cl.SaveStoredData("PasswordHash", newPasswordHash[:]); err != nil {
		fmt.Println("error: wallet update failed(password hash)")
		return false
	}
	if err := cl.SaveStoredData("MasterKey", newMasterKey); err != nil {
		fmt.Println("error: wallet update failed (encrypted master key)")
		return false
	}
	ClearBytes(newPasswordKey, len(newPasswordKey))
	ClearBytes(cl.masterKey, len(cl.masterKey))

	return true
}

func (cl *ClientImpl) CcntmainsAccount(pubKey keypair.PublicKey) bool {
	signatureRedeemScript, err := ccntmract.CreateSignatureRedeemScript(pubKey)
	if err != nil {
		return false
	}
	programHash := common.ToCodeHash(signatureRedeemScript)
	if cl.GetAccountByProgramHash(programHash) != nil {
		return true
	} else {
		return false
	}
}

func (cl *ClientImpl) CreateAccount() (*Account, error) {
	ac := NewAccount()

	cl.mu.Lock()
	cl.accounts[ac.Address] = ac
	cl.mu.Unlock()

	err := cl.SaveAccount(ac)
	if err != nil {
		return nil, err
	}

	ct, err := ccntmract.CreateSignatureCcntmract(ac.PublicKey)
	if err == nil {
		cl.AddCcntmract(ct)
		address := ct.ProgramHash.ToBase58()
		log.Info("[CreateCcntmract] Address: ", address)
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

	cl.accounts[ac.Address] = ac
	err = cl.SaveAccount(ac)
	if err != nil {
		return nil, err
	}
	return ac, nil
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

func (cl *ClientImpl) verifyPasswordKey(passwordKey []byte) bool {
	savedPasswordHash, err := cl.LoadStoredData("PasswordHash")
	if err != nil {
		fmt.Println("error: failed to load password hash")
		return false
	}
	if savedPasswordHash == nil {
		fmt.Println("error: saved password hash is nil")
		return false
	}
	passwordHash := sha256.Sum256(passwordKey)
	///ClearBytes(passwordKey, len(passwordKey))
	if !common.IsEqualBytes(savedPasswordHash, passwordHash[:]) {
		fmt.Println("error: password wrcntm")
		return false
	}
	return true
}

func (cl *ClientImpl) EncryptPrivateKey(prikey []byte) ([]byte, error) {
	enc, err := aes.AesEncrypt(prikey, cl.masterKey, cl.iv)
	if err != nil {
		return nil, err
	}

	return enc, nil
}

func (cl *ClientImpl) DecryptPrivateKey(prikey []byte) ([]byte, error) {
	if prikey == nil {
		return nil, cntmErrors.NewDetailErr(errors.New("The PriKey is nil"), cntmErrors.ErrNoCode, "")
	}

	dec, err := aes.AesDecrypt(prikey, cl.masterKey, cl.iv)
	if err != nil {
		return nil, err
	}

	return dec, nil
}

func (cl *ClientImpl) SaveAccount(ac *Account) error {
	buf := keypair.SerializePrivateKey(ac.PrivateKey)
	encryptedPrivateKey, err := cl.EncryptPrivateKey(buf)
	if err != nil {
		return err
	}

	ClearBytes(buf, len(buf))

	err = cl.SaveAccountData(ac.Address[:], encryptedPrivateKey)
	if err != nil {
		return err
	}

	return nil
}

func (cl *ClientImpl) LoadAccount() map[common.Address]*Account {
	i := 0
	accounts := map[common.Address]*Account{}
	for true {
		_, prikeyenc, err := cl.LoadAccountData(i)
		if err != nil {
			log.Error(err)
			break
		}

		buf, err := cl.DecryptPrivateKey(prikeyenc)
		if err != nil {
			log.Error(err)
			break
		}

		ac, err := NewAccountWithPrivatekey(buf)
		if err != nil {
			log.Error(err)
			break
		}
		accounts[ac.Address] = ac
		i++
		break
	}

	return accounts
}

func (cl *ClientImpl) LoadCcntmracts() map[common.Address]*ct.Ccntmract {
	i := 0
	ccntmracts := map[common.Address]*ct.Ccntmract{}

	for true {
		ph, _, rd, err := cl.LoadCcntmractData(i)
		if err != nil {
			fmt.Println(err)
			break
		}

		rdreader := bytes.NewReader(rd)
		ct := new(ct.Ccntmract)
		ct.Deserialize(rdreader)

		programhash, err := common.AddressParseFromBytes(ph)
		ct.ProgramHash = programhash
		ccntmracts[ct.ProgramHash] = ct
		i++
		break
	}

	return ccntmracts
}

func (cl *ClientImpl) AddCcntmract(ct *ccntmract.Ccntmract) error {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if cl.accounts[ct.ProgramHash] == nil {
		return cntmErrors.NewDetailErr(errors.New("AddCcntmract(): ccntmract.OwnerPubkeyHash not in []accounts"), cntmErrors.ErrNoCode, "")
	}

	cl.ccntmracts[ct.ProgramHash] = ct

	err := cl.SaveCcntmractData(ct)
	return err
}

func (cl *ClientImpl) GetBookkeepers() ([]keypair.PublicKey, error) {
	var pubKeys = []keypair.PublicKey{}
	consensusType := config.Parameters.ConsensusType
	if consensusType == "solo" {
		ac, err := cl.GetDefaultAccount()
		if err != nil {
			return nil, fmt.Errorf("GetDefaultAccount error:%s", err)
		}
		pubKeys = append(pubKeys, ac.PublicKey)
		return pubKeys, nil
	}

	sort.Strings(config.Parameters.Bookkeepers)
	for _, key := range config.Parameters.Bookkeepers {
		//pubKey := []byte(key)
		pubKey, err := hex.DecodeString(key)
		k, err := keypair.DeserializePublicKey(pubKey)
		if err != nil {
			log.Error("Incorrectly book keepers key")
			return nil, fmt.Errorf("Incorrectly book keepers key:%s")
		}
		pubKeys = append(pubKeys, k)
	}

	return pubKeys, nil
}

func clientIsDefaultBookkeeper(publicKey string) bool {
	for _, bookkeeper := range config.Parameters.Bookkeepers {
		if strings.Compare(bookkeeper, publicKey) == 0 {
			return true
		}
	}
	return false
}

func nodeType(typeName string) int {
	if "service" == config.Parameters.NodeType {
		return protocol.SERVICE_NODE
	} else {
		return protocol.VERIFY_NODE
	}
}

func GetClient() Client {
	if !common.FileExisted(WALLET_FILENAME) {
		log.Fatal(fmt.Sprintf("No %s detected, please create a wallet by using command line.", WALLET_FILENAME))
		os.Exit(1)
	}
	passwd, err := password.GetAccountPassword()
	if err != nil {
		log.Fatal("Get password error.")
		os.Exit(1)
	}
	c := Open(WALLET_FILENAME, passwd)
	if c == nil {
		return nil
	}
	return c
}

func GetBookkeepers() []keypair.PublicKey {
	var pubKeys = []keypair.PublicKey{}
	sort.Strings(config.Parameters.Bookkeepers)
	for _, key := range config.Parameters.Bookkeepers {
		pubKey := []byte(key)
		pubKey, err := hex.DecodeString(key)
		// TODO Convert the key string to byte
		k, err := keypair.DeserializePublicKey(pubKey)
		if err != nil {
			log.Error("Incorrectly book keepers key")
			return nil
		}
		pubKeys = append(pubKeys, k)
	}

	return pubKeys
}

func DoubleHash(pwd []byte) []byte {
	pwdhash := sha256.Sum256(pwd)
	pwdhash2 := sha256.Sum256(pwdhash[:])

	// Fixme clean the password buffer
	// ClearBytes(pwd,len(pwd))
	ClearBytes(pwdhash[:], 32)

	return pwdhash2[:]
}

func ClearBytes(arr []byte, len int) {
	for i := 0; i < len; i++ {
		arr[i] = 0
	}
}
