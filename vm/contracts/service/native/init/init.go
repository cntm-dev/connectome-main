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
 * alcntg with The cntm.  If not, see <http://www.gnu.org/licenses/>.
 */

package init

import (
	"bytes"
	"math/big"

	"github.com/conntectome/cntm/common"
	"github.com/conntectome/cntm/smartcontract/service/native/auth"
	"github.com/conntectome/cntm/smartcontract/service/native/cross_chain/cross_chain_manager"
	"github.com/conntectome/cntm/smartcontract/service/native/cross_chain/header_sync"
	"github.com/conntectome/cntm/smartcontract/service/native/cross_chain/lock_proxy"
	params "github.com/conntectome/cntm/smartcontract/service/native/global_params"
	"github.com/conntectome/cntm/smartcontract/service/native/governance"
	"github.com/conntectome/cntm/smartcontract/service/native/cntg"
	"github.com/conntectome/cntm/smartcontract/service/native/cntm"
	"github.com/conntectome/cntm/smartcontract/service/native/cntmid"
	"github.com/conntectome/cntm/smartcontract/service/native/utils"
	"github.com/conntectome/cntm/smartcontract/service/cntmvm"
	vm "github.com/conntectome/cntm/vm/cntmvm"
)

var (
	COMMIT_DPOS_BYTES = InitBytes(utils.GovernanceCcntmractAddress, governance.COMMIT_DPOS)
)

func init() {
	cntg.InitCntg()
	cntm.InitCntm()
	params.InitGlobalParams()
	cntmid.Init()
	auth.Init()
	governance.InitGovernance()
	cross_chain_manager.InitCrossChain()
	header_sync.InitHeaderSync()
	lock_proxy.InitLockProxy()
}

func InitBytes(addr common.Address, method string) []byte {
	bf := new(bytes.Buffer)
	builder := vm.NewParamsBuilder(bf)
	builder.EmitPushByteArray([]byte{})
	builder.EmitPushByteArray([]byte(method))
	builder.EmitPushByteArray(addr[:])
	builder.EmitPushInteger(big.NewInt(0))
	builder.Emit(vm.SYSCALL)
	builder.EmitPushByteArray([]byte(cntmvm.NATIVE_INVOKE_NAME))

	return builder.ToArray()
}
