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

package common

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"github.com/Ontology/common/log"
	. "github.com/Ontology/errors"
	"golang.org/x/crypto/ripemd160"
	"io"
	"math/rand"
	"os"
)

func ToCodeHash(code []byte) Address {
	temp := sha256.Sum256(code)
	md := ripemd160.New()
	io.WriteString(md, string(temp[:]))
	f := md.Sum(nil)

	hash, _ := AddressParseFromBytes(f) // err will never occur
	return hash
}

func GetNonce() uint64 {
	log.Debug()
	// Fixme replace with the real random number generator
	nonce := uint64(rand.Uint32())<<32 + uint64(rand.Uint32())
	return nonce
}

func IntToBytes(n int) []byte {
	tmp := int32(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.LittleEndian, tmp)
	return bytesBuffer.Bytes()
}

func BytesToInt16(b []byte) int16 {
	bytesBuffer := bytes.NewBuffer(b)
	var tmp int16
	binary.Read(bytesBuffer, binary.BigEndian, &tmp)
	return int16(tmp)
}

func IsEqualBytes(b1 []byte, b2 []byte) bool {
	len1 := len(b1)
	len2 := len(b2)
	if len1 != len2 {
		return false
	}

	for i := 0; i < len1; i++ {
		if b1[i] != b2[i] {
			return false
		}
	}

	return true
}

func ToHexString(data []byte) string {
	return hex.EncodeToString(data)
}

func HexToBytes(value string) ([]byte, error) {
	return hex.DecodeString(value)
}

func BytesReverse(u []byte) []byte {
	for i, j := 0, len(u)-1; i < j; i, j = i+1, j-1 {
		u[i], u[j] = u[j], u[i]
	}
	return u
}

func HexToBytesReverse(value string) ([]byte, error) {
	u, err := hex.DecodeString(value)
	if err != nil {
		return u, err
	}
	return BytesReverse(u), err
}

func ClearBytes(arr []byte, len int) {
	for i := 0; i < len; i++ {
		arr[i] = 0
	}
}

func CompareHeight(blockHeight uint64, heights []uint64) bool {
	for _, height := range heights {
		if blockHeight < height {
			return false
		}
	}
	return true
}

func GetUint16Array(source []byte) ([]uint16, error) {
	if source == nil {
		return nil, NewDetailErr(errors.New("[Common] , GetUint16Array err, source = nil"), ErrNoCode, "")
	}

	if len(source)%2 != 0 {
		return nil, NewDetailErr(errors.New("[Common] , GetUint16Array err, length of source is odd."), ErrNoCode, "")
	}

	dst := make([]uint16, len(source)/2)
	for i := 0; i < len(source)/2; i++ {
		dst[i] = uint16(source[i*2]) + uint16(source[i*2+1])*256
	}

	return dst, nil
}

func ToByteArray(source []uint16) []byte {
	dst := make([]byte, len(source)*2)
	for i := 0; i < len(source); i++ {
		dst[i*2] = byte(source[i] % 256)
		dst[i*2+1] = byte(source[i] / 256)
	}

	return dst
}

func SliceRemove(slice []uint32, h uint32) []uint32 {
	for i, v := range slice {
		if v == h {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

func FileExisted(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}

func IsArrayEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
