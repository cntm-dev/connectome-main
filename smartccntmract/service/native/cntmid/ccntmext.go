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
	"bytes"
	"errors"
	"fmt"

	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/core/states"
	"github.com/cntmio/cntmology/smartccntmract/service/native"
	"github.com/cntmio/cntmology/smartccntmract/service/native/utils"
)

//
var _DefaultCcntmexts = [][]byte{[]byte("https://www.w3.org/ns/did/v1"), []byte("https://cntmid.cntm.io/did/v1")}

func addCcntmext(srvc *native.NativeService) ([]byte, error) {
	params := new(Ccntmext)
	if err := params.Deserialization(common.NewZeroCopySource(srvc.Input)); err != nil {
		return utils.BYTE_FALSE, errors.New("addCcntmext error: deserialization params error, " + err.Error())
	}
	encId, err := encodeID(params.OntId)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("addCcntmext error: " + err.Error())
	}
	if !isValid(srvc, encId) {
		return utils.BYTE_FALSE, errors.New("addCcntmext error: have not registered")
	}

	if err := checkWitnessByIndex(srvc, encId, params.Index); err != nil {
		return utils.BYTE_FALSE, errors.New("verify signature failed: " + err.Error())
	}
	key := append(encId, FIELD_CcntmEXT)

	if err := putCcntmexts(srvc, key, params); err != nil {
		return utils.BYTE_FALSE, errors.New("addCcntmext error: putCcntmexts failed: " + err.Error())
	}
	updateTimeAndClearProof(srvc, encId)
	return utils.BYTE_TRUE, nil
}

func removeCcntmext(srvc *native.NativeService) ([]byte, error) {
	params := new(Ccntmext)
	if err := params.Deserialization(common.NewZeroCopySource(srvc.Input)); err != nil {
		return utils.BYTE_FALSE, errors.New("addCcntmext error: deserialization params error, " + err.Error())
	}
	encId, err := encodeID(params.OntId)
	if err != nil {
		return utils.BYTE_FALSE, errors.New("removeCcntmext error: " + err.Error())
	}
	if !isValid(srvc, encId) {
		return utils.BYTE_FALSE, errors.New("removeCcntmext error: have not registered")
	}

	if err := checkWitnessByIndex(srvc, encId, params.Index); err != nil {
		return utils.BYTE_FALSE, errors.New("verify signature failed: " + err.Error())
	}
	key := append(encId, FIELD_CcntmEXT)

	if err := deleteCcntmexts(srvc, key, params); err != nil {
		return utils.BYTE_FALSE, errors.New("removeCcntmext error: deleteCcntmexts failed: " + err.Error())
	}
	updateTimeAndClearProof(srvc, encId)
	return utils.BYTE_TRUE, nil
}

func deleteCcntmexts(srvc *native.NativeService, key []byte, params *Ccntmext) error {
	ccntmexts, err := getCcntmexts(srvc, key)
	if err != nil {
		return fmt.Errorf("deleteCcntmexts error: getCcntmexts error, %s", err)
	}
	coincidence := getCoincidence(ccntmexts, params)
	var remove [][]byte
	var remain [][]byte
	for i := 0; i < len(ccntmexts); i++ {
		if _, ok := coincidence[common.ToHexString(ccntmexts[i])]; !ok {
			remain = append(remain, ccntmexts[i])
		} else {
			remove = append(remove, ccntmexts[i])
		}
	}
	triggerCcntmextEvent(srvc, "remove", params.OntId, remove)
	err = storeCcntmexts(remain, srvc, key)
	if err != nil {
		return fmt.Errorf("deleteCcntmexts error: storeCcntmexts error, %s", err)
	}
	return nil
}

func putCcntmexts(srvc *native.NativeService, key []byte, params *Ccntmext) error {
	ccntmexts, err := getCcntmexts(srvc, key)
	if err != nil {
		return fmt.Errorf("putCcntmexts error: getCcntmexts failed, %s", err)
	}
	var add [][]byte
	removeDuplicate(params)
	coincidence := getCoincidence(ccntmexts, params)
	for i := 0; i < len(params.Ccntmexts); i++ {
		if (!bytes.Equal(params.Ccntmexts[i], _DefaultCcntmexts[0])) && (!bytes.Equal(params.Ccntmexts[i], _DefaultCcntmexts[1])) {
			if _, ok := coincidence[common.ToHexString(params.Ccntmexts[i])]; !ok {
				ccntmexts = append(ccntmexts, params.Ccntmexts[i])
				add = append(add, params.Ccntmexts[i])
			}
		}
	}
	triggerCcntmextEvent(srvc, "add", params.OntId, add)
	err = storeCcntmexts(ccntmexts, srvc, key)
	if err != nil {
		return fmt.Errorf("putCcntmexts error: storeCcntmexts failed, %s", err)
	}
	return nil
}

func getCoincidence(ccntmexts [][]byte, params *Ccntmext) map[string]bool {
	repeat := make(map[string]bool)
	for i := 0; i < len(ccntmexts); i++ {
		for j := 0; j < len(params.Ccntmexts); j++ {
			if bytes.Equal(ccntmexts[i], params.Ccntmexts[j]) {
				repeat[common.ToHexString(params.Ccntmexts[j])] = true
			}
		}
	}
	return repeat
}

func removeDuplicate(params *Ccntmext) {
	repeat := make(map[string]bool)
	var res [][]byte
	for i := 0; i < len(params.Ccntmexts); i++ {
		if _, ok := repeat[common.ToHexString(params.Ccntmexts[i])]; !ok {
			res = append(res, params.Ccntmexts[i])
			repeat[common.ToHexString(params.Ccntmexts[i])] = true
		}
	}
	params.Ccntmexts = res
}

func getCcntmexts(srvc *native.NativeService, key []byte) ([][]byte, error) {
	ccntmextsStore, err := utils.GetStorageItem(srvc.CacheDB, key)
	if err != nil {
		return nil, errors.New("getCcntmexts error:" + err.Error())
	}
	if ccntmextsStore == nil {
		return nil, nil
	}
	ccntmexts := new(Ccntmexts)
	if err := ccntmexts.Deserialization(common.NewZeroCopySource(ccntmextsStore.Value)); err != nil {
		return nil, err
	}
	return *ccntmexts, nil
}

func getCcntmextsWithDefault(srvc *native.NativeService, encId []byte) ([]string, error) {
	key := append(encId, FIELD_CcntmEXT)
	ccntmexts, err := getCcntmexts(srvc, key)
	if err != nil {
		return nil, fmt.Errorf("getCcntmextsWithDefault error, %s", err)
	}
	ccntmexts = append(_DefaultCcntmexts, ccntmexts...)
	var res []string
	for i := 0; i < len(ccntmexts); i++ {
		res = append(res, string(ccntmexts[i]))
	}
	return res, nil
}

func storeCcntmexts(ccntmexts Ccntmexts, srvc *native.NativeService, key []byte) error {
	sink := common.NewZeroCopySink(nil)
	ccntmexts.Serialization(sink)
	item := states.StorageItem{}
	item.Value = sink.Bytes()
	item.StateVersion = _VERSION_0
	srvc.CacheDB.Put(key, item.ToArray())
	return nil
}
