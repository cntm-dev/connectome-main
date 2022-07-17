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

	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/config"
	"github.com/cntmio/cntmology/common/serialization"
	"github.com/cntmio/cntmology/core/payload"
	"github.com/cntmio/cntmology/core/states"
	"github.com/cntmio/cntmology/core/store"
	scommon "github.com/cntmio/cntmology/core/store/common"
	"github.com/cntmio/cntmology/core/store/statestore"
	"github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/smartccntmract"
	"github.com/cntmio/cntmology/smartccntmract/event"
	ninit "github.com/cntmio/cntmology/smartccntmract/service/native/init"
	"github.com/cntmio/cntmology/smartccntmract/service/native/cntm"
	"github.com/cntmio/cntmology/smartccntmract/service/native/utils"
	"github.com/cntmio/cntmology/smartccntmract/service/neovm"
	sstates "github.com/cntmio/cntmology/smartccntmract/states"
	"github.com/cntmio/cntmology/smartccntmract/storage"
	stypes "github.com/cntmio/cntmology/smartccntmract/types"
)

//HandleDeployTransaction deal with smart ccntmract deploy transaction
func (self *StateStore) HandleDeployTransaction(store store.LedgerStore, stateBatch *statestore.StateBatch,
	tx *types.Transaction, block *types.Block, eventStore scommon.EventStore) error {
	deploy := tx.Payload.(*payload.DeployCode)
	txHash := tx.Hash()
	originAddress := deploy.Code.AddressFromVmCode()

	var (
		notifies []*event.NotifyEventInfo
		err      error
	)
	// mapping native ccntmract origin address to target address
	if deploy.Code.VmType == stypes.Native {
		targetAddress, err := common.AddressParseFromBytes(deploy.Code.Code)
		if err != nil {
			return fmt.Errorf("Invalid native ccntmract address:%s", err)
		}
		originAddress = targetAddress
	} else {
		if err := isBalanceSufficient(tx, stateBatch); err != nil {
			return err
		}

		cache := storage.NewCloneCache(stateBatch)

		// init smart ccntmract configuration info
		config := &smartccntmract.Config{
			Time:   block.Header.Timestamp,
			Height: block.Header.Height,
			Tx:     tx,
		}

		notifies, err = costGas(tx.Payer, tx.GasLimit*tx.GasPrice, config, cache, store)
		if err != nil {
			return err
		}
		cache.Commit()
	}

	// store ccntmract message
	err = stateBatch.TryGetOrAdd(scommon.ST_CcntmRACT, originAddress[:], deploy)
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
	code := invoke.Code.Code
	sysTransFlag := bytes.Compare(code, ninit.COMMIT_DPOS_BYTES) == 0 || block.Header.Height == 0

	if !sysTransFlag && tx.GasPrice != 0 {
		if err := isBalanceSufficient(tx, stateBatch); err != nil {
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
		Code:       invoke.Code,
		Gas:        tx.GasLimit,
	}

	//start the smart ccntmract executive function
	_, err := sc.Execute()

	if err != nil {
		return err
	}

	var notifies []*event.NotifyEventInfo
	if !sysTransFlag {
		totalGas := tx.GasLimit - sc.Gas
		if totalGas < neovm.TRANSACTION_GAS {
			totalGas = neovm.TRANSACTION_GAS
		}
		notifies, err = costGas(tx.Payer, totalGas*tx.GasPrice, config, sc.CloneCache, store)
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

//HandleClaimTransaction deal with cntm claim transaction
func (self *StateStore) HandleClaimTransaction(stateBatch *statestore.StateBatch, tx *types.Transaction) error {
	//TODO
	return nil
}

func genNativeTransferCode(ccntmract, from, to common.Address, value uint64) stypes.VmCode {
	transfer := cntm.Transfers{States: []*cntm.State{{From: from, To: to, Value: value}}}
	tr := new(bytes.Buffer)
	transfer.Serialize(tr)
	trans := &sstates.Ccntmract{
		Address: ccntmract,
		Method:  "transfer",
		Args:    tr.Bytes(),
	}
	ts := new(bytes.Buffer)
	trans.Serialize(ts)
	return stypes.VmCode{Code: ts.Bytes(), VmType: stypes.Native}
}

// check whether payer cntm balance sufficient
func isBalanceSufficient(tx *types.Transaction, stateBatch *statestore.StateBatch) error {
	balance, err := getBalance(stateBatch, tx.Payer, utils.OngCcntmractAddress)
	if err != nil {
		return err
	}
	if balance < tx.GasLimit*tx.GasPrice {
		return fmt.Errorf("payer gas insufficient, need %d , only have %d", tx.GasLimit*tx.GasPrice, balance)
	}
	return nil
}

func costGas(payer common.Address, gas uint64, config *smartccntmract.Config,
	cache *storage.CloneCache, store store.LedgerStore) ([]*event.NotifyEventInfo, error) {

	nativeTransferCode := genNativeTransferCode(utils.OngCcntmractAddress, payer,
		utils.GovernanceCcntmractAddress, gas)

	sc := smartccntmract.SmartCcntmract{
		Config:     config,
		CloneCache: cache,
		Store:      store,
		Code:       nativeTransferCode,
		Gas:        math.MaxUint64,
	}

	_, err := sc.Execute()

	if err != nil {
		return nil, err
	}
	return sc.Notifications, nil
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
