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

package common

// DataEntryPrefix
type DataEntryPrefix byte

const (
	// DATA
	DATA_BLOCK_HASH        DataEntryPrefix = 0x00 //Block height => block hash key prefix
	DATA_HEADER                            = 0x01 //Block hash => block header+txhashes key prefix
	DATA_TRANSACTION                       = 0x02 //Transction hash => transaction key prefix
	DATA_STATE_MERKLE_ROOT                 = 0x21 // block height => write set hash + state merkle root

	// Transaction
	ST_BOOKKEEPER DataEntryPrefix = 0x03 //BookKeeper state key prefix
	ST_CcntmRACT   DataEntryPrefix = 0x04 //Smart ccntmract deploy code key prefix
	ST_STORAGE    DataEntryPrefix = 0x05 //Smart ccntmract storage key prefix
	ST_DESTROYED  DataEntryPrefix = 0x06 // record destroyed smart ccntmract: prefix+address -> height

	// eth state
	ST_ETH_CODE    DataEntryPrefix = 0x30 // eth ccntmract code:hash -> bytes
	ST_ETH_ACCOUNT DataEntryPrefix = 0x31 // eth account: address -> [nonce, codeHash]

	IX_HEADER_HASH_LIST DataEntryPrefix = 0x09 //Block height => block hash key prefix

	//SYSTEM
	SYS_CURRENT_BLOCK        DataEntryPrefix = 0x10 //Current block key prefix
	SYS_VERSION              DataEntryPrefix = 0x11 //Store version key prefix
	SYS_CURRENT_CROSS_STATES DataEntryPrefix = 0x12 //Block cross states
	SYS_BLOCK_MERKLE_TREE    DataEntryPrefix = 0x13 // Block merkle tree root key prefix
	SYS_STATE_MERKLE_TREE    DataEntryPrefix = 0x20 // state merkle tree root key prefix
	SYS_CROSS_CHAIN_MSG      DataEntryPrefix = 0x22 // state merkle tree root key prefix

	EVENT_NOTIFY DataEntryPrefix = 0x14 //Event notify key prefix

	DATA_BLOCK_PRUNE_HEIGHT DataEntryPrefix = 0x80 //  last pruned block height, genesis block can not be pruned
)
