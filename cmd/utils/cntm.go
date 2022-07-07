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

package utils

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/cntmio/cntmology-crypto/keypair"
	sig "github.com/cntmio/cntmology-crypto/signature"
	"github.com/cntmio/cntmology/account"
	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/serialization"
	"github.com/cntmio/cntmology/core/payload"
	"github.com/cntmio/cntmology/core/types"
	httpcom "github.com/cntmio/cntmology/http/base/common"
	rpccommon "github.com/cntmio/cntmology/http/base/common"
	"github.com/cntmio/cntmology/smartccntmract/service/native/cntm"
	"github.com/cntmio/cntmology/smartccntmract/service/native/utils"
	"github.com/cntmio/cntmology/smartccntmract/service/wasmvm"
	cstates "github.com/cntmio/cntmology/smartccntmract/states"
	vmtypes "github.com/cntmio/cntmology/smartccntmract/types"
	"github.com/cntmio/cntmology/vm/neovm"
	neotypes "github.com/cntmio/cntmology/vm/neovm/types"
	"github.com/cntmio/cntmology/vm/wasmvm/exec"
	"math/big"
	"strconv"
	"strings"
	"time"
)

const (
	VERSION_TRANSACTION    = 0
	VERSION_CcntmRACT_cntm   = 0
	VERSION_CcntmRACT_cntm   = 0
	CcntmRACT_TRANSFER      = "transfer"
	CcntmRACT_TRANSFER_FROM = "transferFrom"
	CcntmRACT_APPROVE       = "approve"

	ASSET_cntm = "cntm"
	ASSET_cntm = "cntm"
)

//Return balance of address in base58 code
func GetBalance(address string) (*httpcom.BalanceOfRsp, error) {
	result, err := sendRpcRequest("getbalance", []interface{}{address})
	if err != nil {
		return nil, fmt.Errorf("sendRpcRequest error:%s", err)
	}
	balance := &httpcom.BalanceOfRsp{}
	err = json.Unmarshal(result, balance)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal error:%s", err)
	}
	return balance, nil
}

func GetAllowance(asset, from, to string) (string, error) {
	result, err := sendRpcRequest("getallowance", []interface{}{asset, from, to})
	if err != nil {
		return "", fmt.Errorf("sendRpcRequest error:%s", err)
	}
	balance := ""
	err = json.Unmarshal(result, &balance)
	if err != nil {
		return "", fmt.Errorf("json.Unmarshal error:%s", err)
	}
	return balance, nil
}

//Transfer cntm|cntm from account to another account
func Transfer(gasPrice, gasLimit uint64, signer *account.Account, asset, from, to string, amount uint64) (string, error) {
	transferTx, err := TransferTx(gasPrice, gasLimit, asset, signer.Address.ToBase58(), to, amount)
	if err != nil {
		return "", err
	}
	err = SignTransaction(signer, transferTx)
	if err != nil {
		return "", fmt.Errorf("SignTransaction error:%s", err)
	}
	txHash, err := SendRawTransaction(transferTx)
	if err != nil {
		return "", fmt.Errorf("SendTransaction error:%s", err)
	}
	return txHash, nil
}

func TransferFrom(gasPrice, gasLimit uint64, signer *account.Account, asset, sender, from, to string, amount uint64) (string, error) {
	transferFromTx, err := TransferFromTx(gasPrice, gasLimit, asset, sender, from, to, amount)
	if err != nil {
		return "", err
	}
	err = SignTransaction(signer, transferFromTx)
	if err != nil {
		return "", fmt.Errorf("SignTransaction error:%s", err)
	}
	txHash, err := SendRawTransaction(transferFromTx)
	if err != nil {
		return "", fmt.Errorf("SendTransaction error:%s", err)
	}
	return txHash, nil
}

