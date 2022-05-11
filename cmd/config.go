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
	"github.com/cntmio/cntmology/cmd/utils"
	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/config"
	"github.com/urfave/cli"
	"io/ioutil"
)

func SetOntologyConfig(ctx *cli.Ccntmext) (*config.OntologyConfig, error) {
	cfg := config.DefConfig
	err := setGenesis(ctx, cfg.Genesis)
	if err != nil {
		return nil, fmt.Errorf("setGenesis error:%s", err)
	}
	setCommonConfig(ctx, cfg.Common)
	setConsensusConfig(ctx, cfg.Consensus)
	setP2PNodeConfig(ctx, cfg.P2PNode)
	setRpcConfig(ctx, cfg.Rpc)
	setRestfulConfig(ctx, cfg.Restful)
	setWebSocketConfig(ctx, cfg.Ws)
	setCliConfig(ctx, cfg.Cli)
	return cfg, nil
}

func setGenesis(ctx *cli.Ccntmext, cfg *config.GenesisConfig) error {
	if ctx.GlobalBool(utils.GetFlagName(utils.EnableTestModeFlag)) {
		cfg.ConsensusType = config.CONSENSUS_TYPE_SOLO
		cfg.SOLO.GenBlockTime = ctx.Uint(utils.GetFlagName(utils.TestModeGenBlockTimeFlag))
		if cfg.SOLO.GenBlockTime <= 1 {
			cfg.SOLO.GenBlockTime = config.DEFAULT_GEN_BLOCK_TIME
		}
		return nil
	}

	genesisFile := ctx.GlobalString(utils.GetFlagName(utils.ConfigFlag))

	if !common.FileExisted(genesisFile) {
		return nil
	}
	data, err := ioutil.ReadFile(genesisFile)
	if err != nil {
		return fmt.Errorf("ioutil.ReadFile:%s error:%s", genesisFile, err)
	}
	// Remove the UTF-8 Byte Order Mark
	data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))

	err = json.Unmarshal(data, cfg)
	if err != nil {
		return fmt.Errorf("json.Unmarshal GenesisConfig:%s error:%s", data, err)
	}

	switch cfg.ConsensusType {
	case config.CONSENSUS_TYPE_DBFT:
		if len(cfg.DBFT.Bookkeepers) < config.DBFT_MIN_NODE_NUM {
			return fmt.Errorf("DBFT consensus at least need %d bookkeepers in config", config.DBFT_MIN_NODE_NUM)
		}
		if cfg.DBFT.GenBlockTime <= 0 {
			cfg.DBFT.GenBlockTime = config.DEFAULT_GEN_BLOCK_TIME
		}
	case config.CONSENSUS_TYPE_VBFT:
		if len(cfg.VBFT.Peers) < config.VBFT_MIN_NODE_NUM {
			return fmt.Errorf("VBFT consensus at least need %d peers in config", config.VBFT_MIN_NODE_NUM)
		}
	default:
		return fmt.Errorf("Unknow consensus:%s", cfg.ConsensusType)
	}

	return nil
}

func setCommonConfig(ctx *cli.Ccntmext, cfg *config.CommonConfig) {
	cfg.EnableEventLog = !ctx.GlobalBool(utils.GetFlagName(utils.DisableEventLogFlag))
	cfg.GasLimit = ctx.GlobalUint64(utils.GetFlagName(utils.GasLimitFlag))
	cfg.GasPrice = ctx.GlobalUint64(utils.GetFlagName(utils.GasPriceFlag))
}

func setConsensusConfig(ctx *cli.Ccntmext, cfg *config.ConsensusConfig) {
	cfg.EnableConsensus = !ctx.GlobalBool(utils.GetFlagName(utils.DisableConsensusFlag))
	cfg.MaxTxInBlock = ctx.GlobalUint(utils.GetFlagName(utils.MaxTxInBlockFlag))
}

func setP2PNodeConfig(ctx *cli.Ccntmext, cfg *config.P2PNodeConfig) {
	cfg.NodePort = ctx.GlobalUint(utils.GetFlagName(utils.NodePortFlag))
	cfg.NodeConsensusPort = ctx.GlobalUint(utils.GetFlagName(utils.ConsensusPortFlag))
	cfg.DualPortSupport = ctx.GlobalBool(utils.GetFlagName(utils.DualPortSupportFlag))
}

func setRpcConfig(ctx *cli.Ccntmext, cfg *config.RpcConfig) {
	cfg.EnableHttpJsonRpc = true
	cfg.HttpJsonPort = ctx.GlobalUint(utils.GetFlagName(utils.RPCPortFlag))
	cfg.HttpLocalPort = ctx.GlobalUint(utils.GetFlagName(utils.RPCLocalProtFlag))
}

func setRestfulConfig(ctx *cli.Ccntmext, cfg *config.RestfulConfig) {
	cfg.EnableHttpRestful = ctx.GlobalBool(utils.GetFlagName(utils.RestfulEnableFlag))
	cfg.HttpRestPort = ctx.GlobalUint(utils.GetFlagName(utils.RestfulPortFlag))
}

func setWebSocketConfig(ctx *cli.Ccntmext, cfg *config.WebSocketConfig) {
	cfg.EnableHttpWs = ctx.GlobalBool(utils.GetFlagName(utils.WsEnabledFlag))
	cfg.HttpWsPort = ctx.GlobalUint(utils.GetFlagName(utils.WsPortFlag))
}

func setCliConfig(ctx *cli.Ccntmext, cfg *config.CliConfig) {
	cfg.EnableCliRpcServer = ctx.GlobalBool(utils.GetFlagName(utils.CliEnableRpcFlag))
	cfg.CliRpcPort = ctx.GlobalUint(utils.GetFlagName(utils.CliRpcPortFlag))
}
