/*
 * Copyright (C) 2018 The cntm Authors
 * This file is part of The cntm library.
 *
 * The cntm is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The cntm is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * alcntg with The cntm.  If not, see <http://www.gnu.org/licenses/>.
 */

package utils

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/conntectome/cntm-crypto/keypair"
	sig "github.com/conntectome/cntm-crypto/signature"
	"github.com/conntectome/cntm/account"
	"github.com/conntectome/cntm/common"
	"github.com/conntectome/cntm/common/constants"
	"github.com/conntectome/cntm/common/serialization"
	"github.com/conntectome/cntm/core/payload"
	"github.com/conntectome/cntm/core/signature"
	"github.com/conntectome/cntm/core/types"
	cutils "github.com/conntectome/cntm/core/utils"
	httpcom "github.com/conntectome/cntm/http/base/common"
	rpccommon "github.com/conntectome/cntm/http/base/common"
	"github.com/conntectome/cntm/smartcontract/service/native/cntm"
	"github.com/conntectome/cntm/smartcontract/service/native/utils"
)

const (
	VERSION_TRANSACTION    = byte(0)
	VERSION_CCNTMRACT_CNTM   = byte(0)
	VERSION_CCNTMRACT_CNTG   = byte(0)
	CCNTMRACT_TRANSFER      = "transfer"
	CCNTMRACT_TRANSFER_FROM = "transferFrom"
	CCNTMRACT_APPROVE       = "approve"

	ASSET_CNTM = "cntm"
	ASSET_CNTG = "cntg"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

//Return balance of address in base58 code
func GetBalance(address string) (*httpcom.BalanceOfRsp, error) {
	result, cntmErr := sendRpcRequest("getbalance", []interface{}{address})
	if cntmErr != nil {
		switch cntmErr.ErrorCode {
		case ERROR_INVALID_PARAMS:
			return nil, fmt.Errorf("invalid address:%s", address)
		}
		return nil, cntmErr.Error
	}
	balance := &httpcom.BalanceOfRsp{}
	err := json.Unmarshal(result, balance)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal error:%s", err)
	}
	return balance, nil
}

func GetAccountBalance(address, asset string) (uint64, error) {
	balances, err := GetBalance(address)
	if err != nil {
		return 0, err
	}
	var balance uint64
	switch strings.ToLower(asset) {
	case "cntm":
		balance, err = strconv.ParseUint(balances.Cntm, 10, 64)
	case "cntg":
		balance, err = strconv.ParseUint(balances.Cntg, 10, 64)
	default:
		return 0, fmt.Errorf("unsupport asset:%s", asset)
	}
	if err != nil {
		return 0, err
	}
	return balance, nil
}

func GetAllowance(asset, from, to string) (string, error) {
	result, cntmErr := sendRpcRequest("getallowance", []interface{}{asset, from, to})
	if cntmErr != nil {
		return "", cntmErr.Error
	}
	balance := ""
	err := json.Unmarshal(result, &balance)
	if err != nil {
		return "", fmt.Errorf("json.Unmarshal error:%s", err)
	}
	return balance, nil
}

//Transfer cntm|cntg from account to another account
func Transfer(gasPrice, gasLimit uint64, signer *account.Account, asset, from, to string, amount uint64) (string, error) {
	mutable, err := TransferTx(gasPrice, gasLimit, asset, signer.Address.ToBase58(), to, amount)
	if err != nil {
		return "", err
	}
	err = SignTransaction(signer, mutable)
	if err != nil {
		return "", fmt.Errorf("SignTransaction error:%s", err)
	}
	tx, err := mutable.IntoImmutable()
	if err != nil {
		return "", fmt.Errorf("convert immutable transaction error:%s", err)
	}
	txHash, err := SendRawTransaction(tx)
	if err != nil {
		return "", fmt.Errorf("SendTransaction error:%s", err)
	}
	return txHash, nil
}

func TransferFrom(gasPrice, gasLimit uint64, signer *account.Account, asset, sender, from, to string, amount uint64) (string, error) {
	mutable, err := TransferFromTx(gasPrice, gasLimit, asset, sender, from, to, amount)
	if err != nil {
		return "", err
	}
	err = SignTransaction(signer, mutable)
	if err != nil {
		return "", fmt.Errorf("SignTransaction error:%s", err)
	}
	tx, err := mutable.IntoImmutable()
	if err != nil {
		return "", fmt.Errorf("convert to immutable transaction error:%s", err)
	}

	txHash, err := SendRawTransaction(tx)
	if err != nil {
		return "", fmt.Errorf("SendTransaction error:%s", err)
	}
	return txHash, nil
}

