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
	"fmt"
	"math"

	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/log"
	"github.com/cntmio/cntmology/errors"
	"github.com/cntmio/cntmology/smartccntmract/service/native"
	"github.com/cntmio/cntmology/smartccntmract/service/native/utils"
)

func FsGetNodeInfoList(native *native.NativeService) ([]byte, error) {
	var nodesInfoList FsNodeInfoList

	source := common.NewZeroCopySource(native.Input)
	count, err := utils.DecodeVarUint(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsGetNodeInfoList DecodeVarBytes error!")
	}

	nodeList := getNodeAddrList(native)
	if nodeList != nil {
		txHash := native.Tx.Hash()
		seed := txHash.ToArray()
		nodeListLen := len(nodeList)
		randSlice := genRandSlice(uint64(nodeListLen), seed, native.InvokeParam.Address)
		sortByRandSlice(randSlice, nodeList)
	}

	for _, addr := range nodeList {
		nodeInfo := getNodeInfo(native, addr)
		if nodeInfo == nil {
			log.Errorf("[APP SDK] FsGetNodeInfoList getNodeInfo(%v) error", addr)
			ccntminue
		}
		nodesInfoList.NodesInfo = append(nodesInfoList.NodesInfo, *nodeInfo)
		if uint64(len(nodesInfoList.NodesInfo)) == count {
			break
		}
	}

	sink := common.NewZeroCopySink(nil)
	nodesInfoList.Serialization(sink)

	return EncRet(true, sink.Bytes()), nil
}

func FsChallenge(native *native.NativeService) ([]byte, error) {
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress

	var pdpRecord *PdpRecord
	var challenge Challenge
	challengeSrc := common.NewZeroCopySource(native.Input)
	challengeData, err := DecodeVarBytes(challengeSrc)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsChallenge DecodeVarBytes error!")
	}

	globalParam, err := getGlobalParam(native)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsChallenge getGlobalParam error!")
	}

	source := common.NewZeroCopySource(challengeData)
	if err := challenge.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsChallenge Deserialization error!")
	}

	if !native.CcntmextRef.CheckWitness(challenge.FileOwner) {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsChallenge CheckChallenger failed!")
	}

	fileInfo := getFileInfoFromDb(native, challenge.FileOwner, challenge.FileHash)
	if fileInfo == nil {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsChallenge getFileInfo failed!")
	}

	if !fileInfo.ValidFlag {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsChallenge file is invalid!")
	}

	if pdpRecord = getPdpRecord(native, challenge.FileHash, challenge.FileOwner, challenge.NodeAddr); pdpRecord == nil {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsChallenge node has no pdp record!")
	}

	if oldChallenge := getChallenge(native, challenge.NodeAddr, challenge.FileHash); oldChallenge != nil {
		if oldChallenge.State == NoReplyAndExpire {
			return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsChallenge Need to call Judge first!")
		} else if oldChallenge.State == NoReplyAndValid {
			return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsChallenge challenge is already existed!")
		} else if oldChallenge.State == RepliedButVerifyError {
			return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsChallenge challenge state is RepliedButVerifyError!")
		} else if oldChallenge.State == Judged {
			return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsChallenge challenge state is Judged!")
		}
	}

	challenge.State = NoReplyAndValid
	nativeFormatTime := uint64(native.Time)
	if err = checkUint64OverflowWithSum(nativeFormatTime, globalParam.ChallengeInterval); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[APP SDK] FsChallenge error: %s", err.Error())
	}
	challenge.ExpiredTime = nativeFormatTime + globalParam.ChallengeInterval
	challenge.ChallengeHeight = uint64(native.Height)
	if err = checkUint64OverflowWithSum(globalParam.ChallengeReward, globalParam.CcntmractInvokeGasFee); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[APP SDK] FsChallenge error: %s", err.Error())
	}
	challenge.Reward = globalParam.ChallengeReward + globalParam.CcntmractInvokeGasFee

	err = appCallTransfer(native, utils.OngCcntmractAddress, challenge.FileOwner, ccntmract, challenge.Reward)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsChallenge AppCallTransfer, transfer error!")
	}

	addChallenge(native, &challenge)
	return utils.BYTE_TRUE, nil
}

