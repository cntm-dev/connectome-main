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

package ledgerstore

import (
	"bytes"
	"fmt"
	"math"
	"strconv"

	common2 "github.com/ethereum/go-ethereum/common"
	types2 "github.com/ethereum/go-ethereum/core/types"
	"github.com/cntmio/cntmology/common"
	sysconfig "github.com/cntmio/cntmology/common/config"
	"github.com/cntmio/cntmology/common/log"
	"github.com/cntmio/cntmology/core/payload"
	"github.com/cntmio/cntmology/core/store"
	scommon "github.com/cntmio/cntmology/core/store/common"
	"github.com/cntmio/cntmology/core/store/overlaydb"
	"github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/errors"
	"github.com/cntmio/cntmology/smartccntmract"
	"github.com/cntmio/cntmology/smartccntmract/event"
	evm2 "github.com/cntmio/cntmology/smartccntmract/service/evm"
	types3 "github.com/cntmio/cntmology/smartccntmract/service/evm/types"
	"github.com/cntmio/cntmology/smartccntmract/service/evm/witness"
	"github.com/cntmio/cntmology/smartccntmract/service/native/global_params"
	ninit "github.com/cntmio/cntmology/smartccntmract/service/native/init"
	"github.com/cntmio/cntmology/smartccntmract/service/native/cntm"
	"github.com/cntmio/cntmology/smartccntmract/service/native/cntm"
	"github.com/cntmio/cntmology/smartccntmract/service/native/utils"
	"github.com/cntmio/cntmology/smartccntmract/service/neovm"
	"github.com/cntmio/cntmology/smartccntmract/service/wasmvm"
	"github.com/cntmio/cntmology/smartccntmract/storage"
	"github.com/cntmio/cntmology/vm/evm"
	"github.com/cntmio/cntmology/vm/evm/params"
)

func tuneGasFeeByHeight(height uint32, gas uint64, gasRound uint64, curBalance uint64) uint64 {
	gasTuneheight := sysconfig.GetGasRoundTuneHeight(sysconfig.DefConfig.P2PNode.NetworkId)
	if height > gasTuneheight {
		t := (gas + gasRound - 1) / gasRound
		if gas > math.MaxUint64-gasRound {
			return curBalance
		}

		newGas := gasRound * t

		if newGas > curBalance {
			return curBalance
		}

		return newGas
	}

	return gas
}

//HandleDeployTransaction deal with smart ccntmract deploy transaction
func (self *StateStore) HandleDeployTransaction(store store.LedgerStore, overlay *overlaydb.OverlayDB, gasTable map[string]uint64, cache *storage.CacheDB,
	tx *types.Transaction, block *types.Block, notify *event.ExecuteNotify) error {
	deploy := tx.Payload.(*payload.DeployCode)
	var (
		notifies    []*event.NotifyEventInfo
		gasConsumed uint64
		err         error
	)

	if deploy.VmType() == payload.WASMVM_TYPE {
		_, err = wasmvm.ReadWasmModule(deploy.GetRawCode(), sysconfig.DefConfig.Common.WasmVerifyMethod)
		if err != nil {
			return err
		}
	}

	if tx.GasPrice != 0 {
		// init smart ccntmract configuration info
		config := &smartccntmract.Config{
			Time:      block.Header.Timestamp,
			Height:    block.Header.Height,
			Tx:        tx,
			BlockHash: block.Hash(),
		}
		createGasPrice, ok := gasTable[neovm.CcntmRACT_CREATE_NAME]
		if !ok {
			overlay.SetError(errors.NewErr("[HandleDeployTransaction] get CcntmRACT_CREATE_NAME gas failed"))
			return nil
		}

		uintCodePrice, ok := gasTable[neovm.UINT_DEPLOY_CODE_LEN_NAME]
		if !ok {
			overlay.SetError(errors.NewErr("[HandleDeployTransaction] get UINT_DEPLOY_CODE_LEN_NAME gas failed"))
			return nil
		}

		gasLimit := createGasPrice + calcGasByCodeLen(len(deploy.GetRawCode()), uintCodePrice)
		balance, err := isBalanceSufficient(tx.Payer, cache, config, store, gasLimit*tx.GasPrice)
		if err != nil {
			if err := costInvalidGas(tx.Payer, balance, config, overlay, store, notify); err != nil {
				return err
			}
			return err
		}
		if tx.GasLimit < gasLimit {
			if err := costInvalidGas(tx.Payer, tx.GasLimit*tx.GasPrice, config, overlay, store, notify); err != nil {
				return err
			}
			return fmt.Errorf("gasLimit insufficient, need:%d actual:%d", gasLimit, tx.GasLimit)
		}
		gasConsumed = gasLimit * tx.GasPrice
		notifies, err = chargeCostGas(tx.Payer, gasConsumed, config, cache, store)
		if err != nil {
			return err
		}
		cache.Commit()
	}

	address := deploy.Address()
	// store ccntmract message
	dep, destroyed, err := cache.GetCcntmract(address)
	if err != nil {
		return err
	}
	if destroyed {
		return fmt.Errorf("can not redeploy destroyed ccntmract: %s", address.ToHexString())
	}
	if dep == nil {
		log.Infof("deploy ccntmract address:%s", address.ToHexString())
		cache.PutCcntmract(deploy)
		notify.CreatedCcntmract = address
	}
	cache.Commit()

	notify.Notify = append(notify.Notify, notifies...)
	notify.GasConsumed = gasConsumed
	notify.State = event.CcntmRACT_STATE_SUCCESS
	return nil
}

