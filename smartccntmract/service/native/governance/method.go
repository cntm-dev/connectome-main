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

package governance

import (
	"bytes"
	"encoding/hex"
	"sort"

	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/constants"
	cstates "github.com/cntmio/cntmology/core/states"
	scommon "github.com/cntmio/cntmology/core/store/common"
	"github.com/cntmio/cntmology/errors"
	"github.com/cntmio/cntmology/smartccntmract/service/native"
	"github.com/cntmio/cntmology/smartccntmract/service/native/utils"
)

func registerCandidate(native *native.NativeService, flag string) error {
	params := new(RegisterCandidateParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, ccntmract params deserialize error!")
	}
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress

	//check auth of OntID
	err := appCallVerifyToken(native, ccntmract, params.Caller, REGISTER_CANDIDATE, params.KeyNo)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "appCallVerifyToken, verifyToken failed!")
	}

	//check witness
	err = utils.ValidateOwner(native, params.Address)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "validateOwner, checkWitness error!")
	}

	//get current view
	view, err := GetView(native, ccntmract)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "getView, get view error!")
	}

	//check peerPubkey
	if err := validatePeerPubKeyFormat(params.PeerPubkey); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "invalid peer pubkey")
	}

	peerPubkeyPrefix, err := hex.DecodeString(params.PeerPubkey)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
	}
	//get black list
	blackList, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(ccntmract, []byte(BLACK_LIST), peerPubkeyPrefix))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Get, get BlackList error!")
	}
	if blackList != nil {
		return errors.NewErr("registerCandidate, this Peer is in BlackList!")
	}

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, ccntmract, view)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "getPeerPoolMap, get peerPoolMap error!")
	}

	//check if exist in PeerPool
	_, ok := peerPoolMap.PeerPoolMap[params.PeerPubkey]
	if ok {
		return errors.NewErr("registerCandidate, peerPubkey is already in peerPoolMap!")
	}

	peerPoolItem := &PeerPoolItem{
		PeerPubkey: params.PeerPubkey,
		Address:    params.Address,
		InitPos:    params.InitPos,
		Status:     RegisterCandidateStatus,
	}
	peerPoolMap.PeerPoolMap[params.PeerPubkey] = peerPoolItem
	err = putPeerPoolMap(native, ccntmract, view, peerPoolMap)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "putPeerPoolMap, put peerPoolMap error!")
	}

	//get globalParam
	globalParam, err := getGlobalParam(native, ccntmract)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "getGlobalParam, getGlobalParam error!")
	}

	switch flag {
	case "transfer":
		//cntm transfer
		err = appCallTransferOnt(native, params.Address, utils.GovernanceCcntmractAddress, params.InitPos)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferOnt, cntm transfer error!")
		}

		//cntm transfer
		err = appCallTransferOng(native, params.Address, utils.GovernanceCcntmractAddress, globalParam.CandidateFee)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferOng, cntm transfer error!")
		}
	case "transferFrom":
		//cntm transfer from
		err = appCallTransferFromOnt(native, utils.GovernanceCcntmractAddress, params.Address, utils.GovernanceCcntmractAddress, params.InitPos)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferFromOnt, cntm transfer error!")
		}

		//cntm transfer from
		err = appCallTransferFromOng(native, utils.GovernanceCcntmractAddress, params.Address, utils.GovernanceCcntmractAddress, globalParam.CandidateFee)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferFromOng, cntm transfer error!")
		}
	}

	//update total stake
	err = depositTotalStake(native, ccntmract, params.Address, params.InitPos)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "depositTotalStake, depositTotalStake error!")
	}
	return nil
}