func FsJudge(native *native.NativeService) ([]byte, error) {
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress

	var challengeReq Challenge
	challengeSrc := common.NewZeroCopySource(native.Input)
	challengeData, err := DecodeVarBytes(challengeSrc)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsJudge DecodeVarBytes error!")
	}

	source := common.NewZeroCopySource(challengeData)
	if err := challengeReq.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsJudge Deserialization error!")
	}

	globalParam, err := getGlobalParam(native)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsJudge getGlobalParam error!")
	}

	if !native.CcntmextRef.CheckWitness(challengeReq.FileOwner) {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsJudge CheckFileOwner failed!")
	}

	challenge := getChallenge(native, challengeReq.NodeAddr, challengeReq.FileHash)
	if challenge == nil {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsJudge getChallenge challenge is nil!")
	}

	switch challenge.State {
	case RepliedAndSuccess:
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsJudge challenge state is RepliedAndSuccess!")
	case RepliedButVerifyError:
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsJudge challenge state is RepliedButVerifyError!")
	case NoReplyAndValid:
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsJudge challenge state is NoReplyAndValid!")
	case Judged:
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsJudge challenge state is Judged!")
	}

	//go on when challenge has no reply and expired
	nodeInfo := getNodeInfo(native, challenge.NodeAddr)
	if nodeInfo == nil {
		return utils.BYTE_FALSE, fmt.Errorf("[APP SDK] FsJudge getNodeInfo(%v) error", challenge.NodeAddr)
	}

	fileInfo := getFileInfoByHash(native, challenge.FileHash)
	if fileInfo == nil {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsJudge getFileInfoByHash failed!")
	}

	//two ccntmractInvokeGasFee as client Challenge and Judge gas fee
	var punishAmount uint64
	switch fileInfo.StorageType {
	case FileStorageTypeUseFile:
		if err = checkUint64OverflowWithSum(fileInfo.PayAmount, 2*globalParam.CcntmractInvokeGasFee); err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("[APP SDK] FsJudge error: %s", err.Error())
		}
		punishAmount = fileInfo.PayAmount + 2*globalParam.CcntmractInvokeGasFee
	case FileStorageTypeUseSpace:
		spaceInfo := getSpaceInfoFromDb(native, fileInfo.FileOwner)
		if spaceInfo == nil {
			return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsJudge getSpaceRawRealInfo error!")
		}
		punishAmount = calcSpaceModePerServerProfit(spaceInfo.TimeExpired, spaceInfo.TimeExpired, fileInfo)
	default:
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsJudge file StorageType error!")
	}

	if nodeInfo.Profit > punishAmount {
		nodeInfo.Profit -= punishAmount
	} else if nodeInfo.Pledge > punishAmount {
		nodeInfo.Pledge -= punishAmount
	} else {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsJudge node profit or pledge not enough!")
	}
	if err = checkUint64OverflowWithSum(punishAmount, 2*challenge.Reward); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[APP SDK] FsJudge error: %s", err.Error())
	}
	challenge.Reward = punishAmount + challenge.Reward
	err = appCallTransfer(native, utils.OngCcntmractAddress, ccntmract, challenge.FileOwner, challenge.Reward)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsJudge AppCallTransfer, transfer error!")
	}
	challenge.State = Judged

	addNodeInfo(native, nodeInfo)
	addChallenge(native, challenge)

	return utils.BYTE_TRUE, nil
}

func FsGetChallenge(native *native.NativeService) ([]byte, error) {
	var challengeReq Challenge
	challengeSrc := common.NewZeroCopySource(native.Input)
	challengeData, err := DecodeVarBytes(challengeSrc)
	if err != nil {
		return EncRet(false, []byte("[APP SDK] FsGetChallenge DecodeVarBytes error!")), nil
	}

	source := common.NewZeroCopySource(challengeData)
	if err := challengeReq.Deserialization(source); err != nil {
		return EncRet(false, []byte("[APP SDK] FsGetChallenge Deserialization error!")), nil
	}

	challenge := getChallenge(native, challengeReq.NodeAddr, challengeReq.FileHash)
	if challenge == nil {
		return EncRet(false, []byte("[APP SDK] FsGetChallenge challenge is nil!")), nil
	}

	sink := common.NewZeroCopySink(nil)
	challenge.Serialization(sink)

	return utils.BYTE_TRUE, nil
}

func FsGetFileChallengeList(native *native.NativeService) ([]byte, error) {
	var getFileChallengeReq Challenge
	getFileChallengeSrc := common.NewZeroCopySource(native.Input)
	getFileChallengeData, err := DecodeVarBytes(getFileChallengeSrc)
	if err != nil {
		return EncRet(false, []byte("[APP SDK] FsGetFileChallengeList DecodeVarBytes error!")), nil
	}
	source := common.NewZeroCopySource(getFileChallengeData)
	if err := getFileChallengeReq.Deserialization(source); err != nil {
		return EncRet(false, []byte("[APP SDK] FsGetFileChallengeList Deserialization error!")), nil
	}

	pdpRecordList := getPdpRecordList(native, getFileChallengeReq.FileHash, getFileChallengeReq.FileOwner)
	if pdpRecordList == nil {
		return EncRet(false, []byte("[APP SDK] FsGetFileChallengeList Deserialization error!")), nil
	}

	challengeList := getFileChallengeList(native, pdpRecordList)
	if challengeList == nil {
		return EncRet(false, []byte("[Node Business] FsGetFileChallengeList challengeList is nil!")), nil
	}

	sink := common.NewZeroCopySink(nil)
	challengeList.Serialization(sink)

	return EncRet(true, sink.Bytes()), nil
}

