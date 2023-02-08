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
package cntmid

import (
	"errors"

	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/config"
	"github.com/cntmio/cntmology/core/states"
	"github.com/cntmio/cntmology/smartccntmract/service/native"
	"github.com/cntmio/cntmology/smartccntmract/service/native/utils"
)

func updateTime(srvc *native.NativeService, key []byte) {
	if srvc.Height < config.GetNewOntIdHeight() {
		return
	}
	item := states.StorageItem{}
	sink := common.NewZeroCopySink(nil)
	sink.WriteUint32(srvc.Time)
	item.Value = sink.Bytes()
	item.StateVersion = _VERSION_0
	srvc.CacheDB.Put(key, item.ToArray())
}

func getUpdateTime(srvc *native.NativeService, encId []byte) (uint32, error) {
	key := append(encId, FIELD_UPDATED)
	return getTime(srvc, key)
}

func getCreateTime(srvc *native.NativeService, encId []byte) (uint32, error) {
	key := append(encId, FIELD_CREATED)
	return getTime(srvc, key)
}

func getTime(srvc *native.NativeService, key []byte) (uint32, error) {
	timeStore, err := utils.GetStorageItem(srvc.CacheDB, key)
	if err != nil {
		return 0, errors.New("getTime error:" + err.Error())
	}
	var createTime uint32 = 0
	if timeStore != nil {
		source := common.NewZeroCopySource(timeStore.Value)
		createTime, err = utils.DecodeUint32(source)
		if err != nil {
			return 0, errors.New("DecodeUint32 error:" + err.Error())
		}
	}
	return createTime, nil
}
