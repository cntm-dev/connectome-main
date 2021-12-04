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
	"errors"
	"github.com/Ontology/common"
	"github.com/Ontology/common/log"
	"github.com/Ontology/common/serialization"
	"github.com/Ontology/core/types"
	"github.com/Ontology/net/actor"
	. "github.com/Ontology/net/protocol"
)

type headersReq struct {
	hdr msgHdr
	p   struct {
		len       uint8
		hashStart [HASHLEN]byte
		hashEnd   [HASHLEN]byte
	}
}

type blkHeader struct {
	hdr    msgHdr
	cnt    uint32
	blkHdr []ledger.Header
}

func NewHeadersReq(n Noder) ([]byte, error) {
	var h headersReq

	// Fixme correct with the exactly request length
	h.p.len = 1
	buf := n.GetLedger().Blockchain.CurrentBlockHash()
	copy(h.p.hashStart[:], reverse(buf[:]))

	p := new(bytes.Buffer)
	err := binary.Write(p, binary.LittleEndian, &(h.p))
	if err != nil {
		log.Error("Binary Write failed at new headersReq")
		return nil, err
	}

	s := checkSum(p.Bytes())
	h.hdr.init("getheaders", s, uint32(len(p.Bytes())))

	m, err := h.Serialization()
	return m, err
}

func (msg headersReq) Verify(buf []byte) error {
	// TODO Verify the message Ccntment
	err := msg.hdr.Verify(buf)
	return err
}

func (msg blkHeader) Verify(buf []byte) error {
	// TODO Verify the message Ccntment
	err := msg.hdr.Verify(buf)
	return err
}

func (msg headersReq) Serialization() ([]byte, error) {
	hdrBuf, err := msg.hdr.Serialization()
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(hdrBuf)
	err = binary.Write(buf, binary.LittleEndian, msg.p.len)
	if err != nil {
		return nil, err
	}
	err = binary.Write(buf, binary.LittleEndian, msg.p.hashStart)
	if err != nil {
		return nil, err
	}

	err = binary.Write(buf, binary.LittleEndian, msg.p.hashEnd)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

func (msg *headersReq) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, &(msg.hdr))
	if err != nil {
		return err
	}

	err = binary.Read(buf, binary.LittleEndian, &(msg.p.len))
	if err != nil {
		return err
	}

	err = binary.Read(buf, binary.LittleEndian, &(msg.p.hashStart))
	if err != nil {
		return err
	}

	err = binary.Read(buf, binary.LittleEndian, &(msg.p.hashEnd))
	return err
}

func (msg blkHeader) Serialization() ([]byte, error) {
	hdrBuf, err := msg.hdr.Serialization()
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(hdrBuf)
	err = binary.Write(buf, binary.LittleEndian, msg.cnt)
	if err != nil {
		return nil, err
	}

	for _, header := range msg.blkHdr {
		header.Serialize(buf)
	}
	return buf.Bytes(), err
}

func (msg *blkHeader) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, &(msg.hdr))
	if err != nil {
		return err
	}

	err = binary.Read(buf, binary.LittleEndian, &(msg.cnt))
	if err != nil {
		return err
	}

	for i := 0; i < int(msg.cnt); i++ {
		var headers types.Header
		err := (&headers).Deserialize(buf)
		msg.blkHdr = append(msg.blkHdr, headers)
		if err != nil {
			log.Debug("blkHeader Deserialization failed")
			goto blkHdrErr
		}
	}

blkHdrErr:
	return err
}

func (msg headersReq) Handle(node Noder) error {
	common.Trace()
	// lock
	var starthash [HASHLEN]byte //[]common.Uint256
	var stophash [HASHLEN]byte  //common.Uint256
	starthash = msg.p.hashStart
	stophash = msg.p.hashEnd
	//FIXME if HeaderHashCount > 1
	headers, cnt := GetHeadersFromHash(starthash, stophash) //(starthash[0], stophash)
	buf, _ := NewHeaders(headers, cnt)
	go node.Tx(buf)
	return nil
}

func SendMsgSyncHeaders(node Noder) {
	buf, err := NewHeadersReq()
	if err != nil {
		log.Error("failed build a new headersReq")
	} else {
		go node.Tx(buf)
	}
}