func Approve(gasPrice, gasLimit uint64, signer *account.Account, asset, from, to string, amount uint64) (string, error) {
	mutable, err := ApproveTx(gasPrice, gasLimit, asset, from, to, amount)
	if err != nil {
		return "", err
	}
	err = SignTransaction(signer, mutable)
	if err != nil {
		return "", fmt.Errorf("SignTransaction error:%s", err)
	}
	tx, err := mutable.IntoImmutable()
	if err != nil {
		return "", fmt.Errorf("convert to immutable transaction error:%s", err)
	}
	txHash, err := SendRawTransaction(tx)
	if err != nil {
		return "", fmt.Errorf("SendTransaction error:%s", err)
	}
	return txHash, nil
}

func ApproveTx(gasPrice, gasLimit uint64, asset string, from, to string, amount uint64) (*types.MutableTransaction, error) {
	fromAddr, err := common.AddressFromBase58(from)
	if err != nil {
		return nil, fmt.Errorf("from address:%s invalid:%s", from, err)
	}
	toAddr, err := common.AddressFromBase58(to)
	if err != nil {
		return nil, fmt.Errorf("To address:%s invalid:%s", to, err)
	}
	var state = &cntm.State{
		From:  fromAddr,
		To:    toAddr,
		Value: amount,
	}
	var version byte
	var contractAddr common.Address
	switch strings.ToLower(asset) {
	case ASSET_CNTM:
		version = VERSION_CCNTMRACT_CNTM
		contractAddr = utils.CntmCcntmractAddress
	case ASSET_CNTG:
		version = VERSION_CCNTMRACT_CNTG
		contractAddr = utils.CntgCcntmractAddress
	default:
		return nil, fmt.Errorf("Unsupport asset:%s", asset)
	}
	invokeCode, err := cutils.BuildNativeInvokeCode(contractAddr, version, CCNTMRACT_APPROVE, []interface{}{state})
	if err != nil {
		return nil, fmt.Errorf("build invoke code error:%s", err)
	}
	mutableTx := NewInvokeTransaction(gasPrice, gasLimit, invokeCode)
	return mutableTx, nil
}

func TransferTx(gasPrice, gasLimit uint64, asset, from, to string, amount uint64) (*types.MutableTransaction, error) {
	fromAddr, err := common.AddressFromBase58(from)
	if err != nil {
		return nil, fmt.Errorf("from address:%s invalid:%s", from, err)
	}
	toAddr, err := common.AddressFromBase58(to)
	if err != nil {
		return nil, fmt.Errorf("to address:%s invalid:%s", to, err)
	}
	var sts []*cntm.State
	sts = append(sts, &cntm.State{
		From:  fromAddr,
		To:    toAddr,
		Value: amount,
	})
	var version byte
	var contractAddr common.Address
	switch strings.ToLower(asset) {
	case ASSET_CNTM:
		version = VERSION_CCNTMRACT_CNTM
		contractAddr = utils.CntmCcntmractAddress
	case ASSET_CNTG:
		version = VERSION_CCNTMRACT_CNTG
		contractAddr = utils.CntgCcntmractAddress
	default:
		return nil, fmt.Errorf("unsupport asset:%s", asset)
	}
	invokeCode, err := cutils.BuildNativeInvokeCode(contractAddr, version, CCNTMRACT_TRANSFER, []interface{}{sts})
	if err != nil {
		return nil, fmt.Errorf("build invoke code error:%s", err)
	}
	mutableTx := NewInvokeTransaction(gasPrice, gasLimit, invokeCode)
	return mutableTx, nil
}

