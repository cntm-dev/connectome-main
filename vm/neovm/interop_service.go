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

package neovm

import (
	. "github.com/Ontology/vm/neovm/errors"
	"github.com/Ontology/common/log"
)

type IInteropService interface {
	Register(method string, handler func(*ExecutionEngine) (bool, error)) bool
	GetServiceMap() map[string]func(*ExecutionEngine) (bool, error)
}

type InteropService struct {
	serviceMap map[string]func(*ExecutionEngine) (bool, error)
}

func NewInteropService() *InteropService {
	var i InteropService
	i.serviceMap = make(map[string]func(*ExecutionEngine) (bool, error), 0)
	i.Register("System.ExecutionEngine.GetScriptCcntmainer", i.GetCodeCcntmainer)
	i.Register("System.ExecutionEngine.GetExecutingScriptHash", i.GetExecutingCodeHash)
	i.Register("System.ExecutionEngine.GetCallingScriptHash", i.GetCallingCodeHash)
	i.Register("System.ExecutionEngine.GetEntryScriptHash", i.GetEntryCodeHash)
	return &i
}

func (is *InteropService) Register(methodName string, handler func(*ExecutionEngine) (bool, error)) bool {
	if _, ok := is.serviceMap[methodName]; ok {
		return false
	}
	is.serviceMap[methodName] = handler
	return true
}

func (i *InteropService) MergeMap(dictionary map[string]func(*ExecutionEngine) (bool, error)) {
	for k, v := range dictionary {
		if _, ok := i.serviceMap[k]; !ok {
			i.serviceMap[k] = v
		}
	}
}

func (i *InteropService) GetServiceMap() map[string]func(*ExecutionEngine) (bool, error) {
	return i.serviceMap
}

func (i *InteropService) Invoke(methodName string, engine *ExecutionEngine) (bool, error) {
	if v, ok := i.serviceMap[methodName]; ok {
		log.Error("Invoke MethodName:", methodName)
		return v(engine)
	}
	return false, ErrNotSupportService
}

func (i *InteropService) GetCodeCcntmainer(engine *ExecutionEngine) (bool, error) {
	PushData(engine, engine.codeCcntmainer)
	return true, nil
}

func (i *InteropService) GetExecutingCodeHash(engine *ExecutionEngine) (bool, error) {
	ccntmext, err := engine.CurrentCcntmext()
	if err != nil {
		return false, err
	}
	codeHash, err := ccntmext.GetCodeHash()
	if err != nil {
		return false, err
	}
	PushData(engine, codeHash[:])
	return true, nil
}

func (i *InteropService) GetCallingCodeHash(engine *ExecutionEngine) (bool, error) {
	ccntmext, err := engine.CallingCcntmext()
	if err != nil {
		return false, err
	}
	codeHash, err := ccntmext.GetCodeHash()
	if err != nil {
		return false, err
	}
	PushData(engine, codeHash[:])
	return true, nil
}
func (i *InteropService) GetEntryCodeHash(engine *ExecutionEngine) (bool, error) {
	ccntmext, err := engine.EntryCcntmext()
	if err != nil {
		return false, err
	}
	codeHash, err := ccntmext.GetCodeHash()
	if err != nil {
		return false, err
	}
	PushData(engine, codeHash[:])
	return true, nil
}