func voteForPeer(native *native.NativeService, flag string) error {
	params := &VoteForPeerParam{
		PeerPubkeyList: make([]string, 0),
		PosList:        make([]uint64, 0),
	}
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, ccntmract params deserialize error!")
	}
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress

	//check witness
	err := utils.ValidateOwner(native, params.Address)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "validateOwner, checkWitness error!")
	}

	//get current view
	view, err := GetView(native, ccntmract)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "getView, get view error!")
	}

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, ccntmract, view)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "getPeerPoolMap, get peerPoolMap error!")
	}

	//get globalParam
	globalParam, err := getGlobalParam(native, ccntmract)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "getGlobalParam, getGlobalParam error!")
	}

	var total uint64
	for i := 0; i < len(params.PeerPubkeyList); i++ {
		peerPubkey := params.PeerPubkeyList[i]
		pos := params.PosList[i]

		peerPoolItem, ok := peerPoolMap.PeerPoolMap[peerPubkey]
		if !ok {
			return errors.NewErr("voteForPeer, peerPubkey is not in peerPoolMap!")
		}

		if peerPoolItem.Status != CandidateStatus && peerPoolItem.Status != ConsensusStatus {
			return errors.NewErr("voteForPeer, peerPubkey is not candidate and can not be voted!")
		}

		voteInfo, err := getVoteInfo(native, ccntmract, peerPubkey, params.Address)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "getVoteInfo, get voteInfo error!")
		}
		voteInfo.NewPos = voteInfo.NewPos + pos
		total = total + pos
		peerPoolItem.TotalPos = peerPoolItem.TotalPos + pos
		if peerPoolItem.TotalPos > globalParam.PosLimit*peerPoolItem.InitPos {
			return errors.NewErr("voteForPeer, pos of this peer is full!")
		}

		peerPoolMap.PeerPoolMap[peerPubkey] = peerPoolItem
		err = putVoteInfo(native, ccntmract, voteInfo)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "putVoteInfo, put voteInfo error!")
		}
	}
	err = putPeerPoolMap(native, ccntmract, view, peerPoolMap)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "putPeerPoolMap, put peerPoolMap error!")
	}

	switch flag {
	case "transfer":
		//cntm transfer
		err = appCallTransferOnt(native, params.Address, utils.GovernanceCcntmractAddress, total)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferOnt, cntm transfer error!")
		}
	case "transferFrom":
		//cntm transfer from
		err = appCallTransferFromOnt(native, utils.GovernanceCcntmractAddress, params.Address, utils.GovernanceCcntmractAddress, total)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferFromOnt, cntm transfer error!")
		}
	}

	//update total stake
	err = depositTotalStake(native, ccntmract, params.Address, total)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "depositTotalStake, depositTotalStake error!")
	}

	return nil
}

func normalQuit(native *native.NativeService, ccntmract common.Address, peerPoolItem *PeerPoolItem) error {
	peerPubkeyPrefix, err := hex.DecodeString(peerPoolItem.PeerPubkey)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
	}
	flag := false
	//draw back vote pos
	stateValues, err := native.CloneCache.Store.Find(scommon.ST_STORAGE, utils.ConcatKey(ccntmract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Store.Find, get all peerPool error!")
	}
	voteInfo := new(VoteInfo)
	for _, v := range stateValues {
		voteInfoStore, ok := v.Value.(*cstates.StorageItem)
		if !ok {
			return errors.NewErr("voteInfoStore is not available!")
		}
		if err := voteInfo.Deserialize(bytes.NewBuffer(voteInfoStore.Value)); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize voteInfo error!")
		}
		voteInfo.WithdrawUnfreezePos = voteInfo.ConsensusPos + voteInfo.FreezePos + voteInfo.NewPos + voteInfo.WithdrawPos +
			voteInfo.WithdrawFreezePos + voteInfo.WithdrawUnfreezePos
		voteInfo.ConsensusPos = 0
		voteInfo.FreezePos = 0
		voteInfo.NewPos = 0
		voteInfo.WithdrawPos = 0
		voteInfo.WithdrawFreezePos = 0
		if voteInfo.Address == peerPoolItem.Address {
			flag = true
			voteInfo.WithdrawUnfreezePos = voteInfo.WithdrawUnfreezePos + peerPoolItem.InitPos
		}
		err = putVoteInfo(native, ccntmract, voteInfo)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "putVoteInfo, put voteInfo error!")
		}
	}
	if flag == false {
		voteInfo := &VoteInfo{
			PeerPubkey:          peerPoolItem.PeerPubkey,
			Address:             peerPoolItem.Address,
			WithdrawUnfreezePos: peerPoolItem.InitPos,
		}
		err = putVoteInfo(native, ccntmract, voteInfo)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "putVoteInfo, put voteInfo error!")
		}
	}
	return nil
}

