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
	"fmt"
	cmdcom "github.com/cntmio/cntmology/cmd/common"
	"github.com/cntmio/cntmology/cmd/utils"
	"github.com/cntmio/cntmology/smartccntmract/types"
	"github.com/urfave/cli"
	"io/ioutil"
	"strings"
)

var (
	CcntmractCommand = cli.Command{
		Name:         "ccntmract",
		Action:       utils.MigrateFlags(ccntmractCommand),
		Usage:        "Deploy or invoke smart ccntmract",
		ArgsUsage:    " ",
		OnUsageError: ccntmractUsageError,
		Description:  `Deploy or invoke smart ccntmract`,
		Subcommands: []cli.Command{
			{
				Action:       utils.MigrateFlags(invokeCcntmract),
				Name:         "invoke",
				OnUsageError: invokeUsageError,
				Usage:        "Invoke a deployed smart ccntmract",
				ArgsUsage:    " ",
				Flags: []cli.Flag{
					utils.CcntmractAddrFlag,
					utils.CcntmractVmTypeFlag,
					utils.CcntmractParamsFlag,
					utils.WalletFileFlag,
				},
				Description: ``,
			},
			{
				Action:       utils.MigrateFlags(deployCcntmract),
				OnUsageError: deployUsageError,
				Name:         "deploy",
				Usage:        "Deploy a smart ccntmract to the chain",
				ArgsUsage:    " ",
				Flags: []cli.Flag{
					utils.CcntmractVmTypeFlag,
					utils.CcntmractStorageFlag,
					utils.CcntmractCodeFileFlag,
					utils.CcntmractNameFlag,
					utils.CcntmractVersionFlag,
					utils.CcntmractAuthorFlag,
					utils.CcntmractEmailFlag,
					utils.CcntmractDescFlag,
					utils.WalletFileFlag,
				},
				Description: ``,
			},
		},
	}
)

func ccntmractCommand(ctx *cli.Ccntmext) error {
	cli.ShowSubcommandHelp(ctx)
	return nil
}

func ccntmractUsageError(ccntmext *cli.Ccntmext, err error, isSubcommand bool) error {
	fmt.Println(err.Error(), "\n")
	cli.ShowSubcommandHelp(ccntmext)
	return nil
}

func invokeUsageError(ccntmext *cli.Ccntmext, err error, isSubcommand bool) error {
	fmt.Println(err.Error(), "\n")
	cli.ShowSubcommandHelp(ccntmext)
	return nil
}

//
func invokeCcntmract(ctx *cli.Ccntmext) error {
	//	if !ctx.IsSet(utils.CcntmractAddrFlag.Name) {
	//		return fmt.Errorf("Missing ccntmract address argument.\n")
	//	}
	//
	//	wallet, err := cmdcom.OpenWallet(ctx)
	//	if err != nil {
	//		return fmt.Errorf("OpenWallet error:%s", err)
	//	}
	//
	//	acc := wallet.GetDefaultAccount()
	//	if acc == nil {
	//		return fmt.Errorf("Cannot GetDefaultAccount")
	//	}
	//
	//	vmType := ctx.String(utils.CcntmractVmTypeFlag.Name)
	//	ccntmractAddr := ctx.String(utils.CcntmractAddrFlag.Name)
	//
	//	addr, err := common.AddressFromBase58(ccntmractAddr)
	//	if err != nil {
	//		return fmt.Errorf("Invalid ccntmract address")
	//	}
	//
	//	paramsStr := ctx.String(utils.CcntmractParamsFlag.Name)
	//	ps := strings.Split(paramsStr)
	//
	//
	//	txHash, err := cntmSdk.Rpc.InvokeNeoVMSmartCcntmract(acct, new(big.Int), addr, []interface{}{params})
	//	if err != nil {
	//		fmt.Printf("InvokeSmartCcntmract InvokeNeoVMSmartCcntmract error:%s", err)
	//		return err
	//	} else {
	//		fmt.Printf("invoke transaction hash:%s", common.ToHexString(txHash[:]))
	//	}
	//
	//	//WaitForGenerateBlock
	//	_, err = cntmSdk.Rpc.WaitForGenerateBlock(30*time.Second, 1)
	//	if err != nil {
	//		fmt.Printf("InvokeSmartCcntmract WaitForGenerateBlock error:%s", err)
	//	}
	return nil
}

func getVmType(vmType string) types.VmType {
	switch vmType {
	case "neovm":
		return types.NEOVM
	case "wasm":
		return types.WASMVM
	default:
		return types.NEOVM
	}
}

func deployUsageError(ccntmext *cli.Ccntmext, err error, isSubcommand bool) error {
	fmt.Println(err.Error(), "\n")
	cli.ShowSubcommandHelp(ccntmext)
	return nil
}

func deployCcntmract(ctx *cli.Ccntmext) error {
	if !ctx.IsSet(utils.CcntmractCodeFileFlag.Name) ||
		!ctx.IsSet(utils.CcntmractNameFlag.Name) {
		return fmt.Errorf("Missing code or name argument")
	}

	wallet, err := cmdcom.OpenWallet(ctx)
	if err != nil {
		return fmt.Errorf("OpenWallet error:%s", err)
	}

	acc := wallet.GetDefaultAccount()
	if acc == nil {
		return fmt.Errorf("Cannot get default account")
	}

	store := ctx.Bool(utils.CcntmractStorageFlag.Name)
	vmType := getVmType(ctx.String(utils.CcntmractVmTypeFlag.Name))
	codeFile := ctx.String(utils.CcntmractCodeFileFlag.Name)
	if "" == codeFile {
		return fmt.Errorf("Please specific code file")
	}
	data, err := ioutil.ReadFile(codeFile)
	if err != nil {
		return fmt.Errorf("Read code:%s error:%s", codeFile, err)
	}
	code := strings.TrimSpace(string(data))
	name := ctx.String(utils.CcntmractNameFlag.Name)
	version := ctx.String(utils.CcntmractVersionFlag.Name)
	author := ctx.String(utils.CcntmractAuthorFlag.Name)
	email := ctx.String(utils.CcntmractEmailFlag.Name)
	desc := ctx.String(utils.CcntmractDescFlag.Name)

	txHash, err := utils.DeployCcntmract(acc, vmType, store, code, name, version, author, email, desc)
	if err != nil {
		return fmt.Errorf("DeployCcntmract error:%s", err)
	}
	address := utils.GetCcntmractAddress(code, vmType)
	fmt.Printf("Deploy TxHash:%s\n", txHash)
	fmt.Printf("Ccntmract Address:%x\n", address)
	return nil
}
