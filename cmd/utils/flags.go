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
	ConfigUsedFlag = cli.StringFlag{
		Name:  "config",
		Usage: "Use `<filename>` as the config file",
	}

	// RPC settings
	RPCEnabledFlag = cli.BoolFlag{
		Name:  "rpc",
		Usage: "Enable rpc server",
	}

	WsEnabledFlag = cli.BoolFlag{
		Name:  "ws",
		Usage: "Enable websocket server",
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
		Usage: "ccntmract `<address>` of the asset",
	}

	TransactionFromFlag = cli.StringFlag{
		Name:  "from",
		Usage: "`<address>` which sends the asset",
	}
	TransactionToFlag = cli.StringFlag{
		Name:  "to",
		Usage: "`<address>` which receives the asset",
	}
	TransactionValueFlag = cli.Int64Flag{
		Name:  "value",
		Usage: "Specifies `<value>` as the transferred amount",
	}

	DebugLevelFlag = cli.UintFlag{
		Name:  "debuglevel",
		Usage: "Set the log level to `<level>` (0~6)",
	}

	ConsensusFlag = cli.StringFlag{
		Name:  "consensus",
		Usage: "Turn `<on | off>` the consensus",
	}

	//ccntmract deploy
	CcntmractVmTypeFlag = cli.StringFlag{
		Name:  "type",
		Value: "neovm",
		Usage: "Specifies ccntmract type to one of `<neovm|wasm>`",
	}

	CcntmractStorageFlag = cli.BoolFlag{
		Name:  "store",
		Usage: "Store the ccntmract",
	}

	CcntmractCodeFlag = cli.StringFlag{
		Name:  "code",
		Usage: "Input ccntmracts from `<path>`",
	}

	CcntmractNameFlag = cli.StringFlag{
		Name:  "cname",
		Usage: "Specifies ccntmract name to `<name>`",
	}

	CcntmractVersionFlag = cli.StringFlag{
		Name:  "cversion",
		Usage: "Specifies ccntmract version to `<ver>`",
	}

	CcntmractAuthorFlag = cli.StringFlag{
		Name:  "author",
		Usage: "Set `<address>` as the ccntmract owner",
	}

	CcntmractEmailFlag = cli.StringFlag{
		Name:  "email",
		Usage: "Set `<email>` owner's email address",
	}

	CcntmractDescFlag = cli.StringFlag{
		Name:  "desc",
		Usage: "Set `<text>` as the description of the ccntmract",
	}

	CcntmractParamsFlag = cli.StringFlag{
		Name:  "params",
		Usage: "Specifies ccntmract parameters `<list>` when invoked",
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
		Usage: "Specifies the `<key-type>` by signature algorithm",
	}
	AccountKeylenFlag = cli.StringFlag{
		Name:  "bit-length,b",
		Usage: "Specifies the `<bit-length>` of key",
	}
	AccountSigSchemeFlag = cli.StringFlag{
		Name:  "signature-scheme,s",
		Usage: "Specifies the signature scheme `<scheme>`",
	}
	AccountPassFlag = cli.StringFlag{
		Name:   "password,p",
		Hidden: true,
		Usage:  "Specifies `<password>` for the account",
	}
	AccountDefaultFlag = cli.BoolFlag{
		Name:  "default,d",
		Usage: "Use default settings (equal to '-t ecdsa -b 256 -s SHA256withECDSA')",
	}
	AccountSetDefaultFlag = cli.BoolFlag{
		Name:  "as-default,d",
		Usage: "Set the specified account to default",
	}
	AccountFileFlag = cli.StringFlag{
		Name:  "file,f",
		Value: "wallet.dat",
		Usage: "Use `<filename>` as the wallet",
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