func blackQuit(native *native.NativeService, ccntmract common.Address, peerPoolItem *PeerPoolItem) error {
	// cntm transfer to trigger unboundcntm
	err := appCallTransferOnt(native, utils.GovernanceCcntmractAddress, utils.GovernanceCcntmractAddress, peerPoolItem.InitPos)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferOnt, cntm transfer error!")
	}

	//update total stake
	err = withdrawTotalStake(native, ccntmract, peerPoolItem.Address, peerPoolItem.InitPos)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "withdrawTotalStake, withdrawTotalStake error!")
	}

	initPos := peerPoolItem.InitPos
	var votePos uint64

	//get globalParam
	globalParam, err := getGlobalParam(native, ccntmract)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "getGlobalParam, getGlobalParam error!")
	}

	peerPubkeyPrefix, err := hex.DecodeString(peerPoolItem.PeerPubkey)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
	}
	//draw back vote pos
	stateValues, err := native.CloneCache.Store.Find(scommon.ST_STORAGE, utils.ConcatKey(ccntmract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Store.Find, get all peerPool error!")
	}
	voteInfo := new(VoteInfo)
	for _, v := range stateValues {
		voteInfoStore, ok := v.Value.(*cstates.StorageItem)
		if !ok {
			return errors.NewErr("voteInfoStore is not available!")
		}
		if err := voteInfo.Deserialize(bytes.NewBuffer(voteInfoStore.Value)); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize voteInfo error!")
		}
		total := voteInfo.ConsensusPos + voteInfo.FreezePos + voteInfo.NewPos + voteInfo.WithdrawPos +
			voteInfo.WithdrawFreezePos + voteInfo.WithdrawUnfreezePos
		penalty := (globalParam.Penalty*total + 99) / 100
		voteInfo.WithdrawUnfreezePos = total - penalty
		voteInfo.ConsensusPos = 0
		voteInfo.FreezePos = 0
		voteInfo.NewPos = 0
		voteInfo.WithdrawPos = 0
		voteInfo.WithdrawFreezePos = 0
		address := voteInfo.Address
		err = putVoteInfo(native, ccntmract, voteInfo)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "putVoteInfo, put voteInfo error!")
		}

		//update total stake
		err = withdrawTotalStake(native, ccntmract, address, penalty)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "withdrawTotalStake, withdrawTotalStake error!")
		}
		votePos = votePos + penalty
	}

	//add penalty stake
	err = depositPenaltyStake(native, ccntmract, peerPoolItem.PeerPubkey, initPos, votePos)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "depositPenaltyStake, deposit penaltyStake error!")
	}
	return nil
}

func consensusToConsensus(native *native.NativeService, ccntmract common.Address, peerPoolItem *PeerPoolItem) error {
	peerPubkeyPrefix, err := hex.DecodeString(peerPoolItem.PeerPubkey)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
	}
	//update voteInfoPool
	stateValues, err := native.CloneCache.Store.Find(scommon.ST_STORAGE, utils.ConcatKey(ccntmract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Store.Find, get all peerPool error!")
	}
	voteInfo := new(VoteInfo)
	for _, v := range stateValues {
		voteInfoStore, ok := v.Value.(*cstates.StorageItem)
		if !ok {
			return errors.NewErr("voteInfoStore is not available!")
		}
		if err := voteInfo.Deserialize(bytes.NewBuffer(voteInfoStore.Value)); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize voteInfo error!")
		}
		if voteInfo.FreezePos != 0 {
			return errors.NewErr("commitPos, freezePos should be 0!")
		}
		newPos := voteInfo.NewPos
		voteInfo.ConsensusPos = voteInfo.ConsensusPos + newPos
		voteInfo.NewPos = 0
		withdrawPos := voteInfo.WithdrawPos
		withdrawFreezePos := voteInfo.WithdrawFreezePos
		voteInfo.WithdrawFreezePos = withdrawPos
		voteInfo.WithdrawUnfreezePos = voteInfo.WithdrawUnfreezePos + withdrawFreezePos
		voteInfo.WithdrawPos = 0

		err = putVoteInfo(native, ccntmract, voteInfo)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "putVoteInfo, put voteInfo error!")
		}
	}
	return nil
}

