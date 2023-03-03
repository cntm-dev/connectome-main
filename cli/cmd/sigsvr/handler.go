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

package sigsvr

import "github.com/conntectome/cntm/cmd/sigsvr/handlers"

func init() {
	DefCliRpcSvr.RegHandler("createaccount", handlers.CreateAccount)
	DefCliRpcSvr.RegHandler("exportaccount", handlers.ExportAccount)
	DefCliRpcSvr.RegHandler("sigdata", handlers.SigData)
	DefCliRpcSvr.RegHandler("sigrawtx", handlers.SigRawTransaction)
	DefCliRpcSvr.RegHandler("sigmutilrawtx", handlers.SigMutilRawTransaction)
	DefCliRpcSvr.RegHandler("sigtransfertx", handlers.SigTransferTransaction)
	DefCliRpcSvr.RegHandler("sigcntmvminvoketx", handlers.SigCntmVMInvokeTx)
	DefCliRpcSvr.RegHandler("sigcntmvminvokeabitx", handlers.SigCntmVMInvokeAbiTx)
	DefCliRpcSvr.RegHandler("signativeinvoketx", handlers.SigNativeInvokeTx)
}