func Approve(gasPrice, gasLimit uint64, signer *account.Account, asset, from, to string, amount uint64) (string, error) {
	approveTx, err := ApproveTx(gasPrice, gasLimit, asset, from, to, amount)
	if err != nil {
		return "", err
	}
	err = SignTransaction(signer, approveTx)
	if err != nil {
		return "", fmt.Errorf("SignTransaction error:%s", err)
	}
	txHash, err := SendRawTransaction(approveTx)
	if err != nil {
		return "", fmt.Errorf("SendTransaction error:%s", err)
	}
	return txHash, nil
}

func ApproveTx(gasPrice, gasLimit uint64, asset string, from, to string, amount uint64) (*types.Transaction, error) {
	fromAddr, err := common.AddressFromBase58(from)
	if err != nil {
		return nil, fmt.Errorf("To address:%s invalid:%s", from, err)
	}
	toAddr, err := common.AddressFromBase58(to)
	if err != nil {
		return nil, fmt.Errorf("To address:%s invalid:%s", to, err)
	}
	buf := bytes.NewBuffer(nil)
	var state = &cntm.State{
		From:  fromAddr,
		To:    toAddr,
		Value: amount,
	}
	err = state.Serialize(buf)
	if err != nil {
		return nil, fmt.Errorf("transfers.Serialize error %s", err)
	}
	var cversion byte
	var ccntmractAddr common.Address
	switch strings.ToLower(asset) {
	case ASSET_cntm:
		ccntmractAddr = utils.OntCcntmractAddress
		cversion = VERSION_CcntmRACT_cntm
	case ASSET_cntm:
		ccntmractAddr = utils.OngCcntmractAddress
		cversion = VERSION_CcntmRACT_cntm
	default:
		return nil, fmt.Errorf("Unsupport asset:%s", asset)
	}
	return InvokeNativeCcntmractTx(gasPrice, gasLimit, cversion, ccntmractAddr, CcntmRACT_APPROVE, buf.Bytes())
}

func TransferTx(gasPrice, gasLimit uint64, asset, from, to string, amount uint64) (*types.Transaction, error) {
	fromAddr, err := common.AddressFromBase58(from)
	if err != nil {
		return nil, fmt.Errorf("from address:%s invalid:%s", from, err)
	}
	toAddr, err := common.AddressFromBase58(to)
	if err != nil {
		return nil, fmt.Errorf("To address:%s invalid:%s", to, err)
	}
	buf := bytes.NewBuffer(nil)
	var sts []*cntm.State
	sts = append(sts, &cntm.State{
		From:  fromAddr,
		To:    toAddr,
		Value: amount,
	})
	transfers := &cntm.Transfers{
		States: sts,
	}
	err = transfers.Serialize(buf)
	if err != nil {
		return nil, fmt.Errorf("transfers.Serialize error %s", err)
	}
	var cversion byte
	var ccntmractAddr common.Address
	switch strings.ToLower(asset) {
	case ASSET_cntm:
		ccntmractAddr = utils.OntCcntmractAddress
		cversion = VERSION_CcntmRACT_cntm
	case ASSET_cntm:
		ccntmractAddr = utils.OngCcntmractAddress
		cversion = VERSION_CcntmRACT_cntm
	default:
		return nil, fmt.Errorf("Unsupport asset:%s", asset)
	}
	return InvokeNativeCcntmractTx(gasPrice, gasLimit, cversion, ccntmractAddr, CcntmRACT_TRANSFER, buf.Bytes())
}