func FsCreateSpace(native *native.NativeService) ([]byte, error) {
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress

	var spaceInfo SpaceInfo
	spaceInfoSrc := common.NewZeroCopySource(native.Input)
	spaceInfoData, err := DecodeVarBytes(spaceInfoSrc)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsCreateSpace DecodeVarBytes error!")
	}

	source := common.NewZeroCopySource(spaceInfoData)
	if err := spaceInfo.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsCreateSpace Deserialization error!")
	}

	if !native.CcntmextRef.CheckWitness(spaceInfo.SpaceOwner) {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsCreateSpace CheckSpaceOwner failed!")
	}

	if spaceInfoExist(native, spaceInfo.SpaceOwner) {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsCreateSpace Space has been created!")
	}

	globalParam, err := getGlobalParam(native)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsCreateSpace getGlobalParam error!")
	}

	if err = checkUint64OverflowWithSum(uint64(native.Time), globalParam.MinTimeForFileStorage); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[APP SDK] FsCreateSpace error: %s", err.Error())
	}

	if spaceInfo.TimeExpired < uint64(native.Time)+globalParam.MinTimeForFileStorage {
		err = fmt.Errorf("[APP SDK] FsCreateSpace spaceInfo TimeExpired smaller than Native.Time + %d",
			globalParam.MinTimeForFileStorage)
		return utils.BYTE_FALSE, err
	}
	if spaceInfo.Volume < DefaultPerBlockSize {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsCreateSpace space volume smaller than DefaultPerBlockSize(256kb)")
	}

	spaceInfo.Volume = formatVolumeToBlock(spaceInfo.Volume)
	spaceInfo.ValidFlag = true
	spaceInfo.RestVol = spaceInfo.Volume
	spaceInfo.TimeStart = uint64(native.Time)
	spaceInfo.TimeExpired = formatUint64TimeToHour(spaceInfo.TimeExpired)
	spaceInfo.CurrFeeRate = globalParam.SpacePerBlockFeeRate
	spaceInfo.PayAmount = calcTotalPayAmountWithSpace(&spaceInfo)
	spaceInfo.RestAmount = spaceInfo.PayAmount

	err = appCallTransfer(native, utils.OngCcntmractAddress, spaceInfo.SpaceOwner, ccntmract, spaceInfo.PayAmount)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsCreateSpace AppCallTransfer, transfer error!")
	}
	addSpaceInfo(native, &spaceInfo)
	return utils.BYTE_TRUE, nil
}

func FsDeleteSpace(native *native.NativeService) ([]byte, error) {
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress

	source := common.NewZeroCopySource(native.Input)
	spaceOwner, err := utils.DecodeAddress(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsDeleteSpace DecodeAddress error!")
	}

	if !native.CcntmextRef.CheckWitness(spaceOwner) {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsDeleteSpace CheckSpaceOwner failed!")
	}

	spaceInfo := getSpaceInfoFromDb(native, spaceOwner)
	if spaceInfo == nil {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsDeleteSpace getSpaceInfoFromDb error!")
	}
	if spaceInfo.Volume != spaceInfo.RestVol {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsDeleteSpace not allow, check files!")
	}

	err = appCallTransfer(native, utils.OngCcntmractAddress, ccntmract, spaceInfo.SpaceOwner, spaceInfo.RestAmount)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsDeleteSpace AppCallTransfer, transfer error!")
	}

	delSpaceInfo(native, spaceOwner)
	return utils.BYTE_TRUE, nil
}

func FsUpdateSpace(native *native.NativeService) ([]byte, error) {
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress

	var spaceUpdate SpaceUpdate
	spaceUpdateSrc := common.NewZeroCopySource(native.Input)
	spaceInfoData, err := DecodeVarBytes(spaceUpdateSrc)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsUpdateSpace DecodeVarBytes error!")
	}

	source := common.NewZeroCopySource(spaceInfoData)
	if err := spaceUpdate.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsUpdateSpace Deserialization error!")
	}

	spaceUpdate.NewVolume = formatVolumeToBlock(spaceUpdate.NewVolume)

	if spaceUpdate.NewTimeExpired == 0 && spaceUpdate.NewVolume == 0 {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsUpdateSpace Param error!")
	}

	if !native.CcntmextRef.CheckWitness(spaceUpdate.Payer) {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsUpdateSpace CheckPayer failed!")
	}

	globalParam, err := getGlobalParam(native)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsCreateSpace getGlobalParam error!")
	}

	spaceInfo := getAndUpdateSpaceInfo(native, spaceUpdate.SpaceOwner)
	if spaceInfo == nil {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsUpdateSpace getAndUpdateSpaceInfo error!")
	}

	if !spaceInfo.ValidFlag {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsUpdateSpace space timeExpired! please create space again")
	}

	if spaceUpdate.NewTimeExpired != 0 && uint64(native.Time) >= spaceUpdate.NewTimeExpired {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsUpdateSpace NewTimeExpired error!")
	}

	if spaceUpdate.NewTimeExpired != 0 {
		if err = checkUint64OverflowWithSum(spaceInfo.TimeStart, globalParam.MinTimeForFileStorage); err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("[APP SDK] FsUpdateSpace error: %s", err.Error())
		}
		if spaceUpdate.NewTimeExpired < spaceInfo.TimeStart+globalParam.MinTimeForFileStorage {
			err = fmt.Errorf("[APP SDK] FsUpdateSpace spaceInfo NewTimeExpired smaller than TimeStart + %d",
				globalParam.MinTimeForFileStorage)
			return utils.BYTE_FALSE, err
		}
	}

	if spaceUpdate.NewTimeExpired == 0 {
		spaceUpdate.NewTimeExpired = spaceInfo.TimeExpired
	}

	if spaceUpdate.NewVolume == 0 {
		spaceUpdate.NewVolume = spaceInfo.Volume
	}

	if spaceInfo.Volume-spaceInfo.RestVol >= spaceUpdate.NewVolume {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsUpdateSpace NewVolume is not enough!")
	}

	spaceInfo.RestVol = spaceUpdate.NewVolume - (spaceInfo.Volume - spaceInfo.RestVol)
	spaceInfo.Volume = spaceUpdate.NewVolume
	spaceInfo.TimeExpired = formatUint64TimeToHour(spaceUpdate.NewTimeExpired)

	newPayAmount := calcTotalPayAmountWithSpace(spaceInfo)

	var newFee uint64
	var payer, payee common.Address
	if newPayAmount > spaceInfo.PayAmount {
		newFee = newPayAmount - spaceInfo.PayAmount
		payer = spaceUpdate.Payer
		payee = ccntmract
		if err = checkUint64OverflowWithSum(spaceInfo.RestAmount, newFee); err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("[APP SDK] FsUpdateSpace error: %s", err.Error())
		}
		spaceInfo.RestAmount += newFee
	} else if newPayAmount < spaceInfo.PayAmount {
		newFee = spaceInfo.PayAmount - newPayAmount
		payee = spaceUpdate.Payer
		payer = ccntmract
		if spaceInfo.RestAmount < newFee {
			return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsUpdateSpace space RestAmount < newFee error!")
		}
		spaceInfo.RestAmount -= newFee
	} else {
		newFee = 0
	}
	spaceInfo.PayAmount = newPayAmount

	if newFee != 0 {
		err = appCallTransfer(native, utils.OngCcntmractAddress, payer, payee, newFee)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsUpdateSpace AppCallTransfer, transfer error!")
		}
	}

	addSpaceInfo(native, spaceInfo)
	return utils.BYTE_TRUE, nil
}

