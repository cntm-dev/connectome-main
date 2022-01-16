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
	"encoding/binary"

	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/core/store"
	scommon "github.com/cntmio/cntmology/core/store/common"
	ctypes "github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/errors"
	"github.com/cntmio/cntmology/smartccntmract/ccntmext"
	"github.com/cntmio/cntmology/smartccntmract/event"
	"github.com/cntmio/cntmology/smartccntmract/service/native"
	sneovm "github.com/cntmio/cntmology/smartccntmract/service/neovm"
	"github.com/cntmio/cntmology/smartccntmract/service/wasm"
	stypes "github.com/cntmio/cntmology/smartccntmract/types"
	"github.com/cntmio/cntmology/vm/neovm"
	"github.com/cntmio/cntmology/vm/neovm/interfaces"
	vmtypes "github.com/cntmio/cntmology/vm/types"
	"github.com/cntmio/cntmology/vm/wasmvm/exec"
	"github.com/cntmio/cntmology/vm/wasmvm/util"
)

type SmartCcntmract struct {
	Ccntmext       []*ccntmext.Ccntmext
	Config        *Config
	Engine        Engine
	Notifications []*event.NotifyEventInfo
}

type Config struct {
	Time    uint32
	Height  uint32
	Tx      *ctypes.Transaction
	Table   interfaces.CodeTable
	DBCache scommon.StateStore
	Store   store.LedgerStore
}

type Engine interface {
	StepInto()
}

//put current ccntmext to smart ccntmract
func (sc *SmartCcntmract) PushCcntmext(ccntmext *ccntmext.Ccntmext) {
	sc.Ccntmext = append(sc.Ccntmext, ccntmext)
}

//get smart ccntmract current ccntmext
func (sc *SmartCcntmract) CurrentCcntmext() *ccntmext.Ccntmext {
	if len(sc.Ccntmext) < 1 {
		return nil
	}
	return sc.Ccntmext[len(sc.Ccntmext)-1]
}

//get smart ccntmract caller ccntmext
func (sc *SmartCcntmract) CallingCcntmext() *ccntmext.Ccntmext {
	if len(sc.Ccntmext) < 2 {
		return nil
	}
	return sc.Ccntmext[len(sc.Ccntmext)-2]
}

//get smart ccntmract entry entrance ccntmext
func (sc *SmartCcntmract) EntryCcntmext() *ccntmext.Ccntmext {
	if len(sc.Ccntmext) < 1 {
		return nil
	}
	return sc.Ccntmext[0]
}

//pop smart ccntmract current ccntmext
func (sc *SmartCcntmract) PopCcntmext() {
	sc.Ccntmext = sc.Ccntmext[:len(sc.Ccntmext)-1]
}

func (sc *SmartCcntmract) PushNotifications(notifications []*event.NotifyEventInfo) {
	sc.Notifications = append(sc.Notifications, notifications...)
}

func (sc *SmartCcntmract) Execute() error {
	ctx := sc.CurrentCcntmext()
	switch ctx.Code.VmType {
	case vmtypes.Native:
		service := native.NewNativeService(sc.Config.DBCache, sc.Config.Height, sc.Config.Tx, sc)
		if err := service.Invoke(); err != nil {
			return err
		}
	case vmtypes.NEOVM:
		stateMachine := sneovm.NewStateMachine(sc.Config.Store, sc.Config.DBCache, stypes.Application, sc.Config.Time)
		engine := neovm.NewExecutionEngine(
			sc.Config.Tx,
			new(neovm.ECDsaCrypto),
			sc.Config.Table,
			stateMachine,
		)
		engine.LoadCode(ctx.Code.Code, false)
		if err := engine.Execute(); err != nil {
			return err
		}
		stateMachine.CloneCache.Commit()
		sc.Notifications = append(sc.Notifications, stateMachine.Notifications...)
	case vmtypes.WASMVM:
		//todo refactor following code to match Neovm
		stateMachine := wasm.NewWasmStateMachine(sc.Config.Store, sc.Config.DBCache, stypes.Application, sc.Config.Time)
		engine := exec.NewExecutionEngine(
			sc.Config.Tx,
			new(util.ECDsaCrypto),
			sc.Config.Table,
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

		input := ctx.Code.Code[len(ccntmractCode)+1:]
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
		sc.Notifications = append(sc.Notifications, stateMachine.Notifications...)
	}
	return nil
}

func (sc *SmartCcntmract) CheckWitness(address common.Address) bool {
	if vmtypes.IsVmCodeAddress(address) {
		for _, v := range sc.Ccntmext {
			if v.CcntmractAddress == address {
				return true
			}
		}
	} else {
		addresses := sc.Config.Tx.GetSignatureAddresses()
		for _, v := range addresses {
			if v == address {
				return true
			}
		}
	}

	return false
}
