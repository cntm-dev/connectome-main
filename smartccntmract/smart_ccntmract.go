// Copyright 2017 The Ontology Authors
// This file is part of the Ontology library.
//
// The Ontology library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Ontology library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// alcntm with the Ontology library. If not, see <http://www.gnu.org/licenses/>.

package smartccntmract

import (
	"bytes"

	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/core/payload"
	"github.com/cntmio/cntmology/core/store"
	scommon "github.com/cntmio/cntmology/core/store/common"
	ctypes "github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/errors"
	"github.com/cntmio/cntmology/smartccntmract/ccntmext"
	"github.com/cntmio/cntmology/smartccntmract/event"
	"github.com/cntmio/cntmology/smartccntmract/service/native"
	"github.com/cntmio/cntmology/smartccntmract/service/neovm"
	"github.com/cntmio/cntmology/smartccntmract/service/wasmvm"
	"github.com/cntmio/cntmology/smartccntmract/states"
	"github.com/cntmio/cntmology/smartccntmract/storage"
	stypes "github.com/cntmio/cntmology/smartccntmract/types"
	vm "github.com/cntmio/cntmology/vm/neovm"
)

var (
	CcntmRACT_NOT_EXIST    = errors.NewErr("[AppCall] Get ccntmract ccntmext nil")
	DEPLOYCODE_TYPE_ERROR = errors.NewErr("[AppCall] DeployCode type error!")
	INVOKE_CODE_EXIST     = errors.NewErr("[AppCall] Invoke codes exist!")
	ENGINE_NOT_SUPPORT    = errors.NewErr("[Execute] Engine doesn't support!")
)

// SmartCcntmract describe smart ccntmract execute engine
type SmartCcntmract struct {
	Ccntmexts      []*ccntmext.Ccntmext  // all execute smart ccntmract ccntmext
	CloneCache    *storage.CloneCache // state cache
	Store         store.LedgerStore   // ledger store
	Config        *Config
	Engine        Engine
	Code          stypes.VmCode
	Notifications []*event.NotifyEventInfo // all execute smart ccntmract event notify info
}

// Config describe smart ccntmract need parameters configuration
type Config struct {
	Time   uint32              // current block timestamp
	Height uint32              // current block height
	Tx     *ctypes.Transaction // current transaction
}

type Engine interface {
	Invoke() (interface{}, error)
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
	if len(this.Ccntmexts) > 0 {
		this.Ccntmexts = this.Ccntmexts[:len(this.Ccntmexts)-1]
	}
}

// PushNotifications push smart ccntmract event info
func (this *SmartCcntmract) PushNotifications(notifications []*event.NotifyEventInfo) {
	this.Notifications = append(this.Notifications, notifications...)
}

// Execute is smart ccntmract execute manager
// According different vm type to launch different service
func (this *SmartCcntmract) Execute() (interface{}, error) {
	var engine Engine
	switch this.Code.VmType {
	case stypes.Native:
		service := &native.NativeService{
			CloneCache: this.CloneCache,
			Code:       this.Code.Code,
			Tx:         this.Config.Tx,
			Height:     this.Config.Height,
			CcntmextRef: this,
		}
		service.InitService()
		engine = service
	case stypes.NEOVM:
		engine = &neovm.NeoVmService{
			Store:      this.Store,
			CloneCache: this.CloneCache,
			CcntmextRef: this,
			Code:       this.Code.Code,
			Tx:         this.Config.Tx,
			Time:       this.Config.Time,
		}
	case stypes.WASMVM:
		engine = &wasmvm.WasmVmService{
			Store:      this.Store,
			CloneCache: this.CloneCache,
			CcntmextRef: this,
			Code:       this.Code.Code,
			Tx:         this.Config.Tx,
			Time:       this.Config.Time,
		}
	default:
		return nil, ENGINE_NOT_SUPPORT
	}
	return engine.Invoke()
}

