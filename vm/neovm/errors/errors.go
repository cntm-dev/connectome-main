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

package errors

import "errors"

var (
	ERR_BAD_VALUE                = errors.New("bad value")
	ERR_BAD_TYPE                 = errors.New("bad type")
	ERR_OVER_STACK_LEN           = errors.New("the count over the stack length")
	ERR_OVER_CODE_LEN            = errors.New("the count over the code length")
	ERR_UNDER_STACK_LEN          = errors.New("the count under the stack length")
	ERR_FAULT                    = errors.New("the exeution meet fault")
	ERR_NOT_SUPPORT_SERVICE      = errors.New("the service is not registered")
	ERR_NOT_SUPPORT_OPCODE       = errors.New("does not support the operation code")
	ERR_OVER_LIMIT_STACK         = errors.New("the stack over max size")
	ERR_OVER_MAX_ITEM_SIZE       = errors.New("the item over max size")
	ERR_OVER_MAX_ARRAY_SIZE      = errors.New("the array over max size")
	ERR_OVER_MAX_BIGINTEGER_SIZE = errors.New("the biginteger over max size 32bit")
	ERR_OUT_OF_GAS               = errors.New("out of gas")
	ERR_NOT_ARRAY                = errors.New("not array")
	ERR_TABLE_IS_NIL             = errors.New("table is nil")
	ERR_SERVICE_IS_NIL           = errors.New("service is nil")
	ERR_DIV_MOD_BY_ZERO          = errors.New("div or mod by zero")
	ERR_SHIFT_BY_NEG             = errors.New("shift by negtive value")
	ERR_EXECUTION_CcntmEXT_NIL    = errors.New("execution ccntmext is nil")
	ERR_CURRENT_CcntmEXT_NIL      = errors.New("current ccntmext is nil")
	ERR_CALLING_CcntmEXT_NIL      = errors.New("calling ccntmext is nil")
	ERR_ENTRY_CcntmEXT_NIL        = errors.New("entry ccntmext is nil")
	ERR_APPEND_NOT_ARRAY         = errors.New("append not array")
	ERR_NOT_SUPPORT_TYPE         = errors.New("not a supported type")
	ERR_MAP_NOT_EXIST            = errors.New("map not ccntmain key")
)