func FsGetSpaceInfo(native *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(native.Input)
	spaceOwner, err := utils.DecodeAddress(source)
	if err != nil {
		return EncRet(false, []byte("[APP SDK] FsGetSpaceInfo DecodeAddress error!")), nil
	}

	spaceInfo := getSpaceRawRealInfo(native, spaceOwner)
	if spaceInfo == nil {
		return EncRet(false, []byte("[APP SDK] FsGetSpaceInfo getSpaceRawInfo error!")), nil
	}

	return EncRet(true, spaceInfo), nil
}

func FsStoreFiles(native *native.NativeService) ([]byte, error) {
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress

	var errInfos Errors
	var fileInfoList FileInfoList
	source := common.NewZeroCopySource(native.Input)
	fileInfoListData, err := DecodeVarBytes(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsStoreFiles DecodeVarBytes error!")
	}

	fileInfoListDataSrc := common.NewZeroCopySource(fileInfoListData)
	if err := fileInfoList.Deserialization(fileInfoListDataSrc); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsStoreFiles Deserialization error!")
	}

	globalParam, err := getGlobalParam(native)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsStoreFiles getGlobalParam error!")
	}

	for _, fileInfo := range fileInfoList.FilesI {
		if !native.CcntmextRef.CheckWitness(fileInfo.FileOwner) {
			errInfos.AddObjectError(string(fileInfo.FileHash), "[APP SDK] FsStoreFiles CheckFileOwner failed!")
			log.Error("[APP SDK] FsStoreFiles CheckFileOwner failed!")
			ccntminue
		}

		if fileExist := getAndUpdateFileInfo(native, fileInfo.FileOwner, fileInfo.FileHash); fileExist != nil {
			if !fileExist.ValidFlag {
				log.Debug("[APP SDK] FsStoreFiles Delete old fileInfo")
				if !deleteFile(native, fileExist, &errInfos) {
					ccntminue
				}
			} else {
				errInfos.AddObjectError(string(fileInfo.FileHash), "[APP SDK] FsStoreFiles File has stored!")
				log.Debug("[APP SDK] FsStoreFiles File has stored!")
				ccntminue
			}
		}

		fileInfo.ValidFlag = true
		fileInfo.BeginHeight = uint64(native.Height)
		fileInfo.TimeStart = uint64(native.Time)
		fileInfo.TimeExpired = formatUint64TimeToHour(fileInfo.TimeExpired)

		log.Debugf("[APP SDK] FsStoreFiles BlockCount:%d, PayAmount :%d\n", fileInfo.FileBlockCount, fileInfo.PayAmount)

		if fileInfo.StorageType == FileStorageTypeUseSpace {
			spaceInfo := getAndUpdateSpaceInfo(native, fileInfo.FileOwner)
			if spaceInfo == nil {
				errInfos.AddObjectError(string(fileInfo.FileHash), "[APP SDK] FsStoreFiles getAndUpdateSpaceInfo error!")
				ccntminue
			}
			if !spaceInfo.ValidFlag {
				errInfos.AddObjectError(string(fileInfo.FileHash), "[APP SDK] FsStoreFiles space timeExpired!")
				ccntminue
			}
			if spaceInfo.RestVol <= fileInfo.FileBlockCount*DefaultPerBlockSize {
				errInfos.AddObjectError(string(fileInfo.FileHash), "[APP SDK] FsStoreFiles RestVol is not enough error!")
				ccntminue
			}
			fileInfo.CurrFeeRate = spaceInfo.CurrFeeRate
			spaceInfo.RestVol -= fileInfo.FileBlockCount * DefaultPerBlockSize

			serverPdpGasFee := globalParam.FilePerServerPdpTimes * globalParam.CcntmractInvokeGasFee * spaceInfo.CopyNumber
			err = appCallTransfer(native, utils.OngCcntmractAddress, fileInfo.FileOwner, ccntmract, serverPdpGasFee)
			if err != nil {
				errInfos.AddObjectError(string(fileInfo.FileHash), "[APP SDK] FsStoreFiles AppCallTransfer, transfer error!")
				ccntminue
			}
			addSpaceInfo(native, spaceInfo)
		} else if fileInfo.StorageType == FileStorageTypeUseFile {
			if err = checkUint64OverflowWithSum(uint64(native.Time), globalParam.MinTimeForFileStorage); err != nil {
				return utils.BYTE_FALSE, fmt.Errorf("[APP SDK] FsStoreFiles error: %s", err.Error())
			}
			if fileInfo.TimeExpired < uint64(native.Time)+globalParam.MinTimeForFileStorage {
				errInfo := fmt.Sprintf("[APP SDK] FsStoreFiles fileInfo TimeExpired error: "+
					"TimeExpired smaller than Native.Time + %d", globalParam.MinTimeForFileStorage)
				errInfos.AddObjectError(string(fileInfo.FileHash), errInfo)
				log.Error(errInfo)
				ccntminue
			}
			serverPdpGasFee := globalParam.FilePerServerPdpTimes * globalParam.CcntmractInvokeGasFee * fileInfo.CopyNumber
			fileInfo.CurrFeeRate = globalParam.FilePerBlockFeeRate
			fileInfo.PayAmount = calcTotalPayAmountWithFile(&fileInfo)
			if err = checkUint64OverflowWithSum(fileInfo.PayAmount, serverPdpGasFee); err != nil {
				return utils.BYTE_FALSE, fmt.Errorf("[APP SDK] FsStoreFiles error: %s", err.Error())
			}
			err = appCallTransfer(native, utils.OngCcntmractAddress, fileInfo.FileOwner, ccntmract, fileInfo.PayAmount+serverPdpGasFee)
			if err != nil {
				errInfos.AddObjectError(string(fileInfo.FileHash), "[APP SDK] FsStoreFiles AppCallTransfer, transfer error!")
				ccntminue
			}
		} else {
			errInfos.AddObjectError(string(fileInfo.FileHash), "[APP SDK] FsStoreFiles unknown StorageType!")
			ccntminue
		}
		addFileInfo(native, &fileInfo)
		log.Infof("setFileOwner %s %s", fileInfo.FileHash, fileInfo.FileOwner.ToBase58())
		setFileOwner(native, fileInfo.FileHash, fileInfo.FileOwner)
	}

	errInfos.AddErrorsEvent(native)
	return utils.BYTE_TRUE, nil
}

