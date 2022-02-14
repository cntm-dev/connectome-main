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
	"fmt"
	"bytes"
	"encoding/binary"

	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/core/store"
	scommon "github.com/cntmio/cntmology/core/store/common"
	ctypes "github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/errors"
	"github.com/cntmio/cntmology/smartccntmract/ccntmext"
	"github.com/cntmio/cntmology/smartccntmract/event"
	"github.com/cntmio/cntmology/smartccntmract/service/native"
	"github.com/cntmio/cntmology/smartccntmract/service/wasm"
	stypes "github.com/cntmio/cntmology/smartccntmract/types"
	"github.com/cntmio/cntmology/vm/wasmvm/exec"
	"github.com/cntmio/cntmology/vm/wasmvm/util"
	"github.com/cntmio/cntmology/smartccntmract/service/neovm"
	"github.com/cntmio/cntmology/core/payload"
	"github.com/cntmio/cntmology/smartccntmract/states"
	vm "github.com/cntmio/cntmology/vm/neovm"
)

var (
	CcntmRACT_NOT_EXIST = errors.NewErr("[AppCall] Get ccntmract ccntmext nil")
	DEPLOYCODE_TYPE_ERROR = errors.NewErr("[AppCall] DeployCode type error!")
	INVOKE_CODE_EXIST = errors.NewErr("[AppCall] Invoke codes exist!")
)

type SmartCcntmract struct {
	Ccntmexts      []*ccntmext.Ccntmext       // all execute smart ccntmract ccntmext
	Config        *Config
	Engine        Engine
	Notifications []*event.NotifyEventInfo // all execute smart ccntmract event notify info
}

type Config struct {
	Time    uint32              // current block timestamp
	Height  uint32              // current block height
	Tx      *ctypes.Transaction // current transaction
	DBCache scommon.StateStore  // db states cache
	Store   store.LedgerStore   // ledger store
}

type Engine interface {
	Invoke()
}

//put current ccntmext to smart ccntmract
func (this *SmartCcntmract) PushCcntmext(ccntmext *ccntmext.Ccntmext) {
	this.Ccntmexts = append(this.Ccntmexts, ccntmext)
}

//get smart ccntmract current ccntmext
func (this *SmartCcntmract) CurrentCcntmext() *ccntmext.Ccntmext {
	if len(this.Ccntmexts) < 1 {
		return nil
	}
	return this.Ccntmexts[len(this.Ccntmexts) - 1]
}

//get smart ccntmract caller ccntmext
func (this *SmartCcntmract) CallingCcntmext() *ccntmext.Ccntmext {
	if len(this.Ccntmexts) < 2 {
		return nil
	}
	return this.Ccntmexts[len(this.Ccntmexts) - 2]
}

//get smart ccntmract entry entrance ccntmext
func (this *SmartCcntmract) EntryCcntmext() *ccntmext.Ccntmext {
	if len(this.Ccntmexts) < 1 {
		return nil
	}
	return this.Ccntmexts[0]
}

//pop smart ccntmract current ccntmext
func (this *SmartCcntmract) PopCcntmext() {
	if len(this.Ccntmexts) > 0 {
		this.Ccntmexts = this.Ccntmexts[:len(this.Ccntmexts) - 1]
	}
}

// push smart ccntmract event info
func (this *SmartCcntmract) PushNotifications(notifications []*event.NotifyEventInfo) {
	this.Notifications = append(this.Notifications, notifications...)
}