//HandleInvokeTransaction deal with smart ccntmract invoke transaction
func (self *StateStore) HandleInvokeTransaction(store store.LedgerStore, overlay *overlaydb.OverlayDB, gasTable map[string]uint64, cache *storage.CacheDB,
	tx *types.Transaction, block *types.Block, notify *event.ExecuteNotify) ([]common.Uint256, error) {
	invoke := tx.Payload.(*payload.InvokeCode)
	code := invoke.Code
	sysTransFlag := bytes.Compare(code, ninit.COMMIT_DPOS_BYTES) == 0 || block.Header.Height == 0

	isCharge := !sysTransFlag && tx.GasPrice != 0

	// init smart ccntmract configuration info
	config := &smartccntmract.Config{
		Time:      block.Header.Timestamp,
		Height:    block.Header.Height,
		Tx:        tx,
		BlockHash: block.Hash(),
	}

	var (
		costGasLimit      uint64
		costGas           uint64
		oldBalance        uint64
		newBalance        uint64
		codeLenGasLimit   uint64
		availableGasLimit uint64
		minGas            uint64
		err               error
	)

	availableGasLimit = tx.GasLimit
	if isCharge {
		uintCodeGasPrice, ok := gasTable[neovm.UINT_INVOKE_CODE_LEN_NAME]
		if !ok {
			overlay.SetError(errors.NewErr("[HandleInvokeTransaction] get UINT_INVOKE_CODE_LEN_NAME gas failed"))
			return nil, nil
		}

		oldBalance, err = getBalanceFromNative(config, cache, store, tx.Payer)
		if err != nil {
			return nil, err
		}

		minGas = neovm.MIN_TRANSACTION_GAS * tx.GasPrice

		if oldBalance < minGas {
			if err := costInvalidGas(tx.Payer, oldBalance, config, overlay, store, notify); err != nil {
				return nil, err
			}
			return nil, fmt.Errorf("balance gas: %d less than min gas: %d", oldBalance, minGas)
		}

		codeLenGasLimit = calcGasByCodeLen(len(invoke.Code), uintCodeGasPrice)

		if oldBalance < codeLenGasLimit*tx.GasPrice {
			if err := costInvalidGas(tx.Payer, oldBalance, config, overlay, store, notify); err != nil {
				return nil, err
			}
			return nil, fmt.Errorf("balance gas insufficient: balance:%d < code length need gas:%d", oldBalance, codeLenGasLimit*tx.GasPrice)
		}

		if tx.GasLimit < codeLenGasLimit {
			if err := costInvalidGas(tx.Payer, tx.GasLimit*tx.GasPrice, config, overlay, store, notify); err != nil {
				return nil, err
			}
			return nil, fmt.Errorf("invoke transaction gasLimit insufficient: need%d actual:%d", tx.GasLimit, codeLenGasLimit)
		}

		maxAvaGasLimit := oldBalance / tx.GasPrice

		if availableGasLimit > maxAvaGasLimit {
			availableGasLimit = maxAvaGasLimit
		}
	}

	//init smart ccntmract info
	sc := smartccntmract.SmartCcntmract{
		Config:       config,
		CacheDB:      cache,
		Store:        store,
		GasTable:     gasTable,
		Gas:          availableGasLimit - codeLenGasLimit,
		WasmExecStep: sysconfig.DEFAULT_WASM_MAX_STEPCOUNT,
		PreExec:      false,
	}

	//start the smart ccntmract executive function
	engine, _ := sc.NewExecuteEngine(invoke.Code, tx.TxType)

	_, err = engine.Invoke()
	if sc.IsInternalErr() {
		overlay.SetError(fmt.Errorf("[HandleInvokeTransaction] %s", err))
		return nil, nil
	}

	costGasLimit = availableGasLimit - sc.Gas
	if costGasLimit < neovm.MIN_TRANSACTION_GAS {
		costGasLimit = neovm.MIN_TRANSACTION_GAS
	}

	costGas = costGasLimit * tx.GasPrice
	if err != nil {
		if isCharge {
			costGas = tuneGasFeeByHeight(config.Height, costGas, tx.GasPrice*neovm.MIN_TRANSACTION_GAS, oldBalance)
			if err := costInvalidGas(tx.Payer, costGas, config, overlay, store, notify); err != nil {
				return nil, err
			}
		}
		return nil, err
	}

	var notifies []*event.NotifyEventInfo
	if isCharge {
		newBalance, err = getBalanceFromNative(config, cache, store, tx.Payer)
		if err != nil {
			return nil, err
		}

		if newBalance < costGas {
			costGas = tuneGasFeeByHeight(config.Height, costGas, tx.GasPrice*neovm.MIN_TRANSACTION_GAS, oldBalance)
			if err := costInvalidGas(tx.Payer, costGas, config, overlay, store, notify); err != nil {
				return nil, err
			}
			return nil, fmt.Errorf("gas insufficient, balance:%d < costGas:%d", newBalance, costGas)
		}

		costGas = tuneGasFeeByHeight(config.Height, costGas, tx.GasPrice*neovm.MIN_TRANSACTION_GAS, newBalance)
		notifies, err = chargeCostGas(tx.Payer, costGas, config, sc.CacheDB, store)
		if err != nil {
			return nil, err
		}
	}

	notify.Notify = append(notify.Notify, sc.Notifications...)
	notify.Notify = append(notify.Notify, notifies...)
	notify.GasConsumed = costGas
	notify.State = event.CcntmRACT_STATE_SUCCESS
	sc.CacheDB.Commit()
	return sc.CrossHashes, nil
}