func FsRenewFiles(native *native.NativeService) ([]byte, error) {
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress

	var errInfos Errors
	var filesReNew FileReNewList
	filesReNewSrc := common.NewZeroCopySource(native.Input)
	filesReNewData, err := DecodeVarBytes(filesReNewSrc)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsRenewFiles DecodeVarBytes error!")
	}

	source := common.NewZeroCopySource(filesReNewData)
	if err := filesReNew.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsRenewFiles Deserialization error!")
	}

	for _, fileReNew := range filesReNew.FilesReNew {
		if !native.CcntmextRef.CheckWitness(fileReNew.Payer) {
			errInfos.AddObjectError(string(fileReNew.FileHash), "[APP SDK] FsRenewFiles CheckPayer failed!")
			ccntminue
		}

		fileInfo := getAndUpdateFileInfo(native, fileReNew.FileOwner, fileReNew.FileHash)
		if fileInfo == nil {
			errInfos.AddObjectError(string(fileReNew.FileHash), "[APP SDK] FsRenewFiles getAndUpdateFileInfo error!")
			ccntminue
		}

		if fileInfo.StorageType == FileStorageTypeUseFile {
			if !fileInfo.ValidFlag {
				errInfos.AddObjectError(string(fileReNew.FileHash), "[APP SDK] FsRenewFiles File is expired! need to upload again")
				ccntminue
			}

			fileInfo.TimeExpired = formatUint64TimeToHour(fileReNew.NewTimeExpired)
			newFee := calcTotalPayAmountWithFile(fileInfo)
			if newFee < fileInfo.PayAmount {
				errInfos.AddObjectError(string(fileReNew.FileHash), "[APP SDK] FsRenewFiles newFee < fileInfo.PayAmount")
				ccntminue
			}

			renewFee := newFee - fileInfo.PayAmount
			err = appCallTransfer(native, utils.OngCcntmractAddress, fileReNew.Payer, ccntmract, renewFee)
			if err != nil {
				errInfos.AddObjectError(string(fileReNew.FileHash), "[APP SDK] FsRenewFiles AppCallTransfer, transfer error!")
				ccntminue
			}

			fileInfo.PayAmount = newFee
			addFileInfo(native, fileInfo)
		} else {
			errInfos.AddObjectError(string(fileReNew.FileHash), "[APP SDK] FsRenewFiles StorageType is not FileStorageTypeUseFile!")
		}
	}

	errInfos.AddErrorsEvent(native)
	return utils.BYTE_TRUE, nil
}

