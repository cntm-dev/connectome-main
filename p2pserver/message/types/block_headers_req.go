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
	"io"

	"github.com/cntmio/cntmology/common"
	comm "github.com/cntmio/cntmology/p2pserver/common"
)

type HeadersReq struct {
	Len       uint8
	HashStart common.Uint256
	HashEnd   common.Uint256
}

//Serialize message payload
func (this *HeadersReq) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint8(this.Len)
	sink.WriteHash(this.HashStart)
	sink.WriteHash(this.HashEnd)
}

func (this *HeadersReq) CmdType() string {
	return comm.GET_HEADERS_TYPE
}

//Deserialize message payload
func (this *HeadersReq) Deserialization(source *common.ZeroCopySource) error {
	var eof bool
	this.Len, eof = source.NextUint8()
	if eof {
		return io.ErrUnexpectedEOF
	}
	this.HashStart, eof = source.NextHash()
	if eof {
		return io.ErrUnexpectedEOF
	}
	this.HashEnd, eof = source.NextHash()
	if eof {
		return io.ErrUnexpectedEOF
	}

	return nil
}
