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

package init

import (
	"bytes"
	"math/big"

	"github.com/cntmio/cntmology/smartccntmract/service/native/system"

	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/smartccntmract/service/native/auth"
	"github.com/cntmio/cntmology/smartccntmract/service/native/cross_chain/cross_chain_manager"
	"github.com/cntmio/cntmology/smartccntmract/service/native/cross_chain/header_sync"
	"github.com/cntmio/cntmology/smartccntmract/service/native/cross_chain/lock_proxy"
	params "github.com/cntmio/cntmology/smartccntmract/service/native/global_params"
	"github.com/cntmio/cntmology/smartccntmract/service/native/governance"
	"github.com/cntmio/cntmology/smartccntmract/service/native/cntm"
	"github.com/cntmio/cntmology/smartccntmract/service/native/cntm"
	"github.com/cntmio/cntmology/smartccntmract/service/native/cntmfs"
	"github.com/cntmio/cntmology/smartccntmract/service/native/cntmid"
	"github.com/cntmio/cntmology/smartccntmract/service/native/utils"
	"github.com/cntmio/cntmology/smartccntmract/service/neovm"
	vm "github.com/cntmio/cntmology/vm/neovm"
)

var (
	COMMIT_DPOS_BYTES = InitBytes(utils.GovernanceCcntmractAddress, governance.COMMIT_DPOS)
)

func init() {
	cntm.InitOng()
	cntm.InitOnt()
	params.InitGlobalParams()
	cntmid.Init()
	auth.Init()
	governance.InitGovernance()
	cross_chain_manager.InitCrossChain()
	header_sync.InitHeaderSync()
	lock_proxy.InitLockProxy()
	cntmfs.InitFs()
	system.InitSystem()
}

func InitBytes(addr common.Address, method string) []byte {
	bf := new(bytes.Buffer)
	builder := vm.NewParamsBuilder(bf)
	builder.EmitPushByteArray([]byte{})
	builder.EmitPushByteArray([]byte(method))
	builder.EmitPushByteArray(addr[:])
	builder.EmitPushInteger(big.NewInt(0))
	builder.Emit(vm.SYSCALL)
	builder.EmitPushByteArray([]byte(neovm.NATIVE_INVOKE_NAME))

	return builder.ToArray()
}