func FsDeleteFiles(native *native.NativeService) ([]byte, error) {
	var errInfos Errors
	var fileDelList FileDelList
	fileDelListSrc := common.NewZeroCopySource(native.Input)
	fileDelListData, err := DecodeVarBytes(fileDelListSrc)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsDeleteFiles DecodeVarBytes error!")
	}
	source := common.NewZeroCopySource(fileDelListData)
	if err := fileDelList.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsDeleteFiles Deserialization error!")
	}

	for _, fileDel := range fileDelList.FilesDel {
		fileInfo := getFileInfoByHash(native, fileDel.FileHash)
		if fileInfo == nil {
			errInfos.AddObjectError(string(fileDel.FileHash), "[APP SDK] FsDeleteFiles fileInfo is nil")
			ccntminue
		}
		if !native.CcntmextRef.CheckWitness(fileInfo.FileOwner) {
			errInfos.AddObjectError(string(fileDel.FileHash), "[APP SDK] FsDeleteFiles CheckFileOwner failed!")
			ccntminue
		}
		deleteFile(native, fileInfo, &errInfos)
	}

	errInfos.AddErrorsEvent(native)
	return utils.BYTE_TRUE, nil
}

func deleteChallenge(native *native.NativeService, nodeAddress common.Address, fileInfo *FileInfo) error {
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	chl := getChallenge(native, nodeAddress, fileInfo.FileHash)
	if chl == nil {
		return nil
	}

	switch chl.State {
	case NoReplyAndValid, NoReplyAndExpire:
		if err := appCallTransfer(native, utils.OngCcntmractAddress, ccntmract, fileInfo.FileOwner, chl.Reward); err != nil {
			return fmt.Errorf("deleteChallenge AppCallTransfer, transfer error: %s", err.Error())
		}
	}
	delChallenge(native, nodeAddress, fileInfo.FileHash)
	return nil
}

func deleteFile(native *native.NativeService, fileInfo *FileInfo, errInfos *Errors) bool {
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	pdpRecordList := getPdpRecordList(native, fileInfo.FileHash, fileInfo.FileOwner)

	var err error
	for _, pdpRecord := range pdpRecordList.PdpRecords {
		if err = deleteChallenge(native, pdpRecord.NodeAddr, fileInfo); err != nil {
			errInfos.AddObjectError(string(fileInfo.FileHash), "[APP SDK] DeleteFile deleteChallenge error")
			ccntminue
		}

		if pdpRecord.SettleFlag {
			ccntminue
		}
		nodeInfo := getNodeInfo(native, pdpRecord.NodeAddr)
		if nodeInfo == nil {
			errInfos.AddObjectError(string(fileInfo.FileHash), "[APP SDK] DeleteFile getNodeInfo error")
			ccntminue
		}

		var nodeProfit uint64
		switch fileInfo.StorageType {
		case FileStorageTypeUseFile:
			nodeProfit = calcFileModePerServerProfit(uint64(native.Time), fileInfo)
			if err = checkUint64OverflowWithSum(nodeInfo.Profit, nodeProfit); err != nil {
				errInfos.AddObjectError(string(fileInfo.FileHash), "[APP SDK] DeleteFile checkUint64OverflowWithSum error: "+err.Error())
				ccntminue
			}
			nodeInfo.Profit += nodeProfit
		case FileStorageTypeUseSpace:
			spaceInfo := getSpaceInfoFromDb(native, fileInfo.FileOwner)
			if spaceInfo == nil {
				errInfos.AddObjectError(string(fileInfo.FileHash), "[APP SDK] DeleteFile getSpaceInfoFromDb error!")
				ccntminue
			}
			nodeProfit = calcSpaceModePerServerProfit(uint64(native.Time), spaceInfo.TimeExpired, fileInfo)
			if err = checkUint64OverflowWithSum(nodeInfo.Profit, nodeProfit); err != nil {
				errInfos.AddObjectError(string(fileInfo.FileHash), "[APP SDK] DeleteFile checkUint64OverflowWithSum error: "+err.Error())
				ccntminue
			}
			nodeInfo.Profit += nodeProfit
		default:
			errInfos.AddObjectError(string(fileInfo.FileHash), "[APP SDK] DeleteFile file StorageType error")
			ccntminue
		}
		fileSize := fileInfo.FileBlockCount * DefaultPerBlockSize
		if err = checkUint64OverflowWithSum(nodeInfo.RestVol, fileSize); err != nil {
			errInfos.AddObjectError(string(fileInfo.FileHash), "[APP SDK] DeleteFile checkUint64OverflowWithSum error: "+err.Error())
			ccntminue
		}
		if fileInfo.RestAmount < nodeProfit {
			errInfos.AddObjectError(string(fileInfo.FileHash), "[APP SDK] DeleteFile fileInfo.RestAmount not enough")
			ccntminue
		}
		fileInfo.RestAmount -= nodeProfit
		nodeInfo.RestVol += fileSize
		addNodeInfo(native, nodeInfo)
	}

	switch fileInfo.StorageType {
	case FileStorageTypeUseFile:
		if fileInfo.RestAmount > 0 {
			err := appCallTransfer(native, utils.OngCcntmractAddress, ccntmract, fileInfo.FileOwner, fileInfo.RestAmount)
			if err != nil {
				errInfos.AddObjectError(string(fileInfo.FileHash), "[APP SDK] DeleteFile AppCallTransfer, transfer error!")
				return false
			}
		}
	case FileStorageTypeUseSpace:
		spaceInfo := getSpaceInfoFromDb(native, fileInfo.FileOwner)
		if spaceInfo == nil {
			errInfos.AddObjectError(string(fileInfo.FileHash), "[APP SDK] DeleteFile getSpaceInfoFromDb error!")
			return false
		}
		fileSize := fileInfo.FileBlockCount * DefaultPerBlockSize
		if err = checkUint64OverflowWithSum(spaceInfo.RestVol, fileSize); err != nil {
			errInfos.AddObjectError(string(fileInfo.FileHash), "[APP SDK] DeleteFile checkUint64OverflowWithSum error: "+err.Error())
			return false
		}
		spaceInfo.RestVol += fileSize
		addSpaceInfo(native, spaceInfo)
	}

	delFileInfo(native, fileInfo.FileOwner, fileInfo.FileHash)
	delFileOwner(native, fileInfo.FileHash)
	delPdpRecordList(native, fileInfo.FileHash, fileInfo.FileOwner)
	return true
}

