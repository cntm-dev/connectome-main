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

	"github.com/cntmio/cntmology/common/log"
	ct "github.com/cntmio/cntmology/core/types"
)

type Block struct {
	MsgHdr
	Blk ct.Block
}

//Check whether header is correct
func (msg Block) Verify(buf []byte) error {
	err := msg.MsgHdr.Verify(buf)
	return err
}

//Serialize message payload
func (msg Block) Serialization() ([]byte, error) {

	tmpBuffer := bytes.NewBuffer([]byte{})
	msg.Blk.Serialize(tmpBuffer)

	checkSumBuf := CheckSum(tmpBuffer.Bytes())
	msg.Init("block", checkSumBuf, uint32(len(tmpBuffer.Bytes())))
	log.Debug("The message payload length is ", msg.Length)

	hdrBuf, err := msg.MsgHdr.Serialization()
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(hdrBuf)
	err = binary.Write(buf, binary.LittleEndian, tmpBuffer.Bytes())
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), err
}

//Deserialize message payload
func (msg *Block) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, &(msg.MsgHdr))
	if err != nil {
		log.Warn("Parse block message hdr error")
		return errors.New("Parse block message hdr error ")
	}

	err = msg.Blk.Deserialize(buf)
	if err != nil {
		log.Warn("Parse block message error")
		return errors.New("Parse block message error ")
	}

	return err
}