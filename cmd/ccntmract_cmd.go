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

package cmd

import (
	"encoding/json"
	"fmt"
	cmdcom "github.com/cntmio/cntmology/cmd/common"
	"github.com/cntmio/cntmology/cmd/utils"
	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/config"
	"github.com/cntmio/cntmology/core/payload"
	httpcom "github.com/cntmio/cntmology/http/base/common"
	"github.com/urfave/cli"
	"io/ioutil"
	"strings"
)

var (
	CcntmractCommand = cli.Command{
		Name:        "ccntmract",
		Action:      cli.ShowSubcommandHelp,
		Usage:       "Deploy or invoke smart ccntmract",
		ArgsUsage:   " ",
		Description: `Smart ccntmract operations support the deployment of NeoVM / WasmVM smart ccntmract, and the pre-execution and execution of NeoVM / WasmVM smart ccntmract.`,
		Subcommands: []cli.Command{
			{
				Action:    deployCcntmract,
				Name:      "deploy",
				Usage:     "Deploy a smart ccntmract to cntmology",
				ArgsUsage: " ",
				Flags: []cli.Flag{
					utils.RPCPortFlag,
					utils.TransactionGasPriceFlag,
					utils.TransactionGasLimitFlag,
					utils.CcntmractVmTypeFlag,
					utils.CcntmractCodeFileFlag,
					utils.CcntmractNameFlag,
					utils.CcntmractVersionFlag,
					utils.CcntmractAuthorFlag,
					utils.CcntmractEmailFlag,
					utils.CcntmractDescFlag,
					utils.CcntmractPrepareDeployFlag,
					utils.WalletFileFlag,
					utils.AccountAddressFlag,
				},
			},
			{
				Action: invokeCcntmract,
				Name:   "invoke",
				Usage:  "Invoke smart ccntmract",
				ArgsUsage: `Ontology ccntmract support bytearray(need encode to hex string), string, integer, boolean parameter type.

  Parameter 
     Ccntmract parameters separate with comma ',' to split params. and must add type prefix to params.
     For example:string:foo,int:0,bool:true 
     If parameter is an object array, enclose array with '[]'. 
     For example: string:foo,[int:0,bool:true]

  Note that if string ccntmain some special char like :,[,] and so one, please use '/' char to escape. 
  For example: string:did/:ed1e25c9dccae0c694ee892231407afa20b76008

  Return type
     When invoke ccntmract with --prepare flag, you need specifies return type by --return flag, to decode the return value.
     Return type support bytearray(encoded to hex string), string, integer, boolean. 
     If return type is object array, enclose array with '[]'. 
     For example: [string,int,bool,string]
`,
				Flags: []cli.Flag{
					utils.RPCPortFlag,
					utils.TransactionGasPriceFlag,
					utils.TransactionGasLimitFlag,
					utils.CcntmractAddrFlag,
					utils.CcntmractVmTypeFlag,
					utils.CcntmractParamsFlag,
					utils.CcntmractVersionFlag,
					utils.CcntmractPrepareInvokeFlag,
					utils.CcntmractReturnTypeFlag,
					utils.WalletFileFlag,
					utils.AccountAddressFlag,
				},
			},
			{
				Action:    invokeCodeCcntmract,
				Name:      "invokecode",
				Usage:     "Invoke smart ccntmract by code",
				ArgsUsage: " ",
				Flags: []cli.Flag{
					utils.RPCPortFlag,
					utils.CcntmractCodeFileFlag,
					utils.TransactionGasPriceFlag,
					utils.TransactionGasLimitFlag,
					utils.WalletFileFlag,
					utils.CcntmractPrepareInvokeFlag,
					utils.AccountAddressFlag,
				},
			},
		},
	}
)