func unConsensusToConsensus(native *native.NativeService, ccntmract common.Address, peerPoolItem *PeerPoolItem) error {
	peerPubkeyPrefix, err := hex.DecodeString(peerPoolItem.PeerPubkey)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
	}
	//update voteInfoPool
	stateValues, err := native.CloneCache.Store.Find(scommon.ST_STORAGE, utils.ConcatKey(ccntmract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Store.Find, get all peerPool error!")
	}
	voteInfo := new(VoteInfo)
	for _, v := range stateValues {
		voteInfoStore, ok := v.Value.(*cstates.StorageItem)
		if !ok {
			return errors.NewErr("voteInfoStore is not available!")
		}
		if err := voteInfo.Deserialize(bytes.NewBuffer(voteInfoStore.Value)); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize voteInfo error!")
		}
		if voteInfo.ConsensusPos != 0 {
			return errors.NewErr("consensusPos, freezePos should be 0!")
		}

		voteInfo.ConsensusPos = voteInfo.ConsensusPos + voteInfo.FreezePos + voteInfo.NewPos
		voteInfo.NewPos = 0
		voteInfo.FreezePos = 0
		withdrawPos := voteInfo.WithdrawPos
		withdrawFreezePos := voteInfo.WithdrawFreezePos
		voteInfo.WithdrawFreezePos = withdrawPos
		voteInfo.WithdrawUnfreezePos = voteInfo.WithdrawUnfreezePos + withdrawFreezePos
		voteInfo.WithdrawPos = 0

		err = putVoteInfo(native, ccntmract, voteInfo)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "putVoteInfo, put voteInfo error!")
		}
	}
	return nil
}

func consensusToUnConsensus(native *native.NativeService, ccntmract common.Address, peerPoolItem *PeerPoolItem) error {
	peerPubkeyPrefix, err := hex.DecodeString(peerPoolItem.PeerPubkey)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
	}
	//update voteInfoPool
	stateValues, err := native.CloneCache.Store.Find(scommon.ST_STORAGE, utils.ConcatKey(ccntmract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Store.Find, get all peerPool error!")
	}
	voteInfo := new(VoteInfo)
	for _, v := range stateValues {
		voteInfoStore, ok := v.Value.(*cstates.StorageItem)
		if !ok {
			return errors.NewErr("voteInfoStore is not available!")
		}
		if err := voteInfo.Deserialize(bytes.NewBuffer(voteInfoStore.Value)); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize voteInfo error!")
		}
		if voteInfo.FreezePos != 0 {
			return errors.NewErr("commitPos, freezePos should be 0!")
		}

		voteInfo.FreezePos = voteInfo.ConsensusPos + voteInfo.NewPos
		voteInfo.NewPos = 0
		voteInfo.ConsensusPos = 0
		withdrawPos := voteInfo.WithdrawPos
		withdrawFreezePos := voteInfo.WithdrawFreezePos
		voteInfo.WithdrawFreezePos = withdrawPos
		voteInfo.WithdrawUnfreezePos = voteInfo.WithdrawUnfreezePos + withdrawFreezePos
		voteInfo.WithdrawPos = 0

		err = putVoteInfo(native, ccntmract, voteInfo)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "putVoteInfo, put voteInfo error!")
		}
	}
	return nil
}

func unConsensusToUnConsensus(native *native.NativeService, ccntmract common.Address, peerPoolItem *PeerPoolItem) error {
	peerPubkeyPrefix, err := hex.DecodeString(peerPoolItem.PeerPubkey)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
	}
	//update voteInfoPool
	stateValues, err := native.CloneCache.Store.Find(scommon.ST_STORAGE, utils.ConcatKey(ccntmract, []byte(VOTE_INFO_POOL), peerPubkeyPrefix))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "native.CloneCache.Store.Find, get all peerPool error!")
	}
	voteInfo := new(VoteInfo)
	for _, v := range stateValues {
		voteInfoStore, ok := v.Value.(*cstates.StorageItem)
		if !ok {
			return errors.NewErr("voteInfoStore is not available!")
		}
		if err := voteInfo.Deserialize(bytes.NewBuffer(voteInfoStore.Value)); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "deserialize, deserialize voteInfo error!")
		}
		if voteInfo.ConsensusPos != 0 {
			return errors.NewErr("consensusPos, freezePos should be 0!")
		}

		newPos := voteInfo.NewPos
		freezePos := voteInfo.FreezePos
		voteInfo.NewPos = 0
		voteInfo.FreezePos = newPos + freezePos
		withdrawPos := voteInfo.WithdrawPos
		withdrawFreezePos := voteInfo.WithdrawFreezePos
		voteInfo.WithdrawFreezePos = withdrawPos
		voteInfo.WithdrawUnfreezePos = voteInfo.WithdrawUnfreezePos + withdrawFreezePos
		voteInfo.WithdrawPos = 0

		err = putVoteInfo(native, ccntmract, voteInfo)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "putVoteInfo, put voteInfo error!")
		}
	}
	return nil
}