func FsTransferFiles(native *native.NativeService) ([]byte, error) {
	//Note: May cause storage node not to find PdpInfo, so when an error occurs,
	//the storage node needs to try to commit more than once

	var errInfos Errors
	var fileTransferList FileTransferList
	fileTransferListSrc := common.NewZeroCopySource(native.Input)
	fileTransferListData, err := DecodeVarBytes(fileTransferListSrc)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsTransferFiles DecodeVarBytes error!")
	}
	source := common.NewZeroCopySource(fileTransferListData)
	if err := fileTransferList.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsTransferFiles OwnerChange Deserialization error!")
	}

	for _, fileTransfer := range fileTransferList.FilesTransfer {
		if native.CcntmextRef.CheckWitness(fileTransfer.OriOwner) == false {
			errInfos.AddObjectError(string(fileTransfer.FileHash), "[APP SDK] FsTransferFiles CheckFileOwner failed!")
			ccntminue
		}

		fileInfo := getAndUpdateFileInfo(native, fileTransfer.OriOwner, fileTransfer.FileHash)
		if fileInfo == nil {
			errInfos.AddObjectError(string(fileTransfer.FileHash), "[APP SDK] FsTransferFiles GetFsFileInfo error!")
			ccntminue
		}

		if !fileInfo.ValidFlag {
			errInfos.AddObjectError(string(fileTransfer.FileHash), "[APP SDK] FsTransferFiles File is expired!")
			ccntminue
		}

		if fileInfo.StorageType != FileStorageTypeUseFile {
			errInfos.AddObjectError(string(fileTransfer.FileHash), "[APP SDK] FsTransferFiles file StorageType is not FileStorageTypeUseFile error!")
			ccntminue
		}

		if fileInfo.FileOwner != fileTransfer.OriOwner {
			errInfos.AddObjectError(string(fileTransfer.FileHash), "[APP SDK] FsTransferFiles Caller is not file's owner!")
			ccntminue
		}

		fileInfo.FileOwner = fileTransfer.NewOwner
		delFileInfo(native, fileTransfer.OriOwner, fileTransfer.FileHash)
		addFileInfo(native, fileInfo)

		pdpRecordList := getPdpRecordList(native, fileTransfer.FileHash, fileTransfer.OriOwner)
		for _, pdpInfo := range pdpRecordList.PdpRecords {
			delPdpRecord(native, pdpInfo.FileHash, pdpInfo.FileOwner, pdpInfo.NodeAddr)
			pdpInfo.FileOwner = fileTransfer.NewOwner
			addPdpRecord(native, &pdpInfo)
		}
		delFileOwner(native, fileInfo.FileHash)
		setFileOwner(native, fileInfo.FileHash, fileInfo.FileOwner)
	}

	errInfos.AddErrorsEvent(native)
	return utils.BYTE_TRUE, nil
}

func FsGetFileHashList(native *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(native.Input)
	passportData, err := DecodeVarBytes(source)
	if err != nil {
		return EncRet(false, []byte("[APP SDK] FsGetFileHashList DecodeVarBytes error!")), nil
	}

	globalParam, err := getGlobalParam(native)
	if err != nil {
		errInfo := fmt.Sprintf("[APP SDK] FsGetFileHashList getGlobalParam error: %s", err.Error())
		return EncRet(false, []byte(errInfo)), nil
	}

	walletAddr, err := CheckPassport(uint64(native.Height), globalParam.PassportExpire, passportData)
	if err != nil {
		errInfo := fmt.Sprintf("[APP SDK] FsGetFileHashList CheckFileListOwner error: %s", err.Error())
		return EncRet(false, []byte(errInfo)), nil
	}

	fileHashList := getFileHashList(native, walletAddr)
	sink := common.NewZeroCopySink(nil)
	fileHashList.Serialization(sink)
	return EncRet(true, sink.Bytes()), nil
}

func FsGetFileInfo(native *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(native.Input)
	fileHash, err := DecodeVarBytes(source)
	if err != nil {
		return EncRet(false, []byte("[APP SDK] FsGetFileInfo DecodeBytes error!")), nil
	}

	owner, err := getFileOwner(native, fileHash)
	if err != nil {
		return EncRet(false, []byte("[APP SDK] FsGetFileInfo getFileOwner error!")), nil
	}

	fileRawInfo := getFileRawRealInfo(native, owner, fileHash)
	return EncRet(true, fileRawInfo), nil
}