func TransferFromTx(gasPrice, gasLimit uint64, asset, sender, from, to string, amount uint64) (*types.MutableTransaction, error) {
	senderAddr, err := common.AddressFromBase58(sender)
	if err != nil {
		return nil, fmt.Errorf("sender address:%s invalid:%s", to, err)
	}
	fromAddr, err := common.AddressFromBase58(from)
	if err != nil {
		return nil, fmt.Errorf("from address:%s invalid:%s", from, err)
	}
	toAddr, err := common.AddressFromBase58(to)
	if err != nil {
		return nil, fmt.Errorf("to address:%s invalid:%s", to, err)
	}
	transferFrom := &cntm.TransferFrom{
		Sender: senderAddr,
		From:   fromAddr,
		To:     toAddr,
		Value:  amount,
	}
	var version byte
	var contractAddr common.Address
	switch strings.ToLower(asset) {
	case ASSET_CNTM:
		version = VERSION_CCNTMRACT_CNTM
		contractAddr = utils.CntmCcntmractAddress
	case ASSET_CNTG:
		version = VERSION_CCNTMRACT_CNTG
		contractAddr = utils.CntgCcntmractAddress
	default:
		return nil, fmt.Errorf("unsupport asset:%s", asset)
	}
	invokeCode, err := cutils.BuildNativeInvokeCode(contractAddr, version, CCNTMRACT_TRANSFER_FROM, []interface{}{transferFrom})
	if err != nil {
		return nil, fmt.Errorf("build invoke code error:%s", err)
	}
	mutableTx := NewInvokeTransaction(gasPrice, gasLimit, invokeCode)
	return mutableTx, nil
}

//NewInvokeTransaction return smart contract invoke transaction
func NewInvokeTransaction(gasPrice, gasLimit uint64, invokeCode []byte) *types.MutableTransaction {
	invokePayload := &payload.InvokeCode{
		Code: invokeCode,
	}
	tx := &types.MutableTransaction{
		GasPrice: gasPrice,
		GasLimit: gasLimit,
		TxType:   types.InvokeCntm,
		Nonce:    rand.Uint32(),
		Payload:  invokePayload,
		Sigs:     make([]types.Sig, 0, 0),
	}
	return tx
}

func SignTransaction(signer *account.Account, tx *types.MutableTransaction) error {
	if tx.Payer == common.ADDRESS_EMPTY {
		tx.Payer = signer.Address
	}
	txHash := tx.Hash()
	sigData, err := Sign(txHash.ToArray(), signer)
	if err != nil {
		return fmt.Errorf("sign error:%s", err)
	}
	hasSig := false
	for i, sig := range tx.Sigs {
		if len(sig.PubKeys) == 1 && pubKeysEqual(sig.PubKeys, []keypair.PublicKey{signer.PublicKey}) {
			if hasAlreadySig(txHash.ToArray(), signer.PublicKey, sig.SigData) {
				//has already signed
				return nil
			}
			hasSig = true
			//replace
			tx.Sigs[i].SigData = [][]byte{sigData}
		}
	}
	if !hasSig {
		tx.Sigs = append(tx.Sigs, types.Sig{
			PubKeys: []keypair.PublicKey{signer.PublicKey},
			M:       1,
			SigData: [][]byte{sigData},
		})
	}
	return nil
}

func MultiSigTransaction(mutTx *types.MutableTransaction, m uint16, pubKeys []keypair.PublicKey, signer *account.Account) error {
	pkSize := len(pubKeys)
	if m == 0 || int(m) > pkSize || pkSize > constants.MULTI_SIG_MAX_PUBKEY_SIZE {
		return fmt.Errorf("invalid params")
	}
	validPubKey := false
	for _, pk := range pubKeys {
		if keypair.ComparePublicKey(pk, signer.PublicKey) {
			validPubKey = true
			break
		}
	}
	if !validPubKey {
		return fmt.Errorf("invalid signer")
	}
	if mutTx.Payer == common.ADDRESS_EMPTY {
		payer, err := types.AddressFromMultiPubKeys(pubKeys, int(m))
		if err != nil {
			return fmt.Errorf("AddressFromMultiPubKeys error:%s", err)
		}
		mutTx.Payer = payer
	}

	if len(mutTx.Sigs) == 0 {
		mutTx.Sigs = make([]types.Sig, 0)
	}

	txHash := mutTx.Hash()
	sigData, err := Sign(txHash.ToArray(), signer)
	if err != nil {
		return fmt.Errorf("sign error:%s", err)
	}

	hasMutilSig := false
	for i, sigs := range mutTx.Sigs {
		if !pubKeysEqual(sigs.PubKeys, pubKeys) {
			ccntminue
		}
		hasMutilSig = true
		if hasAlreadySig(txHash.ToArray(), signer.PublicKey, sigs.SigData) {
			break
		}
		sigs.SigData = append(sigs.SigData, sigData)
		mutTx.Sigs[i] = sigs
		break
	}
	if !hasMutilSig {
		mutTx.Sigs = append(mutTx.Sigs, types.Sig{
			PubKeys: pubKeys,
			M:       uint16(m),
			SigData: [][]byte{sigData},
		})
	}
	return nil
}