func SaveNotify(eventStore scommon.EventStore, txHash common.Uint256, notify *event.ExecuteNotify) error {
	if !sysconfig.DefConfig.Common.EnableEventLog {
		return nil
	}
	if err := eventStore.SaveEventNotifyByTx(txHash, notify); err != nil {
		return fmt.Errorf("SaveEventNotifyByTx error %s", err)
	}
	event.PushSmartCodeEvent(txHash, 0, event.EVENT_NOTIFY, notify)
	return nil
}

func genNativeTransferCode(from, to common.Address, value uint64) []byte {
	transfer := &cntm.TransferStates{States: []cntm.TransferState{{From: from, To: to, Value: value}}}
	return common.SerializeToBytes(transfer)
}

// check whether payer cntm balance sufficient
func isBalanceSufficient(payer common.Address, cache *storage.CacheDB, config *smartccntmract.Config, store store.LedgerStore, gas uint64) (uint64, error) {
	balance, err := getBalanceFromNative(config, cache, store, payer)
	if err != nil {
		return 0, err
	}
	if balance < gas {
		return 0, fmt.Errorf("payer gas insufficient, need %d , only have %d", gas, balance)
	}
	return balance, nil
}

func chargeCostGas(payer common.Address, gas uint64, config *smartccntmract.Config,
	cache *storage.CacheDB, store store.LedgerStore) ([]*event.NotifyEventInfo, error) {

	params := genNativeTransferCode(payer, utils.GovernanceCcntmractAddress, gas)

	sc := smartccntmract.SmartCcntmract{
		Config:  config,
		CacheDB: cache,
		Store:   store,
		Gas:     math.MaxUint64,
	}

	service, _ := sc.NewNativeService()
	_, err := service.NativeCall(utils.OngCcntmractAddress, "transfer", params)
	if err != nil {
		return nil, err
	}
	return sc.Notifications, nil
}

func getEvmSystemWitnessAddress(config *smartccntmract.Config, cache *storage.CacheDB, store store.LedgerStore) common2.Address {
	sink := common.NewZeroCopySink(nil)
	utils.EncodeVarUint(sink, 1)
	sink.WriteString(witness.WitnessGlobalParamKey)

	sc := smartccntmract.SmartCcntmract{
		Config:  config,
		CacheDB: cache,
		Store:   store,
		Gas:     math.MaxUint64,
	}

	service, _ := sc.NewNativeService()
	result, err := service.NativeCall(utils.ParamCcntmractAddress, "getGlobalParam", sink.Bytes())
	if err != nil {
		log.Errorf("get witness address error: %s", err)
		return common2.Address{}
	}
	params := new(global_params.Params)
	if err := params.Deserialization(common.NewZeroCopySource(result)); err != nil {
		log.Errorf("deserialize global params error:%s", err)
		return common2.Address{}
	}
	n, ps := params.GetParam(witness.WitnessGlobalParamKey)
	if n != -1 && ps.Value != "" {
		return common2.HexToAddress(ps.Value)
	}

	return common2.Address{}
}

