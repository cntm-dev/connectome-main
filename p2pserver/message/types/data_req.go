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
	"errors"

	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/log"
)

type DataReq struct {
	MsgHdr
	DataType common.InventoryType
	Hash     common.Uint256
}

//Serialize message payload
func (this DataReq) Serialization() ([]byte, error) {
	p := bytes.NewBuffer([]byte{})
	err := binary.Write(p, binary.LittleEndian, &(this.DataType))
	this.Hash.Serialize(p)
	if err != nil {
		return nil, err
	}

	checkSumBuf := CheckSum(p.Bytes())
	this.Init("getdata", checkSumBuf, uint32(len(p.Bytes())))

	hdrBuf, err := this.MsgHdr.Serialization()
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(hdrBuf)
	data := append(buf.Bytes(), p.Bytes()...)
	return data, nil
}

//Deserialize message payload
func (this *DataReq) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, &(this.MsgHdr))
	if err != nil {
		log.Warn("Parse dataReq message hdr error")
		return errors.New("Parse dataReq message hdr error ")
	}

	err = binary.Read(buf, binary.LittleEndian, &(this.DataType))
	if err != nil {
		log.Warn("Parse dataReq message dataType error")
		return errors.New("Parse dataReq message dataType error ")
	}

	err = this.Hash.Deserialize(buf)
	if err != nil {
		log.Warn("Parse dataReq message hash error")
		return errors.New("Parse dataReq message hash error ")
	}
	return nil
}
