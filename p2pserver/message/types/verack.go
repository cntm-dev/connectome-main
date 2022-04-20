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

	"github.com/cntmio/cntmology/common/serialization"
)

type VerACK struct {
	MsgHdr
	IsConsensus bool
}

//Serialize message payload
func (msg VerACK) Serialization() ([]byte, error) {
	tmpBuffer := bytes.NewBuffer([]byte{})
	serialization.WriteBool(tmpBuffer, msg.IsConsensus)

	checkSumBuf := CheckSum(tmpBuffer.Bytes())
	msg.Init("verack", checkSumBuf, uint32(len(tmpBuffer.Bytes())))

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
func (msg *VerACK) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, &(msg.MsgHdr))
	if err != nil {
		return err
	}

	msg.IsConsensus, err = serialization.ReadBool(buf)
	return err
}