func hasAlreadySig(data []byte, pk keypair.PublicKey, sigDatas [][]byte) bool {
	for _, sigData := range sigDatas {
		err := signature.Verify(pk, data, sigData)
		if err == nil {
			return true
		}
	}
	return false
}

func pubKeysEqual(pks1, pks2 []keypair.PublicKey) bool {
	if len(pks1) != len(pks2) {
		return false
	}
	size := len(pks1)
	if size == 0 {
		return true
	}
	pkstr1 := make([]string, 0, size)
	for _, pk := range pks1 {
		pkstr1 = append(pkstr1, hex.EncodeToString(keypair.SerializePublicKey(pk)))
	}
	pkstr2 := make([]string, 0, size)
	for _, pk := range pks2 {
		pkstr2 = append(pkstr2, hex.EncodeToString(keypair.SerializePublicKey(pk)))
	}
	sort.Strings(pkstr1)
	sort.Strings(pkstr2)
	for i := 0; i < size; i++ {
		if pkstr1[i] != pkstr2[i] {
			return false
		}
	}
	return true
}

//Sign sign return the signature to the data of private key
func Sign(data []byte, signer *account.Account) ([]byte, error) {
	s, err := sig.Sign(signer.SigScheme, signer.PrivateKey, data, nil)
	if err != nil {
		return nil, err
	}
	sigData, err := sig.Serialize(s)
	if err != nil {
		return nil, fmt.Errorf("sig.Serialize error:%s", err)
	}
	return sigData, nil
}

//SendRawTransaction send a transaction to cntm network, and return hash of the transaction
func SendRawTransaction(tx *types.Transaction) (string, error) {
	txData := hex.EncodeToString(common.SerializeToBytes(tx))
	return SendRawTransactionData(txData)
}

func SendRawTransactionData(txData string) (string, error) {
	data, cntmErr := sendRpcRequest("sendrawtransaction", []interface{}{txData})
	if cntmErr != nil {
		return "", cntmErr.Error
	}
	hexHash := ""
	err := json.Unmarshal(data, &hexHash)
	if err != nil {
		return "", fmt.Errorf("json.Unmarshal hash:%s error:%s", data, err)
	}
	return hexHash, nil
}

func PrepareSendRawTransaction(txData string) (*rpccommon.PreExecuteResult, error) {
	data, cntmErr := sendRpcRequest("sendrawtransaction", []interface{}{txData, 1})
	if cntmErr != nil {
		return nil, cntmErr.Error
	}
	preResult := &rpccommon.PreExecuteResult{}
	err := json.Unmarshal(data, &preResult)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal PreExecResult:%s error:%s", data, err)
	}
	return preResult, nil
}

//GetSmartCcntmractEvent return smart contract event execute by invoke transaction by hex string code
func GetSmartCcntmractEvent(txHash string) (*rpccommon.ExecuteNotify, error) {
	data, cntmErr := sendRpcRequest("getsmartcodeevent", []interface{}{txHash})
	if cntmErr != nil {
		switch cntmErr.ErrorCode {
		case ERROR_INVALID_PARAMS:
			return nil, fmt.Errorf("invalid TxHash:%s", txHash)
		}
		return nil, cntmErr.Error
	}
	notifies := &rpccommon.ExecuteNotify{}
	err := json.Unmarshal(data, &notifies)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal SmartCcntmactEvent:%s error:%s", data, err)
	}
	return notifies, nil
}

func GetSmartCcntmractEventInfo(txHash string) ([]byte, error) {
	data, cntmErr := sendRpcRequest("getsmartcodeevent", []interface{}{txHash})
	if cntmErr == nil {
		return data, nil
	}
	switch cntmErr.ErrorCode {
	case ERROR_INVALID_PARAMS:
		return nil, fmt.Errorf("invalid TxHash:%s", txHash)
	}
	return nil, cntmErr.Error
}

func GetRawTransaction(txHash string) ([]byte, error) {
	data, cntmErr := sendRpcRequest("getrawtransaction", []interface{}{txHash, 1})
	if cntmErr == nil {
		return data, nil
	}
	switch cntmErr.ErrorCode {
	case ERROR_INVALID_PARAMS:
		return nil, fmt.Errorf("invalid TxHash:%s", txHash)
	}
	return nil, cntmErr.Error
}

