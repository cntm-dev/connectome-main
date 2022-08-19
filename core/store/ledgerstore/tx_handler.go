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

	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/config"
	"github.com/cntmio/cntmology/common/log"
	"github.com/cntmio/cntmology/common/serialization"
	"github.com/cntmio/cntmology/core/payload"
	"github.com/cntmio/cntmology/core/states"
	"github.com/cntmio/cntmology/core/store"
	scommon "github.com/cntmio/cntmology/core/store/common"
	"github.com/cntmio/cntmology/core/store/statestore"
	"github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/smartccntmract"
	"github.com/cntmio/cntmology/smartccntmract/event"
	"github.com/cntmio/cntmology/smartccntmract/service/native/global_params"
	ninit "github.com/cntmio/cntmology/smartccntmract/service/native/init"
	"github.com/cntmio/cntmology/smartccntmract/service/native/cntm"
	"github.com/cntmio/cntmology/smartccntmract/service/native/utils"
	"github.com/cntmio/cntmology/smartccntmract/service/neovm"
	"github.com/cntmio/cntmology/smartccntmract/storage"
)

//HandleDeployTransaction deal with smart ccntmract deploy transaction
func (self *StateStore) HandleDeployTransaction(store store.LedgerStore, stateBatch *statestore.StateBatch,
	tx *types.Transaction, block *types.Block, eventStore scommon.EventStore) error {
	deploy := tx.Payload.(*payload.DeployCode)
	txHash := tx.Hash()
	address := types.AddressFromVmCode(deploy.Code)
	var (
		notifies []*event.NotifyEventInfo
		err      error
	)

	if tx.GasPrice != 0 {
		gas, overflow := common.SafeMul(tx.GasLimit, tx.GasPrice)
		if overflow {
			return fmt.Errorf("gaslimit:%d*gasprice:%d overflow!", tx.GasLimit, tx.GasPrice)
		}

		if err := isBalanceSufficient(tx.Payer, stateBatch, gas); err != nil {
			return err
		}

		cache := storage.NewCloneCache(stateBatch)

		// init smart ccntmract configuration info
		config := &smartccntmract.Config{
			Time:   block.Header.Timestamp,
			Height: block.Header.Height,
			Tx:     tx,
		}

		notifies, err = costGas(tx.Payer, gas, config, cache, store)
		if err != nil {
			return err
		}
		cache.Commit()
	}

	log.Infof("deploy ccntmract address:%s", address.ToHexString())
	// store ccntmract message
	err = stateBatch.TryGetOrAdd(scommon.ST_CcntmRACT, address[:], deploy)
	if err != nil {
		return err
	}

	SaveNotify(eventStore, txHash, notifies, true)
	return nil
}

//HandleInvokeTransaction deal with smart ccntmract invoke transaction
func (self *StateStore) HandleInvokeTransaction(store store.LedgerStore, stateBatch *statestore.StateBatch,
	tx *types.Transaction, block *types.Block, eventStore scommon.EventStore) error {
	invoke := tx.Payload.(*payload.InvokeCode)
	txHash := tx.Hash()
	code := invoke.Code
	sysTransFlag := bytes.Compare(code, ninit.COMMIT_DPOS_BYTES) == 0 || block.Header.Height == 0

	isCharge := !sysTransFlag && tx.GasPrice != 0

	gas, overflow := common.SafeMul(tx.GasLimit, tx.GasPrice)
	if overflow {
		return fmt.Errorf("gaslimit:%d*gasprice:%d overflow!", tx.GasLimit, tx.GasPrice)
	}
	if isCharge {
		if err := isBalanceSufficient(tx.Payer, stateBatch, gas); err != nil {
			return err
		}
	}

	// init smart ccntmract configuration info
	config := &smartccntmract.Config{
		Time:   block.Header.Timestamp,
		Height: block.Header.Height,
		Tx:     tx,
	}

	cache := storage.NewCloneCache(stateBatch)

	//init smart ccntmract info
	sc := smartccntmract.SmartCcntmract{
		Config:     config,
		CloneCache: cache,
		Store:      store,
		Gas:        tx.GasLimit,
	}

	//start the smart ccntmract executive function
	engine, err := sc.NewExecuteEngine(invoke.Code)
	if err != nil {
		return err
	}

	_, err = engine.Invoke()
	if err != nil {
		return err
	}

	var notifies []*event.NotifyEventInfo
	if isCharge {
		totalGas := tx.GasLimit - sc.Gas
		mixGas := neovm.MIN_TRANSACTION_GAS
		if totalGas < mixGas {
			totalGas = mixGas
		}
		notifies, err = costGas(tx.Payer, gas, config, sc.CloneCache, store)
		if err != nil {
			return err
		}
	}

	SaveNotify(eventStore, txHash, append(sc.Notifications, notifies...), true)
	sc.CloneCache.Commit()
	return nil
}

