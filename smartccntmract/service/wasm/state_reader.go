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

package wasm

import (
	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/log"
	"github.com/cntmio/cntmology/core/store"
	"github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/errors"
	"github.com/cntmio/cntmology/smartccntmract/event"
	"github.com/cntmio/cntmology/vm/wasmvm/exec"
)

type WasmStateReader struct {
	serviceMap    map[string]func(*exec.ExecutionEngine) (bool, error)
	Notifications []*event.NotifyEventInfo
	ldgerStore    store.LedgerStore
}

func NewWasmStateReader(ldgerStore store.LedgerStore) *WasmStateReader {
	i := &WasmStateReader{
		ldgerStore: ldgerStore,
		serviceMap: make(map[string]func(*exec.ExecutionEngine) (bool, error)),
	}

	i.Register("GetBlockHeight", i.Getblockheight)
	i.Register("RuntimeNotify", i.RuntimeNotify)

	return i
}

func (i *WasmStateReader) Register(name string, handler func(*exec.ExecutionEngine) (bool, error)) bool {
	if _, ok := i.serviceMap[name]; ok {
		return false
	}
	i.serviceMap[name] = handler
	return true
}

func (i *WasmStateReader) Invoke(methodName string, engine *exec.ExecutionEngine) (bool, error) {

	if v, ok := i.serviceMap[methodName]; ok {
		return v(engine)
	}
	return true, errors.NewErr("Not supported method:" + methodName)
}

func (i *WasmStateReader) MergeMap(mMap map[string]func(*exec.ExecutionEngine) (bool, error)) bool {

	for k, v := range mMap {
		if _, ok := i.serviceMap[k]; !ok {
			i.serviceMap[k] = v
		}
	}
	return true
}

func (i *WasmStateReader) GetServiceMap() map[string]func(*exec.ExecutionEngine) (bool, error) {
	return i.serviceMap
}

func (i *WasmStateReader) Exists(name string) bool {
	_, ok := i.serviceMap[name]
	return ok
}

//============================block apis here============================/

func (i *WasmStateReader) Getblockheight(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()

	h := i.ldgerStore.GetCurrentBlockHeight()
	vm.RestoreCtx()
	if vm.GetEnvCall().GetReturns() {
		vm.PushResult(uint64(h))
	}
	return true, nil
}

func (i *WasmStateReader) RuntimeNotify(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()

	if len(params) != 1 {
		return false, errors.NewErr("[RuntimeNotify] get Parameter count error!")
	}

	returnStr, err := vm.GetPointerMemory(params[0])
	if err != nil {
		return false, err
	}

	tran, ok := engine.CodeCcntmainer.(*types.Transaction)
	if !ok {
		log.Error("[RuntimeNotify] Ccntmainer not transaction!")
		return false, errors.NewErr("[RuntimeNotify] Ccntmainer not transaction!")
	}

	hash := engine.GetVM().CodeHash

	txid := tran.Hash()

	i.Notifications = append(i.Notifications, &event.NotifyEventInfo{TxHash: txid, CcntmractAddress: hash, States: []interface{}{common.ToHexString([]byte(returnStr))}})
	vm.RestoreCtx()

	return true, nil

}
