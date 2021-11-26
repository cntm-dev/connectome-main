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

package rpc

import (
	Err "github.com/Ontology/http/base/error"
)

var (
	RpcInvalidHash        = responsePacking(Err.INVALID_PARAMS, "invalid hash")
	RpcInvalidBlock       = responsePacking(Err.INVALID_BLOCK, "invalid block")
	RpcInvalidTransaction = responsePacking(Err.INVALID_TRANSACTION, "invalid transaction")
	RpcInvalidParameter   = responsePacking(Err.INVALID_PARAMS, "invalid parameter")

	RpcUnknownBlock       = responsePacking(Err.UNKNOWN_BLOCK, "unknown block")
	RpcUnknownTransaction = responsePacking(Err.UNKNOWN_TRANSACTION, "unknown transaction")
	RpcUnKnownCcntmact     = responsePacking(Err.UNKNWN_CcntmRACT, "unknow ccntmract")

	RpcNil           = responsePacking(Err.INVALID_PARAMS, nil)
	RpcUnsupported   = responsePacking(Err.INTERNAL_ERROR, "Unsupported")
	RpcInternalError = responsePacking(Err.INTERNAL_ERROR, "internal error")
	RpcIOError       = responsePacking(Err.INTERNAL_ERROR, "internal IO error")
	RpcAPIError      = responsePacking(Err.INTERNAL_ERROR, "internal API error")

	RpcFailed          = responsePacking(Err.INTERNAL_ERROR, false)
	RpcAccountNotFound = responsePacking(Err.INTERNAL_ERROR, "Account not found")

	RpcSuccess = responsePacking(Err.SUCCESS, true)
	Rpc        = responseSuccess
)

func responseSuccess(result interface{}) map[string]interface{} {
	return responsePacking(Err.SUCCESS, result)
}
func responsePacking(errcode int64, result interface{}) map[string]interface{} {
	resp := map[string]interface{}{
		"error":  errcode,
		"desc":   Err.ErrMap[errcode],
		"result": result,
	}
	return resp
}
