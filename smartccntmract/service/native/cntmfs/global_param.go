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
	"github.com/cntmio/cntmology/errors"
	"github.com/cntmio/cntmology/smartccntmract/service/native"
	"github.com/cntmio/cntmology/smartccntmract/service/native/utils"
)

type FsGlobalParam struct {
	MinTimeForFileStorage uint64
	CcntmractInvokeGasFee  uint64
	ChallengeReward       uint64
	FilePerServerPdpTimes uint64
	PassportExpire        uint64
	ChallengeInterval     uint64
	NodeMinVolume         uint64 //min total volume with fsNode
	NodePerKbPledge       uint64 //fsNode's pledge for participant
	FeePerBlockForRead    uint64 //cost for cntmfs-sdk read from fsNode
	FilePerBlockFeeRate   uint64 //cost for cntmfs-sdk save from fsNode
	SpacePerBlockFeeRate  uint64 //cost for cntmfs-sdk save from fsNode
}

func (this *FsGlobalParam) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarUint(sink, this.MinTimeForFileStorage)
	utils.EncodeVarUint(sink, this.CcntmractInvokeGasFee)
	utils.EncodeVarUint(sink, this.ChallengeReward)
	utils.EncodeVarUint(sink, this.FilePerServerPdpTimes)
	utils.EncodeVarUint(sink, this.PassportExpire)
	utils.EncodeVarUint(sink, this.ChallengeInterval)
	utils.EncodeVarUint(sink, this.NodeMinVolume)
	utils.EncodeVarUint(sink, this.NodePerKbPledge)
	utils.EncodeVarUint(sink, this.FeePerBlockForRead)
	utils.EncodeVarUint(sink, this.FilePerBlockFeeRate)
	utils.EncodeVarUint(sink, this.SpacePerBlockFeeRate)
}

func (this *FsGlobalParam) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.MinTimeForFileStorage, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.CcntmractInvokeGasFee, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.ChallengeReward, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.FilePerServerPdpTimes, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.PassportExpire, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.NodeMinVolume, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.NodePerKbPledge, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.FeePerBlockForRead, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.FilePerBlockFeeRate, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.SpacePerBlockFeeRate, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	return err
}

func setGlobalParam(native *native.NativeService, globalParam *FsGlobalParam) {
	globalParamKey := GenGlobalParamKey(native.CcntmextRef.CurrentCcntmext().CcntmractAddress)
	sink := common.NewZeroCopySink(nil)
	globalParam.Serialization(sink)
	utils.PutBytes(native, globalParamKey, sink.Bytes())
}

func getGlobalParam(native *native.NativeService) (*FsGlobalParam, error) {
	var globalParam FsGlobalParam

	globalParamKey := GenGlobalParamKey(native.CcntmextRef.CurrentCcntmext().CcntmractAddress)
	item, err := utils.GetStorageItem(native.CacheDB, globalParamKey)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "getGlobalParam GetStorageItem error!")
	}
	if item == nil {
		globalParam = FsGlobalParam{
			MinTimeForFileStorage: DefaultMinTimeForFileStorage,
			CcntmractInvokeGasFee:  DefaultCcntmractInvokeGasFee,
			ChallengeReward:       DefaultChallengeReward,
			FilePerServerPdpTimes: DefaultFilePerServerPdpTimes,
			PassportExpire:        DefaultPassportExpire,
			ChallengeInterval:     DefaultChallengeInterval,
			NodeMinVolume:         DefaultNodeMinVolume,
			NodePerKbPledge:       DefaultNodePerKbPledge,
			FeePerBlockForRead:    DefaultGasPerBlockForRead,
			FilePerBlockFeeRate:   DefaultFilePerBlockFeeRate,
			SpacePerBlockFeeRate:  DefaultSpacePerBlockFeeRate,
		}
		return &globalParam, nil
	}

	source := common.NewZeroCopySource(item.Value)
	if err := globalParam.Deserialization(source); err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "getGlobalParam Deserialization error!")
	}
	return &globalParam, nil
}
