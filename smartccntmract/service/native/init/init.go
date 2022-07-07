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

	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/smartccntmract/service/native/auth"
	params "github.com/cntmio/cntmology/smartccntmract/service/native/global_params"
	"github.com/cntmio/cntmology/smartccntmract/service/native/governance"
	"github.com/cntmio/cntmology/smartccntmract/service/native/cntm"
	"github.com/cntmio/cntmology/smartccntmract/service/native/cntm"
	"github.com/cntmio/cntmology/smartccntmract/service/native/cntmid"
	"github.com/cntmio/cntmology/smartccntmract/service/native/utils"
	"github.com/cntmio/cntmology/smartccntmract/states"
)

var (
	cntm_INIT_BYTES    = InitBytes(utils.OntCcntmractAddress, cntm.INIT_NAME)
	cntm_INIT_BYTES    = InitBytes(utils.OngCcntmractAddress, cntm.INIT_NAME)
	PARAM_INIT_BYTES  = InitBytes(utils.ParamCcntmractAddress, params.INIT_NAME)
	COMMIT_DPOS_BYTES = InitBytes(utils.GovernanceCcntmractAddress, governance.COMMIT_DPOS)
	INIT_CONFIG_BYTES = InitBytes(utils.GovernanceCcntmractAddress, governance.INIT_CONFIG)
)

func init() {
	cntm.InitOng()
	cntm.InitOnt()
	params.InitGlobalParams()
	cntmid.Init()
	auth.Init()
	governance.InitGovernance()
}

func InitBytes(addr common.Address, method string) []byte {
	init := states.Ccntmract{Address: addr, Method: method}
	bf := new(bytes.Buffer)
	init.Serialize(bf)
	return bf.Bytes()
}
