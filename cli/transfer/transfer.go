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

package transfer

import (
	"github.com/urfave/cli"
	"fmt"
	"os"
	"github.com/Ontology/http/base/rpc"
	. "github.com/Ontology/cli/common"
	cutils "github.com/Ontology/core/utils"
	vmtypes "github.com/Ontology/vm/types"
	ctypes "github.com/Ontology/core/types"
	"github.com/Ontology/smartccntmract/service/native/states"
	"github.com/Ontology/crypto"
	"math/big"
	"github.com/Ontology/common"
	"bytes"
	"github.com/Ontology/account"
	"encoding/json"
	"encoding/hex"
	"time"
)


func transferAction(c *cli.Ccntmext) error {
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}
	// wallet name is wallet.dat by default
	ccntmract := c.String("ccntmract")
	if ccntmract == "" {
		fmt.Println("Invalid ccntmract address: ", ccntmract)
		os.Exit(1)
	}
	ct, _ := common.HexToBytes(ccntmract)
	ctu, _ := common.AddressParseFromBytes(ct)
	from := c.String("from")
	if from == "" {
		fmt.Println("Invalid sender address: ", from)
		os.Exit(1)
	}
	f, _ := common.HexToBytes(from)
	fu, _ := common.AddressParseFromBytes(f)
	to := c.String("to")
	if to == "" {
		fmt.Println("Invalid revicer address: ", to)
		os.Exit(1)
	}
	t, _ := common.HexToBytes(to)
	tu, _ := common.AddressParseFromBytes(t)
	value := c.Int64("value")
	if value <= 0 {
		fmt.Println("Invalid cntm amount: ", value)
		os.Exit(1)
	}

	var sts []*states.State
	sts = append(sts, &states.State{
		From: fu,
		To: tu,
		Value: big.NewInt(value),
	})
	transfers := &states.Transfers{
		States: sts,
	}
	bf := new(bytes.Buffer)

	if err := transfers.Serialize(bf); err != nil {
		fmt.Println("Serialize transfers struct error.")
		os.Exit(1)
	}

	ccntm := &states.Ccntmract{
		Address: ctu,
		Method: "transfer",
		Args: bf.Bytes(),
	}

	ff := new(bytes.Buffer)

	if err := ccntm.Serialize(ff); err != nil {
		fmt.Println("Serialize ccntmract struct error.")
		os.Exit(1)
	}

	tx := cutils.NewInvokeTransaction(vmtypes.VmCode{
		VmType: vmtypes.Native,
		Code: ff.Bytes(),
	})

	tx.Nonce = uint32(time.Now().Unix())

	acct := account.Open(account.WalletFileName, []byte("passwordtest"))
	acc, err := acct.GetDefaultAccount(); if err != nil {
		fmt.Println("GetDefaultAccount error:", err)
		os.Exit(1)
	}
	if err := signTransaction(acc, tx); err != nil {
		fmt.Println("signTransaction error:", err)
		os.Exit(1)
	}

	txbf := new(bytes.Buffer)
	if err := tx.Serialize(txbf); err != nil {
		fmt.Println("Serialize transaction error.")
		os.Exit(1)
	}
	resp, err := rpc.Call(RpcAddress(), "sendrawtransaction", 0,
		[]interface{}{hex.EncodeToString(txbf.Bytes())})

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	r := make(map[string]interface{})
	err = json.Unmarshal(resp, &r)
	if err != nil {
		fmt.Println("Unmarshal JSON failed")
		os.Exit(1)
	}
	switch r["result"].(type) {
	case map[string]interface{}:

	case string:
		fmt.Println(r["result"].(string))
		os.Exit(1)
	}
	return nil
}

func signTransaction(signer *account.Account, tx *ctypes.Transaction) error {
	hash := tx.Hash()
	sign, _ := crypto.Sign(signer.PrivateKey, hash[:])
	tx.Sigs = append(tx.Sigs, &ctypes.Sig{
		PubKeys: []*crypto.PubKey{signer.PublicKey},
		M: 1,
		SigData: [][]byte{sign},
	})
	return nil
}

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:        "transfer",
		Usage:       "user cntm transfer",
		Description: "With nodectl transfer, you could transfer cntm.",
		ArgsUsage:   "[args]",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "ccntmract, c",
				Usage: "ccntmract address",
			},
			cli.StringFlag{
				Name:  "from, f",
				Usage: "sender address",
			},
			cli.StringFlag{
				Name:  "to, t",
				Usage: "revicer address",
			},
			cli.Int64Flag{
				Name:  "value, v",
				Usage: "cntm amount",
			},
		},
		Action: transferAction,
		OnUsageError: func(c *cli.Ccntmext, err error, isSubcommand bool) error {
			return cli.NewExitError("", 1)
		},
	}
}