func depositTotalStake(native *native.NativeService, ccntmract common.Address, address common.Address, stake uint64) error {
	totalStake, err := getTotalStake(native, ccntmract, address)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "getTotalStake, get totalStake error!")
	}

	preStake := totalStake.Stake
	preTimeOffset := totalStake.TimeOffset
	timeOffset := native.Time - constants.GENESIS_BLOCK_TIMESTAMP

	amount := utils.CalcUnbindOng(preStake, preTimeOffset, timeOffset)
	err = appCallTransferFromOng(native, utils.GovernanceCcntmractAddress, utils.OntCcntmractAddress, totalStake.Address, amount)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferFromOng, transfer from cntm error!")
	}

	totalStake.Stake = preStake + stake
	totalStake.TimeOffset = timeOffset

	err = putTotalStake(native, ccntmract, totalStake)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "putTotalStake, put totalStake error!")
	}
	return nil
}

func withdrawTotalStake(native *native.NativeService, ccntmract common.Address, address common.Address, stake uint64) error {
	totalStake, err := getTotalStake(native, ccntmract, address)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "getTotalStake, get totalStake error!")
	}
	if totalStake.Stake < stake {
		return errors.NewErr("withdraw, cntm deposit is not enough!")
	}

	preStake := totalStake.Stake
	preTimeOffset := totalStake.TimeOffset
	timeOffset := native.Time - constants.GENESIS_BLOCK_TIMESTAMP

	amount := utils.CalcUnbindOng(preStake, preTimeOffset, timeOffset)
	err = appCallTransferFromOng(native, utils.GovernanceCcntmractAddress, utils.OntCcntmractAddress, totalStake.Address, amount)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferFromOng, transfer from cntm error!")
	}

	totalStake.Stake = preStake - stake
	totalStake.TimeOffset = timeOffset

	err = putTotalStake(native, ccntmract, totalStake)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "putTotalStake, put totalStake error!")
	}
	return nil
}

func depositPenaltyStake(native *native.NativeService, ccntmract common.Address, peerPubkey string, initPos uint64, votePos uint64) error {
	penaltyStake, err := getPenaltyStake(native, ccntmract, peerPubkey)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "getPenaltyStake, get penaltyStake error!")
	}

	preInitPos := penaltyStake.InitPos
	preVotePos := penaltyStake.VotePos
	preStake := preInitPos + preVotePos
	preTimeOffset := penaltyStake.TimeOffset
	preAmount := penaltyStake.Amount
	timeOffset := native.Time - constants.GENESIS_BLOCK_TIMESTAMP

	amount := utils.CalcUnbindOng(preStake, preTimeOffset, timeOffset)

	penaltyStake.Amount = preAmount + amount
	penaltyStake.InitPos = preInitPos + initPos
	penaltyStake.VotePos = preVotePos + votePos
	penaltyStake.TimeOffset = timeOffset

	err = putPenaltyStake(native, ccntmract, penaltyStake)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "putPenaltyStake, put penaltyStake error!")
	}
	return nil
}

func withdrawPenaltyStake(native *native.NativeService, ccntmract common.Address, peerPubkey string, address common.Address) error {
	penaltyStake, err := getPenaltyStake(native, ccntmract, peerPubkey)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "getPenaltyStake, get penaltyStake error!")
	}

	preStake := penaltyStake.InitPos + penaltyStake.VotePos
	preTimeOffset := penaltyStake.TimeOffset
	preAmount := penaltyStake.Amount
	timeOffset := native.Time - constants.GENESIS_BLOCK_TIMESTAMP

	amount := utils.CalcUnbindOng(preStake, preTimeOffset, timeOffset)

	//cntm transfer
	err = appCallTransferOnt(native, utils.GovernanceCcntmractAddress, address, preStake)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferOnt, cntm transfer error!")
	}
	//cntm approve
	err = appCallTransferFromOng(native, utils.GovernanceCcntmractAddress, utils.OntCcntmractAddress, address, amount+preAmount)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransferFromOng, transfer from cntm error!")
	}

	peerPubkeyPrefix, err := hex.DecodeString(peerPubkey)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "hex.DecodeString, peerPubkey format error!")
	}
	native.CloneCache.Delete(scommon.ST_STORAGE, utils.ConcatKey(ccntmract, []byte(PENALTY_STAKE), peerPubkeyPrefix))
	return nil
}