func TransferFromTx(gasPrice, gasLimit uint64, asset, sender, from, to string, amount uint64) (*types.Transaction, error) {
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
		return nil, fmt.Errorf("To address:%s invalid:%s", to, err)
	}
	transferFrom := &cntm.TransferFrom{
		Sender: senderAddr,
		From:   fromAddr,
		To:     toAddr,
		Value:  amount,
	}
	buf := bytes.NewBuffer(nil)
	err = transferFrom.Serialize(buf)
	if err != nil {
		return nil, fmt.Errorf("transferFrom.Serialize error:%s", err)
	}
	var cversion byte
	var ccntmractAddr common.Address
	switch strings.ToLower(asset) {
	case ASSET_cntm:
		ccntmractAddr = utils.OntCcntmractAddress
		cversion = VERSION_CcntmRACT_cntm
	case ASSET_cntm:
		ccntmractAddr = utils.OngCcntmractAddress
		cversion = VERSION_CcntmRACT_cntm
	default:
		return nil, fmt.Errorf("Unsupport asset:%s", asset)
	}
	return InvokeNativeCcntmractTx(gasPrice, gasLimit, cversion, ccntmractAddr, CcntmRACT_TRANSFER_FROM, buf.Bytes())
}

func SignTransaction(signer *account.Account, tx *types.Transaction) error {
	tx.Payer = signer.Address
	txHash := tx.Hash()
	sigData, err := sign(signer.SigScheme.Name(), txHash.ToArray(), signer)
	if err != nil {
		return fmt.Errorf("sign error:%s", err)
	}
	sig := &types.Sig{
		PubKeys: []keypair.PublicKey{signer.PublicKey},
		M:       1,
		SigData: [][]byte{sigData},
	}
	tx.Sigs = []*types.Sig{sig}
	return nil
}

//Sign sign return the signature to the data of private key
func sign(cryptScheme string, data []byte, signer *account.Account) ([]byte, error) {
	scheme, err := sig.GetScheme(cryptScheme)
	if err != nil {
		return nil, fmt.Errorf("GetScheme by:%s error:%s", cryptScheme, err)
	}
	s, err := sig.Sign(scheme, signer.PrivateKey, data, nil)
	if err != nil {
		return nil, err
	}
	sigData, err := sig.Serialize(s)
	if err != nil {
		return nil, fmt.Errorf("sig.Serialize error:%s", err)
	}
	return sigData, nil
}

//NewInvokeTransaction return smart ccntmract invoke transaction
func NewInvokeTransaction(gasPirce, gasLimit uint64, vmType vmtypes.VmType, code []byte) *types.Transaction {
	invokePayload := &payload.InvokeCode{
		Code: vmtypes.VmCode{
			VmType: vmType,
			Code:   code,
		},
	}
	tx := &types.Transaction{
		Version:    VERSION_TRANSACTION,
		GasPrice:   gasPirce,
		GasLimit:   gasLimit,
		TxType:     types.Invoke,
		Nonce:      uint32(time.Now().Unix()),
		Payload:    invokePayload,
		Attributes: make([]*types.TxAttribute, 0, 0),
		Sigs:       make([]*types.Sig, 0, 0),
	}
	return tx
}

//SendRawTransaction send a transaction to cntmology network, and return hash of the transaction
func SendRawTransaction(tx *types.Transaction) (string, error) {
	var buffer bytes.Buffer
	err := tx.Serialize(&buffer)
	if err != nil {
		return "", fmt.Errorf("Serialize error:%s", err)
	}
	txData := hex.EncodeToString(buffer.Bytes())
	data, err := sendRpcRequest("sendrawtransaction", []interface{}{txData})
	if err != nil {
		return "", err
	}
	hexHash := ""
	err = json.Unmarshal(data, &hexHash)
	if err != nil {
		return "", fmt.Errorf("json.Unmarshal hash:%s error:%s", data, err)
	}
	return hexHash, nil
}

//GetSmartCcntmractEvent return smart ccntmract event execute by invoke transaction by hex string code
func GetSmartCcntmractEvent(txHash string) (*rpccommon.ExecuteNotify, error) {
	data, err := sendRpcRequest("getsmartcodeevent", []interface{}{txHash})
	if err != nil {
		return nil, fmt.Errorf("sendRpcRequest error:%s", err)
	}
	notifies := &rpccommon.ExecuteNotify{}
	err = json.Unmarshal(data, &notifies)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal SmartCcntmactEvent:%s error:%s", data, err)
	}
	return notifies, nil
}

