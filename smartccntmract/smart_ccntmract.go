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
package smartccntmract

import (
	"errors"
	"fmt"

	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/config"
	"github.com/cntmio/cntmology/common/log"
	"github.com/cntmio/cntmology/core/store"
	ctypes "github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/smartccntmract/ccntmext"
	"github.com/cntmio/cntmology/smartccntmract/event"
	"github.com/cntmio/cntmology/smartccntmract/service/native"
	"github.com/cntmio/cntmology/smartccntmract/service/neovm"
	"github.com/cntmio/cntmology/smartccntmract/service/wasmvm"
	"github.com/cntmio/cntmology/smartccntmract/storage"
	vm "github.com/cntmio/cntmology/vm/neovm"
)

const (
	MAX_EXECUTE_ENGINE = 128
)

// SmartCcntmract describe smart ccntmract execute engine
type SmartCcntmract struct {
	Ccntmexts      []*ccntmext.Ccntmext // all execute smart ccntmract ccntmext
	CacheDB       *storage.CacheDB   // state cache
	Store         store.LedgerStore  // ledger store
	Config        *Config
	Notifications []*event.NotifyEventInfo // all execute smart ccntmract event notify info
	GasTable      map[string]uint64
	Gas           uint64
	ExecStep      int
	PreExec       bool
}

// Config describe smart ccntmract need parameters configuration
type Config struct {
	Time      uint32              // current block timestamp
	Height    uint32              // current block height
	BlockHash common.Uint256      // current block hash
	Tx        *ctypes.Transaction // current transaction
}

// PushCcntmext push current ccntmext to smart ccntmract
func (this *SmartCcntmract) PushCcntmext(ccntmext *ccntmext.Ccntmext) {
	this.Ccntmexts = append(this.Ccntmexts, ccntmext)
}

// CurrentCcntmext return smart ccntmract current ccntmext
func (this *SmartCcntmract) CurrentCcntmext() *ccntmext.Ccntmext {
	if len(this.Ccntmexts) < 1 {
		return nil
	}
	return this.Ccntmexts[len(this.Ccntmexts)-1]
}

// CallingCcntmext return smart ccntmract caller ccntmext
func (this *SmartCcntmract) CallingCcntmext() *ccntmext.Ccntmext {
	if len(this.Ccntmexts) < 2 {
		return nil
	}
	return this.Ccntmexts[len(this.Ccntmexts)-2]
}

// EntryCcntmext return smart ccntmract entry entrance ccntmext
func (this *SmartCcntmract) EntryCcntmext() *ccntmext.Ccntmext {
	if len(this.Ccntmexts) < 1 {
		return nil
	}
	return this.Ccntmexts[0]
}

// PopCcntmext pop smart ccntmract current ccntmext
func (this *SmartCcntmract) PopCcntmext() {
	if len(this.Ccntmexts) > 1 {
		this.Ccntmexts = this.Ccntmexts[:len(this.Ccntmexts)-1]
	}
}

// PushNotifications push smart ccntmract event info
func (this *SmartCcntmract) PushNotifications(notifications []*event.NotifyEventInfo) {
	this.Notifications = append(this.Notifications, notifications...)
}

func (this *SmartCcntmract) CheckExecStep() bool {
	if this.ExecStep >= neovm.VM_STEP_LIMIT {
		return false
	}
	this.ExecStep += 1
	return true
}

func (this *SmartCcntmract) CheckUseGas(gas uint64) bool {
	if this.Gas < gas {
		return false
	}
	this.Gas -= gas
	return true
}

func (this *SmartCcntmract) checkCcntmexts() bool {
	if len(this.Ccntmexts) > MAX_EXECUTE_ENGINE {
		return false
	}
	return true
}

func NewVmFeatureFlag(blockHeight uint32) vm.VmFeatureFlag {
	var feature vm.VmFeatureFlag
	enableHeight := config.GetOpcodeUpdateCheckHeight(config.DefConfig.P2PNode.NetworkId)
	feature.DisableHasKey = blockHeight <= enableHeight

	return feature
}

// Execute is smart ccntmract execute manager
// According different vm type to launch different service
func (this *SmartCcntmract) NewExecuteEngine(code []byte, txtype ctypes.TransactionType) (ccntmext.Engine, error) {
	if !this.checkCcntmexts() {
		return nil, fmt.Errorf("%s", "engine over max limit!")
	}

	var service ccntmext.Engine
	switch txtype {
	case ctypes.InvokeNeo:
		feature := NewVmFeatureFlag(this.Config.Height)
		service = &neovm.NeoVmService{
			Store:      this.Store,
			CacheDB:    this.CacheDB,
			CcntmextRef: this,
			GasTable:   this.GasTable,
			Code:       code,
			Tx:         this.Config.Tx,
			Time:       this.Config.Time,
			Height:     this.Config.Height,
			BlockHash:  this.Config.BlockHash,
			Engine:     vm.NewExecutor(code, feature),
			PreExec:    this.PreExec,
		}
	case ctypes.InvokeWasm:
		gasFactor := this.GasTable[config.WASM_GAS_FACTOR]
		if gasFactor == 0 {
			gasFactor = config.DEFAULT_WASM_GAS_FACTOR
		}

		service = &wasmvm.WasmVmService{
			Store:      this.Store,
			CacheDB:    this.CacheDB,
			CcntmextRef: this,
			Code:       code,
			Tx:         this.Config.Tx,
			Time:       this.Config.Time,
			Height:     this.Config.Height,
			BlockHash:  this.Config.BlockHash,
			PreExec:    this.PreExec,
			GasLimit:   &this.Gas,
			GasFactor:  gasFactor,
		}
	default:
		return nil, errors.New("failed to construct execute engine, wrcntm transaction type")
	}

	return service, nil
}

func (this *SmartCcntmract) NewNativeService() (*native.NativeService, error) {
	if !this.checkCcntmexts() {
		return nil, fmt.Errorf("%s", "engine over max limit!")
	}
	service := &native.NativeService{
		CacheDB:    this.CacheDB,
		CcntmextRef: this,
		Tx:         this.Config.Tx,
		Time:       this.Config.Time,
		Height:     this.Config.Height,
		BlockHash:  this.Config.BlockHash,
		ServiceMap: make(map[string]native.Handler),
	}
	return service, nil
}

// CheckWitness check whether authorization correct
// If address is wallet address, check whether in the signature addressed list
// Else check whether address is calling ccntmract address
// Param address: wallet address or ccntmract address
func (this *SmartCcntmract) CheckWitness(address common.Address) bool {
	if this.checkAccountAddress(address) || this.checkCcntmractAddress(address) {
		return true
	}
	return false
}

func (this *SmartCcntmract) checkAccountAddress(address common.Address) bool {
	addresses, err := this.Config.Tx.GetSignatureAddresses()
	if err != nil {
		log.Errorf("get signature address error:%v", err)
		return false
	}
	for _, v := range addresses {
		if v == address {
			return true
		}
	}
	return false
}

func (this *SmartCcntmract) checkCcntmractAddress(address common.Address) bool {
	if this.CallingCcntmext() != nil && this.CallingCcntmext().CcntmractAddress == address {
		return true
	}
	return false
}
