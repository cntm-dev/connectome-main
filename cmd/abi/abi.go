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
	NATIVE_PARAM_TYPE_BOOL      = "bool"
	NATIVE_PARAM_TYPE_BYTE      = "byte"
	NATIVE_PARAM_TYPE_INTEGER   = "int"
	NATIVE_PARAM_TYPE_STRING    = "string"
	NATIVE_PARAM_TYPE_BYTEARRAY = "bytearray"
	NATIVE_PARAM_TYPE_ARRAY     = "array"
	NATIVE_PARAM_TYPE_ADDRESS   = "address"
	NATIVE_PARAM_TYPE_UINT256   = "uint256"
	NATIVE_PARAM_TYPE_STRUCT    = "struct"
)

type NativeCcntmractAbi struct {
	Address   string                       `json:"hash"`
	Functions []*NativeCcntmractFunctionAbi `json:"functions"`
	Events    []*NativeCcntmractEventAbi    `json:"events"`
}

type NativeCcntmractFunctionAbi struct {
	Name       string                    `json:"name"`
	Parameters []*NativeCcntmractParamAbi `json:"parameters"`
	ReturnType string                    `json:"returnType"`
}

type NativeCcntmractParamAbi struct {
	Name    string                    `json:"name"`
	Type    string                    `json:"type"`
	SubType []*NativeCcntmractParamAbi `json:"subType"`
}

type NativeCcntmractEventAbi struct {
	Name       string                    `json:"name"`
	Parameters []*NativeCcntmractParamAbi `json:"parameters"`
}

func (this *NativeCcntmractAbi) GetFunc(name string) *NativeCcntmractFunctionAbi {
	name = strings.ToLower(name)
	for _, funcAbi := range this.Functions {
		if strings.ToLower(funcAbi.Name) == name {
			return funcAbi
		}
	}
	return nil
}

func (this *NativeCcntmractAbi) GetEvent(name string) *NativeCcntmractEventAbi {
	name = strings.ToLower(name)
	for _, evtAbi := range this.Events {
		if strings.ToLower(evtAbi.Name) == name {
			return evtAbi
		}
	}
	return nil
}
