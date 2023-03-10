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

package states

import "math/big"

type OracleNodeStatus int

type RegisterOracleNodeParam struct {
	Address  string `json:"address"`
	Guaranty uint64 `json:"guaranty"`
}

type ApproveOracleNodeParam struct {
	Address string `json:"address"`
}

type OracleNode struct {
	Address  string           `json:"address"`
	Guaranty uint64           `json:"guaranty"`
	Status   OracleNodeStatus `json:"status"`
}

type QuitOracleNodeParam struct {
	Address string `json:"address"`
}

type CreateOracleRequestParam struct {
	Request   string   `json:"request"`
	OracleNum *big.Int `json:"oracleNum"`
	Address   string   `json:"address"`
}

type UndoRequests struct {
	Requests map[string]struct{} `json:"requests"`
}

type SetOracleOutcomeParam struct {
	TxHash  string      `json:"txHash"`
	Address string      `json:"owner"`
	Outcome interface{} `json:"outcome"`
}

type OutcomeRecord struct {
	OutcomeRecord map[string]interface{} `json:"outcomeRecord"`
}

type SetOracleCronOutcomeParam struct {
	TxHash  string      `json:"txHash"`
	Address string      `json:"owner"`
	Outcome interface{} `json:"outcome"`
}

type CronOutcomeRecord struct {
	CronOutcomeRecord map[string]interface{} `json:"cronOutcomeRecord"`
}

type ChangeCronViewParam struct {
	TxHash  string `json:"txHash"`
	Address string `json:"owner"`
}