func SaveNotify(eventStore scommon.EventStore, txHash common.Uint256, notifies []*event.NotifyEventInfo, execSucc bool) error {
	if !config.DefConfig.Common.EnableEventLog {
		return nil
	}
	var notifyInfo *event.ExecuteNotify
	if execSucc {
		notifyInfo = &event.ExecuteNotify{TxHash: txHash,
			State: event.CcntmRACT_STATE_SUCCESS, Notify: notifies}
	} else {
		notifyInfo = &event.ExecuteNotify{TxHash: txHash,
			State: event.CcntmRACT_STATE_FAIL, Notify: notifies}
	}
	if err := eventStore.SaveEventNotifyByTx(txHash, notifyInfo); err != nil {
		return fmt.Errorf("SaveEventNotifyByTx error %s", err)
	}
	event.PushSmartCodeEvent(txHash, 0, event.EVENT_NOTIFY, notifyInfo)
	return nil
}

func genNativeTransferCode(from, to common.Address, value uint64) []byte {
	transfer := cntm.Transfers{States: []*cntm.State{{From: from, To: to, Value: value}}}
	tr := new(bytes.Buffer)
	transfer.Serialize(tr)
	return tr.Bytes()
}

// check whether payer cntm balance sufficient
func isBalanceSufficient(payer common.Address, stateBatch *statestore.StateBatch, gas uint64) error {
	balance, err := getBalance(stateBatch, payer, utils.OngCcntmractAddress)
	if err != nil {
		return err
	}
	if balance < gas {
		return fmt.Errorf("payer gas insufficient, need %d , only have %d", gas, balance)
	}
	return nil
}

func costGas(payer common.Address, gas uint64, config *smartccntmract.Config,
	cache *storage.CloneCache, store store.LedgerStore) ([]*event.NotifyEventInfo, error) {

	params := genNativeTransferCode(payer, utils.GovernanceCcntmractAddress, gas)

	sc := smartccntmract.SmartCcntmract{
		Config:     config,
		CloneCache: cache,
		Store:      store,
		Gas:        math.MaxUint64,
	}

	service, _ := sc.NewNativeService()
	_, err := service.NativeCall(utils.OngCcntmractAddress, "transfer", params)
	if err != nil {
		return nil, err
	}
	return sc.Notifications, nil
}

func refreshGlobalParam(config *smartccntmract.Config, cache *storage.CloneCache, store store.LedgerStore) error {
	bf := new(bytes.Buffer)
	if err := utils.WriteVarUint(bf, uint64(len(neovm.GAS_TABLE_KEYS))); err != nil {
		return fmt.Errorf("write gas_table_keys length error:%s", err)
	}
	for _, value := range neovm.GAS_TABLE_KEYS {
		if err := serialization.WriteString(bf, value); err != nil {
			return fmt.Errorf("serialize param name error:%s", value)
		}
	}

	sc := smartccntmract.SmartCcntmract{
		Config:     config,
		CloneCache: cache,
		Store:      store,
		Gas:        math.MaxUint64,
	}

	service, _ := sc.NewNativeService()
	result, err := service.NativeCall(utils.ParamCcntmractAddress, "getGlobalParam", bf.Bytes())
	if err != nil {
		return err
	}
	params := new(global_params.Params)
	if err := params.Deserialize(bytes.NewBuffer(result.([]byte))); err != nil {
		fmt.Errorf("deserialize global params error:%s", err)
	}

	for k, _ := range neovm.GAS_TABLE {
		n, ps := params.GetParam(k)
		if n != -1 && ps.Value != "" {
			pu, err := strconv.ParseUint(ps.Value, 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse uint %v", err)
			}
			neovm.GAS_TABLE[k] = pu
		}
	}
	return nil
}

func getBalance(stateBatch *statestore.StateBatch, address, ccntmract common.Address) (uint64, error) {
	bl, err := stateBatch.TryGet(scommon.ST_STORAGE, append(ccntmract[:], address[:]...))
	if err != nil {
		return 0, fmt.Errorf("get balance error:%s", err)
	}
	if bl == nil || bl.Value == nil {
		return 0, fmt.Errorf("get %s balance fail from %s", address.ToHexString(), ccntmract.ToHexString())
	}
	item, ok := bl.Value.(*states.StorageItem)
	if !ok {
		return 0, fmt.Errorf("%s", "instance doesn't StorageItem!")
	}
	balance, err := serialization.ReadUint64(bytes.NewBuffer(item.Value))
	if err != nil {
		return 0, fmt.Errorf("read balance error:%s", err)
	}
	return balance, nil
}