func GetBlock(hashOrHeight interface{}) ([]byte, error) {
	data, cntmErr := sendRpcRequest("getblock", []interface{}{hashOrHeight, 1})
	if cntmErr == nil {
		return data, nil
	}
	switch cntmErr.ErrorCode {
	case ERROR_INVALID_PARAMS:
		return nil, fmt.Errorf("invalid block hash or block height:%v", hashOrHeight)
	}
	return nil, cntmErr.Error
}

func GetNetworkId() (uint32, error) {
	data, cntmErr := sendRpcRequest("getnetworkid", []interface{}{})
	if cntmErr != nil {
		return 0, cntmErr.Error
	}
	var networkId uint32
	err := json.Unmarshal(data, &networkId)
	if err != nil {
		return 0, fmt.Errorf("json.Unmarshal networkId error:%s", err)
	}
	return networkId, nil
}

func GetBlockData(hashOrHeight interface{}) ([]byte, error) {
	data, cntmErr := sendRpcRequest("getblock", []interface{}{hashOrHeight})
	if cntmErr != nil {
		switch cntmErr.ErrorCode {
		case ERROR_INVALID_PARAMS:
			return nil, fmt.Errorf("invalid block hash or block height:%v", hashOrHeight)
		}
		return nil, cntmErr.Error
	}
	hexStr := ""
	err := json.Unmarshal(data, &hexStr)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal error:%s", err)
	}
	blockData, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, fmt.Errorf("hex.DecodeString error:%s", err)
	}
	return blockData, nil
}

func GetCrossChainMsg(height uint32) ([]byte, error) {
	data, cntmErr := sendRpcRequest("getcrosschainmsg", []interface{}{height})
	if cntmErr != nil {
		switch cntmErr.ErrorCode {
		case ERROR_INVALID_PARAMS:
			return nil, fmt.Errorf("invalid block hash or block height:%d", height)
		}
		return nil, cntmErr.Error
	}
	hexStr := ""
	err := json.Unmarshal(data, &hexStr)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal error:%s", err)
	}
	crossChainMsg, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, fmt.Errorf("hex.DecodeString error:%s", err)
	}
	return crossChainMsg, nil
}

func GetBlockCount() (uint32, error) {
	data, cntmErr := sendRpcRequest("getblockcount", []interface{}{})
	if cntmErr != nil {
		return 0, cntmErr.Error
	}
	num := uint32(0)
	err := json.Unmarshal(data, &num)
	if err != nil {
		return 0, fmt.Errorf("json.Unmarshal:%s error:%s", data, err)
	}
	return num, nil
}

func GetTxHeight(txHash string) (uint32, error) {
	data, cntmErr := sendRpcRequest("getblockheightbytxhash", []interface{}{txHash})
	if cntmErr != nil {
		switch cntmErr.ErrorCode {
		case ERROR_INVALID_PARAMS:
			return 0, fmt.Errorf("cannot find tx by:%s", txHash)
		}
		return 0, cntmErr.Error
	}
	height := uint32(0)
	err := json.Unmarshal(data, &height)
	if err != nil {
		return 0, fmt.Errorf("json.Unmarshal error:%s", err)
	}
	return height, nil
}

func DeployCcntmract(
	gasPrice,
	gasLimit uint64,
	signer *account.Account,
	vmtype payload.VmType,
	code,
	cname,
	cversion,
	cauthor,
	cemail,
	cdesc string) (string, error) {

	c, err := hex.DecodeString(code)
	if err != nil {
		return "", fmt.Errorf("hex.DecodeString error:%s", err)
	}
	mutable, err := NewDeployCodeTransaction(gasPrice, gasLimit, c, vmtype, cname, cversion, cauthor, cemail, cdesc)
	if err != nil {
		return "", err
	}

	err = SignTransaction(signer, mutable)
	if err != nil {
		return "", err
	}
	tx, err := mutable.IntoImmutable()
	if err != nil {
		return "", fmt.Errorf("convert to immutable transation error:%v", err)
	}
	txHash, err := SendRawTransaction(tx)
	if err != nil {
		return "", fmt.Errorf("SendRawTransaction error:%s", err)
	}
	return txHash, nil
}