func (this *SmartCcntmract) Execute() error {
	ctx := this.CurrentCcntmext()
	switch ctx.Code.VmType {
	case stypes.Native:
		service := native.NewNativeService(this.Config.DBCache, this.Config.Height, this.Config.Tx, this)
		if err := service.Invoke(); err != nil {
			return err
		}
	case stypes.NEOVM:
		service := neovm.NewNeoVmService(this.Config.Store, this.Config.DBCache, this.Config.Tx, this.Config.Time, this)
		if err := service.Invoke(); err != nil {
			fmt.Println("execute neovm error:", err)
			return err
		}
	case stypes.WASMVM:
		stateMachine := wasm.NewWasmStateMachine(this.Config.Store, this.Config.DBCache, this.Config.Time)

		engine := exec.NewExecutionEngine(
			this.Config.Tx,
			new(util.ECDsaCrypto),
			stateMachine,
		)

		ccntmract := &states.Ccntmract{}
		ccntmract.Deserialize(bytes.NewBuffer(ctx.Code.Code))
		addr := ccntmract.Address

		dpcode, err := stateMachine.GetCcntmractCodeFromAddress(addr)
		if err != nil {
			return errors.NewErr("get ccntmract  error")
		}

		var caller common.Address
		if this.CallingCcntmext() == nil {
			caller = common.Address{}
		} else {
			caller = this.CallingCcntmext().CcntmractAddress
		}
		res, err := engine.Call(caller, dpcode, ccntmract.Method, ccntmract.Args, ccntmract.Version)

		if err != nil {
			return err
		}

		//get the return message
		_, err = engine.GetVM().GetPointerMemory(uint64(binary.LittleEndian.Uint32(res)))
		if err != nil {
			return err
		}

		stateMachine.CloneCache.Commit()
		this.Notifications = append(this.Notifications, stateMachine.Notifications...)
	}
	return nil
}

// When you want to call a ccntmract use this function, if ccntmract exist in block chain, you should set isLoad true,
// Otherwise, you can set execute code, and set isLoad false.
// param address: smart ccntmract address
// param method: invoke smart ccntmract method name
// param codes: invoke smart ccntmract whether need to load code
// param args: invoke smart ccntmract args
func (this *SmartCcntmract) AppCall(address common.Address, method string, codes, args []byte) error {
	var code []byte

	vmType := stypes.VmType(address[0])

	switch vmType {
	case stypes.Native:
		bf := new(bytes.Buffer)
		c := states.Ccntmract{
			Address: address,
			Method: method,
			Args: args,
		}
		if err := c.Serialize(bf); err != nil {
			return err
		}
		code = bf.Bytes()
	case stypes.NEOVM:
		code, err := this.loadCode(address, codes);
		if err != nil {
			return nil
		}
		var temp []byte
		build := vm.NewParamsBuilder(new(bytes.Buffer))
		if method != "" {
			build.EmitPushByteArray([]byte(method))
		}
		temp = append(args, build.ToArray()...)
		code = append(temp, code...)
	case stypes.WASMVM:
	}

	this.PushCcntmext(&ccntmext.Ccntmext{
		Code: stypes.VmCode{
			Code: code,
			VmType: vmType,
		},
		CcntmractAddress: address,
	})

	if err := this.Execute(); err != nil {
		return err
	}

	this.PopCcntmext()
	return nil
}

// check authorization correct
// if address is wallet address, check whether in the signature addressed list
// else check whether address is calling ccntmract address
// param address: wallet address or ccntmract address
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

// load smart ccntmract execute code
// param address, invoke cntm smart ccntmract address
// param codes, invoke offchain smart ccntmract code
// if you invoke offchain smart ccntmract, you can set address is codes address
// but this address cann't find in blockchain
func (this *SmartCcntmract) loadCode(address common.Address, codes []byte) ([]byte, error) {
	isLoad := false
	if len(codes) == 0 {
		isLoad = true
	}
	item, err := this.getCcntmract(address[:]); if err != nil {
		return nil, err
	}
	if isLoad {
		if item == nil {
			return nil, CcntmRACT_NOT_EXIST
		}
		ccntmract, ok := item.Value.(*payload.DeployCode); if !ok {
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
	item, err := this.Config.DBCache.TryGet(scommon.ST_CcntmRACT, address[:]);
	if err != nil {
		return nil, errors.NewErr("[getCcntmract] Get ccntmract ccntmext error!")
	}
	return item, nil
}