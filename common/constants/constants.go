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

package constants

import (
	"time"
)

// genesis constants
var (
	//TODO: modify this when on mainnet
	GENESIS_BLOCK_TIMESTAMP = uint32(time.Date(2018, time.June, 12, 0, 0, 0, 0, time.UTC).Unix())
)

// cntm constants
const (
	cntm_NAME         = "cntm Token"
	cntm_SYMBOL       = "cntm"
	cntm_DECIMALS     = 1
	cntm_TOTAL_SUPPLY = uint64(1000000000)
)

// cntm constants
const (
	cntm_NAME         = "cntm Token"
	cntm_SYMBOL       = "cntm"
	cntm_DECIMALS     = 9
	cntm_TOTAL_SUPPLY = uint64(1000000000000000000)
)

// cntm/cntm unbound model constants
const UNBOUND_TIME_INTERVAL = uint32(31536000)

var UNBOUND_GENERATION_AMOUNT = [18]uint64{5, 4, 3, 3, 2, 2, 2, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}

// the end of unbound timestamp offset from genesis block's timestamp
var UNBOUND_DEADLINE = (func() uint32 {
	count := uint64(0)
	for _, m := range UNBOUND_GENERATION_AMOUNT {
		count += m
	}
	count *= uint64(UNBOUND_TIME_INTERVAL)

	numInterval := len(UNBOUND_GENERATION_AMOUNT)

	if UNBOUND_GENERATION_AMOUNT[numInterval-1] != 1 ||
		!(count-uint64(UNBOUND_TIME_INTERVAL) < cntm_TOTAL_SUPPLY && cntm_TOTAL_SUPPLY <= count) {
		panic("incompatible constants setting")
	}

	return UNBOUND_TIME_INTERVAL*uint32(numInterval) - uint32(count-uint64(cntm_TOTAL_SUPPLY))
})()

// multi-sig constants
const MULTI_SIG_MAX_PUBKEY_SIZE = 1024

const (
	NETWORK_MAIGIC_MAINNET = 0x74746e41
	NETWORK_MAIGIC_POLARIS = 0x74746e41
)
