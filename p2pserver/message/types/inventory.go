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

package types

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/serialization"
	"github.com/cntmio/cntmology/errors"
	p2pCommon "github.com/cntmio/cntmology/p2pserver/common"
)

var LastInvHash common.Uint256

type InvPayload struct {
	InvType common.InventoryType
	Blk     []common.Uint256
}

type Inv struct {
	P InvPayload
}

func (this Inv) invType() common.InventoryType {
	return this.P.InvType
}

func (this *Inv) CmdType() string {
	return p2pCommon.INV_TYPE
}

//Serialize message payload
func (this Inv) Serialization() ([]byte, error) {
	p := bytes.NewBuffer([]byte{})
	err := serialization.WriteUint8(p, uint8(this.P.InvType))
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNetPackFail, fmt.Sprintf("write error. InvType:%v", this.P.InvType))
	}
	blkCnt := uint32(len(this.P.Blk))
	err = serialization.WriteUint32(p, blkCnt)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNetPackFail, fmt.Sprintf("write error. Cnt:%v", blkCnt))
	}
	err = binary.Write(p, binary.LittleEndian, this.P.Blk)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNetPackFail, fmt.Sprintf("write error. Blk:%v", this.P.Blk))
	}

	return p.Bytes(), nil
}

//Deserialize message payload
func (this *Inv) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	invType, err := serialization.ReadUint8(buf)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNetUnPackFail, fmt.Sprintf("read invType error. buf:%v", buf))
	}
	this.P.InvType = common.InventoryType(invType)
	blkCnt, err := serialization.ReadUint32(buf)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNetUnPackFail, fmt.Sprintf("read Cnt error. buf:%v", buf))
	}
	if blkCnt > p2pCommon.MAX_INV_BLK_CNT {
		blkCnt = p2pCommon.MAX_INV_BLK_CNT
	}
	for i := 0; i < int(blkCnt); i++ {
		var blk common.Uint256
		err := binary.Read(buf, binary.LittleEndian, &blk)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNetUnPackFail, fmt.Sprintf("read inv blk error. buf:%v", buf))
		}
		this.P.Blk = append(this.P.Blk, blk)
	}
	return nil
}
