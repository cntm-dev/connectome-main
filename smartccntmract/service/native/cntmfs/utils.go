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

	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/config"
	"github.com/cntmio/cntmology/errors"
	"github.com/cntmio/cntmology/smartccntmract/service/native"
	"github.com/cntmio/cntmology/smartccntmract/service/native/cntm"
	"github.com/cntmio/cntmology/smartccntmract/service/native/utils"
)

const (
	FS_SET_GLOBAL_PARAM        = "FsSetGlobalParam"
	FS_GET_GLOBAL_PARAM        = "FsGetGlobalParam"
	FS_NODE_REGISTER           = "FsNodeRegister"
	FS_NODE_QUERY              = "FsNodeQuery"
	FS_NODE_UPDATE             = "FsNodeUpdate"
	FS_NODE_CANCEL             = "FsNodeCancel"
	FS_FILE_PROVE              = "FsFileProve"
	FS_NODE_WITHDRAW_PROFIT    = "FsNodeWithdrawProfit"
	FS_CHALLENGE               = "FsChallenge"
	FS_GET_CHALLENGE           = "FsGetChallenge"
	FS_GET_FILE_CHALLENGE_LIST = "FsGetFileChallengeList"
	FS_GET_NODE_CHALLENGE_LIST = "FsGetNodeChallengeList"
	FS_RESPONSE                = "FsResponse"
	FS_JUDGE                   = "FsJudge"
	FS_GET_NODE_LIST           = "FsGetNodeList"
	FS_GET_PDP_INFO_LIST       = "FsGetPdpInfoList"
	FS_STORE_FILES             = "FsStoreFiles"
	FS_RENEW_FILES             = "FsRenewFiles"
	FS_DELETE_FILES            = "FsDeleteFiles"
	FS_TRANSFER_FILES          = "FsTransferFiles"
	FS_GET_FILE_INFO           = "FsGetFileInfo"
	FS_GET_FILE_LIST           = "FsGetFileList"
	FS_READ_FILE_PLEDGE        = "FsReadFilePledge"
	FS_READ_FILE_SETTLE        = "FsReadFileSettle"
	FS_GET_READ_PLEDGE         = "FsGetReadPledge"
	FS_CREATE_SPACE            = "FsCreateSpace"
	FS_DELETE_SPACE            = "FsDeleteSpace"
	FS_UPDATE_SPACE            = "FsUpdateSpace"
	FS_GET_SPACE_INFO          = "FsGetSpaceInfo"
)

const (
	cntmFS_GLOBAL_PARAM     = "cntmFsGlobalParam"
	cntmFS_CHALLENGE        = "cntmFsChallenge"
	cntmFS_RESPONSE         = "cntmFsResponse"
	cntmFS_NODE_INFO        = "cntmFsNodeInfo"
	cntmFS_FILE_INFO        = "cntmFsFileInfo"
	cntmFS_FILE_PDP         = "cntmFsFilePdp"
	cntmFS_FILE_OWNER       = "cntmFsFileOwner"
	cntmFS_FILE_READ_PLEDGE = "cntmFsFileReadPledge"
	cntmFS_FILE_SPACE       = "cntmFsFileSpace"
)

func GenGlobalParamKey(ccntmract common.Address) []byte {
	return append(ccntmract[:], cntmFS_GLOBAL_PARAM...)
}

func GenFsNodeInfoPrefix(ccntmract common.Address) []byte {
	prefix := append(ccntmract[:], cntmFS_NODE_INFO...)
	return prefix
}

func GenFsNodeInfoKey(ccntmract common.Address, nodeAddr common.Address) []byte {
	prefix := GenFsNodeInfoPrefix(ccntmract)
	return append(prefix, nodeAddr[:]...)
}

func GenFsFileInfoPrefix(ccntmract common.Address, fileOwner common.Address) []byte {
	prefix := append(ccntmract[:], cntmFS_FILE_INFO...)
	prefix = append(prefix, fileOwner[:]...)
	return prefix
}

