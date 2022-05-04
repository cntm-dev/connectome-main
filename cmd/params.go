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
	"github.com/cntmio/cntmology/cmd/utils"
	"github.com/urfave/cli"
)

var (
	NodeFlags = []cli.Flag{
		utils.LogLevelFlag,
		//utils.ConsensusFlag,
		utils.WalletFileFlag,
		utils.AccountPassFlag,
	}

	CcntmractFlags = []cli.Flag{
		utils.CcntmractVmTypeFlag,
		utils.CcntmractStorageFlag,
		utils.CcntmractCodeFileFlag,
		utils.CcntmractNameFlag,
		utils.CcntmractVersionFlag,
		utils.CcntmractAuthorFlag,
		utils.CcntmractEmailFlag,
		utils.CcntmractDescFlag,
		utils.CcntmractParamsFlag,
		utils.CcntmractAddrFlag,
	}

	InfoFlags = []cli.Flag{
		utils.BlockHeightInfoFlag,
		utils.BlockHashInfoFlag,
	}

	listFlags = []cli.Flag{
		utils.AccountVerboseFlag,
		utils.WalletFileFlag,
		utils.AccountFileFlag,
		utils.AccountLabelFlag,
	}

	setFlags = []cli.Flag{
		utils.AccountSigSchemeFlag,
		utils.AccountSetDefaultFlag,
		utils.WalletFileFlag,
		utils.AccountFileFlag,
		utils.AccountLabelFlag,
	}
	addFlags = []cli.Flag{
		utils.AccountQuantityFlag,
		utils.AccountTypeFlag,
		utils.AccountKeylenFlag,
		utils.AccountSigSchemeFlag,
		utils.AccountPassFlag,
		utils.AccountDefaultFlag,
		utils.AccountFileFlag,
		utils.AccountLabelFlag,
		utils.WalletFileFlag,
	}
	fileFlags = []cli.Flag{
		utils.WalletFileFlag,
		utils.AccountFileFlag,
	}
	importFlags = []cli.Flag{
		utils.AccountFileFlag,
		utils.AccountSourceFileFlag,
		utils.AccountKeyFlag,
	}
)
