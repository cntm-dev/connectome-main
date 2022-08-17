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
package abi

import "strings"

const (
	NEOVM_PARAM_TYPE_BOOL       = "boolean"
	NEOVM_PARAM_TYPE_STRING     = "string"
	NEOVM_PARAM_TYPE_INTEGER    = "integer"
	NEOVM_PARAM_TYPE_ARRAY      = "array"
	NEOVM_PARAM_TYPE_BYTE_ARRAY = "bytearray"
	NEOVM_PARAM_TYPE_VOID       = "void"
	NEOVM_PARAM_TYPE_ANY        = "any"
)

type NeovmCcntmractAbi struct {
	Address    string                      `json:"hash"`
	EntryPoint string                      `json:"entrypoint"`
	Functions  []*NeovmCcntmractFunctionAbi `json:"functions"`
	Events     []*NeovmCcntmractEventAbi    `json:"events"`
}

func (this *NeovmCcntmractAbi) GetFunc(method string) *NeovmCcntmractFunctionAbi {
	method = strings.ToLower(method)
	for _, funcAbi := range this.Functions {
		if strings.ToLower(funcAbi.Name) == method {
			return funcAbi
		}
	}
	return nil
}

func (this *NeovmCcntmractAbi) GetEvent(evt string) *NeovmCcntmractEventAbi {
	evt = strings.ToLower(evt)
	for _, evtAbi := range this.Events {
		if strings.ToLower(evtAbi.Name) == evt {
			return evtAbi
		}
	}
	return nil
}

type NeovmCcntmractFunctionAbi struct {
	Name       string                    `json:"name"`
	Parameters []*NeovmCcntmractParamsAbi `json:"parameters"`
	ReturnType string                    `json:"returntype"`
}

type NeovmCcntmractParamsAbi struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type NeovmCcntmractEventAbi struct {
	Name       string                    `json:"name"`
	Parameters []*NeovmCcntmractParamsAbi `json:"parameters"`
	ReturnType string                    `json:"returntype"`
}