func deployCcntmract(ctx *cli.Ccntmext) error {
	SetRpcPort(ctx)
	if !ctx.IsSet(utils.GetFlagName(utils.CcntmractCodeFileFlag)) ||
		!ctx.IsSet(utils.GetFlagName(utils.CcntmractNameFlag)) {
		PrintErrorMsg("Missing %s or %s argument.", utils.CcntmractCodeFileFlag.Name, utils.CcntmractNameFlag.Name)
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	vmtypeFlag := ctx.Uint(utils.GetFlagName(utils.CcntmractVmTypeFlag))
	vmtype, err := payload.VmTypeFromByte(byte(vmtypeFlag))
	if err != nil {
		return err
	}

	codeFile := ctx.String(utils.GetFlagName(utils.CcntmractCodeFileFlag))
	if "" == codeFile {
		return fmt.Errorf("please specific code file")
	}
	codeStr, err := ioutil.ReadFile(codeFile)
	if err != nil {
		return fmt.Errorf("read code:%s error:%s", codeFile, err)
	}

	name := ctx.String(utils.GetFlagName(utils.CcntmractNameFlag))
	version := ctx.String(utils.GetFlagName(utils.CcntmractVersionFlag))
	author := ctx.String(utils.GetFlagName(utils.CcntmractAuthorFlag))
	email := ctx.String(utils.GetFlagName(utils.CcntmractEmailFlag))
	desc := ctx.String(utils.GetFlagName(utils.CcntmractDescFlag))
	code := strings.TrimSpace(string(codeStr))
	gasPrice := ctx.Uint64(utils.GetFlagName(utils.TransactionGasPriceFlag))
	gasLimit := ctx.Uint64(utils.GetFlagName(utils.TransactionGasLimitFlag))
	networkId, err := utils.GetNetworkId()
	if err != nil {
		return err
	}
	if networkId == config.NETWORK_ID_SOLO_NET {
		gasPrice = 0
	}

	cversion := fmt.Sprintf("%s", version)

	if ctx.IsSet(utils.GetFlagName(utils.CcntmractPrepareDeployFlag)) {
		preResult, err := utils.PrepareDeployCcntmract(vmtype, code, name, cversion, author, email, desc)
		if err != nil {
			return fmt.Errorf("PrepareDeployCcntmract error:%s", err)
		}
		if preResult.State == 0 {
			return fmt.Errorf("ccntmract pre-deploy failed")
		}
		PrintInfoMsg("Ccntmract pre-deploy successfully.")
		PrintInfoMsg("Gas consumed:%d.", preResult.Gas)
		return nil
	}

	signer, err := cmdcom.GetAccount(ctx)
	if err != nil {
		return fmt.Errorf("get signer account error:%s", err)
	}

	txHash, err := utils.DeployCcntmract(gasPrice, gasLimit, signer, vmtype, code, name, cversion, author, email, desc)
	if err != nil {
		return fmt.Errorf("DeployCcntmract error:%s", err)
	}
	c, _ := common.HexToBytes(code)
	address := common.AddressFromVmCode(c)
	PrintInfoMsg("Deploy ccntmract:")
	PrintInfoMsg("  Ccntmract Address:%s", address.ToHexString())
	PrintInfoMsg("  TxHash:%s", txHash)
	PrintInfoMsg("\nTip:")
	PrintInfoMsg("  Using './cntmology info status %s' to query transaction status.", txHash)
	return nil
}

func invokeCodeCcntmract(ctx *cli.Ccntmext) error {
	SetRpcPort(ctx)
	if !ctx.IsSet(utils.GetFlagName(utils.CcntmractCodeFileFlag)) {
		PrintErrorMsg("Missing %s or %s argument.", utils.CcntmractCodeFileFlag.Name, utils.CcntmractNameFlag.Name)
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	codeFile := ctx.String(utils.GetFlagName(utils.CcntmractCodeFileFlag))
	if "" == codeFile {
		return fmt.Errorf("please specific code file")
	}
	codeStr, err := ioutil.ReadFile(codeFile)
	if err != nil {
		return fmt.Errorf("read code:%s error:%s", codeFile, err)
	}
	code := strings.TrimSpace(string(codeStr))
	c, err := common.HexToBytes(code)
	if err != nil {
		return fmt.Errorf("ccntmrace code convert hex to bytes error:%s", err)
	}

	if ctx.IsSet(utils.GetFlagName(utils.CcntmractPrepareInvokeFlag)) {
		preResult, err := utils.PrepareInvokeCodeNeoVMCcntmract(c)
		if err != nil {
			return fmt.Errorf("PrepareInvokeCodeNeoVMCcntmract error:%s", err)
		}
		if preResult.State == 0 {
			return fmt.Errorf("ccntmract pre-invoke failed")
		}
		PrintInfoMsg("Ccntmract pre-invoke successfully")
		PrintInfoMsg("  Gas limit:%d", preResult.Gas)

		rawReturnTypes := ctx.String(utils.GetFlagName(utils.CcntmractReturnTypeFlag))
		if rawReturnTypes == "" {
			PrintInfoMsg("Return:%s (raw value)", preResult.Result)
			return nil
		}
		values, err := utils.ParseReturnValue(preResult.Result, rawReturnTypes, payload.NEOVM_TYPE)
		if err != nil {
			return fmt.Errorf("parseReturnValue values:%+v types:%s error:%s", values, rawReturnTypes, err)
		}
		switch len(values) {
		case 0:
			PrintInfoMsg("Return: nil")
		case 1:
			PrintInfoMsg("Return:%+v", values[0])
		default:
			PrintInfoMsg("Return:%+v", values)
		}
		return nil
	}
	gasPrice := ctx.Uint64(utils.GetFlagName(utils.TransactionGasPriceFlag))
	gasLimit := ctx.Uint64(utils.GetFlagName(utils.TransactionGasLimitFlag))
	networkId, err := utils.GetNetworkId()
	if err != nil {
		return err
	}
	if networkId == config.NETWORK_ID_SOLO_NET {
		gasPrice = 0
	}

	invokeTx, err := httpcom.NewSmartCcntmractTransaction(gasPrice, gasLimit, c)
	if err != nil {
		return err
	}

	signer, err := cmdcom.GetAccount(ctx)
	if err != nil {
		return fmt.Errorf("get signer account error:%s", err)
	}

	err = utils.SignTransaction(signer, invokeTx)
	if err != nil {
		return fmt.Errorf("SignTransaction error:%s", err)
	}
	tx, err := invokeTx.IntoImmutable()
	if err != nil {
		return err
	}

	txHash, err := utils.SendRawTransaction(tx)
	if err != nil {
		return fmt.Errorf("SendTransaction error:%s", err)
	}

	PrintInfoMsg("TxHash:%s", txHash)
	PrintInfoMsg("\nTip:")
	PrintInfoMsg("  Using './cntmology info status %s' to query transaction status.", txHash)
	return nil
}

func invokeCcntmract(ctx *cli.Ccntmext) error {
	SetRpcPort(ctx)
	if !ctx.IsSet(utils.GetFlagName(utils.CcntmractAddrFlag)) {
		PrintErrorMsg("Missing %s argument.", utils.CcntmractAddrFlag.Name)
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	ccntmractAddrStr := ctx.String(utils.GetFlagName(utils.CcntmractAddrFlag))
	ccntmractAddr, err := common.AddressFromHexString(ccntmractAddrStr)
	if err != nil {
		return fmt.Errorf("invalid ccntmract address error:%s", err)
	}
	vmtypeFlag := ctx.Uint(utils.GetFlagName(utils.CcntmractVmTypeFlag))
	vmtype, err := payload.VmTypeFromByte(byte(vmtypeFlag))
	if err != nil {
		return err
	}
	paramsStr := ctx.String(utils.GetFlagName(utils.CcntmractParamsFlag))
	params, err := utils.ParseParams(paramsStr)
	if err != nil {
		return fmt.Errorf("parseParams error:%s", err)
	}

	paramData, _ := json.Marshal(params)
	PrintInfoMsg("Invoke:%x Params:%s", ccntmractAddr[:], paramData)
	if ctx.IsSet(utils.GetFlagName(utils.CcntmractPrepareInvokeFlag)) {

		var preResult *httpcom.PreExecuteResult
		if vmtype == payload.NEOVM_TYPE {
			preResult, err = utils.PrepareInvokeNeoVMCcntmract(ccntmractAddr, params)

		}
		if vmtype == payload.WASMVM_TYPE {
			preResult, err = utils.PrepareInvokeWasmVMCcntmract(ccntmractAddr, params)
		}

		if err != nil {
			return fmt.Errorf("PrepareInvokeNeoVMSmartCcntmact error:%s", err)
		}
		if preResult.State == 0 {
			return fmt.Errorf("ccntmract invoke failed")
		}

		PrintInfoMsg("Ccntmract invoke successfully")
		PrintInfoMsg("  Gas limit:%d", preResult.Gas)

		rawReturnTypes := ctx.String(utils.GetFlagName(utils.CcntmractReturnTypeFlag))
		if rawReturnTypes == "" {
			PrintInfoMsg("  Return:%s (raw value)", preResult.Result)
			return nil
		}
		values, err := utils.ParseReturnValue(preResult.Result, rawReturnTypes, vmtype)
		if err != nil {
			return fmt.Errorf("parseReturnValue values:%+v types:%s error:%s", values, rawReturnTypes, err)
		}
		switch len(values) {
		case 0:
			PrintInfoMsg("  Return: nil")
		case 1:
			PrintInfoMsg("  Return:%+v", values[0])
		default:
			PrintInfoMsg("  Return:%+v", values)
		}
		return nil
	}
	signer, err := cmdcom.GetAccount(ctx)
	if err != nil {
		return fmt.Errorf("get signer account error:%s", err)
	}
	gasPrice := ctx.Uint64(utils.GetFlagName(utils.TransactionGasPriceFlag))
	gasLimit := ctx.Uint64(utils.GetFlagName(utils.TransactionGasLimitFlag))
	networkId, err := utils.GetNetworkId()
	if err != nil {
		return err
	}
	if networkId == config.NETWORK_ID_SOLO_NET {
		gasPrice = 0
	}

	var txHash string
	if vmtype == payload.NEOVM_TYPE {
		txHash, err = utils.InvokeNeoVMCcntmract(gasPrice, gasLimit, signer, ccntmractAddr, params)
		if err != nil {
			return fmt.Errorf("invoke NeoVM ccntmract error:%s", err)
		}
	}
	if vmtype == payload.WASMVM_TYPE {
		txHash, err = utils.InvokeWasmVMCcntmract(gasPrice, gasLimit, signer, ccntmractAddr, params)
		if err != nil {
			return fmt.Errorf("invoke NeoVM ccntmract error:%s", err)
		}
	}

	PrintInfoMsg("  TxHash:%s", txHash)
	PrintInfoMsg("\nTips:")
	PrintInfoMsg("  Using './cntmology info status %s' to query transaction status.", txHash)
	return nil
}
