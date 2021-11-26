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

package message

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/Ontology/common/log"
	. "github.com/Ontology/net/protocol"
	"net"
	"strconv"
)

type addrReq struct {
	Hdr msgHdr
	// No payload
}

type addr struct {
	hdr       msgHdr
	nodeCnt   uint64
	nodeAddrs []NodeAddr
}

const (
	NODEADDRSIZE = 30
)

func newGetAddr() ([]byte, error) {
	var msg addrReq
	// Fixme the check is the []byte{0} instead of 0
	var sum []byte
	sum = []byte{0x5d, 0xf6, 0xe0, 0xe2}
	msg.Hdr.init("getaddr", sum, 0)

	buf, err := msg.Serialization()
	if err != nil {
		return nil, err
	}

	str := hex.EncodeToString(buf)
	log.Debug("The message get addr length is: ", len(buf), " ", str)

	return buf, err
}

func NewAddrs(nodeaddrs []NodeAddr, count uint64) ([]byte, error) {
	var msg addr
	msg.nodeAddrs = nodeaddrs
	msg.nodeCnt = count
	msg.hdr.Magic = NETMAGIC
	cmd := "addr"
	copy(msg.hdr.CMD[0:7], cmd)
	p := new(bytes.Buffer)
	err := binary.Write(p, binary.LittleEndian, msg.nodeCnt)
	if err != nil {
		log.Error("Binary Write failed at new Msg: ", err.Error())
		return nil, err
	}

	err = binary.Write(p, binary.LittleEndian, msg.nodeAddrs)
	if err != nil {
		log.Error("Binary Write failed at new Msg: ", err.Error())
		return nil, err
	}
	s := sha256.Sum256(p.Bytes())
	s2 := s[:]
	s = sha256.Sum256(s2)
	buf := bytes.NewBuffer(s[:4])
	binary.Read(buf, binary.LittleEndian, &(msg.hdr.Checksum))
	msg.hdr.Length = uint32(len(p.Bytes()))
	log.Debug("The message payload length is ", msg.hdr.Length)

	m, err := msg.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}

	return m, nil
}

func (msg addrReq) Verify(buf []byte) error {
	// TODO Verify the message Ccntment
	err := msg.Hdr.Verify(buf)
	return err
}

func (msg addrReq) Handle(node Noder) error {
	log.Debug()
	// lock
	var addrstr []NodeAddr
	var count uint64
	addrstr, count = node.LocalNode().GetNeighborAddrs()
	buf, err := NewAddrs(addrstr, count)
	if err != nil {
		return err
	}
	go node.Tx(buf)
	return nil
}

func (msg addrReq) Serialization() ([]byte, error) {
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.LittleEndian, msg)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

func (msg *addrReq) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, msg)
	return err
}

func (msg addr) Serialization() ([]byte, error) {
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.LittleEndian, msg.hdr)

	if err != nil {
		return nil, err
	}
	err = binary.Write(&buf, binary.LittleEndian, msg.nodeCnt)
	if err != nil {
		return nil, err
	}
	for _, v := range msg.nodeAddrs {
		err = binary.Write(&buf, binary.LittleEndian, v)
		if err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), err
}

func (msg *addr) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, &(msg.hdr))
	err = binary.Read(buf, binary.LittleEndian, &(msg.nodeCnt))
	log.Debug("The address count is ", msg.nodeCnt)
	msg.nodeAddrs = make([]NodeAddr, msg.nodeCnt)
	for i := 0; i < int(msg.nodeCnt); i++ {
		err := binary.Read(buf, binary.LittleEndian, &(msg.nodeAddrs[i]))
		if err != nil {
			goto err
		}
	}
err:
	return err
}

func (msg addr) Verify(buf []byte) error {
	err := msg.hdr.Verify(buf)
	// TODO Verify the message Ccntment, check the ipaddr number
	return err
}

func (msg addr) Handle(node Noder) error {
	log.Debug()
	for _, v := range msg.nodeAddrs {
		var ip net.IP
		ip = v.IpAddr[:]
		//address := ip.To4().String() + ":" + strconv.Itoa(int(v.Port))
		address := ip.To16().String() + ":" + strconv.Itoa(int(v.Port))
		log.Info(fmt.Sprintf("The ip address is %s id is 0x%x", address, v.ID))

		if v.ID == node.LocalNode().GetID() {
			ccntminue
		}
		if node.LocalNode().NodeEstablished(v.ID) {
			ccntminue
		}

		if v.Port == 0 {
			ccntminue
		}

		go node.LocalNode().Connect(address, false)
	}
	return nil
}
