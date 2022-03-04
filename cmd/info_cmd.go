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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/cntmio/cntmology/cmd/common"
	"github.com/cntmio/cntmology/cmd/utils"
	"github.com/urfave/cli"
)

var (
	InfoCommand = cli.Command{
		Action: utils.MigrateFlags(infoCommand),
		Name:   "info",
		Usage:  "Display informations about the chain",
		Flags:  append(NodeFlags, InfoFlags...),
		Subcommands: []cli.Command{
			blockCommandSet,
			txCommandSet,
			versionCommand,
		},
		Description: ``,
	}
)

func infoCommand(ccntmext *cli.Ccntmext) error {
	showInfoHelp()
	return nil
}

func blockInfoUsageError(ccntmext *cli.Ccntmext, err error, isSubcommand bool) error {
	fmt.Println("Error:", err.Error())
	showBlockInfoHelp()
	return nil
}

func getCurrentBlockHeight(ctx *cli.Ccntmext) error {
	height, err := cntmSdk.Rpc.GetBlockCount()
	if nil != err {
		fmt.Printf("Get block height information is error:  %s", err.Error())
		return err
	}
	fmt.Println("Current blockchain height: ", height)
	return nil
}

var blockCommandSet = cli.Command{
	Action:       utils.MigrateFlags(blockInfoCommand),
	Name:         "block",
	Usage:        "Display block informations",
	Flags:        append(NodeFlags, InfoFlags...),
	OnUsageError: blockInfoUsageError,
	Description:  ``,
	Subcommands: []cli.Command{
		{
			Action:      utils.MigrateFlags(getCurrentBlockHeight),
			Name:        "count",
			Usage:       "issue asset by command",
			Description: ``,
		},
	},
}

func txInfoUsageError(ccntmext *cli.Ccntmext, err error, isSubcommand bool) error {
	fmt.Println("Error:", err.Error())
	showTxInfoHelp()
	return nil
}

var txCommandSet = cli.Command{
	Action:       utils.MigrateFlags(txInfoCommand),
	Name:         "tx",
	Usage:        "Display transaction informations",
	Flags:        append(NodeFlags, InfoFlags...),
	OnUsageError: txInfoUsageError,
	Description:  ``,
}

func versionInfoUsageError(ccntmext *cli.Ccntmext, err error, isSubcommand bool) error {
	fmt.Println("Error:", err.Error())
	showVersionInfoHelp()
	return nil
}

var versionCommand = cli.Command{
	Action:       utils.MigrateFlags(versionInfoCommand),
	Name:         "version",
	Usage:        "Display the version",
	OnUsageError: versionInfoUsageError,
	Description:  ``,
}

func versionInfoCommand(ctx *cli.Ccntmext) error {
	version, err := cntmSdk.Rpc.GetVersion()
	if nil != err {
		fmt.Printf("Get version information is error:  %s", err.Error())
		return err
	}
	fmt.Println("Node version: ", version)
	return nil
}

func txInfoCommand(ctx *cli.Ccntmext) error {
	if ctx.IsSet(utils.HashInfoFlag.Name) {
		txHash := ctx.String(utils.HashInfoFlag.Name)
		resp, err := request("GET", nil, restfulAddr()+"/api/v1/transaction/"+txHash)
		if err != nil {
			return err
		}
		common.EchoJsonDataGracefully(resp)
		return nil
	}
	showTxInfoHelp()
	return nil
}

func blockInfoCommand(ctx *cli.Ccntmext) error {
	if ctx.IsSet(utils.HeightInfoFlag.Name) {
		height := ctx.Int(utils.HeightInfoFlag.Name)
		if height >= 0 {
			resp, err := request("GET", nil, restfulAddr()+"/api/v1/block/details/height/"+strconv.Itoa(height))
			if err != nil {
				return err
			}
			common.EchoJsonDataGracefully(resp)
			return nil
		}
	} else if ctx.IsSet(utils.HashInfoFlag.Name) {
		blockHash := ctx.String(utils.HashInfoFlag.Name)
		if "" != blockHash {
			resp, err := request("GET", nil, restfulAddr()+"/api/v1/block/details/hash/"+blockHash)
			if err != nil {
				return err
			}
			common.EchoJsonDataGracefully(resp)
			return nil
		}
	}
	showBlockInfoHelp()
	return nil
}

func request(method string, cmd map[string]interface{}, url string) (map[string]interface{}, error) {
	hClient := &http.Client{}
	var repMsg = make(map[string]interface{})
	var response *http.Response
	var err error
	switch method {
	case "GET":
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return repMsg, err
		}
		response, err = hClient.Do(req)
	case "POST":
		data, err := json.Marshal(cmd)
		if err != nil {
			return repMsg, err
		}
		reqData := bytes.NewBuffer(data)
		req, err := http.NewRequest("POST", url, reqData)
		if err != nil {
			return repMsg, err
		}
		req.Header.Set("Ccntment-type", "application/json")
		response, err = hClient.Do(req)
	default:
		return repMsg, err
	}
	if response != nil {
		defer response.Body.Close()

		body, _ := ioutil.ReadAll(response.Body)
		if err := json.Unmarshal(body, &repMsg); err == nil {
			return repMsg, err
		}
	}
	if err != nil {
		return repMsg, err
	}
	return repMsg, err
}