func (msg blkHeader) Handle(node Noder) error {
	//log.Debug()
	//err := ledger.DefaultLedger.Store.AddHeaders(msg.blkHdr, ledger.DefaultLedger)
	//if err != nil {
	//	log.Warn("Add block Header error")
	//	return errors.New("Add block Header error, send new header request to another node\n")
	//}
	var blkHdr []*types.Header
	var i uint32
	for i = 0; i < msg.cnt; i++ {
		blkHdr = append(blkHdr, &msg.blkHdr[i])
	}
	actor.AddHeaders(blkHdr)
	return nil
}

func GetHeadersFromHash(startHash common.Uint256, stopHash common.Uint256) ([]types.Header, uint32, error) {
	var count uint32 = 0
	var empty [HASHLEN]byte
	headers := []types.Header{}
	var startHeight uint32
	var stopHeight uint32
	//curHeight := ledger.DefaultLedger.Store.GetHeaderHeight()
	curHeight, _ := actor.GetCurrentHeaderHeight()
	if startHash == empty {
		if stopHash == empty {
			if curHeight > MAXBLKHDRCNT {
				count = MAXBLKHDRCNT
			} else {
				count = curHeight
			}
		} else {
			//bkstop, err := ledger.DefaultLedger.Store.GetHeader(stopHash)
			bkstop, err := actor.GetHeaderByHash(stopHash)
			if err != nil || bkstop == nil {
				return nil, 0, err
			}
			stopHeight = bkstop.Height
			count = curHeight - stopHeight
			if count > MAXBLKHDRCNT {
				count = MAXBLKHDRCNT
			}
		}
	} else {
		bkstart, err := actor.GetHeaderByHash(startHash)
		if err != nil || bkstart == nil {
			return nil, 0, err
		}
		startHeight = bkstart.Height
		if stopHash != empty {
			bkstop, err := actor.GetHeaderByHash(stopHash)
			if err != nil || bkstop == nil {
				return nil, 0, err
			}
			stopHeight = bkstop.Height

			// avoid unsigned integer underflow
			if startHeight < stopHeight {
				return nil, 0, errors.New("do not have header to send")
			}
			count = startHeight - stopHeight

			if count >= MAXBLKHDRCNT {
				count = MAXBLKHDRCNT
				stopHeight = startHeight - MAXBLKHDRCNT
			}
		} else {

			if startHeight > MAXBLKHDRCNT {
				count = MAXBLKHDRCNT
			} else {
				count = startHeight
			}
		}
	}

	var i uint32
	for i = 1; i <= count; i++ {
		hash, err := actor.GetBlockHashByHeight(stopHeight + i)
		if err != nil {
			log.Errorf("GetBlockHashByHeight failed with err=%s, hash=%x,height=%d\n", err.Error(), hash, stopHeight+i)
			return nil, 0, err
		}
		hd, err := actor.GetHeaderByHash(hash)
		if err != nil || hd == nil {
			log.Errorf("GetHeaderByHash failed with err=%s, hash=%x,height=%d\n", err.Error(), hash, stopHeight+i)
			return nil, 0, err
		}
		headers = append(headers, *hd)
	}

	return headers, count, nil
}

func NewHeaders(headers []types.Header, count uint32) ([]byte, error) {
	var msg blkHeader
	msg.cnt = count
	msg.blkHdr = headers
	msg.hdr.Magic = NETMAGIC
	cmd := "headers"
	copy(msg.hdr.CMD[0:len(cmd)], cmd)

	tmpBuffer := bytes.NewBuffer([]byte{})
	serialization.WriteUint32(tmpBuffer, msg.cnt)
	for _, header := range headers {
		header.Serialize(tmpBuffer)
	}
	b := new(bytes.Buffer)
	err := binary.Write(b, binary.LittleEndian, tmpBuffer.Bytes())
	if err != nil {
		log.Error("Binary Write failed at new Msg")
		return nil, err
	}
	s := sha256.Sum256(b.Bytes())
	s2 := s[:]
	s = sha256.Sum256(s2)
	buf := bytes.NewBuffer(s[:4])
	binary.Read(buf, binary.LittleEndian, &(msg.hdr.Checksum))
	msg.hdr.Length = uint32(len(b.Bytes()))

	m, err := msg.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}
	return m, nil
}
