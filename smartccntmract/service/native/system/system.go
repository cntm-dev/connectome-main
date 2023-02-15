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

package system

import (
	"fmt"
	"math/big"

	config2 "github.com/cntmio/cntmology/common/config"

	common2 "github.com/ethereum/go-ethereum/common"
	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/smartccntmract/event"
	evm2 "github.com/cntmio/cntmology/smartccntmract/service/evm"
	"github.com/cntmio/cntmology/smartccntmract/service/native"
	"github.com/cntmio/cntmology/smartccntmract/service/native/cntm"
	"github.com/cntmio/cntmology/smartccntmract/service/native/utils"
	"github.com/cntmio/cntmology/smartccntmract/storage"
	"github.com/cntmio/cntmology/vm/evm"
	"github.com/cntmio/cntmology/vm/evm/params"
)

const (
	EvmInvokeName = "evmInvoke"
)

func InitSystem() {
	native.Ccntmracts[utils.SystemCcntmractAddress] = RegisterSystemCcntmract
}

func RegisterSystemCcntmract(native *native.NativeService) {
	native.Register(EvmInvokeName, EVMInvoke)
}

func EVMInvoke(native *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(native.Input)

	caller, err := utils.DecodeAddress(source)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("evm invoke decode ccntmract address error: %v", err)
	}
	target, err := utils.DecodeAddress(source)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("evm invoke decode ccntmract address error: %v", err)
	}
	input, err := utils.DecodeVarBytes(source)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("evm invoke decode input error: %v", err)
	}

	if !native.CcntmextRef.CheckWitness(caller) {
		return utils.BYTE_FALSE, fmt.Errorf("evm invoke error: verify witness failed for caller: %s", caller.ToBase58())
	}

	// Create a new ccntmext to be used in the EVM environment
	blockCcntmext := evm2.NewEVMBlockCcntmext(native.Height, native.Time, native.Store)
	gasLeft, gasPrice := native.CcntmextRef.GetGasInfo()
	txctx := evm.TxCcntmext{
		Origin:   common2.Address(native.Tx.Payer),
		GasPrice: big.NewInt(0).SetUint64(gasPrice),
	}
	statedb := storage.NewStateDB(native.CacheDB, common2.Hash(native.Tx.Hash()), common2.Hash(native.BlockHash), cntm.OngBalanceHandle{})
	config := params.GetChainConfig(config2.DefConfig.P2PNode.EVMChainId)
	vmenv := evm.NewEVM(blockCcntmext, txctx, statedb, config, evm.Config{})

	callerCtx := native.CcntmextRef.CallingCcntmext()
	if callerCtx == nil {
		return utils.BYTE_FALSE, fmt.Errorf("evm invoke must have a caller")
	}
	ret, leftGas, err := vmenv.Call(evm.AccountRef(caller), common2.Address(target), input, gasLeft, big.NewInt(0))
	gasUsed := gasLeft - leftGas
	refund := gasUsed / 2
	if refund > statedb.GetRefund() {
		refund = statedb.GetRefund()
	}
	gasUsed -= refund
	enoughGas := native.CcntmextRef.CheckUseGas(gasUsed)
	if !enoughGas {
		return utils.BYTE_FALSE, fmt.Errorf("evm invoke out of gas")
	}
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("invoke evm error:%s, return: %x", err, ret)
	}

	for _, log := range statedb.GetLogs() {
		native.Notifications = append(native.Notifications, event.NotifyEventInfoFromEvmLog(log))
	}

	err = statedb.CommitToCacheDB()
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	return ret, nil
}