func GetSmartCcntmractEventInfo(txHash string) ([]byte, error) {
	return sendRpcRequest("getsmartcodeevent", []interface{}{txHash})
}

func GetRawTransaction(txHash string) ([]byte, error) {
	return sendRpcRequest("getrawtransaction", []interface{}{txHash, 1})
}

func GetBlock(hashOrHeight interface{}) ([]byte, error) {
	return sendRpcRequest("getblock", []interface{}{hashOrHeight, 1})
}

func DeployCcntmract(
	gasPrice,
	gasLimit uint64,
	signer *account.Account,
	vmType vmtypes.VmType,
	needStorage bool,
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
	tx := NewDeployCodeTransaction(gasPrice, gasLimit, vmType, c, needStorage, cname, cversion, cauthor, cemail, cdesc)

	err = SignTransaction(signer, tx)
	if err != nil {
		return "", err
	}
	txHash, err := SendRawTransaction(tx)
	if err != nil {
		return "", fmt.Errorf("SendRawTransaction error:%s", err)
	}
	return txHash, nil
}

func InvokeNativeCcntmract(
	gasPrice,
	gasLimit uint64,
	signer *account.Account,
	cversion byte,
	ccntmractAddress common.Address,
	method string,
	args []byte,
) (string, error) {
	return InvokeSmartCcntmract(gasPrice, gasLimit, signer, vmtypes.Native, cversion, ccntmractAddress, method, args)
}

func InvokeNativeCcntmractTx(gasPrice,
	gasLimit uint64,
	cversion byte,
	ccntmractAddress common.Address,
	method string,
	args []byte) (*types.Transaction, error) {
	return InvokeSmartCcntmractTx(gasPrice, gasLimit, vmtypes.Native, cversion, ccntmractAddress, method, args)
}

//Invoke wasm smart ccntmract
//methodName is wasm ccntmract action name
//paramType  is Json or Raw format
//version should be greater than 0 (0 is reserved for test)
func InvokeWasmVMCcntmract(
	gasPrice,
	gasLimit uint64,
	siger *account.Account,
	cversion byte, //version of ccntmract
	ccntmractAddress common.Address,
	method string,
	paramType wasmvm.ParamType,
	params []interface{}) (string, error) {

	args, err := buildWasmCcntmractParam(params, paramType)
	if err != nil {
		return "", fmt.Errorf("buildWasmCcntmractParam error:%s", err)
	}
	return InvokeSmartCcntmract(gasPrice, gasLimit, siger, vmtypes.WASMVM, cversion, ccntmractAddress, method, args)
}

//Invoke neo vm smart ccntmract. if isPreExec is true, the invoke will not really execute
func InvokeNeoVMCcntmract(
	gasPrice,
	gasLimit uint64,
	signer *account.Account,
	cversion byte,
	smartcodeAddress common.Address,
	params []interface{}) (string, error) {

	builder := neovm.NewParamsBuilder(new(bytes.Buffer))
	err := buildNeoVMParamInter(builder, params)
	if err != nil {
		return "", err
	}
	args := builder.ToArray()

	return InvokeSmartCcntmract(gasPrice, gasLimit, signer, vmtypes.NEOVM, cversion, smartcodeAddress, "", args)
}

func InvokeNeoVMCcntmractTx(gasPrice,
	gasLimit uint64,
	cversion byte,
	smartcodeAddress common.Address,
	params []interface{}) (*types.Transaction, error) {
	builder := neovm.NewParamsBuilder(new(bytes.Buffer))
	err := buildNeoVMParamInter(builder, params)
	if err != nil {
		return nil, err
	}
	args := builder.ToArray()
	return InvokeSmartCcntmractTx(gasPrice, gasLimit, vmtypes.NEOVM, cversion, smartcodeAddress, "", args)
}

