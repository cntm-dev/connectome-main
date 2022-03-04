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
	"github.com/urfave/cli"
)

var (
	WalletAddrFlag = cli.StringFlag{
		Name:  "addr",
		Usage: "wallet address string(base58)",
	}

	WalletNameFlag = cli.StringFlag{
		Name:  "name",
		Usage: "wallet name",
	}

	WalletUsedFlag = cli.StringFlag{
		Name:  "wallet",
		Usage: "which wallet will be used",
	}

	ConfigUsedFlag = cli.StringFlag{
		Name:  "config",
		Usage: "which config file will be used",
	}

	// RPC settings
	RPCEnabledFlag = cli.BoolFlag{
		Name:  "rpc",
		Usage: "enable rpc server? true or false",
	}

	WsEnabledFlag = cli.BoolFlag{
		Name:  "ws",
		Usage: "enable websocket server? true or false",
	}

	//information cmd settings
	HashInfoFlag = cli.StringFlag{
		Name:  "hash",
		Usage: "transaction or block hash value",
	}

	HeightInfoFlag = cli.StringFlag{
		Name:  "height",
		Usage: "block height value",
	}

	//send raw transaction
	CcntmractAddrFlag = cli.StringFlag{
		Name:  "caddr",
		Usage: "ccntmract address that will be used when send raw transaction",
	}

	TransactionFromFlag = cli.StringFlag{
		Name:  "from",
		Usage: "address which transfer from",
	}
	TransactionToFlag = cli.StringFlag{
		Name:  "to",
		Usage: "address which transfer to",
	}
	TransactionValueFlag = cli.Int64Flag{
		Name:  "value",
		Usage: "transfer value",
	}
	UserPasswordFlag = cli.StringFlag{
		Name:  "password",
		Usage: "used when transfer",
	}

	DebugLevelFlag = cli.UintFlag{
		Name:  "debuglevel",
		Usage: "debug level(0~6) will be set",
	}

	ConsensusFlag = cli.StringFlag{
		Name:  "consensus",
		Usage: "consensus turn on/off",
	}

	//ccntmract deploy
	CcntmractVmTypeFlag = cli.UintFlag{
		Name:  "type",
		Usage: "ccntmract type ,value: NEOVM | WASM",
	}

	CcntmractStorageFlag = cli.BoolFlag{
		Name:  "store",
		Usage: "does this ccntmract will be stored, value: true or false",
	}

	CcntmractCodeFlag = cli.StringFlag{
		Name:  "code",
		Usage: "directory of smart ccntmract that will be deployed",
	}

	CcntmractNameFlag = cli.StringFlag{
		Name:  "cname",
		Usage: "ccntmract name that will be deployed",
	}

	CcntmractVersionFlag = cli.StringFlag{
		Name:  "cversion",
		Usage: "ccntmract version which will be deployed",
	}

	CcntmractAuthorFlag = cli.StringFlag{
		Name:  "author",
		Usage: "owner of deployed smart ccntmract",
	}

	CcntmractEmailFlag = cli.StringFlag{
		Name:  "email",
		Usage: "owner email when deploy a smart ccntmract",
	}

	CcntmractDescFlag = cli.StringFlag{
		Name:  "desc",
		Usage: "ccntmract description when deploy one",
	}

	CcntmractParamsFlag = cli.StringFlag{
		Name:  "params",
		Usage: "ccntmract parameter needed when invoded",
	}
	NonOptionFlag = cli.StringFlag{
		Name:  "optiion",
		Usage: "this command does not need option, please run directly",
	}
	//account management
	AccountVerboseFlag = cli.BoolFlag{
		Name:  "verbose,v",
		Usage: "Display accounts with details",
	}
	AccountTypeFlag = cli.StringFlag{
		Name:  "type,t",
		Value: "ecdsa",
		Usage: "Specifies the `<key-type>` by signature algorithm",
	}
	AccountKeylenFlag = cli.StringFlag{
		Name:  "bit-length,b",
		Usage: "Specifies the `<bit-length>` of key",
	}
	AccountSigSchemeFlag = cli.StringFlag{
		Name:  "signature-scheme,s",
		Usage: "Specifies the `<scheme>`",
	}
	AccountPassFlag = cli.StringFlag{
		Name:  "password,p",
		Usage: "Specifies the `<password>` for encrypting the private key",
	}
	AccountDefaultFlag = cli.BoolFlag{
		Name:  "default,d",
		Usage: "Set to default account",
	}
	AccountFileFlag = cli.StringFlag{
		Name:  "file,f",
		Usage: "Specifies the `<filename>` of wallet.",
	}
)

func MigrateFlags(action func(ctx *cli.Ccntmext) error) func(*cli.Ccntmext) error {
	return func(ctx *cli.Ccntmext) error {
		for _, name := range ctx.FlagNames() {
			if ctx.IsSet(name) {
				ctx.GlobalSet(name, ctx.String(name))
			}
		}
		return action(ctx)
	}
}