func FsGetPdpInfoList(native *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(native.Input)
	fileHash, err := DecodeVarBytes(source)
	if err != nil {
		return EncRet(false, []byte("[APP SDK] FsGetPdpInfoList DecodeBytes error!")), nil
	}

	owner, err := getFileOwner(native, fileHash)
	if err != nil {
		return EncRet(false, []byte("[APP SDK] FsGetPdpInfoList getFileOwner error!")), nil
	}

	pdpInfoList := getPdpRecordList(native, fileHash, owner)
	sink := common.NewZeroCopySink(nil)
	pdpInfoList.Serialization(sink)

	return EncRet(true, sink.Bytes()), nil
}

func FsReadFilePledge(native *native.NativeService) ([]byte, error) {
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress

	var err error
	var readPledge ReadPledge

	readPledgeSrc := common.NewZeroCopySource(native.Input)
	readPledgeSrcData, err := DecodeVarBytes(readPledgeSrc)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsReadFilePledge DecodeVarBytes error!")
	}

	source := common.NewZeroCopySource(readPledgeSrcData)
	if err := readPledge.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsReadFilePledge deserialization error!")
	}

	globalParam, err := getGlobalParam(native)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsReadFilePledge getGlobalParam error!")
	}

	fileInfo := getFileInfoByHash(native, readPledge.FileHash)
	if fileInfo == nil {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsReadFilePledge getFsFileInfo error!")
	}

	if !fileInfo.ValidFlag {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsReadFilePledge file out of date!")
	}

	//validation authority
	if !native.CcntmextRef.CheckWitness(readPledge.Downloader) {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsReadFilePledge CheckDownloader failed!")
	}

	//oriPlan ==> newPlan
	var totalAddMaxBlockNumToRead uint64
	for index, readPlan := range readPledge.ReadPlans {
		if err = checkUint64OverflowWithSum(totalAddMaxBlockNumToRead, readPlan.MaxReadBlockNum); err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("[APP SDK] FsReadFilePledge checkUint64OverflowWithSum error: %s", err.Error())
		}
		totalAddMaxBlockNumToRead += readPlan.MaxReadBlockNum
		readPledge.ReadPlans[index].HaveReadBlockNum = 0
		readPledge.ReadPlans[index].NumOfSettlements = 0

	}
	var samePlanCount = uint64(0)
	oriPledge, err := getReadPledge(native, readPledge.Downloader, readPledge.FileHash)
	if err == nil && oriPledge != nil {
		for _, oriReadPlan := range oriPledge.ReadPlans {
			foundSamePlan := false
			for index, readPlan := range readPledge.ReadPlans {
				if readPlan.NodeAddr == oriReadPlan.NodeAddr {
					samePlanCount++
					foundSamePlan = true
					if err = checkUint64OverflowWithSum(readPledge.ReadPlans[index].MaxReadBlockNum, oriReadPlan.MaxReadBlockNum); err != nil {
						return utils.BYTE_FALSE, fmt.Errorf("[APP SDK] FsReadFilePledge checkUint64OverflowWithSum error: %s", err.Error())
					}
					readPledge.ReadPlans[index].MaxReadBlockNum += oriReadPlan.MaxReadBlockNum
					readPledge.ReadPlans[index].HaveReadBlockNum = oriReadPlan.HaveReadBlockNum
					readPledge.ReadPlans[index].NumOfSettlements = oriReadPlan.NumOfSettlements
				}
			}
			if !foundSamePlan {
				readPledge.ReadPlans = append(readPledge.ReadPlans, oriReadPlan)
			}
		}
		readPledge.RestMoney = oriPledge.RestMoney
	} else {
		readPledge.RestMoney = 0
	}

	newPledgeFee := totalAddMaxBlockNumToRead * globalParam.FeePerBlockForRead
	if err = checkUint64OverflowWithSum(readPledge.RestMoney, newPledgeFee); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[APP SDK] FsReadFilePledge checkUint64OverflowWithSum error: %s", err.Error())
	}
	readPledge.RestMoney += newPledgeFee

	newPlanCount := uint64(len(readPledge.ReadPlans)) - samePlanCount
	err = appCallTransfer(native, utils.OngCcntmractAddress, readPledge.Downloader, ccntmract,
		newPledgeFee+newPlanCount*globalParam.CcntmractInvokeGasFee)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[APP SDK] FsReadFilePledge AppCallTransfer, transfer error!")
	}

	addReadPledge(native, &readPledge)
	return utils.BYTE_TRUE, nil
}

func FsGetReadPledge(native *native.NativeService) ([]byte, error) {
	var getPledge GetReadPledge
	source := common.NewZeroCopySource(native.Input)
	if err := getPledge.Deserialization(source); err != nil {
		return EncRet(false, []byte("[APP SDK] FsGetReadPledge Deserialization error!")), nil
	}

	rawPledge, err := getRawReadPledge(native, getPledge.Downloader, getPledge.FileHash)
	if err != nil {
		return EncRet(false, []byte("[APP SDK] FsGetReadPledge getRawReadPledge error!")), nil
	}
	return EncRet(true, rawPledge), nil
}

func formatUint32TimeToHour(time uint32) uint64 {
	return uint64(time - time%Hour)
}

func formatUint64TimeToHour(time uint64) uint64 {
	return time - time%Hour
}

func formatVolumeToBlock(volume uint64) uint64 {
	return volume - volume%DefaultPerBlockSize
}

func checkUint64OverflowWithSum(a, b uint64) error {
	if math.MaxUint64-a < b {
		return fmt.Errorf("checkUint64OverflowWithSum (%d, %d)", a, b)
	}
	return nil
}
