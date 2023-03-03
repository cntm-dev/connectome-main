/*
 * Copyright (C) 2018 The cntm Authors
 * This file is part of The cntm library.
 *
 * The cntm is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The cntm is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The cntm.  If not, see <http://www.gnu.org/licenses/>.
 */
package abi

import "strings"

const (
	CNTMVM_PARAM_TYPE_BOOL       = "boolean"
	CNTMVM_PARAM_TYPE_STRING     = "string"
	CNTMVM_PARAM_TYPE_INTEGER    = "integer"
	CNTMVM_PARAM_TYPE_ARRAY      = "array"
	CNTMVM_PARAM_TYPE_BYTE_ARRAY = "bytearray"
	CNTMVM_PARAM_TYPE_VOID       = "void"
	CNTMVM_PARAM_TYPE_ANY        = "any"
)

type CntmvmContractAbi struct {
	Address    string                      `json:"hash"`
	EntryPoint string                      `json:"entrypoint"`
	Functions  []*CntmvmContractFunctionAbi `json:"functions"`
	Events     []*CntmvmContractEventAbi    `json:"events"`
}

func (this *CntmvmContractAbi) GetFunc(method string) *CntmvmContractFunctionAbi {
	method = strings.ToLower(method)
	for _, funcAbi := range this.Functions {
		if strings.ToLower(funcAbi.Name) == method {
			return funcAbi
		}
	}
	return nil
}

func (this *CntmvmContractAbi) GetEvent(evt string) *CntmvmContractEventAbi {
	evt = strings.ToLower(evt)
	for _, evtAbi := range this.Events {
		if strings.ToLower(evtAbi.Name) == evt {
			return evtAbi
		}
	}
	return nil
}

type CntmvmContractFunctionAbi struct {
	Name       string                    `json:"name"`
	Parameters []*CntmvmContractParamsAbi `json:"parameters"`
	ReturnType string                    `json:"returntype"`
}

type CntmvmContractParamsAbi struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type CntmvmContractEventAbi struct {
	Name       string                    `json:"name"`
	Parameters []*CntmvmContractParamsAbi `json:"parameters"`
	ReturnType string                    `json:"returntype"`
}
