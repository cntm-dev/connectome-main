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
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"time"

	"github.com/cntmio/cntmology/account"
	"github.com/cntmio/cntmology/cmd/utils"
	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/smartccntmract/types"
	"github.com/urfave/cli"
)

var (
	CcntmractCommand = cli.Command{
		Name:         "ccntmract",
		Action:       utils.MigrateFlags(ccntmractCommand),
		Usage:        "cntmology ccntmract [invoke|deploy] [OPTION]",
		Category:     "CcntmRACT COMMANDS",
		OnUsageError: ccntmractUsageError,
		Description:  `account command`,
		Subcommands: []cli.Command{
			{
				Action:       utils.MigrateFlags(invokeCcntmract),
				Name:         "invoke",
				OnUsageError: invokeUsageError,
				Usage:        "cntmology invoke [OPTION]\n",
				Flags:        append(NodeFlags, CcntmractFlags...),
				Category:     "CcntmRACT COMMANDS",
				Description:  ``,
			},
			{
				Action:       utils.MigrateFlags(deployCcntmract),
				OnUsageError: deployUsageError,
				Name:         "deploy",
				Usage:        "cntmology deploy [OPTION]\n",
				Flags:        append(NodeFlags, CcntmractFlags...),
				Category:     "CcntmRACT COMMANDS",
				Description:  ``,
			},
		},
	}
)

func ccntmractCommand(ctx *cli.Ccntmext) error {
	showCcntmractHelp()
	return nil
}

func ccntmractUsageError(ccntmext *cli.Ccntmext, err error, isSubcommand bool) error {
	fmt.Println(err.Error())
	showCcntmractHelp()
	return nil
}

func invokeUsageError(ccntmext *cli.Ccntmext, err error, isSubcommand bool) error {
	fmt.Println(err.Error())
	showInvokeHelp()
	return nil
}

func invokeCcntmract(ctx *cli.Ccntmext) error {
	if !ctx.IsSet(utils.CcntmractAddrFlag.Name) || !ctx.IsSet(utils.CcntmractParamsFlag.Name) {
		showInvokeHelp()
		return nil
	}

	client := account.GetClient(ctx)
	if client == nil {
		fmt.Println("Can't get local account")
		return errors.New("Get client is nil")
	}

	acct := client.GetDefaultAccount()

	ccntmractAddr := ctx.String(utils.CcntmractAddrFlag.Name)
	params := ctx.String(utils.CcntmractParamsFlag.Name)
	if "" == ccntmractAddr {
		fmt.Println("ccntmract address does not allow empty")
	}

	addr, err := common.AddressFromBase58(ccntmractAddr)
	if err != nil {
		fmt.Println("Parase ccntmract address error, please use correct smart ccntmract address")
		return err
	}

	txHash, err := cntmSdk.Rpc.InvokeNeoVMSmartCcntmract(acct, new(big.Int), addr, []interface{}{params})
	if err != nil {
		fmt.Printf("InvokeSmartCcntmract InvokeNeoVMSmartCcntmract error:%s", err)
		return err
	} else {
		fmt.Printf("invoke transaction hash:%s", common.ToHexString(txHash[:]))
	}

	//WaitForGenerateBlock
	_, err = cntmSdk.Rpc.WaitForGenerateBlock(30*time.Second, 1)
	if err != nil {
		fmt.Printf("InvokeSmartCcntmract WaitForGenerateBlock error:%s", err)
	}
	return err
}

func getVmType(vmType uint) types.VmType {
	switch vmType {
	case 1:
		return types.NEOVM
	case 2:
		return types.WASMVM
	default:
		return types.NEOVM
	}
}

func deployUsageError(ccntmext *cli.Ccntmext, err error, isSubcommand bool) error {
	fmt.Println(err.Error())
	showDeployHelp()
	return nil
}

func deployCcntmract(ctx *cli.Ccntmext) error {
	if !ctx.IsSet(utils.CcntmractStorageFlag.Name) || !ctx.IsSet(utils.CcntmractVmTypeFlag.Name) ||
		!ctx.IsSet(utils.CcntmractCodeFlag.Name) || !ctx.IsSet(utils.CcntmractNameFlag.Name) ||
		!ctx.IsSet(utils.CcntmractVersionFlag.Name) || !ctx.IsSet(utils.CcntmractAuthorFlag.Name) ||
		!ctx.IsSet(utils.CcntmractEmailFlag.Name) || !ctx.IsSet(utils.CcntmractDescFlag.Name) {
		showDeployHelp()
		return errors.New("Parameter is err")
	}

	client := account.GetClient(ctx)
	if nil == client {
		fmt.Println("Can't get local account.")
		return errors.New("Get client return nil")
	}

	acct := client.GetDefaultAccount()

	store := ctx.Bool(utils.CcntmractStorageFlag.Name)
	vmType := getVmType(ctx.Uint(utils.CcntmractVmTypeFlag.Name))

	codeDir := ctx.String(utils.CcntmractCodeFlag.Name)
	if "" == codeDir {
		fmt.Println("Code dir is error, value does not allow null")
		return errors.New("Smart ccntmract code dir does not allow empty")
	}
	code, err := ioutil.ReadFile(codeDir)
	if err != nil {
		fmt.Printf("Error in read file,%s", err.Error())
		return err
	}

	name := ctx.String(utils.CcntmractNameFlag.Name)
	version := ctx.String(utils.CcntmractVersionFlag.Name)
	author := ctx.String(utils.CcntmractAuthorFlag.Name)
	email := ctx.String(utils.CcntmractEmailFlag.Name)
	desc := ctx.String(utils.CcntmractDescFlag.Name)

	trHash, err := cntmSdk.Rpc.DeploySmartCcntmract(acct, vmType, store, fmt.Sprintf("%s", code), name, version, author, email, desc)
	if err != nil {
		fmt.Printf("Deploy smart error: %s", err.Error())
		return err
	}
	//WaitForGenerateBlock
	_, err = cntmSdk.Rpc.WaitForGenerateBlock(30*time.Second, 1)
	if err != nil {
		fmt.Printf("DeploySmartCcntmract WaitForGenerateBlock error:%s", err.Error())
		return err
	} else {
		fmt.Printf("Deploy smartCcntmract transaction hash: %s\n", common.ToHexString(trHash[:]))
	}

	return nil
}
