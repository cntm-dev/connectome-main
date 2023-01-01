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
package util

import (
	"errors"
	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/core/utils"
	"github.com/cntmio/cntmology/vm/crossvm_codec"
)

//create paramters for neovm ccntmract
func CreateNeoInvokeParam(ccntmractAddress common.Address, input []byte) ([]byte, error) {
	params, err := crossvm_codec.DeserializeCallParam(input)
	if err != nil {
		return nil, err
	}

	list, ok := params.([]interface{})
	if ok == false {
		return nil, errors.New("invoke neovm param is not list type")
	}

	return utils.BuildNeoVMInvokeCode(ccntmractAddress, list)
}