func executeCommitDpos(native *native.NativeService, ccntmract common.Address, config *Configuration) error {
	//get governace view
	governanceView, err := GetGovernanceView(native, ccntmract)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "getGovernanceView, get GovernanceView error!")
	}

	//get current view
	view := governanceView.View
	newView := view + 1

	//get peerPoolMap
	peerPoolMapSplit, err := GetPeerPoolMap(native, ccntmract, view-1)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "getPeerPoolMap, get peerPoolMap error!")
	}

	//feeSplit first
	err = executeSplit(native, ccntmract, peerPoolMapSplit)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "executeSplit, executeSplit error!")
	}

	//get peerPoolMap
	peerPoolMap, err := GetPeerPoolMap(native, ccntmract, view)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "getPeerPoolMap, get peerPoolMap error!")
	}

	var peers []*PeerStakeInfo
	for _, peerPoolItem := range peerPoolMap.PeerPoolMap {
		if peerPoolItem.Status == QuitingStatus {
			err = normalQuit(native, ccntmract, peerPoolItem)
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "normalQuit, normalQuit error!")
			}
			delete(peerPoolMap.PeerPoolMap, peerPoolItem.PeerPubkey)
		}
		if peerPoolItem.Status == BlackStatus {
			err = blackQuit(native, ccntmract, peerPoolItem)
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "blackQuit, blackQuit error!")
			}
			delete(peerPoolMap.PeerPoolMap, peerPoolItem.PeerPubkey)
		}
		if peerPoolItem.Status == QuitConsensusStatus {
			peerPoolItem.Status = QuitingStatus
			peerPoolMap.PeerPoolMap[peerPoolItem.PeerPubkey] = peerPoolItem
		}

		if peerPoolItem.Status == CandidateStatus || peerPoolItem.Status == ConsensusStatus {
			stake := peerPoolItem.TotalPos + peerPoolItem.InitPos
			peers = append(peers, &PeerStakeInfo{
				Index:      peerPoolItem.Index,
				PeerPubkey: peerPoolItem.PeerPubkey,
				Stake:      stake,
			})
		}
	}
	if len(peers) < int(config.K) {
		return errors.NewErr("commitDpos, num of peers is less than K!")
	}

	// sort peers by stake
	sort.SliceStable(peers, func(i, j int) bool {
		if peers[i].Stake > peers[j].Stake {
			return true
		} else if peers[i].Stake == peers[j].Stake {
			return peers[i].PeerPubkey > peers[j].PeerPubkey
		}
		return false
	})

	// consensus peers
	for i := 0; i < int(config.K); i++ {
		peerPoolItem, ok := peerPoolMap.PeerPoolMap[peers[i].PeerPubkey]
		if !ok {
			return errors.NewErr("commitDpos, peerPubkey is not in peerPoolMap!")
		}

		if peerPoolItem.Status == ConsensusStatus {
			err = consensusToConsensus(native, ccntmract, peerPoolItem)
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "consensusToConsensus, consensusToConsensus error!")
			}
		} else {
			err = unConsensusToConsensus(native, ccntmract, peerPoolItem)
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "unConsensusToConsensus, unConsensusToConsensus error!")
			}
		}
		peerPoolItem.Status = ConsensusStatus
		peerPoolMap.PeerPoolMap[peers[i].PeerPubkey] = peerPoolItem
	}

	//non consensus peers
	for i := int(config.K); i < len(peers); i++ {
		peerPoolItem, ok := peerPoolMap.PeerPoolMap[peers[i].PeerPubkey]
		if !ok {
			return errors.NewErr("voteForPeer, peerPubkey is not in peerPoolMap!")
		}

		if peerPoolItem.Status == ConsensusStatus {
			err = consensusToUnConsensus(native, ccntmract, peerPoolItem)
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "consensusToUnConsensus, consensusToUnConsensus error!")
			}
		} else {
			err = unConsensusToUnConsensus(native, ccntmract, peerPoolItem)
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "unConsensusToUnConsensus, unConsensusToUnConsensus error!")
			}
		}
		peerPoolItem.Status = CandidateStatus
		peerPoolMap.PeerPoolMap[peers[i].PeerPubkey] = peerPoolItem
	}
	err = putPeerPoolMap(native, ccntmract, newView, peerPoolMap)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "putPeerPoolMap, put peerPoolMap error!")
	}
	oldView := view - 1
	oldViewBytes, err := GetUint32Bytes(oldView)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "GetUint32Bytes, get oldViewBytes error!")
	}
	native.CloneCache.Delete(scommon.ST_STORAGE, utils.ConcatKey(ccntmract, []byte(PEER_POOL), oldViewBytes))

	//update view
	governanceView = &GovernanceView{
		View:   newView,
		Height: native.Height,
		TxHash: native.Tx.Hash(),
	}
	err = putGovernanceView(native, ccntmract, governanceView)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "putGovernanceView, put governanceView error!")
	}

	return nil
}

