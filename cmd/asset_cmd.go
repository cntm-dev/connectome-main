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
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/cntmio/cntmology-crypto/keypair"
	"github.com/cntmio/cntmology/account"
	cmdCom "github.com/cntmio/cntmology/cmd/common"
	"github.com/cntmio/cntmology/cmd/utils"
	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/password"
	"github.com/cntmio/cntmology/core/signature"
	ctypes "github.com/cntmio/cntmology/core/types"
	cutils "github.com/cntmio/cntmology/core/utils"
	jrpc "github.com/cntmio/cntmology/http/base/rpc"
	nstates "github.com/cntmio/cntmology/smartccntmract/service/native/states"
	"github.com/cntmio/cntmology/smartccntmract/states"
	vmtypes "github.com/cntmio/cntmology/smartccntmract/types"
	"github.com/urfave/cli"
)

var (
	AssetCommand = cli.Command{
		Name:         "asset",
		Action:       utils.MigrateFlags(assetCommand),
		Usage:        "Handle assets",
		ArgsUsage:    "",
		OnUsageError: assetUsageError,
		Description:  `asset ccntmrol`,
		Subcommands: []cli.Command{
			{
				Action:       utils.MigrateFlags(transferAsset),
				OnUsageError: transferAssetUsageError,
				Name:         "transfer",
				Usage:        "Transfer asset to another account",
				Flags:        append(NodeFlags, CcntmractFlags...),
				Description:  ``,
			},
			{
				Action:       utils.MigrateFlags(queryTransferStatus),
				OnUsageError: transferAssetUsageError,
				Name:         "status",
				Usage:        "Display asset status",
				Flags:        append(append(NodeFlags, CcntmractFlags...), InfoFlags...),
				Description:  ``,
			},
			{
				Action:       cntmBalance,
				OnUsageError: balanceUsageError,
				Name:         "cntm-balance",
				Usage:        "Show balance of cntm and cntm of specified account",
				ArgsUsage:    "[address]",
				Flags: []cli.Flag{
					utils.UserPasswordFlag,
					utils.AccountFileFlag,
				},
			},
		},
	}
)

func assetUsageError(ccntmext *cli.Ccntmext, err error, isSubcommand bool) error {
	fmt.Println(err.Error())
	cli.ShowSubcommandHelp(ccntmext)
	return nil
}

func assetCommand(ctx *cli.Ccntmext) error {
	showAssetHelp()
	return nil
}

func transferAssetUsageError(ccntmext *cli.Ccntmext, err error, isSubcommand bool) error {
	fmt.Println(err.Error())
	showAssetTransferHelp()
	return nil
}

func balanceUsageError(ccntmext *cli.Ccntmext, err error, isSubcommand bool) error {
	fmt.Println(err)
	showAssetTransferHelp()
	return nil
}

func signTransaction(signer *account.Account, tx *ctypes.Transaction) error {
	hash := tx.Hash()
	sign, _ := signature.Sign(signer, hash[:])
	tx.Sigs = append(tx.Sigs, &ctypes.Sig{
		PubKeys: []keypair.PublicKey{signer.PublicKey},
		M:       1,
		SigData: [][]byte{sign},
	})
	return nil
}