func PrepareDeployCcntmract(
	vmtype payload.VmType,
	code,
	cname,
	cversion,
	cauthor,
	cemail,
	cdesc string) (*httpcom.PreExecuteResult, error) {
	c, err := hex.DecodeString(code)
	if err != nil {
		return nil, fmt.Errorf("hex.DecodeString error:%s", err)
	}
	mutable, err := NewDeployCodeTransaction(0, 0, c, vmtype, cname, cversion, cauthor, cemail, cdesc)
	if err != nil {
		return nil, fmt.Errorf("NewDeployCodeTransaction error:%s", err)
	}
	tx, err := mutable.IntoImmutable()
	if err != nil {
		return nil, err
	}
	txData := hex.EncodeToString(common.SerializeToBytes(tx))
	return PrepareSendRawTransaction(txData)
}

//Invoke cntm vm smart contract. if isPreExec is true, the invoke will not really execute
func InvokeCntmVMCcntmract(
	gasPrice,
	gasLimit uint64,
	signer *account.Account,
	smartcodeAddress common.Address,
	params []interface{}) (string, error) {
	tx, err := httpcom.NewCntmvmInvokeTransaction(gasPrice, gasLimit, smartcodeAddress, params)
	if err != nil {
		return "", err
	}
	return InvokeSmartCcntmract(signer, tx)
}

//Invoke wasm vm smart contract. if isPreExec is true, the invoke will not really execute
func InvokeWasmVMCcntmract(
	gasPrice,
	gasLimit uint64,
	signer *account.Account,
	smartcodeAddress common.Address,
	params []interface{}) (string, error) {
	tx, err := cutils.NewWasmVMInvokeTransaction(gasPrice, gasLimit, smartcodeAddress, params)
	if err != nil {
		return "", err
	}
	return InvokeSmartCcntmract(signer, tx)
}

//InvokeSmartCcntmract is low level method to invoke ccntmact.
func InvokeSmartCcntmract(signer *account.Account, tx *types.MutableTransaction) (string, error) {
	err := SignTransaction(signer, tx)
	if err != nil {
		return "", fmt.Errorf("SignTransaction error:%s", err)
	}
	immut, err := tx.IntoImmutable()
	if err != nil {
		return "", err
	}
	txHash, err := SendRawTransaction(immut)
	if err != nil {
		return "", fmt.Errorf("SendTransaction error:%s", err)
	}
	return txHash, nil
}

func PrepareInvokeCntmVMCcntmract(
	contractAddress common.Address,
	params []interface{},
) (*rpccommon.PreExecuteResult, error) {
	mutable, err := httpcom.NewCntmvmInvokeTransaction(0, 0, contractAddress, params)
	if err != nil {
		return nil, err
	}

	tx, err := mutable.IntoImmutable()
	if err != nil {
		return nil, err
	}

	txData := hex.EncodeToString(common.SerializeToBytes(tx))
	return PrepareSendRawTransaction(txData)
}

func PrepareInvokeCodeCntmVMCcntmract(code []byte) (*rpccommon.PreExecuteResult, error) {
	mutable, err := httpcom.NewSmartCcntmractTransaction(0, 0, code)
	if err != nil {
		return nil, err
	}
	tx, err := mutable.IntoImmutable()
	if err != nil {
		return nil, err
	}
	txData := hex.EncodeToString(common.SerializeToBytes(tx))
	return PrepareSendRawTransaction(txData)
}

//prepare invoke wasm
func PrepareInvokeWasmVMCcntmract(contractAddress common.Address, params []interface{}) (*rpccommon.PreExecuteResult, error) {
	mutable, err := cutils.NewWasmVMInvokeTransaction(0, 0, contractAddress, params)
	if err != nil {
		return nil, err
	}

	tx, err := mutable.IntoImmutable()
	if err != nil {
		return nil, err
	}

	txData := hex.EncodeToString(common.SerializeToBytes(tx))
	return PrepareSendRawTransaction(txData)
}

func PrepareInvokeNativeCcntmract(
	contractAddress common.Address,
	version byte,
	method string,
	params []interface{}) (*httpcom.PreExecuteResult, error) {
	mutable, err := httpcom.NewNativeInvokeTransaction(0, 0, contractAddress, version, method, params)
	if err != nil {
		return nil, err
	}
	tx, err := mutable.IntoImmutable()
	if err != nil {
		return nil, err
	}
	txData := hex.EncodeToString(common.SerializeToBytes(tx))
	return PrepareSendRawTransaction(txData)
}