func executeSplit(native *native.NativeService, ccntmract common.Address, peerPoolMap *PeerPoolMap) error {
	balance, err := getOngBalance(native, utils.GovernanceCcntmractAddress)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "executeSplit, getOngBalance error!")
	}
	//get globalParam
	globalParam, err := getGlobalParam(native, ccntmract)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "getGlobalParam, getGlobalParam error!")
	}

	peersCandidate := []*CandidateSplitInfo{}

	for _, peerPoolItem := range peerPoolMap.PeerPoolMap {
		if peerPoolItem.Status == CandidateStatus || peerPoolItem.Status == ConsensusStatus {
			stake := peerPoolItem.TotalPos + peerPoolItem.InitPos
			peersCandidate = append(peersCandidate, &CandidateSplitInfo{
				PeerPubkey: peerPoolItem.PeerPubkey,
				InitPos:    peerPoolItem.InitPos,
				Address:    peerPoolItem.Address,
				Stake:      stake,
			})
		}
	}

	// get config
	config, err := getConfig(native, ccntmract)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "getConfig, get config error!")
	}

	// sort peers by stake
	sort.SliceStable(peersCandidate, func(i, j int) bool {
		if peersCandidate[i].Stake > peersCandidate[j].Stake {
			return true
		} else if peersCandidate[i].Stake == peersCandidate[j].Stake {
			return peersCandidate[i].PeerPubkey > peersCandidate[j].PeerPubkey
		}
		return false
	})

	// cal s of each consensus node
	var sum uint64
	for i := 0; i < int(config.K); i++ {
		sum += peersCandidate[i].Stake
	}
	// if sum = 0, means consensus peer in config, do not split
	if sum < uint64(config.K) {
		return nil
	}
	avg := sum / uint64(config.K)
	var sumS uint64
	for i := 0; i < int(config.K); i++ {
		peersCandidate[i].S, err = splitCurve(native, ccntmract, peersCandidate[i].Stake, avg, globalParam.Yita)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "splitCurve, calculate splitCurve error!")
		}
		sumS += peersCandidate[i].S
	}
	if sumS == 0 {
		return errors.NewErr("executeSplit, sumS is 0!")
	}

	//fee split of consensus peer
	for i := int(config.K) - 1; i >= 0; i-- {
		nodeAmount := balance * globalParam.A / 100 * peersCandidate[i].S / sumS
		address := peersCandidate[i].Address
		err = appCallTransferOng(native, utils.GovernanceCcntmractAddress, address, nodeAmount)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "executeSplit, cntm transfer error!")
		}
	}

	//fee split of candidate peer
	// cal s of each candidate node
	sum = 0
	for i := int(config.K); i < len(peersCandidate); i++ {
		sum += peersCandidate[i].Stake
	}
	if sum == 0 {
		return nil
	}
	for i := int(config.K); i < len(peersCandidate); i++ {
		nodeAmount := balance * globalParam.B / 100 * peersCandidate[i].Stake / sum
		address := peersCandidate[i].Address
		err = appCallTransferOng(native, utils.GovernanceCcntmractAddress, address, nodeAmount)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "executeSplit, cntm transfer error!")
		}
	}

	return nil
}
