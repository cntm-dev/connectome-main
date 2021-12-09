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
	vmtypes "github.com/Ontology/vm/types"
	"github.com/Ontology/vm/neovm/interfaces"
	ctypes "github.com/Ontology/core/types"
	"github.com/Ontology/smartccntmract/service/native"
	scommon "github.com/Ontology/core/store/common"
	sneovm "github.com/Ontology/smartccntmract/service/neovm"
	"github.com/Ontology/core/store"
	stypes "github.com/Ontology/smartccntmract/types"
	"github.com/Ontology/vm/neovm"
	"github.com/Ontology/smartccntmract/ccntmext"
	"github.com/Ontology/smartccntmract/event"
)

type SmartCcntmract struct {
	Ccntmext []*ccntmext.Ccntmext
	Config *Config
	Engine Engine
	Notifications []*event.NotifyEventInfo
}

type Config struct {
	Time uint32
	Height uint32
	Tx *ctypes.Transaction
	Table interfaces.ICodeTable
	DBCache scommon.IStateStore
	Store store.ILedgerStore
}

type Engine interface {
	StepInto()
}

//put current ccntmext to smart ccntmract
func(sc *SmartCcntmract) PushCcntmext(ccntmext *ccntmext.Ccntmext) {
	sc.Ccntmext = append(sc.Ccntmext, ccntmext)
}

//get smart ccntmract current ccntmext
func(sc *SmartCcntmract) CurrentCcntmext() *ccntmext.Ccntmext {
	if len(sc.Ccntmext) < 1 {
		return nil
	}
	return sc.Ccntmext[len(sc.Ccntmext) - 1]
}

//get smart ccntmract caller ccntmext
func(sc *SmartCcntmract) CallingCcntmext() *ccntmext.Ccntmext {
	if len(sc.Ccntmext) < 2 {
		return nil
	}
	return sc.Ccntmext[len(sc.Ccntmext) - 2]
}

//get smart ccntmract entry entrance ccntmext
func(sc *SmartCcntmract) EntryCcntmext() *ccntmext.Ccntmext {
	if len(sc.Ccntmext) < 1 {
		return nil
	}
	return sc.Ccntmext[0]
}

//pop smart ccntmract current ccntmext
func(sc *SmartCcntmract) PopCcntmext() {
	sc.Ccntmext = sc.Ccntmext[:len(sc.Ccntmext) - 1]
}

func (sc *SmartCcntmract) Execute() error {
	ctx := sc.CurrentCcntmext()
	switch ctx.Code.VmType {
	case vmtypes.Native:
		service := native.NewNativeService(sc.Config.DBCache, sc.Config.Height, sc.Config.Tx, sc)
		if err := service.Invoke(); err != nil {
			return err
		}
		service.CloneCache.Commit()
		sc.Notifications = append(sc.Notifications, service.Notifications...)
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
		sc.Notifications = append(sc.Notifications, stateMachine.Notifications...)
	case vmtypes.WASMVM:
	}
	return nil
}
