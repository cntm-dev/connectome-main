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
package sigsvr

import (
	"fmt"

	"github.com/cntmio/cntmology/account"
	"github.com/cntmio/cntmology/cmd"
	"github.com/cntmio/cntmology/cmd/sigsvr/store"
	"github.com/cntmio/cntmology/cmd/utils"
	"github.com/cntmio/cntmology/common"
	"github.com/urfave/cli"
)

var ImportWalletCommand = cli.Command{
	Name:      "import",
	Usage:     "Import accounts from a wallet file",
	ArgsUsage: "",
	Action:    importWallet,
	Flags: []cli.Flag{
		utils.CliWalletDirFlag,
		utils.WalletFileFlag,
	},
	Description: "",
}

func importWallet(ctx *cli.Ccntmext) error {
	walletDirPath := ctx.String(utils.GetFlagName(utils.CliWalletDirFlag))
	walletFilePath := ctx.String(utils.GetFlagName(utils.WalletFileFlag))
	if walletDirPath == "" || walletFilePath == "" {
		cmd.PrintErrorMsg("Missing %s or %s flag.", utils.CliWalletDirFlag.Name, utils.WalletFileFlag.Name)
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	if !common.FileExisted(walletFilePath) {
		return fmt.Errorf("wallet file:%s does not exist", walletFilePath)
	}
	walletStore, err := store.NewWalletStore(walletDirPath)
	if err != nil {
		return fmt.Errorf("NewWalletStore dir path:%s error:%s", walletDirPath, err)
	}
	wallet, err := account.Open(walletFilePath)
	if err != nil {
		return fmt.Errorf("open wallet:%s error:%s", walletFilePath, err)
	}
	walletData := wallet.GetWalletData()
	if *walletStore.WalletScrypt != *walletData.Scrypt {
		return fmt.Errorf("import account failed, wallet scrypt:%+v != %+v", walletData.Scrypt, walletStore.WalletScrypt)
	}
	addNum := 0
	updateNum := 0
	for i := 0; i < len(walletData.Accounts); i++ {
		ok, err := walletStore.AddAccountData(walletData.Accounts[i])
		if err != nil {
			return fmt.Errorf("import account address:%s error:%s", walletData.Accounts[i].Address, err)
		}
		if ok {
			addNum++
		} else {
			updateNum++
		}
	}
	cmd.PrintInfoMsg("Import account success.")
	cmd.PrintInfoMsg("Total account number:%d", len(walletData.Accounts))
	cmd.PrintInfoMsg("Add account number:%d", addNum)
	cmd.PrintInfoMsg("Update account number:%d", updateNum)
	return nil
}