func GenFsFileInfoKey(ccntmract common.Address, fileOwner common.Address, fileHash []byte) []byte {
	prefix := GenFsFileInfoPrefix(ccntmract, fileOwner)
	return append(prefix, fileHash...)
}

func GenChallengePrefix(ccntmract common.Address, nodeAddr common.Address) []byte {
	prefix := append(ccntmract[:], cntmFS_CHALLENGE...)
	prefix = append(prefix, nodeAddr[:]...)
	return prefix
}

func GenChallengeKey(ccntmract common.Address, nodeAddr common.Address, fileHash []byte) []byte {
	prefix := GenChallengePrefix(ccntmract, nodeAddr)
	return append(prefix, fileHash...)
}

func GenResponsePrefix(ccntmract common.Address, fileOwner common.Address) []byte {
	prefix := append(ccntmract[:], cntmFS_RESPONSE...)
	prefix = append(prefix, fileOwner[:]...)
	return prefix
}

func GenResponseKey(ccntmract common.Address, fileOwner common.Address, fileHash []byte) []byte {
	prefix := GenResponsePrefix(ccntmract, fileOwner)
	return append(prefix, fileHash...)
}

func GenFsPdpRecordPrefix(ccntmract common.Address, fileHash []byte, fileOwner common.Address) []byte {
	prefix := append(ccntmract[:], cntmFS_FILE_PDP...)
	prefix = append(prefix, fileHash...)
	prefix = append(prefix, fileOwner[:]...)
	return prefix
}

func GenFsPdpRecordKey(ccntmract common.Address, fileHash []byte, fileOwner common.Address, nodeAddr common.Address) []byte {
	prefix := GenFsPdpRecordPrefix(ccntmract, fileHash, fileOwner)
	return append(prefix, nodeAddr[:]...)
}

func GenFsFileOwnerKey(ccntmract common.Address, fileHash []byte) []byte {
	prefix := append(ccntmract[:], cntmFS_FILE_OWNER...)
	return append(prefix, fileHash...)
}

func GenFsReadPledgeKey(ccntmract common.Address, downloader common.Address, fileHash []byte) []byte {
	key := append(ccntmract[:], cntmFS_FILE_READ_PLEDGE...)
	key = append(key[:], downloader[:]...)
	return append(key, fileHash[:]...)
}

func GenFsSpaceKey(ccntmract common.Address, spaceOwner common.Address) []byte {
	key := append(ccntmract[:], cntmFS_FILE_SPACE...)
	return append(key, spaceOwner[:]...)
}

func appCallTransfer(native *native.NativeService, ccntmract common.Address, from common.Address, to common.Address, amount uint64) error {
	var sts []cntm.State
	sts = append(sts, cntm.State{
		From:  from,
		To:    to,
		Value: amount,
	})
	transfers := cntm.Transfers{
		States: sts,
	}

	sink := common.NewZeroCopySink(nil)
	transfers.Serialization(sink)

	if _, err := native.NativeCall(ccntmract, "transfer", sink.Bytes()); err != nil {
		return fmt.Errorf("appCallTransfer, appCall error: %v", err)
	}
	return nil
}

func DecodeVarBytes(source *common.ZeroCopySource) ([]byte, error) {
	var err error
	buf, _, irregular, eof := source.NextVarBytes()
	if eof {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadVarBytes, ccntmract params deserialize error!")
	}
	if irregular {
		return utils.BYTE_FALSE, common.ErrIrregularData
	}
	return buf, err
}

func DecodeBool(source *common.ZeroCopySource) (bool, error) {
	var err error
	ret, irregular, eof := source.NextBool()
	if eof {
		return false, errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadBool, ccntmract params deserialize error!")
	}
	if irregular {
		return false, common.ErrIrregularData
	}
	return ret, err
}

func CheckOntFsAvailability(service *native.NativeService) error {
	if service.Height < config.GetOntFsHeight() {
		return fmt.Errorf("OntFs ccntmract is not avaliable")
	}
	return nil
}