//NewDeployCodeTransaction return a smart contract deploy transaction instance
func NewDeployCodeTransaction(gasPrice, gasLimit uint64, code []byte, vmType payload.VmType,
	cname, cversion, cauthor, cemail, cdesc string) (*types.MutableTransaction, error) {

	deployPayload, err := payload.NewDeployCode(code, vmType, cname, cversion, cauthor, cemail, cdesc)
	if err != nil {
		return nil, err
	}
	tx := &types.MutableTransaction{
		Version:  VERSION_TRANSACTION,
		TxType:   types.Deploy,
		Nonce:    uint32(time.Now().Unix()),
		Payload:  deployPayload,
		GasPrice: gasPrice,
		GasLimit: gasLimit,
		Sigs:     make([]types.Sig, 0, 0),
	}
	return tx, nil
}

//ParseCntmVMCcntmractReturnTypeBool return bool value of smart contract execute code.
func ParseCntmVMCcntmractReturnTypeBool(hexStr string) (bool, error) {
	return hexStr == "01", nil
}

//ParseCntmVMCcntmractReturnTypeInteger return integer value of smart contract execute code.
func ParseCntmVMCcntmractReturnTypeInteger(hexStr string) (int64, error) {
	data, err := hex.DecodeString(hexStr)
	if err != nil {
		return 0, fmt.Errorf("hex.DecodeString error:%s", err)
	}
	return common.BigIntFromCntmBytes(data).Int64(), nil
}

//ParseCntmVMCcntmractReturnTypeByteArray return []byte value of smart contract execute code.
func ParseCntmVMCcntmractReturnTypeByteArray(hexStr string) (string, error) {
	return hexStr, nil
}

//ParseCntmVMCcntmractReturnTypeString return string value of smart contract execute code.
func ParseCntmVMCcntmractReturnTypeString(hexStr string) (string, error) {
	data, err := hex.DecodeString(hexStr)
	if err != nil {
		return "", fmt.Errorf("hex.DecodeString:%s error:%s", hexStr, err)
	}
	return string(data), nil
}

func ParseWasmVMCcntmractReturnTypeByteArray(hexStr string) (string, error) {
	hexbs, err := common.HexToBytes(hexStr)
	if err != nil {
		return "", fmt.Errorf("common.HexToBytes:%s error:%s", hexStr, err)
	}
	source := common.NewZeroCopySource(hexbs)
	bs, _, irregular, eof := source.NextVarBytes()
	if irregular {
		return "", fmt.Errorf("ParseWasmVMCcntmractReturnTypeByteArray:%s error:%s", hexStr, common.ErrIrregularData)
	}
	if eof {
		return "", fmt.Errorf("ParseWasmVMCcntmractReturnTypeByteArray:%s error:%s", hexStr, io.ErrUnexpectedEOF)
	}
	return common.ToHexString(bs), nil
}

//ParseWasmVMCcntmractReturnTypeString return string value of smart contract execute code.
func ParseWasmVMCcntmractReturnTypeString(hexStr string) (string, error) {
	hexbs, err := common.HexToBytes(hexStr)
	if err != nil {
		return "", fmt.Errorf("common.HexToBytes:%s error:%s", hexStr, err)
	}
	source := common.NewZeroCopySource(hexbs)
	data, _, irregular, eof := source.NextString()
	if irregular {
		return "", common.ErrIrregularData
	}
	if eof {
		return "", io.ErrUnexpectedEOF
	}
	return data, nil
}

//ParseWasmVMCcntmractReturnTypeInteger return integer value of smart contract execute code.
func ParseWasmVMCcntmractReturnTypeInteger(hexStr string) (int64, error) {
	hexbs, err := common.HexToBytes(hexStr)
	if err != nil {
		return 0, fmt.Errorf("common.HexToBytes:%s error:%s", hexStr, err)
	}
	bf := bytes.NewBuffer(hexbs)
	res, err := serialization.ReadUint64(bf)
	return int64(res), err
}

//ParseWasmVMCcntmractReturnTypeBool return bool value of smart contract execute code.
func ParseWasmVMCcntmractReturnTypeBool(hexStr string) (bool, error) {
	hexbs, err := common.HexToBytes(hexStr)
	if err != nil {
		return false, fmt.Errorf("common.HexToBytes:%s error:%s", hexStr, err)
	}
	bf := bytes.NewBuffer(hexbs)
	return serialization.ReadBool(bf)
}