func transferAsset(ctx *cli.Ccntmext) error {
	if !ctx.IsSet(utils.CcntmractAddrFlag.Name) || !ctx.IsSet(utils.TransactionFromFlag.Name) || !ctx.IsSet(utils.TransactionToFlag.Name) || !ctx.IsSet(utils.TransactionValueFlag.Name) {
		showAssetTransferHelp()
		return nil
	}
	ccntmract := ctx.GlobalString(utils.CcntmractAddrFlag.Name)
	ct, err := common.HexToBytes(ccntmract)
	if err != nil {
		fmt.Println("Parase ccntmract address error, from hex to bytes")
		return err
	}

	ctu, err := common.AddressParseFromBytes(ct)
	if err != nil {
		fmt.Println("Parase ccntmract address error, please use correct smart ccntmract address")
		return err
	}

	from := ctx.GlobalString(utils.TransactionFromFlag.Name)
	fu, err := common.AddressFromBase58(from)
	if err != nil {
		fmt.Println("Parase transfer-from address error, make sure you are using base58 address")
		return err
	}

	to := ctx.GlobalString(utils.TransactionToFlag.Name)
	tu, err := common.AddressFromBase58(to)
	if err != nil {
		fmt.Println("Parase transfer-to address error, make sure you are using base58 address")
		return err
	}

	value := ctx.Int64(utils.TransactionValueFlag.Name)
	if value <= 0 {
		fmt.Println("Value must be int type and bigger than zero. Invalid cntm amount: ", value)
		return errors.New("Value is invalid")
	}

	var sts []*nstates.State
	sts = append(sts, &nstates.State{
		From:  fu,
		To:    tu,
		Value: big.NewInt(value),
	})
	transfers := &nstates.Transfers{
		States: sts,
	}
	bf := new(bytes.Buffer)

	if err := transfers.Serialize(bf); err != nil {
		fmt.Println("Serialize transfers struct error.")
		return err
	}

	ccntm := &states.Ccntmract{
		Address: ctu,
		Method:  "transfer",
		Args:    bf.Bytes(),
	}

	ff := new(bytes.Buffer)

	if err := ccntm.Serialize(ff); err != nil {
		fmt.Println("Serialize ccntmract struct error.")
		return err
	}

	tx := cutils.NewInvokeTransaction(vmtypes.VmCode{
		VmType: vmtypes.Native,
		Code:   ff.Bytes(),
	})

	tx.Nonce = uint32(time.Now().Unix())

	var passwd []byte
	if ctx.IsSet(utils.UserPasswordFlag.Name) {
		passwd = []byte(ctx.GlobalString(utils.UserPasswordFlag.Name))
	} else {
		passwd, err = password.GetAccountPassword()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return errors.New("input password error")
		}
	}

	var filename string = account.WALLET_FILENAME
	if ctx.IsSet(utils.AccountFileFlag.Name) {
		filename = ctx.String(utils.AccountFileFlag.Name)
	}
	acct := account.Open(filename, passwd)
	if acct != nil {
		return errors.New("open wallet error")
	}
	acc := acct.GetDefaultAccount()
	if acc == nil {
		return errors.New("cannot get the default account")
	}

	if err := signTransaction(acc, tx); err != nil {
		fmt.Println("signTransaction error:", err)
		return err
	}

	txbf := new(bytes.Buffer)
	if err := tx.Serialize(txbf); err != nil {
		fmt.Println("Serialize transaction error.")
		return err
	}

	resp, err := jrpc.Call(rpcAddress(), "sendrawtransaction", 0,
		[]interface{}{hex.EncodeToString(txbf.Bytes())})

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	r := make(map[string]interface{})
	err = json.Unmarshal(resp, &r)
	if err != nil {
		fmt.Println("Unmarshal JSON failed")
		return err
	}

	switch r["result"].(type) {
	case map[string]interface{}:

	case string:
		time.Sleep(10 * time.Second)
		resp, err := cntmSdk.Rpc.GetSmartCcntmractEventWithHexString(r["result"].(string))
		if err != nil {
			fmt.Printf("Please query transfer status manually by hash :%s", r["result"].(string))
			return err
		}
		fmt.Println("\nAsset Transfer Result:")
		cmdCom.EchoJsonDataGracefully(resp)
		return nil
	}

	fmt.Printf("Please query transfer status manually by hash :%s", r["result"].(string))
	return nil
}

func queryTransferStatus(ctx *cli.Ccntmext) error {
	if !ctx.IsSet(utils.HashInfoFlag.Name) {
		showQueryAssetTransferHelp()
	}

	trHash := ctx.GlobalString(utils.HashInfoFlag.Name)
	resp, err := cntmSdk.Rpc.GetSmartCcntmractEventWithHexString(trHash)
	if err != nil {
		fmt.Println("Parase ccntmract address error, from hex to bytes")
		return err
	}
	cmdCom.EchoJsonDataGracefully(resp)
	return nil
}

func cntmBalance(ctx *cli.Ccntmext) error {
	var filename string = account.WALLET_FILENAME
	if ctx.IsSet(utils.AccountFileFlag.Name) {
		filename = ctx.String(utils.AccountFileFlag.Name)
	}

	var base58Addr string
	if ctx.NArg() == 0 {
		var passwd []byte
		var err error
		if ctx.IsSet(utils.UserPasswordFlag.Name) {
			passwd = []byte(ctx.GlobalString(utils.UserPasswordFlag.Name))
		} else {
			passwd, err = password.GetAccountPassword()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return errors.New("input password error")
			}
		}
		acct := account.Open(filename, passwd)
		if acct == nil {
			return errors.New("open wallet error")
		}
		dac := acct.GetDefaultAccount()
		if dac == nil {
			return errors.New("cannot get the default account")
		}
		base58Addr = dac.Address.ToBase58()
	} else {
		base58Addr = ctx.Args().First()
	}
	balance, err := cntmSdk.Rpc.GetBalanceWithBase58(base58Addr)
	if nil != err {
		fmt.Printf("Get Balance with base58 err: %s", err.Error())
		return err
	}
	fmt.Printf("cntm: %d; cntm: %d; cntmAppove: %d\n Address(base58): %s\n", balance.Ont.Int64(), balance.Ong.Int64(), balance.OngAppove.Int64(), base58Addr)
	return nil
}
