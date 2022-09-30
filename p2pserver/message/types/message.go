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
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/cntmio/cntmology/common/config"
	"github.com/cntmio/cntmology/p2pserver/common"
)

type Message interface {
	Serialization() ([]byte, error)
	Deserialization([]byte) error
	CmdType() string
}

//MsgPayload in link channel
type MsgPayload struct {
	Id          uint64  //peer ID
	Addr        string  //link address
	PayloadSize uint32  //payload size
	Payload     Message //msg payload
}

type messageHeader struct {
	Magic    uint32
	CMD      [common.MSG_CMD_LEN]byte // The message type
	Length   uint32
	Checksum [common.CHECKSUM_LEN]byte
}

func readMessageHeader(reader io.Reader) (messageHeader, error) {
	msgh := messageHeader{}
	err := binary.Read(reader, binary.LittleEndian, &msgh)
	return msgh, err
}

func writeMessageHeader(writer io.Writer, msgh messageHeader) error {
	return binary.Write(writer, binary.LittleEndian, msgh)
}

func newMessageHeader(cmd string, length uint32, checksum [common.CHECKSUM_LEN]byte) messageHeader {
	msgh := messageHeader{}
	msgh.Magic = config.DefConfig.P2PNode.NetworkMagic
	copy(msgh.CMD[:], cmd)
	msgh.Checksum = checksum
	msgh.Length = length
	return msgh
}

func WriteMessage(writer io.Writer, msg Message) error {
	buf, err := msg.Serialization()
	if err != nil {
		return err
	}
	checksum := CheckSum(buf)

	hdr := newMessageHeader(msg.CmdType(), uint32(len(buf)), checksum)

	err = writeMessageHeader(writer, hdr)
	if err != nil {
		return err
	}

	_, err = writer.Write(buf)
	return err
}

func ReadMessage(reader io.Reader) (Message, uint32, error) {
	hdr, err := readMessageHeader(reader)
	if err != nil {
		return nil, 0, err
	}

	magic := config.DefConfig.P2PNode.NetworkMagic
	if hdr.Magic != magic {
		return nil, 0, fmt.Errorf("unmatched magic number %d, expected %d", hdr.Magic, magic)
	}

	if int(hdr.Length) > common.MAX_PAYLOAD_LEN {
		return nil, 0, fmt.Errorf("msg payload length:%d exceed max payload size: %d",
			hdr.Length, common.MAX_PAYLOAD_LEN)
	}

	buf := make([]byte, hdr.Length)
	_, err = io.ReadFull(reader, buf)
	if err != nil {
		return nil, 0, err
	}

	checksum := CheckSum(buf)
	if checksum != hdr.Checksum {
		return nil, 0, fmt.Errorf("message checksum mismatch: %x != %x ", hdr.Checksum, checksum)
	}

	cmdType := string(bytes.TrimRight(hdr.CMD[:], string(0)))
	msg, err := MakeEmptyMessage(cmdType)
	if err != nil {
		return nil, 0, err
	}

	err = msg.Deserialization(buf)
	if err != nil {
		return nil, 0, err
	}

	return msg, hdr.Length, nil
}

func MakeEmptyMessage(cmdType string) (Message, error) {
	switch cmdType {
	case common.PING_TYPE:
		return &Ping{}, nil
	case common.VERSION_TYPE:
		return &Version{}, nil
	case common.VERACK_TYPE:
		return &VerACK{}, nil
	case common.ADDR_TYPE:
		return &Addr{}, nil
	case common.GetADDR_TYPE:
		return &AddrReq{}, nil
	case common.Pcntm_TYPE:
		return &Pcntm{}, nil
	case common.GET_HEADERS_TYPE:
		return &HeadersReq{}, nil
	case common.HEADERS_TYPE:
		return &BlkHeader{}, nil
	case common.INV_TYPE:
		return &Inv{}, nil
	case common.GET_DATA_TYPE:
		return &DataReq{}, nil
	case common.BLOCK_TYPE:
		return &Block{}, nil
	case common.TX_TYPE:
		return &Trn{}, nil
	case common.CONSENSUS_TYPE:
		return &Consensus{}, nil
	case common.NOT_FOUND_TYPE:
		return &NotFound{}, nil
	case common.DISCONNECT_TYPE:
		return &Disconnected{}, nil
	case common.GET_BLOCKS_TYPE:
		return &BlocksReq{}, nil
	default:
		return nil, errors.New("unsupported cmd type:" + cmdType)
	}

}

//caculate checksum value
func CheckSum(p []byte) [common.CHECKSUM_LEN]byte {
	var checksum [common.CHECKSUM_LEN]byte
	t := sha256.Sum256(p)
	s := sha256.Sum256(t[:])

	copy(checksum[:], s[:])

	return checksum
}