func refreshGlobalParam(config *smartccntmract.Config, cache *storage.CacheDB, store store.LedgerStore) error {
	sink := common.NewZeroCopySink(nil)
	utils.EncodeVarUint(sink, uint64(len(neovm.GAS_TABLE_KEYS)))
	for _, value := range neovm.GAS_TABLE_KEYS {
		sink.WriteString(value)
	}

	sc := smartccntmract.SmartCcntmract{
		Config:  config,
		CacheDB: cache,
		Store:   store,
		Gas:     math.MaxUint64,
	}

	service, _ := sc.NewNativeService()
	result, err := service.NativeCall(utils.ParamCcntmractAddress, "getGlobalParam", sink.Bytes())
	if err != nil {
		return err
	}
	params := new(global_params.Params)
	if err := params.Deserialization(common.NewZeroCopySource(result)); err != nil {
		return fmt.Errorf("deserialize global params error:%s", err)
	}
	neovm.GAS_TABLE.Range(func(key, value interface{}) bool {
		n, ps := params.GetParam(key.(string))
		if n != -1 && ps.Value != "" {
			pu, err := strconv.ParseUint(ps.Value, 10, 64)
			if err != nil {
				log.Errorf("[refreshGlobalParam] failed to parse uint %v\n", ps.Value)
			} else {
				neovm.GAS_TABLE.Store(key, pu)
			}
		}
		return true
	})
	return nil
}

func getBalanceFromNative(config *smartccntmract.Config, cache *storage.CacheDB, store store.LedgerStore, address common.Address) (uint64, error) {
	bf := common.NewZeroCopySink(nil)
	utils.EncodeAddress(bf, address)
	sc := smartccntmract.SmartCcntmract{
		Config:  config,
		CacheDB: cache,
		Store:   store,
		Gas:     math.MaxUint64,
	}

	service, _ := sc.NewNativeService()
	result, err := service.NativeCall(utils.OngCcntmractAddress, cntm.BALANCEOF_NAME, bf.Bytes())
	if err != nil {
		return 0, err
	}
	return common.BigIntFromNeoBytes(result).Uint64(), nil
}

func costInvalidGas(address common.Address, gas uint64, config *smartccntmract.Config, overlay *overlaydb.OverlayDB,
	store store.LedgerStore, notify *event.ExecuteNotify) error {
	cache := storage.NewCacheDB(overlay)
	notifies, err := chargeCostGas(address, gas, config, cache, store)
	if err != nil {
		return err
	}
	cache.Commit()
	notify.GasConsumed = gas
	notify.Notify = append(notify.Notify, notifies...)
	return nil
}

func calcGasByCodeLen(codeLen int, codeGas uint64) uint64 {
	return uint64(codeLen/neovm.PER_UNIT_CODE_LEN) * codeGas
}

type Eip155Ccntmext struct {
	BlockHash common.Uint256
	TxIndex   uint32
	Height    uint32
	Timestamp uint32
}

func (self *StateStore) HandleEIP155Transaction(store store.LedgerStore, cache *storage.CacheDB,
	tx *types2.Transaction, ctx Eip155Ccntmext, notify *event.ExecuteNotify, checkNonce bool) (*types3.ExecutionResult, *types.Receipt, error) {
	usedGas := uint64(0)
	config := params.GetChainConfig(sysconfig.DefConfig.P2PNode.EVMChainId)
	statedb := storage.NewStateDB(cache, tx.Hash(), common2.Hash(ctx.BlockHash), cntm.OngBalanceHandle{})
	result, receipt, err := evm2.ApplyTransaction(config, store, statedb, ctx.Height, ctx.Timestamp, tx, &usedGas,
		utils.GovernanceCcntmractAddress, evm.Config{}, checkNonce)

	if err != nil {
		cache.SetDbErr(err)
		return nil, nil, err
	}
	if err = statedb.DbErr(); err != nil {
		cache.SetDbErr(err)
		return nil, nil, err
	}
	receipt.TxIndex = ctx.TxIndex

	*notify = *event.ExecuteNotifyFromEthReceipt(receipt)

	return result, receipt, nil
}
