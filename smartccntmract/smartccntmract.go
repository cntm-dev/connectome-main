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
	"encoding/binary"
)

type SmartCcntmract struct {
	Ccntmexts      []*ccntmext.Ccntmext
	Config        *Config
	Engine        Engine
	Notifications []*event.NotifyEventInfo
}

type Config struct {
	Time    uint32
	Height  uint32
	Tx      *ctypes.Transaction
	DBCache scommon.StateStore
	Store   store.LedgerStore
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
		//todo refactor following code to match Neovm
		stateMachine := wasm.NewWasmStateMachine(this.Config.Store, this.Config.DBCache, this.Config.Time)
		engine := exec.NewExecutionEngine(
			this.Config.Tx,
			new(util.ECDsaCrypto),
			stateMachine,
			"product",
		)

		tmpcodes := bytes.Split(ctx.Code.Code, []byte(exec.PARAM_SPLITER))
		if len(tmpcodes) != 3 {
			return errors.NewErr("Wasm paramter count error")
		}
		ccntmractCode := tmpcodes[0]

		addr, err := common.AddressParseFromBytes(ccntmractCode)
		if err != nil {
			return errors.NewErr("get ccntmract address error")
		}

		dpcode, err := stateMachine.GetCcntmractCodeFromAddress(addr)
		if err != nil {
			return errors.NewErr("get ccntmract  error")
		}

		input := ctx.Code.Code[len(ccntmractCode) + 1:]
		res, err := engine.Call(ctx.CcntmractAddress, dpcode, input)
		if err != nil {
			return err
		}

		//todo how to deal with the result???
		_, err = engine.GetVM().GetPointerMemory(uint64(binary.LittleEndian.Uint32(res)))
		if err != nil {
			return err
		}

		stateMachine.CloneCache.Commit()
		this.Notifications = append(this.Notifications, stateMachine.Notifications...)
	}
	return nil
}

func (this *SmartCcntmract) AppCall(address common.Address, method string, codes, args []byte, isLoad bool) error {
	var code []byte
	if isLoad {
		c, err := this.getCcntmract(address[:]); if err != nil {
			return err
		}
		code = c.Code.Code
	}

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
		var temp []byte
		build := vm.NewParamsBuilder(new(bytes.Buffer))
		if method != "" {
			build.EmitPushByteArray([]byte(method))
		}
		temp = append(args, build.ToArray()...)
		if isLoad {
			code = append(temp, code...)
		} else {
			code = append(temp, codes...)
		}
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

func (this *SmartCcntmract) getCcntmract(address []byte) (*payload.DeployCode, error) {
	item, err := this.Config.DBCache.TryGet(scommon.ST_CcntmRACT, address[:]);
	if err != nil || item == nil || item.Value == nil {
		return nil, errors.NewErr("[getCcntmract] Get ccntmext doesn't exist!")
	}
	ccntmract, ok := item.Value.(*payload.DeployCode); if !ok {
		return nil, errors.NewErr("[getCcntmract] Type error!")
	}
	return ccntmract, nil
}