//InvokeSmartCcntmract is low level method to invoke ccntmact.
func InvokeSmartCcntmract(
	gasPrice,
	gasLimit uint64,
	signer *account.Account,
	vmType vmtypes.VmType,
	cversion byte,
	ccntmractAddress common.Address,
	method string,
	args []byte,
) (string, error) {
	invokeTx, err := InvokeSmartCcntmractTx(gasPrice, gasLimit, vmType, cversion, ccntmractAddress, method, args)
	if err != nil {
		return "", err
	}
	err = SignTransaction(signer, invokeTx)
	if err != nil {
		return "", fmt.Errorf("SignTransaction error:%s", err)
	}
	txHash, err := SendRawTransaction(invokeTx)
	if err != nil {
		return "", fmt.Errorf("SendTransaction error:%s", err)
	}
	return txHash, nil
}

func InvokeSmartCcntmractTx(gasPrice,
	gasLimit uint64,
	vmType vmtypes.VmType,
	cversion byte,
	ccntmractAddress common.Address,
	method string,
	args []byte) (*types.Transaction, error) {
	crt := &cstates.Ccntmract{
		Version: cversion,
		Address: ccntmractAddress,
		Method:  method,
		Args:    args,
	}
	buf := bytes.NewBuffer(nil)
	err := crt.Serialize(buf)
	if err != nil {
		return nil, fmt.Errorf("Serialize ccntmract error:%s", err)
	}
	invokCode := buf.Bytes()
	if vmType == vmtypes.NEOVM {
		invokCode = append([]byte{0x67}, invokCode[:]...)
	}
	return NewInvokeTransaction(gasPrice, gasLimit, vmType, invokCode), nil
}

func PrepareInvokeNeoVMCcntmract(
	cversion byte,
	ccntmractAddress common.Address,
	params []interface{},
) (*cstates.PreExecResult, error) {
	code, err := BuildNeoVMInvokeCode(cversion, ccntmractAddress, params)
	if err != nil {
		return nil, fmt.Errorf("BuildNVMInvokeCode error:%s", err)
	}
	tx := NewInvokeTransaction(0, 0, vmtypes.NEOVM, code)
	var buffer bytes.Buffer
	err = tx.Serialize(&buffer)
	if err != nil {
		return nil, fmt.Errorf("Serialize error:%s", err)
	}
	txData := hex.EncodeToString(buffer.Bytes())
	data, err := sendRpcRequest("sendrawtransaction", []interface{}{txData, 1})
	if err != nil {
		return nil, err
	}
	preResult := &cstates.PreExecResult{}
	err = json.Unmarshal(data, &preResult)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal PreExecResult:%s error:%s", data, err)
	}
	return preResult, nil
}

func PrepareInvokeNativeCcntmract(
	ccntmractAddress common.Address,
	code []byte) (*cstates.PreExecResult, error) {
	tx := NewInvokeTransaction(0, 0, vmtypes.Native, code)
	var buffer bytes.Buffer
	err := tx.Serialize(&buffer)
	if err != nil {
		return nil, fmt.Errorf("Serialize error:%s", err)
	}
	txData := hex.EncodeToString(buffer.Bytes())
	data, err := sendRpcRequest("sendrawtransaction", []interface{}{txData, 1})
	if err != nil {
		return nil, err
	}
	preResult := &cstates.PreExecResult{}
	err = json.Unmarshal(data, &preResult)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal PreExecResult:%s error:%s", data, err)
	}
	return preResult, nil
}

//NewDeployCodeTransaction return a smart ccntmract deploy transaction instance
func NewDeployCodeTransaction(
	gasPrice,
	gasLimit uint64,
	vmType vmtypes.VmType,
	code []byte,
	needStorage bool,
	cname, cversion, cauthor, cemail, cdesc string) *types.Transaction {

	vmCode := vmtypes.VmCode{
		VmType: vmType,
		Code:   code,
	}
	deployPayload := &payload.DeployCode{
		Code:        vmCode,
		NeedStorage: needStorage,
		Name:        cname,
		Version:     cversion,
		Author:      cauthor,
		Email:       cemail,
		Description: cdesc,
	}
	tx := &types.Transaction{
		Version:    VERSION_TRANSACTION,
		TxType:     types.Deploy,
		Nonce:      uint32(time.Now().Unix()),
		Payload:    deployPayload,
		Attributes: make([]*types.TxAttribute, 0, 0),
		GasPrice:   gasPrice,
		GasLimit:   gasLimit,
		Sigs:       make([]*types.Sig, 0, 0),
	}
	return tx
}

