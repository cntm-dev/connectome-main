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
package wasmvm

var (
	TIMESTAMP_GAS        uint64 = 1
	BLOCK_HEGHT_GAS      uint64 = 1
	SELF_ADDRESS_GAS     uint64 = 1
	CALLER_ADDRESS_GAS   uint64 = 1
	ENTRY_ADDRESS_GAS    uint64 = 1
	CHECKWITNESS_GAS     uint64 = 200
	CALL_CCNTMRACT_GAS    uint64 = 10
	CCNTMRACT_CREATE_GAS  uint64 = 20000000
	CCNTMRACT_MIGRATE_GAS uint64 = 20000000
	NATIVE_INVOKE_GAS    uint64 = 1000

	CURRENT_BLOCK_HASH_GAS uint64 = 100
	CURRENT_TX_HASH_GAS    uint64 = 100

	STORAGE_GET_GAS          uint64 = 200
	STORAGE_PUT_GAS          uint64 = 4000
	STORAGE_DELETE_GAS       uint64 = 100
	UINT_DEPLOY_CODE_LEN_GAS uint64 = 200000
	PER_UNIT_CODE_LEN        uint64 = 1024

	SHA256_GAS uint64 = 10
)
