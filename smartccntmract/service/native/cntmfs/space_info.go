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

package cntmfs

import (
	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/smartccntmract/service/native"
	"github.com/cntmio/cntmology/smartccntmract/service/native/utils"
)

type SpaceInfo struct {
	SpaceOwner  common.Address
	Volume      uint64
	RestVol     uint64
	CopyNumber  uint64
	PayAmount   uint64
	RestAmount  uint64
	TimeStart   uint64
	TimeExpired uint64
	CurrFeeRate uint64
	ValidFlag   bool
}

func (this *SpaceInfo) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeAddress(sink, this.SpaceOwner)
	utils.EncodeVarUint(sink, this.Volume)
	utils.EncodeVarUint(sink, this.RestVol)
	utils.EncodeVarUint(sink, this.CopyNumber)
	utils.EncodeVarUint(sink, this.PayAmount)
	utils.EncodeVarUint(sink, this.RestAmount)
	utils.EncodeVarUint(sink, this.TimeStart)
	utils.EncodeVarUint(sink, this.TimeExpired)
	utils.EncodeVarUint(sink, this.CurrFeeRate)
	sink.WriteBool(this.ValidFlag)
}

func (this *SpaceInfo) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.SpaceOwner, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.Volume, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.RestVol, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.CopyNumber, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.PayAmount, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.RestAmount, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.TimeStart, err = utils.DecodeVarUint(source)
	if err != nil {
		return nil
	}
	this.TimeExpired, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.CurrFeeRate, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.ValidFlag, err = DecodeBool(source)
	if err != nil {
		return err
	}
	return nil
}

func addSpaceInfo(native *native.NativeService, spaceInfo *SpaceInfo) {
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	spaceInfoKey := GenFsSpaceKey(ccntmract, spaceInfo.SpaceOwner)

	sink := common.NewZeroCopySink(nil)
	spaceInfo.Serialization(sink)

	utils.PutBytes(native, spaceInfoKey, sink.Bytes())
}

func delSpaceInfo(native *native.NativeService, spaceOwner common.Address) {
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	spaceInfoKey := GenFsSpaceKey(ccntmract, spaceOwner)
	native.CacheDB.Delete(spaceInfoKey)
}

func spaceInfoExist(native *native.NativeService, spaceOwner common.Address) bool {
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	spaceInfoKey := GenFsSpaceKey(ccntmract, spaceOwner)

	item, err := utils.GetStorageItem(native.CacheDB, spaceInfoKey)
	if err != nil || item == nil || item.Value == nil {
		return false
	}
	return true
}

func getSpaceInfoFromDb(native *native.NativeService, fileOwner common.Address) *SpaceInfo {
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	spaceInfoKey := GenFsSpaceKey(ccntmract, fileOwner)

	item, err := utils.GetStorageItem(native.CacheDB, spaceInfoKey)
	if err != nil || item == nil || item.Value == nil {
		return nil
	}

	var spaceInfo SpaceInfo
	source := common.NewZeroCopySource(item.Value)
	if err := spaceInfo.Deserialization(source); err != nil {
		return nil
	}
	if uint64(native.Time) > spaceInfo.TimeExpired {
		spaceInfo.ValidFlag = false
	}
	return &spaceInfo
}

func getSpaceRawRealInfo(native *native.NativeService, fileOwner common.Address) []byte {
	spaceInfo := getSpaceInfoFromDb(native, fileOwner)
	if spaceInfo == nil {
		return nil
	}

	sink := common.NewZeroCopySink(nil)
	spaceInfo.Serialization(sink)
	return sink.Bytes()
}

func getAndUpdateSpaceInfo(native *native.NativeService, fileOwner common.Address) *SpaceInfo {
	spaceInfo := getSpaceInfoFromDb(native, fileOwner)
	if spaceInfo == nil {
		return nil
	}

	if !spaceInfo.ValidFlag {
		addSpaceInfo(native, spaceInfo)
	}

	return spaceInfo
}