//buildNeoVMParamInter build neovm invoke param code
func buildNeoVMParamInter(builder *neovm.ParamsBuilder, smartCcntmractParams []interface{}) error {
	//VM load params in reverse order
	for i := len(smartCcntmractParams) - 1; i >= 0; i-- {
		switch v := smartCcntmractParams[i].(type) {
		case bool:
			builder.EmitPushBool(v)
		case int:
			builder.EmitPushInteger(big.NewInt(int64(v)))
		case uint:
			builder.EmitPushInteger(big.NewInt(int64(v)))
		case int32:
			builder.EmitPushInteger(big.NewInt(int64(v)))
		case uint32:
			builder.EmitPushInteger(big.NewInt(int64(v)))
		case int64:
			builder.EmitPushInteger(big.NewInt(int64(v)))
		case common.Fixed64:
			builder.EmitPushInteger(big.NewInt(int64(v.GetData())))
		case uint64:
			val := big.NewInt(0)
			builder.EmitPushInteger(val.SetUint64(uint64(v)))
		case string:
			builder.EmitPushByteArray([]byte(v))
		case *big.Int:
			builder.EmitPushInteger(v)
		case []byte:
			builder.EmitPushByteArray(v)
		case []interface{}:
			err := buildNeoVMParamInter(builder, v)
			if err != nil {
				return err
			}
			builder.EmitPushInteger(big.NewInt(int64(len(v))))
			builder.Emit(neovm.PACK)
		default:
			return fmt.Errorf("unsupported param:%s", v)
		}
	}
	return nil
}

//BuildNeoVMInvokeCode build NeoVM Invoke code for params
func BuildNeoVMInvokeCode(cversion byte, smartCcntmractAddress common.Address, params []interface{}) ([]byte, error) {
	builder := neovm.NewParamsBuilder(new(bytes.Buffer))
	err := buildNeoVMParamInter(builder, params)
	if err != nil {
		return nil, err
	}
	args := builder.ToArray()

	crt := &cstates.Ccntmract{
		Version: cversion,
		Address: smartCcntmractAddress,
		Args:    args,
	}
	crtBuf := bytes.NewBuffer(nil)
	err = crt.Serialize(crtBuf)
	if err != nil {
		return nil, fmt.Errorf("Serialize ccntmract error:%s", err)
	}

	buf := bytes.NewBuffer(nil)
	buf.Write(append([]byte{0x67}, crtBuf.Bytes()[:]...))
	return buf.Bytes(), nil
}

