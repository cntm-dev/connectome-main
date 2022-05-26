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

package native

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"sort"

	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/core/genesis"
	cstates "github.com/cntmio/cntmology/core/states"
	scommon "github.com/cntmio/cntmology/core/store/common"
	"github.com/cntmio/cntmology/errors"
	"github.com/cntmio/cntmology/smartccntmract/service/native/states"
)

const (
	EXECUTE_SPLIT = "executeSplit"
	a             = 0.75
	b             = 0.2
	c             = 0.05
	TOTAL_cntm     = 10000000000
)

func init() {
	Ccntmracts[genesis.FeeSplitCcntmractAddress] = RegisterFeeSplitCcntmract
}

func RegisterFeeSplitCcntmract(native *NativeService) {
	native.Register(EXECUTE_SPLIT, ExecuteSplit)
}

func ExecuteSplit(native *NativeService) error {
	ccntmract := genesis.GovernanceCcntmractAddress
	//get current view
	cView, err := getGovernanceView(native, ccntmract)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Get view error!")
	}
	view := new(big.Int).Sub(cView, new(big.Int).SetInt64(1))

	//get all peerPool
	stateValues, err := native.CloneCache.Store.Find(scommon.ST_STORAGE, concatKey(ccntmract, []byte(PEER_POOL), view.Bytes()))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Get all peerPool error!")
	}
	peersCandidate := []*states.CandidateSplitInfo{}
	peersSyncNode := []*states.SyncNodeSplitInfo{}
	peerPool := new(states.PeerPool)
	var syncNodePosSum uint64
	for _, v := range stateValues {
		peerPoolStore, _ := v.Value.(*cstates.StorageItem)
		err = json.Unmarshal(peerPoolStore.Value, peerPool)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Unmarshal peerPool error!")
		}
		if peerPool.Status == CandidateStatus || peerPool.Status == ConsensusStatus {
			stake := peerPool.TotalPos + peerPool.InitPos
			peersCandidate = append(peersCandidate, &states.CandidateSplitInfo{
				PeerPubkey: peerPool.PeerPubkey,
				InitPos:    peerPool.InitPos,
				Address:    peerPool.Address,
				Stake:      float64(stake),
			})
		}
		if peerPool.Status == SyncNodeStatus || peerPool.Status == RegisterCandidateStatus {
			syncNodePosSum += peerPool.InitPos
			peersSyncNode = append(peersSyncNode, &states.SyncNodeSplitInfo{
				PeerPubkey: peerPool.PeerPubkey,
				InitPos:    peerPool.InitPos,
				Address:    peerPool.Address,
			})
		}
	}

	//fee split of syncNode peer
	fmt.Println("###############################################################")
	var splitSyncNodeAmount uint64
	for _, v := range stateValues {
		peerPoolStore, _ := v.Value.(*cstates.StorageItem)
		err = json.Unmarshal(peerPoolStore.Value, peerPool)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Unmarshal peerPool error!")
		}
		if peerPool.Status == SyncNodeStatus || peerPool.Status == RegisterCandidateStatus {
			amount := TOTAL_cntm * c * peerPool.InitPos / syncNodePosSum
			addressBytes, err := hex.DecodeString(peerPool.Address)
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Address format error!")
			}
			address, err := common.AddressParseFromBytes(addressBytes)
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Address format error!")
			}
			err = appCallApproveOng(native, genesis.FeeSplitCcntmractAddress, address, new(big.Int).SetUint64(amount))
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Ong transfer error!")
			}
			fmt.Printf("Amount of address %v is: %d \n", peerPool.Address, amount)
			splitSyncNodeAmount += amount
		}
	}
	remainSyncNodeAmount := TOTAL_cntm*c - splitSyncNodeAmount
	fmt.Println("RemainSyncNodeAmount is : ", remainSyncNodeAmount)

	// sort peers by peerPubkey
	sort.Slice(peersSyncNode, func(i, j int) bool {
		return peersSyncNode[i].PeerPubkey > peersSyncNode[j].PeerPubkey
	})
	//TODO: how if initPos is the same
	addressBytes, err := hex.DecodeString(peersSyncNode[0].Address)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Address format error!")
	}
	address, err := common.AddressParseFromBytes(addressBytes)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Address format error!")
	}
	err = appCallApproveOng(native, genesis.FeeSplitCcntmractAddress, address, new(big.Int).SetUint64(remainSyncNodeAmount))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Ong transfer error!")
	}
	fmt.Printf("Amount of address %v is: %d \n", peersSyncNode[0].Address, remainSyncNodeAmount)

	// get config
	config := new(states.Configuration)
	configBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, concatKey(ccntmract, []byte(VBFT_CONFIG)))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Get configBytes error!")
	}
	if configBytes == nil {
		return errors.NewErr("[executeSplit] ConfigBytes is nil!")
	}
	configStore, _ := configBytes.(*cstates.StorageItem)
	err = json.Unmarshal(configStore.Value, config)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Unmarshal config error!")
	}

	// sort peers by stake
	sort.Slice(peersCandidate, func(i, j int) bool {
		return peersCandidate[i].Stake > peersCandidate[j].Stake
	})

	// cal s of each consensus node
	var sum float64
	for i := 0; i < int(config.K); i++ {
		sum += peersCandidate[i].Stake
	}
	avg := sum / float64(config.K)
	var sumS float64
	for i := 0; i < int(config.K); i++ {
		peersCandidate[i].S = (0.5 * peersCandidate[i].Stake) / (2 * avg)
		sumS += peersCandidate[i].S
	}

	//fee split of consensus peer
	fmt.Println("###############################################################")
	var splitAmount uint64
	remainCandidate := peersCandidate[0]
	for i := int(config.K) - 1; i >= 0; i-- {
		if peersCandidate[i].PeerPubkey > remainCandidate.PeerPubkey {
			remainCandidate = peersCandidate[i]
		}

		nodeAmount := TOTAL_cntm * a * peersCandidate[i].S / sumS
		fmt.Printf("Amount of node %v is %v: \n", i, nodeAmount)
		peerPubkeyPrefix, err := hex.DecodeString(peersCandidate[i].PeerPubkey)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] PeerPubkey format error!")
		}
		stateValues, err = native.CloneCache.Store.Find(scommon.ST_STORAGE, concatKey(ccntmract, []byte(VOTE_INFO_POOL),
			view.Bytes(), peerPubkeyPrefix))
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Get all peerPool error!")
		}

		//init pos cntm transfer
		initAmount := uint64(nodeAmount * float64(peersCandidate[i].InitPos) / peersCandidate[i].Stake)
		initAddressBytes, err := hex.DecodeString(peersCandidate[i].Address)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Address format error!")
		}
		initAddress, err := common.AddressParseFromBytes(initAddressBytes)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Address format error!")
		}
		err = appCallApproveOng(native, genesis.FeeSplitCcntmractAddress, initAddress, new(big.Int).SetUint64(initAmount))
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Ong transfer error!")
		}
		fmt.Printf("Amount of address %v is: %d \n", peersCandidate[i].Address, initAmount)
		splitAmount += initAmount

		//vote pos cntm transfer
		voteInfoPool := new(states.VoteInfoPool)
		for _, v := range stateValues {
			voteInfoPoolStore, _ := v.Value.(*cstates.StorageItem)
			err = json.Unmarshal(voteInfoPoolStore.Value, voteInfoPool)
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Unmarshal voteInfoPool error!")
			}
			addressBytes, err := hex.DecodeString(voteInfoPool.Address)
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Address format error!")
			}
			address, err := common.AddressParseFromBytes(addressBytes)
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Address format error!")
			}
			pos := voteInfoPool.PrePos + voteInfoPool.PreFreezePos + voteInfoPool.FreezePos + voteInfoPool.NewPos
			amount := uint64(nodeAmount * float64(pos) / peersCandidate[i].Stake)

			//cntm transfer
			err = appCallApproveOng(native, genesis.FeeSplitCcntmractAddress, address, new(big.Int).SetUint64(amount))
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Ong transfer error!")
			}
			fmt.Printf("Amount of address %v is: %d \n", voteInfoPool.Address, amount)
			splitAmount += amount
		}
	}
	//split remained amount
	remainAmount := TOTAL_cntm*a - splitAmount
	fmt.Println("Remained Amount is : ", remainAmount)
	remainAddressBytes, err := hex.DecodeString(remainCandidate.Address)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Address format error!")
	}
	remainAddress, err := common.AddressParseFromBytes(remainAddressBytes)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Address format error!")
	}
	err = appCallApproveOng(native, genesis.FeeSplitCcntmractAddress, remainAddress, new(big.Int).SetUint64(remainAmount))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Ong transfer error!")
	}
	fmt.Printf("Amount of address %v is: %d \n", remainCandidate.Address, remainAmount)

	//fee split of candidate peer
	fmt.Println("###############################################################")
	// cal s of each candidate node
	sum = 0
	for i := int(config.K); i < len(peersCandidate); i++ {
		sum += peersCandidate[i].Stake
	}
	avg = sum / float64(config.K)
	sumS = 0
	for i := int(config.K); i < len(peersCandidate); i++ {
		peersCandidate[i].S = (0.5 * peersCandidate[i].Stake) / (2 * avg)
		sumS += peersCandidate[i].S
	}
	splitAmount = 0
	remainCandidate = peersCandidate[int(config.K)]
	for i := int(config.K); i < len(peersCandidate); i++ {
		if peersCandidate[i].PeerPubkey > remainCandidate.PeerPubkey {
			remainCandidate = peersCandidate[i]
		}

		nodeAmount := TOTAL_cntm * b * peersCandidate[i].S / sumS
		fmt.Printf("Amount of node %v is %v: \n", i, nodeAmount)
		peerPubkeyPrefix, err := hex.DecodeString(peersCandidate[i].PeerPubkey)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] PeerPubkey format error!")
		}
		stateValues, err = native.CloneCache.Store.Find(scommon.ST_STORAGE, concatKey(ccntmract, []byte(VOTE_INFO_POOL),
			view.Bytes(), peerPubkeyPrefix))
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Get all peerPool error!")
		}

		//init pos cntm transfer
		initAmount := uint64(nodeAmount * float64(peersCandidate[i].InitPos) / peersCandidate[i].Stake)
		initAddressBytes, err := hex.DecodeString(peersCandidate[i].Address)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Address format error!")
		}
		initAddress, err := common.AddressParseFromBytes(initAddressBytes)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Address format error!")
		}
		err = appCallApproveOng(native, genesis.FeeSplitCcntmractAddress, initAddress, new(big.Int).SetUint64(initAmount))
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Ong transfer error!")
		}
		fmt.Printf("Amount of address %v is: %d \n", peersCandidate[i].Address, initAmount)
		splitAmount += initAmount

		//vote pos cntm transfer
		voteInfoPool := new(states.VoteInfoPool)
		for _, v := range stateValues {
			voteInfoPoolStore, _ := v.Value.(*cstates.StorageItem)
			err = json.Unmarshal(voteInfoPoolStore.Value, voteInfoPool)
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Unmarshal voteInfoPool error!")
			}
			addressBytes, err := hex.DecodeString(voteInfoPool.Address)
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Address format error!")
			}
			address, err := common.AddressParseFromBytes(addressBytes)
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Address format error!")
			}
			pos := voteInfoPool.PrePos + voteInfoPool.PreFreezePos + voteInfoPool.FreezePos + voteInfoPool.NewPos
			amount := uint64(nodeAmount * float64(pos) / peersCandidate[i].Stake)

			//cntm transfer
			err = appCallApproveOng(native, genesis.FeeSplitCcntmractAddress, address, new(big.Int).SetUint64(amount))
			if err != nil {
				return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Ong transfer error!")
			}
			fmt.Printf("Amount of address %v is: %d \n", voteInfoPool.Address, amount)
			splitAmount += amount
		}
	}
	//split remained amount
	remainAmount = TOTAL_cntm*b - splitAmount
	fmt.Println("Remained Amount is : ", remainAmount)
	remainAddressBytes, err = hex.DecodeString(remainCandidate.Address)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Address format error!")
	}
	remainAddress, err = common.AddressParseFromBytes(remainAddressBytes)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Address format error!")
	}
	err = appCallApproveOng(native, genesis.FeeSplitCcntmractAddress, remainAddress, new(big.Int).SetUint64(remainAmount))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[executeSplit] Ong transfer error!")
	}
	fmt.Printf("Amount of address %v is: %d \n", remainCandidate.Address, remainAmount)

	addCommonEvent(native, genesis.FeeSplitCcntmractAddress, EXECUTE_SPLIT, true)

	return nil
}