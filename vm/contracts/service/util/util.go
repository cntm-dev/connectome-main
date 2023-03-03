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
package util

import (
	"bytes"
	"errors"

	"github.com/conntectome/cntm/common"
	"github.com/conntectome/cntm/core/utils"
	"github.com/conntectome/cntm/smartcontract/context"
	cntmvms "github.com/conntectome/cntm/smartcontract/service/cntmvm"
	"github.com/conntectome/cntm/vm/crossvm_codec"
	"github.com/conntectome/cntm/vm/cntmvm"
)

func BuildCntmVMParamEvalStack(params []interface{}) (*cntmvm.ValueStack, error) {
	builder := cntmvm.NewParamsBuilder(new(bytes.Buffer))
	err := utils.BuildCntmVMParam(builder, params)
	if err != nil {
		return nil, err
	}

	exec := cntmvm.NewExecutor(builder.ToArray(), cntmvm.VmFeatureFlag{true, true})
	err = exec.Execute()
	if err != nil {
		return nil, err
	}
	return exec.EvalStack, nil
}

//create paramters for cntmvm contract
func GenerateCntmVMParamEvalStack(input []byte) (*cntmvm.ValueStack, error) {
	params, err := crossvm_codec.DeserializeCallParam(input)
	if err != nil {
		return nil, err
	}

	list, ok := params.([]interface{})
	if ok == false {
		return nil, errors.New("invoke cntmvm param is not list type")
	}

	stack, err := BuildCntmVMParamEvalStack(list)
	if err != nil {
		return nil, err
	}

	return stack, nil
}

func SetCntmServiceParamAndEngine(addr common.Address, engine context.Engine, stack *cntmvm.ValueStack) error {
	service, ok := engine.(*cntmvms.CntmVmService)
	if ok == false {
		return errors.New("engine should be CntmVmService")
	}

	code, err := service.GetCntmContract(addr)
	if err != nil {
		return err
	}

	feature := service.Engine.Features
	service.Engine = cntmvm.NewExecutor(code, feature)
	service.Code = code

	service.Engine.EvalStack = stack

	return nil
}