// AppCall a smart ccntmract, if ccntmract exist on blockchain, you should set the address
// Param address: invoke smart ccntmract on blockchain according ccntmract address
// Param method: invoke smart ccntmract method name
// Param codes: invoke smart ccntmract off blockchain
// Param args: invoke smart ccntmract args
func (this *SmartCcntmract) AppCall(address common.Address, method string, codes, args []byte) (interface{}, error) {
	var code []byte
	vmType := stypes.VmType(address[0])
	switch vmType {
	case stypes.Native:
		bf := new(bytes.Buffer)
		c := states.Ccntmract{
			Address: address,
			Method:  method,
			Args:    args,
		}
		if err := c.Serialize(bf); err != nil {
			return nil, err
		}
		code = bf.Bytes()
	case stypes.NEOVM:
		c, err := this.loadCode(address, codes)
		if err != nil {
			return nil, err
		}
		var temp []byte
		build := vm.NewParamsBuilder(new(bytes.Buffer))
		if method != "" {
			build.EmitPushByteArray([]byte(method))
		}
		temp = append(args, build.ToArray()...)
		code = append(temp, c...)
		vmCode := stypes.VmCode{Code: c, VmType: stypes.NEOVM}
		this.PushCcntmext(&ccntmext.Ccntmext{CcntmractAddress: vmCode.AddressFromVmCode()})
	case stypes.WASMVM:
		c, err := this.loadCode(address, codes)
		if err != nil {
			return nil, err
		}
		bf := new(bytes.Buffer)
		ccntmract := states.Ccntmract{
			Version: 1, //fix to > 0
			Address: address,
			Method:  method,
			Args:    args,
			Code:    c,
		}
		if err := ccntmract.Serialize(bf); err != nil {
			return nil, err
		}
		code = bf.Bytes()
	}

	this.Code = stypes.VmCode{Code: code, VmType: vmType}
	res, err := this.Execute()
	if err != nil {
		return nil, err
	}

	return res, nil
}

// CheckWitness check whether authorization correct
// If address is wallet address, check whether in the signature addressed list
// Else check whether address is calling ccntmract address
// Param address: wallet address or ccntmract address
func (this *SmartCcntmract) CheckWitness(address common.Address) bool {
	if stypes.IsVmCodeAddress(address) {
		if this.CallingCcntmext() != nil && this.CallingCcntmext().CcntmractAddress == address {
			return true
		}
	} else {
		addresses := this.Config.Tx.GetSignatureAddresses()
		for _, v := range addresses {
			if v == address {
				return true
			}
		}
	}

	return false
}

// loadCode load smart ccntmract execute code
// Param address, invoke on blockchain smart ccntmract address
// Param codes, invoke off blockchain smart ccntmract code
// If you invoke off blockchain smart ccntmract, you can set address is codes address
// But this address doesn't deployed on blockchain
func (this *SmartCcntmract) loadCode(address common.Address, codes []byte) ([]byte, error) {
	isLoad := false
	if len(codes) == 0 {
		isLoad = true
	}
	item, err := this.getCcntmract(address[:])
	if err != nil {
		return nil, err
	}
	if isLoad {
		if item == nil {
			return nil, CcntmRACT_NOT_EXIST
		}
		ccntmract, ok := item.Value.(*payload.DeployCode)
		if !ok {
			return nil, DEPLOYCODE_TYPE_ERROR
		}
		return ccntmract.Code.Code, nil
	} else {
		if item != nil {
			return nil, INVOKE_CODE_EXIST
		}
		return codes, nil
	}
}

func (this *SmartCcntmract) getCcntmract(address []byte) (*scommon.StateItem, error) {
	item, err := this.CloneCache.Store.TryGet(scommon.ST_CcntmRACT, address[:])
	if err != nil {
		return nil, errors.NewErr("[getCcntmract] Get ccntmract ccntmext error!")
	}
	return item, nil
}
