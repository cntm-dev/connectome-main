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
	"github.com/cntmio/cntmology/core/types"
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
		Description: `Deploy or invoke smart ccntmract`,
		Subcommands: []cli.Command{
			{
				Action:    deployCcntmract,
				Name:      "deploy",
				Usage:     "Deploy a smart ccntmract to cntmolgoy",
				ArgsUsage: " ",
				Flags: []cli.Flag{
					utils.RPCPortFlag,
					utils.TransactionGasPriceFlag,
					utils.TransactionGasLimitFlag,
					utils.CcntmractStorageFlag,
					utils.CcntmractCodeFileFlag,
					utils.CcntmractNameFlag,
					utils.CcntmractVersionFlag,
					utils.CcntmractAuthorFlag,
					utils.CcntmractEmailFlag,
					utils.CcntmractDescFlag,
					utils.WalletFileFlag,
					utils.AccountAddressFlag,
				},
			},
			{
				Action:    invokeCcntmract,
				Name:      "invoke",
				Usage:     "Invoke smart ccntmract",
				ArgsUsage: " ",
				Flags: []cli.Flag{
					utils.RPCPortFlag,
					utils.TransactionGasPriceFlag,
					utils.TransactionGasLimitFlag,
					utils.CcntmractAddrFlag,
					utils.CcntmractParamsFlag,
					utils.CcntmractVersionFlag,
					utils.CcntmractPrepareInvokeFlag,
					utils.CcntmranctReturnTypeFlag,
					utils.WalletFileFlag,
					utils.AccountAddressFlag,
				},
			},
			{
				Action:    invokeCodeCcntmract,
				Name:      "invokeCode",
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
		fmt.Errorf("Missing code or name argument\n")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	signer, err := cmdcom.GetAccount(ctx)
	if err != nil {
		return fmt.Errorf("Get signer account error:%s", err)
	}

	store := ctx.Bool(utils.GetFlagName(utils.CcntmractStorageFlag))
	codeFile := ctx.String(utils.GetFlagName(utils.CcntmractCodeFileFlag))
	if "" == codeFile {
		return fmt.Errorf("Please specific code file")
	}
	codeStr, err := ioutil.ReadFile(codeFile)
	if err != nil {
		return fmt.Errorf("Read code:%s error:%s", codeFile, err)
	}

	name := ctx.String(utils.GetFlagName(utils.CcntmractNameFlag))
	version := ctx.Int(utils.GetFlagName(utils.CcntmractVersionFlag))
	author := ctx.String(utils.GetFlagName(utils.CcntmractAuthorFlag))
	email := ctx.String(utils.GetFlagName(utils.CcntmractEmailFlag))
	desc := ctx.String(utils.GetFlagName(utils.CcntmractDescFlag))
	code := strings.TrimSpace(string(codeStr))
	gasPrice := ctx.Uint64(utils.GetFlagName(utils.TransactionGasPriceFlag))
	gasLimit := ctx.Uint64(utils.GetFlagName(utils.TransactionGasLimitFlag))
	cversion := fmt.Sprintf("%s", version)

	txHash, err := utils.DeployCcntmract(gasPrice, gasLimit, signer, store, code, name, cversion, author, email, desc)
	if err != nil {
		return fmt.Errorf("DeployCcntmract error:%s", err)
	}
	c, _ := common.HexToBytes(code)
	address := types.AddressFromVmCode(c)
	fmt.Printf("Deploy ccntmract:\n")
	fmt.Printf("  Ccntmract Address:%s\n", address.ToHexString())
	fmt.Printf("  TxHash:%s\n", txHash)
	fmt.Printf("\nTip:\n")
	fmt.Printf("  Using './cntmology info status %s' to query transaction status\n", txHash)
	return nil
}

func invokeCodeCcntmract(ctx *cli.Ccntmext) error {
	SetRpcPort(ctx)
	if !ctx.IsSet(utils.GetFlagName(utils.CcntmractCodeFileFlag)) {
		fmt.Printf("Missing code or name argument\n")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	signer, err := cmdcom.GetAccount(ctx)
	if err != nil {
		return fmt.Errorf("Get signer account error:%s", err)
	}
	codeFile := ctx.String(utils.GetFlagName(utils.CcntmractCodeFileFlag))
	if "" == codeFile {
		return fmt.Errorf("Please specific code file")
	}
	codeStr, err := ioutil.ReadFile(codeFile)
	if err != nil {
		return fmt.Errorf("Read code:%s error:%s", codeFile, err)
	}
	code := strings.TrimSpace(string(codeStr))
	c, err := common.HexToBytes(code)
	if err != nil {
		return fmt.Errorf("hex to bytes error:%s", err)
	}

	if ctx.IsSet(utils.GetFlagName(utils.CcntmractPrepareInvokeFlag)) {
		preResult, err := utils.PrepareInvokeCodeNeoVMCcntmract(c)
		if err != nil {
			return fmt.Errorf("PrepareInvokeCodeNeoVMCcntmract error:%s", err)
		}
		if preResult.State == 0 {
			return fmt.Errorf("Ccntmract invoke failed\n")
		}
		fmt.Printf("Ccntmract invoke successfully\n")
		fmt.Printf("Gas consumed:%d\n", preResult.Gas)

		rawReturnTypes := ctx.String(utils.GetFlagName(utils.CcntmranctReturnTypeFlag))
		if rawReturnTypes == "" {
			fmt.Printf("Return:%s (raw value)\n", preResult.Result)
			return nil
		}
		values, err := utils.ParseReturnValue(preResult.Result, rawReturnTypes)
		if err != nil {
			return fmt.Errorf("parseReturnValue values:%+v types:%s error:%s", values, rawReturnTypes, err)
		}
		switch len(values) {
		case 0:
			fmt.Printf("Return: nil\n")
		case 1:
			fmt.Printf("Return:%+v\n", values[0])
		default:
			fmt.Printf("Return:%+v\n", values)
		}
		return nil
	}
	gasPrice := ctx.Uint64(utils.GetFlagName(utils.TransactionGasPriceFlag))
	gasLimit := ctx.Uint64(utils.GetFlagName(utils.TransactionGasLimitFlag))

	invokeTx, err := httpcom.NewSmartCcntmractTransaction(gasPrice, gasLimit, c)
	if err != nil {
		return err
	}
	err = utils.SignTransaction(signer, invokeTx)
	if err != nil {
		return fmt.Errorf("SignTransaction error:%s", err)
	}
	txHash, err := utils.SendRawTransaction(invokeTx)
	if err != nil {
		return fmt.Errorf("SendTransaction error:%s", err)
	}

	fmt.Printf("TxHash:%s\n", txHash)
	fmt.Printf("\nTip:\n")
	fmt.Printf("  Using './cntmology info status %s' to query transaction status\n", txHash)
	return nil
}

func invokeCcntmract(ctx *cli.Ccntmext) error {
	SetRpcPort(ctx)
	if !ctx.IsSet(utils.GetFlagName(utils.CcntmractAddrFlag)) {
		fmt.Printf("Missing ccntmract address argument.\n")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	ccntmractAddrStr := ctx.String(utils.GetFlagName(utils.CcntmractAddrFlag))
	ccntmractAddr, err := common.AddressFromHexString(ccntmractAddrStr)
	if err != nil {
		return fmt.Errorf("Invalid ccntmract address error:%s", err)
	}

	paramsStr := ctx.String(utils.GetFlagName(utils.CcntmractParamsFlag))
	params, err := utils.ParseParams(paramsStr)
	if err != nil {
		return fmt.Errorf("parseParams error:%s", err)
	}

	paramData, _ := json.Marshal(params)
	fmt.Printf("Invoke:%x Params:%s\n", ccntmractAddr[:], paramData)

	if ctx.IsSet(utils.GetFlagName(utils.CcntmractPrepareInvokeFlag)) {
		preResult, err := utils.PrepareInvokeNeoVMCcntmract(ccntmractAddr, params)
		if err != nil {
			return fmt.Errorf("PrepareInvokeNeoVMSmartCcntmact error:%s", err)
		}
		if preResult.State == 0 {
			return fmt.Errorf("Ccntmract invoke failed\n")
		}
		fmt.Printf("Ccntmract invoke successfully\n")
		fmt.Printf("Gas consumed:%d\n", preResult.Gas)

		rawReturnTypes := ctx.String(utils.GetFlagName(utils.CcntmranctReturnTypeFlag))
		if rawReturnTypes == "" {
			fmt.Printf("Return:%s (raw value)\n", preResult.Result)
			return nil
		}
		values, err := utils.ParseReturnValue(preResult.Result, rawReturnTypes)
		if err != nil {
			return fmt.Errorf("parseReturnValue values:%+v types:%s error:%s", values, rawReturnTypes, err)
		}
		switch len(values) {
		case 0:
			fmt.Printf("Return: nil\n")
		case 1:
			fmt.Printf("Return:%+v\n", values[0])
		default:
			fmt.Printf("Return:%+v\n", values)
		}
		return nil
	}
	signer, err := cmdcom.GetAccount(ctx)
	if err != nil {
		return fmt.Errorf("Get signer account error:%s", err)
	}
	gasPrice := ctx.Uint64(utils.GetFlagName(utils.TransactionGasPriceFlag))
	gasLimit := ctx.Uint64(utils.GetFlagName(utils.TransactionGasLimitFlag))

	txHash, err := utils.InvokeNeoVMCcntmract(gasPrice, gasLimit, signer, ccntmractAddr, params)
	if err != nil {
		return fmt.Errorf("Invoke NeoVM ccntmract error:%s", err)
	}

	fmt.Printf("TxHash:%s\n", txHash)
	fmt.Printf("\nTip:\n")
	fmt.Printf("  Using './cntmology info status %s' to query transaction status\n", txHash)
	return nil
}
