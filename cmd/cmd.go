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
	"io"
	"strconv"
	"strings"

	sdk "github.com/cntmio/cntmology-go-sdk"
	"github.com/cntmio/cntmology/cmd/utils"
	"github.com/cntmio/cntmology/common/config"
	"github.com/urfave/cli"
)

// AppHelpTemplate is the test template for the default, global app help topic.
var AppHelpTemplate = `NAME:
   {{.Name}} - {{.Usage}}

USAGE:
   {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}{{if .Commands}} command [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{end}}{{end}}{{if .Description}}

DESCRIPTION:
   {{.Description}}{{end}}{{if .VisibleCommands}}

COMMANDS:{{range .Commands}}
   {{join .Names ", "}}{{ "\t" }}{{.Usage}}{{end}}{{end}}{{if .Version}}

VERSION:
   {{.Version}}{{end}}{{if .Copyright }}

COPYRIGHT:
   {{.Copyright}}{{end}}
`

var cntmSdk *sdk.OntologySdk

func localRpcAddress() string {
	return "http://localhost:" + strconv.Itoa(config.Parameters.HttpJsonPort)
}

func rpcAddress() string {
	return "http://localhost:" + strconv.Itoa(config.Parameters.HttpJsonPort)
}

func restfulAddr() string {
	return "http://localhost:" + strconv.Itoa(config.Parameters.HttpRestPort)
}

func init() {
	cntmSdk = sdk.NewOntologySdk()
	cntmSdk.Rpc.SetAddress(rpcAddress())
	//cli.AppHelpTemplate = AppHelpTemplate
}

// flagGroup is a collection of flags belcntming to a single topic.
type flagGroup struct {
	Name  string
	Flags []cli.Flag
}

// AppHelpFlagGroups is the application flags, grouped by functionality.
var AppHelpFlagGroups = []flagGroup{
	{
		Name: "cntmOLOGY INFO BLOCK",
		Flags: []cli.Flag{
			utils.HashInfoFlag,
			utils.HeightInfoFlag,
		},
	},

	{
		Name: "cntmOLOGY INFO TRANSACTION",
		Flags: []cli.Flag{
			utils.HashInfoFlag,
		},
	},

	{
		Name: "cntmOLOGY INFO VERSION",
		Flags: []cli.Flag{
			utils.NonOptionFlag,
		},
	},

	{
		Name: "cntmOLOGY INFO BLOCK HEIGHT",
		Flags: []cli.Flag{
			utils.NonOptionFlag,
		},
	},

	{
		Name: "cntmOLOGY ASSET TRANSFER",
		Flags: []cli.Flag{
			utils.TransactionFromFlag,
			utils.TransactionToFlag,
			utils.TransactionValueFlag,
			utils.CcntmractAddrFlag,
			utils.AccountPassFlag,
		},
	},

	{
		Name: "cntmOLOGY SET DEBUG",
		Flags: []cli.Flag{
			utils.DebugLevelFlag,
		},
	},

	{
		Name: "cntmOLOGY SET CONSENSUS",
		Flags: []cli.Flag{
			utils.ConsensusFlag,
		},
	},

	{
		Name: "cntmOLOGY CcntmRACT DEPLOY",
		Flags: []cli.Flag{
			utils.CcntmractVmTypeFlag,
			utils.CcntmractStorageFlag,
			utils.CcntmractCodeFlag,
			utils.CcntmractNameFlag,
			utils.CcntmractVersionFlag,
			utils.CcntmractAuthorFlag,
			utils.CcntmractDescFlag,
			utils.CcntmractEmailFlag,
		},
	},

	{
		Name: "cntmOLOGY CcntmRACT INVOKE",
		Flags: []cli.Flag{
			utils.CcntmractAddrFlag,
			utils.CcntmractParamsFlag,
		},
	},

	{
		Name: "MISC",
	},
}

type byCategory []flagGroup

func (a byCategory) Len() int      { return len(a) }
func (a byCategory) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byCategory) Less(i, j int) bool {
	iCat, jCat := a[i].Name, a[j].Name
	iIdx, jIdx := len(AppHelpFlagGroups), len(AppHelpFlagGroups) // ensure non categorized flags come last

	for i, group := range AppHelpFlagGroups {
		if iCat == group.Name {
			iIdx = i
		}
		if jCat == group.Name {
			jIdx = i
		}
	}

	return iIdx < jIdx
}

func HelpUsage() {
	cli.AppHelpTemplate = AppHelpTemplate

	// Define a one shot struct to pass to the usage template
	type helpData struct {
		App        interface{}
		FlagGroups []flagGroup
	}

	// Override the default app help printer, but only for the global app help
	originalHelpPrinter := cli.HelpPrinter
	cli.HelpPrinter = func(w io.Writer, tmpl string, data interface{}) {
		if tmpl == AppHelpTemplate {
			// Iterate over all the flags and add any uncategorized ones
			categorized := make(map[string]struct{})
			for _, group := range AppHelpFlagGroups {
				for _, flag := range group.Flags {
					categorized[flag.String()] = struct{}{}
				}
			}
			uncategorized := []cli.Flag{}
			for _, flag := range data.(*cli.App).Flags {
				if _, ok := categorized[flag.String()]; !ok {
					if strings.HasPrefix(flag.GetName(), "dashboard") {
						ccntminue
					}
					uncategorized = append(uncategorized, flag)
				}
			}
			if len(uncategorized) > 0 {
				// Append all ungategorized options to the misc group
				miscs := len(AppHelpFlagGroups[len(AppHelpFlagGroups)-1].Flags)
				AppHelpFlagGroups[len(AppHelpFlagGroups)-1].Flags = append(AppHelpFlagGroups[len(AppHelpFlagGroups)-1].Flags, uncategorized...)

				// Make sure they are removed afterwards
				defer func() {
					AppHelpFlagGroups[len(AppHelpFlagGroups)-1].Flags = AppHelpFlagGroups[len(AppHelpFlagGroups)-1].Flags[:miscs]
				}()
			}
			// Render out custom usage screen
			originalHelpPrinter(w, tmpl, helpData{data, AppHelpFlagGroups})
		}
	}
}