//for wasm vm
//build param bytes for wasm ccntmract
func buildWasmCcntmractParam(params []interface{}, paramType wasmvm.ParamType) ([]byte, error) {
	switch paramType {
	case wasmvm.Json:
		args := make([]exec.Param, len(params))

		for i, param := range params {
			switch param.(type) {
			case string:
				arg := exec.Param{Ptype: "string", Pval: param.(string)}
				args[i] = arg
			case int:
				arg := exec.Param{Ptype: "int", Pval: strconv.Itoa(param.(int))}
				args[i] = arg
			case int64:
				arg := exec.Param{Ptype: "int64", Pval: strconv.FormatInt(param.(int64), 10)}
				args[i] = arg
			case []int:
				bf := bytes.NewBuffer(nil)
				array := param.([]int)
				for i, tmp := range array {
					bf.WriteString(strconv.Itoa(tmp))
					if i != len(array)-1 {
						bf.WriteString(",")
					}
				}
				arg := exec.Param{Ptype: "int_array", Pval: bf.String()}
				args[i] = arg
			case []int64:
				bf := bytes.NewBuffer(nil)
				array := param.([]int64)
				for i, tmp := range array {
					bf.WriteString(strconv.FormatInt(tmp, 10))
					if i != len(array)-1 {
						bf.WriteString(",")
					}
				}
				arg := exec.Param{Ptype: "int_array", Pval: bf.String()}
				args[i] = arg
			default:
				return nil, fmt.Errorf("not a supported type :%v\n", param)
			}
		}

		bs, err := json.Marshal(exec.Args{args})
		if err != nil {
			return nil, err
		}
		return bs, nil
	case wasmvm.Raw:
		bf := bytes.NewBuffer(nil)
		for _, param := range params {
			switch param.(type) {
			case string:
				tmp := bytes.NewBuffer(nil)
				serialization.WriteString(tmp, param.(string))
				bf.Write(tmp.Bytes())

			case int:
				tmpBytes := make([]byte, 4)
				binary.LittleEndian.PutUint32(tmpBytes, uint32(param.(int)))
				bf.Write(tmpBytes)

			case int64:
				tmpBytes := make([]byte, 8)
				binary.LittleEndian.PutUint64(tmpBytes, uint64(param.(int64)))
				bf.Write(tmpBytes)

			default:
				return nil, fmt.Errorf("not a supported type :%v\n", param)
			}
		}
		return bf.Bytes(), nil
	default:
		return nil, fmt.Errorf("unsupported type")
	}
}

//BuildWasmVMInvokeCode return wasn vm invoke code
func BuildWasmVMInvokeCode(smartcodeAddress common.Address, methodName string, paramType wasmvm.ParamType, version byte, params []interface{}) ([]byte, error) {
	ccntmract := &cstates.Ccntmract{}
	ccntmract.Address = smartcodeAddress
	ccntmract.Method = methodName
	ccntmract.Version = version

	argbytes, err := buildWasmCcntmractParam(params, paramType)

	if err != nil {
		return nil, fmt.Errorf("build wasm ccntmract param failed:%s", err)
	}
	ccntmract.Args = argbytes
	bf := bytes.NewBuffer(nil)
	ccntmract.Serialize(bf)
	return bf.Bytes(), nil
}

//GetCcntmractAddress return ccntmract address
func GetCcntmractAddress(code string, vmType vmtypes.VmType) common.Address {
	data, _ := hex.DecodeString(code)
	vmCode := &vmtypes.VmCode{
		VmType: vmType,
		Code:   data,
	}
	return vmCode.AddressFromVmCode()
}

//ParseNeoVMCcntmractReturnTypeBool return bool value of smart ccntmract execute code.
func ParseNeoVMCcntmractReturnTypeBool(hexStr string) (bool, error) {
	return hexStr == "01", nil
}

//ParseNeoVMCcntmractReturnTypeInteger return integer value of smart ccntmract execute code.
func ParseNeoVMCcntmractReturnTypeInteger(hexStr string) (int64, error) {
	data, err := hex.DecodeString(hexStr)
	if err != nil {
		return 0, fmt.Errorf("hex.DecodeString error:%s", err)
	}
	return neotypes.ConvertBytesToBigInteger(data).Int64(), nil
}

//ParseNeoVMCcntmractReturnTypeByteArray return []byte value of smart ccntmract execute code.
func ParseNeoVMCcntmractReturnTypeByteArray(hexStr string) (string, error) {
	return hexStr, nil
}

//ParseNeoVMCcntmractReturnTypeString return string value of smart ccntmract execute code.
func ParseNeoVMCcntmractReturnTypeString(hexStr string) (string, error) {
	data, err := hex.DecodeString(hexStr)
	if err != nil {
		return "", fmt.Errorf("hex.DecodeString:%s error:%s", hexStr, err)
	}
	return string(data), nil
}
