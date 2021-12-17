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

package actor

import (
	"errors"
	"github.com/Ontology/common"
	"github.com/Ontology/core/types"
	cntmerr "github.com/Ontology/errors"
	"github.com/cntmio/cntmology-eventbus/actor"
	tcomn "github.com/Ontology/txnpool/common"
	"time"
	"github.com/Ontology/common/log"
)

var txnPid *actor.PID
var txnPoolPid *actor.PID

func SetTxPid(actr *actor.PID) {
	txnPid = actr
}
func SetTxnPoolPid(actr *actor.PID) {
	txnPoolPid = actr
}
func AppendTxToPool(txn *types.Transaction) cntmerr.ErrCode {
	txnPid.Tell(txn)
	return cntmerr.ErrNoError
	/*
	future := txnPid.RequestFuture(txn, ReqTimeout*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ErrActorComm, err)
		return cntmerr.ErrUnknown
	}
	if result, ok := result.(*tcomn.TxRsp); !ok {
		return cntmerr.ErrUnknown
	} else if result.Hash != txn.Hash() {
		return cntmerr.ErrUnknown
	} else {
		return result.ErrCode
	}
	*/
}

func GetTxsFromPool(byCount bool) (map[common.Uint256]*types.Transaction, common.Fixed64) {
	future := txnPoolPid.RequestFuture(&tcomn.GetTxnPoolReq{ByCount: byCount}, ReqTimeout*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ErrActorComm, err)
		return nil, 0
	}
	txpool, ok := result.(*tcomn.GetTxnPoolRsp)
	if !ok {
		return nil, 0
	}
	txMap := make(map[common.Uint256]*types.Transaction)
	var networkFeeSum common.Fixed64
	for _, v := range txpool.TxnPool {
		txMap[v.Tx.Hash()] = v.Tx
		networkFeeSum += v.Fee
	}
	return txMap, networkFeeSum

}

func GetTxFromPool(hash common.Uint256) (tcomn.TXEntry, error) {

	future := txnPid.RequestFuture(&tcomn.GetTxnReq{hash}, ReqTimeout*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ErrActorComm, err)
		return tcomn.TXEntry{}, err
	}
	txn, ok := result.(*tcomn.GetTxnRsp)
	if !ok {
		return tcomn.TXEntry{}, errors.New("fail")
	}
	if txn == nil {
		return tcomn.TXEntry{}, errors.New("fail")
	}

	future = txnPid.RequestFuture(&tcomn.GetTxnStatusReq{hash}, ReqTimeout*time.Second)
	result, err = future.Result()
	if err != nil {
		log.Errorf(ErrActorComm, err)
		return tcomn.TXEntry{}, err
	}
	txStatus, ok := result.(*tcomn.GetTxnStatusRsp)
	if !ok {
		return tcomn.TXEntry{}, errors.New("fail")
	}
	txnEntry := tcomn.TXEntry{txn.Txn, 0, txStatus.TxStatus}
	return txnEntry, nil
}

func GetTxnCnt() ([]uint64, error) {
	future := txnPid.RequestFuture(&tcomn.GetTxnStats{}, ReqTimeout*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ErrActorComm, err)
		return []uint64{}, err
	}
	txnCnt, ok := result.(*tcomn.GetTxnStatsRsp)
	if !ok {
		return []uint64{}, errors.New("fail")
	}
	return txnCnt.Count, nil
}
