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
	"github.com/cntmio/cntmology/smartccntmract/types"
	"github.com/urfave/cli"
	"io/ioutil"
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
				Usage:     "Deploy a smart ccntmract to the chain",
				ArgsUsage: " ",
				Flags: []cli.Flag{
					utils.TransactionGasPrice,
					utils.TransactionGasLimit,
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
				Usage:     "Invoke neovm smart ccntmract",
				ArgsUsage: " ",
				Flags: []cli.Flag{
					utils.TransactionGasPrice,
					utils.TransactionGasLimit,
					utils.CcntmractAddrFlag,
					utils.CcntmractParamsFlag,
					utils.CcntmractVersionFlag,
					utils.CcntmractPrepareInvokeFlag,
					utils.CcntmranctReturnTypeFlag,
					utils.WalletFileFlag,
					utils.AccountAddressFlag,
				},
			},
		},
	}
)

func deployCcntmract(ctx *cli.Ccntmext) error {
	if !ctx.IsSet(utils.GetFlagName(utils.CcntmractCodeFileFlag)) ||
		!ctx.IsSet(utils.GetFlagName(utils.CcntmractNameFlag)) {
		return fmt.Errorf("Missing code or name argument")
	}

	singer, err := cmdcom.GetAccount(ctx)
	if err != nil {
		return fmt.Errorf("Get singer account error:%s", err)
	}

	store := ctx.Bool(utils.GetFlagName(utils.CcntmractStorageFlag))
	codeFile := ctx.String(utils.GetFlagName(utils.CcntmractCodeFileFlag))
	if "" == codeFile {
		return fmt.Errorf("Please specific code file")
	}
	code, err := ioutil.ReadFile(codeFile)
	if err != nil {
		return fmt.Errorf("Read code:%s error:%s", codeFile, err)
	}

	name := ctx.String(utils.GetFlagName(utils.CcntmractNameFlag))
	version := ctx.Int(utils.GetFlagName(utils.CcntmractVersionFlag))
	author := ctx.String(utils.GetFlagName(utils.CcntmractAuthorFlag))
	email := ctx.String(utils.GetFlagName(utils.CcntmractEmailFlag))
	desc := ctx.String(utils.GetFlagName(utils.CcntmractDescFlag))

	gasPrice := ctx.Uint64(utils.GetFlagName(utils.TransactionGasPrice))
	gasLimit := ctx.Uint64(utils.GetFlagName(utils.TransactionGasLimit))
	vmType := types.NEOVM
	cversion := fmt.Sprintf("%s", version)

	txHash, err := utils.DeployCcntmract(gasPrice, gasLimit, singer, vmType, store, string(code), name, cversion, author, email, desc)
	if err != nil {
		return fmt.Errorf("DeployCcntmract error:%s", err)
	}
	address := utils.GetCcntmractAddress(string(code), vmType)
	fmt.Printf("Deploy TxHash:%s\n", txHash)
	fmt.Printf("Ccntmract Address:%s\n", address.ToBase58())
	return nil
}

func invokeCcntmract(ctx *cli.Ccntmext) error {
	if !ctx.IsSet(utils.GetFlagName(utils.CcntmractAddrFlag)) {
		return fmt.Errorf("Missing ccntmract address argument.\n")
	}
	ccntmractAddrStr := ctx.String(utils.GetFlagName(utils.CcntmractAddrFlag))
	ccntmractAddr, err := common.AddressFromBase58(ccntmractAddrStr)
	if err != nil {
		return fmt.Errorf("Invalid ccntmract address")
	}
	cversion := byte(ctx.Int(utils.GetFlagName(utils.CcntmractVersionFlag)))

	paramsStr := ctx.String(utils.GetFlagName(utils.CcntmractParamsFlag))
	params, err := utils.ParseParams(paramsStr)
	if err != nil {
		return fmt.Errorf("parseParams error:%s", err)
	}

	singer, err := cmdcom.GetAccount(ctx)
	if err != nil {
		return fmt.Errorf("Get singer account error:%s", err)
	}
	gasPrice := ctx.Uint64(utils.GetFlagName(utils.TransactionGasPrice))
	gasLimit := ctx.Uint64(utils.GetFlagName(utils.TransactionGasLimit))

	paramData, _ := json.Marshal(params)
	fmt.Printf("Invoke:%s Params:%s\n", ccntmractAddr.ToBase58(), paramData)

	if ctx.IsSet(utils.GetFlagName(utils.CcntmractPrepareInvokeFlag)) {
		res, err := utils.PrepareInvokeNeoVMCcntmract(gasPrice, gasLimit, cversion, ccntmractAddr, params)
		if err != nil {
			return fmt.Errorf("PrepareInvokeNeoVMSmartCcntmact error:%s", err)
		}
		rawReturnTypes := ctx.String(utils.GetFlagName(utils.CcntmranctReturnTypeFlag))
		if rawReturnTypes == "" {
			fmt.Printf("Return:%s (raw value)\n", res)
			return nil
		}
		values, err := utils.ParseReturnValue(res, rawReturnTypes)
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
	txHash, err := utils.InvokeNeoVMCcntmract(gasPrice, gasLimit, singer, cversion, ccntmractAddr, params)
	if err != nil {
		return fmt.Errorf("Invoke NeoVM ccntmract error:%s", err)
	}

	fmt.Printf("TxHash:%s\n", txHash)
	return nil
}
