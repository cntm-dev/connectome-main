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

// Package proc provides functions for handle messages from
// consensus/ledger/net/http/validators
package proc

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/cntmio/cntmology-eventbus/actor"
	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/log"
	"github.com/cntmio/cntmology/core/ledger"
	ctypes "github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/errors"
	msgpack "github.com/cntmio/cntmology/p2pserver/message/msg_pack"
	p2p "github.com/cntmio/cntmology/p2pserver/net/protocol"
	tc "github.com/cntmio/cntmology/txnpool/common"
	"github.com/cntmio/cntmology/validator/stateful"
	"github.com/cntmio/cntmology/validator/stateless"
	"github.com/cntmio/cntmology/validator/types"
)

type serverPendingTx struct {
	tx             *ctypes.Transaction // Pending
	sender         tc.SenderType       // Indicate which sender tx is from
	ch             chan *tc.TxResult   // channel to send tx result
	checkingStatus *tc.CheckingStatus
}

// TXPoolServer ccntmains all api to external modules
type TXPoolServer struct {
	mu                    sync.RWMutex                        // Sync mutex
	txPool                *tc.TXPool                          // The tx pool that holds the valid transaction
	allPendingTxs         map[common.Uint256]*serverPendingTx // The txs that server is processing
	actor                 *actor.PID
	Net                   p2p.P2P
	slots                 chan struct{} // The limited slots for the new transaction
	height                uint32        // The current block height
	gasPrice              uint64        // Gas price to enforce for acceptance into the pool
	disablePreExec        bool          // Disbale PreExecute a transaction
	disableBroadcastNetTx bool          // Disable broadcast tx from network

	stateless *stateless.ValidatorPool
	stateful  *stateful.ValidatorPool
	rspCh     chan *types.CheckResponse // The channel of verified response
	stopCh    chan bool                 // stop routine
}

// NewTxPoolServer creates a new tx pool server to schedule workers to
// handle and filter inbound transactions from the network, http, and consensus.
func NewTxPoolServer(disablePreExec, disableBroadcastNetTx bool) *TXPoolServer {
	s := &TXPoolServer{}
	// Initial txnPool
	s.txPool = tc.NewTxPool()
	s.allPendingTxs = make(map[common.Uint256]*serverPendingTx)

	s.slots = make(chan struct{}, tc.MAX_LIMITATION)
	for i := 0; i < tc.MAX_LIMITATION; i++ {
		s.slots <- struct{}{}
	}

	s.gasPrice = getGasPriceConfig()
	log.Infof("tx pool: the current local gas price is %d", s.gasPrice)

	s.disablePreExec = disablePreExec
	s.disableBroadcastNetTx = disableBroadcastNetTx
	// Create the given concurrent workers
	s.wg.Add(1)
	s.worker = NewTxPoolWoker(s)
	go s.worker.start()
}

// checkPendingBlockOk checks whether a block from consensus is verified.
// If some transaction is invalid, return the result directly at once, no
// need to wait for verifying the complete block.
func (s *TXPoolServer) checkPendingBlockOk(hash common.Uint256,
	err errors.ErrCode) {

	// Check if the tx is in pending block, if yes, move it to
	// the verified tx list
	s.pendingBlock.mu.Lock()
	defer s.pendingBlock.mu.Unlock()

	tx, ok := s.pendingBlock.unProcessedTxs[hash]
	if !ok {
		return
	}

	// Todo:
	entry := &tc.VerifyTxResult{
		Height:  s.pendingBlock.height,
		Tx:      tx,
		ErrCode: err,
	}

	s.pendingBlock.processedTxs[hash] = entry
	delete(s.pendingBlock.unProcessedTxs, hash)

	// if the tx is invalid, send the response at once
	if err != errors.ErrNoError || len(s.pendingBlock.unProcessedTxs) == 0 {
		s.sendBlkResult2Consensus()
	}
}

// getPendingListSize return the length of the pending tx list.
func (s *TXPoolServer) getPendingListSize() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.allPendingTxs)
}

func (s *TXPoolServer) getHeight() uint32 {
	return atomic.LoadUint32(&s.height)
}

func (s *TXPoolServer) setHeight(height uint32) {
	if height == 0 {
		return
	}
	atomic.StoreUint32(&s.height, height)
}

// getGasPrice returns the current gas price enforced by the transaction pool
func (s *TXPoolServer) getGasPrice() uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.gasPrice
}

func (s *TXPoolServer) GetPendingTx(hash common.Uint256) *serverPendingTx {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.allPendingTxs[hash]
}

// removePendingTx removes a transaction from the pending list
// when it is handled. And if the submitter of the valid transaction
// is from http, broadcast it to the network. Meanwhile, check if it
// is in the block from consensus.
func (s *TXPoolServer) removePendingTx(hash common.Uint256, err errors.ErrCode) {
	s.mu.Lock()

	pt, ok := s.allPendingTxs[hash]
	if !ok {
		s.mu.Unlock()
		return
	}

	if err == errors.ErrNoError && ((pt.sender == tc.HttpSender) ||
		(pt.sender == tc.NetSender && !s.disableBroadcastNetTx)) {
		if s.Net != nil {
			msg := msgpack.NewTxn(pt.tx)
			go s.Net.Broadcast(msg)
		}
	}

	replyTxResult(pt.ch, hash, err, err.Error())

	delete(s.allPendingTxs, hash)

	if len(s.allPendingTxs) < tc.MAX_LIMITATION {
		select {
		case s.slots <- struct{}{}:
		default:
			log.Debug("removePendingTx: slots is full")
		}
	}

	s.mu.Unlock()

	// Check if the tx is in the pending block and
	// the pending block is verified
	s.checkPendingBlockOk(hash, err)
}

// setPendingTx adds a transaction to the pending list, if the
// transaction is already in the pending list, just return false.
func (s *TXPoolServer) setPendingTx(tx *tx.Transaction, sender tc.SenderType, txResultCh chan *tc.TxResult) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if ok := s.allPendingTxs[tx.Hash()]; ok != nil {
		log.Debugf("setPendingTx: transaction %x already in the verifying process",
			tx.Hash())
		return false
	}

	pt := &serverPendingTx{
		tx:     tx,
		sender: sender,
		ch:     txResultCh,
	}

	s.allPendingTxs[tx.Hash()] = pt
	return true
}

// assignTxToWorker assigns a new transaction to a worker by LB
func (s *TXPoolServer) assignTxToWorker(tx *tx.Transaction,
	sender tc.SenderType, txResultCh chan *tc.TxResult) bool {

	if tx == nil {
		return false
	}

	if ok := s.setPendingTx(tx, sender, txResultCh); !ok {
		s.increaseStats(tc.DuplicateStats)
		if sender == tc.HttpSender && txResultCh != nil {
			replyTxResult(txResultCh, tx.Hash(), errors.ErrDuplicateInput,
				"duplicated transaction input detected")
		}
		return false
	}
	// Add the rcvTxn to the worker
	lb := make(tc.LBSlice, len(s.workers))
	for i := 0; i < len(s.workers); i++ {
		entry := tc.LB{Size: len(s.workers[i].pendingTxList),
			WorkerID: uint8(i),
		}
		lb[i] = entry
	}
	sort.Sort(lb)
	s.workers[lb[0].WorkerID].rcvTXCh <- tx
	return true
}

// assignRspToWorker assigns a check response from the validator to
// the correct worker.
func (s *TXPoolServer) assignRspToWorker(rsp *types.CheckResponse) bool {

	if rsp == nil {
		return false
	}

	if rsp.WorkerId >= 0 && rsp.WorkerId < uint8(len(s.workers)) {
		s.workers[rsp.WorkerId].rspCh <- rsp
	}

	if rsp.ErrCode == errors.ErrNoError {
		s.increaseStats(tc.SuccessStats)
	} else {
		s.increaseStats(tc.FailureStats)
		if rsp.Type == types.Stateless {
			s.increaseStats(tc.SigErrStats)
		} else {
			s.increaseStats(tc.StateErrStats)
		}
	}
	return true
}

// GetPID returns an actor pid with the actor type, If the type
// doesn't exist, return nil.
func (s *TXPoolServer) GetPID(actor tc.ActorType) *actor.PID {
	if actor < tc.TxActor || actor >= tc.MaxActor {
		return nil
	}

	return s.actors[actor]
}

// RegisterActor registers an actor with the actor type and pid.
func (s *TXPoolServer) RegisterActor(actor tc.ActorType, pid *actor.PID) {
	s.actors[actor] = pid
}

// UnRegisterActor cancels the actor with the actor type.
func (s *TXPoolServer) UnRegisterActor(actor tc.ActorType) {
	delete(s.actors, actor)
}

// Stop stops server and workers.
func (s *TXPoolServer) Stop() {
	for _, v := range s.actors {
		v.Stop()
	}
	//Stop worker
	s.worker.stop()
	s.wg.Wait()

	if s.slots != nil {
		close(s.slots)
	}
}

// getTransaction returns a transaction with the transaction hash.
func (s *TXPoolServer) getTransaction(hash common.Uint256) *tx.Transaction {
	return s.txPool.GetTransaction(hash)
}

// getTxPool returns a tx list for consensus.
func (s *TXPoolServer) getTxPool(byCount bool, height uint32) []*tc.TXEntry {
	s.setHeight(height)

	avlTxList, oldTxList := s.txPool.GetTxPool(byCount, height)

	for _, t := range oldTxList {
		s.delTransaction(t)
		s.reVerifyStateful(t, tc.NilSender)
	}

	return avlTxList
}

// getTxCount returns current tx count, including pending and verified
func (s *TXPoolServer) getTxCount() []uint32 {
	ret := make([]uint32, 0)
	ret = append(ret, uint32(s.txPool.GetTransactionCount()))
	ret = append(ret, uint32(s.getPendingListSize()))
	return ret
}

// getTxHashList returns a currently pending tx hash list
func (s *TXPoolServer) getTxHashList() []common.Uint256 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	txHashPool := s.txPool.GetTransactionHashList()
	ret := make([]common.Uint256, 0, len(s.allPendingTxs)+len(txHashPool))
	existedTxHash := make(map[common.Uint256]bool)
	for _, hash := range txHashPool {
		ret = append(ret, hash)
		existedTxHash[hash] = true
	}
	for _, v := range s.allPendingTxs {
		hash := v.tx.Hash()
		if !existedTxHash[hash] {
			ret = append(ret, hash)
			existedTxHash[hash] = true
		}
	}
	return ret
}

// cleanTransactionList cleans the txs in the block from the ledger
func (s *TXPoolServer) cleanTransactionList(txs []*ctypes.Transaction, height uint32) {
	s.txPool.CleanTransactionList(txs)

	// Check whether to update the gas price and remove txs below the threshold
	if height%tc.UPDATE_FREQUENCY == 0 {
		gasPrice := getGasPriceConfig()
		s.mu.Lock()
		oldGasPrice := s.gasPrice
		s.gasPrice = gasPrice
		s.mu.Unlock()
		if oldGasPrice != gasPrice {
			log.Infof("Transaction pool price threshold updated from %d to %d",
				oldGasPrice, gasPrice)
		}

		if oldGasPrice < gasPrice {
			s.txPool.RemoveTxsBelowGasPrice(gasPrice)
		}
	}
	// Cleanup tx pool
	if !s.disablePreExec {
		remain := s.txPool.Remain()
		for _, t := range remain {
			if ok, _ := preExecCheck(t); !ok {
				log.Debugf("cleanTransactionList: preExecCheck tx %x failed", t.Hash())
				ccntminue
			}
			s.reVerifyStateful(t, tc.NilSender)
		}
	}
}

// delTransaction deletes a transaction in the tx pool.
func (s *TXPoolServer) delTransaction(t *ctypes.Transaction) {
	s.txPool.DelTxList(t)
}

// adds a valid transaction to the tx pool.
func (s *TXPoolServer) addTxList(txEntry *tc.VerifiedTx) bool {
	ret := s.txPool.AddTxList(txEntry)
	return ret
}

// getTxStatusReq returns a transaction's status with the transaction hash.
func (s *TXPoolServer) getTxStatusReq(hash common.Uint256) *tc.TxStatus {
	if ret := s.GetPendingTx(hash); ret != nil {
		return &tc.TxStatus{
			Hash:  hash,
			Attrs: ret.checkingStatus.GetTxAttr(),
		}
	}

	return s.txPool.GetTxStatus(hash)
}

// getTransactionCount returns the tx size of the transaction pool.
func (s *TXPoolServer) getTransactionCount() int {
	return s.txPool.GetTransactionCount()
}

// reVerifyStateful re-verify a transaction's stateful data.
func (s *TXPoolServer) reVerifyStateful(tx *tx.Transaction, sender tc.SenderType) {
	if ok := s.setPendingTx(tx, sender, nil); !ok {
		s.increaseStats(tc.DuplicateStats)
		return
	}

	// Add the rcvTxn to the worker
	lb := make(tc.LBSlice, len(s.workers))
	for i := 0; i < len(s.workers); i++ {
		entry := tc.LB{Size: len(s.workers[i].pendingTxList),
			WorkerID: uint8(i),
		}
		lb[i] = entry
	}

	sort.Sort(lb)
	s.workers[lb[0].WorkerID].stfTxCh <- tx
}

// sendBlkResult2Consensus sends the result of verifying block to  consensus
func (s *TXPoolServer) sendBlkResult2Consensus() {
	rsp := &tc.VerifyBlockRsp{
		TxnPool: make([]*tc.VerifyTxResult,
			0, len(s.pendingBlock.processedTxs)),
	}
	for _, v := range s.pendingBlock.processedTxs {
		rsp.TxnPool = append(rsp.TxnPool, v)
	}

	if s.pendingBlock.sender != nil {
		s.pendingBlock.sender.Tell(rsp)
	}

	// Clear the processedTxs for the next block verify req
	for k := range s.pendingBlock.processedTxs {
		delete(s.pendingBlock.processedTxs, k)
	}
}

// verifyBlock verifies the block from consensus.
// There are three cases to handle.
// 1, for those unverified txs, assign them to the available worker;
// 2, for those verified txs whose height >= block's height, nothing to do;
// 3, for those verified txs whose height < block's height, re-verify their
// stateful data.
func (s *TXPoolServer) verifyBlock(req *tc.VerifyBlockReq, sender *actor.PID) {
	if req == nil || len(req.Txs) == 0 {
		return
	}

	s.setHeight(req.Height)

	processedTxs := make([]*tc.VerifyTxResult, len(req.Txs))

	// Check whether a tx's gas price is lower than the required, if yes, just return error
	txs := make(map[common.Uint256]*ctypes.Transaction, len(req.Txs))
	for _, t := range req.Txs {
		if t.GasPrice < s.gasPrice {
			entry := &tc.VerifyTxResult{
				Height:  req.Height,
				Tx:      t,
				ErrCode: errors.ErrGasPrice,
			}
			processedTxs = append(processedTxs, entry)
			sender.Tell(&tc.VerifyBlockRsp{TxnPool: processedTxs})
			return
		}
		// Check whether double spent
		if _, ok := txs[t.Hash()]; ok {
			entry := &tc.VerifyTxResult{
				Height:  req.Height,
				Tx:      t,
				ErrCode: errors.ErrDoubleSpend,
			}
			processedTxs = append(processedTxs, entry)
			sender.Tell(&tc.VerifyBlockRsp{TxnPool: processedTxs})
			return
		}
		txs[t.Hash()] = t
	}

	checkBlkResult := s.txPool.GetUnverifiedTxs(req.Txs, req.Height)

	if len(checkBlkResult.UnverifiedTxs) > 0 {
		ch := make(chan *types.CheckResponse, len(checkBlkResult.UnverifiedTxs))
		validator := stateless.NewValidatorPool(5)
		for _, t := range checkBlkResult.UnverifiedTxs {
			validator.SubmitVerifyTask(t, ch)
		}
		for i := 0; i < len(checkBlkResult.UnverifiedTxs); i++ {
			response := <-ch
			if response.ErrCode != errors.ErrNoError {
				processedTxs = append(processedTxs, &tc.VerifyTxResult{
					Height:  req.Height,
					Tx:      txs[response.Hash],
					ErrCode: response.ErrCode,
				})
				sender.Tell(&tc.VerifyBlockRsp{TxnPool: processedTxs})
				return
			}
		}
	}

	lenStateFul := len(checkBlkResult.UnverifiedTxs) + len(checkBlkResult.OldTxs)
	if lenStateFul > 0 {
		currHeight := ledger.DefLedger.GetCurrentBlockHeight()
		for currHeight < req.Height {
			// wait ledger sync up
			log.Warnf("ledger need sync up for tx verification, curr height: %d, expected:%d", currHeight, req.Height)
			time.Sleep(time.Second)
			currHeight = ledger.DefLedger.GetCurrentBlockHeight()
		}

		ch := make(chan *types.CheckResponse, lenStateFul)
		validator := stateful.NewValidatorPool(1)
		for _, tx := range checkBlkResult.UnverifiedTxs {
			validator.SubmitVerifyTask(tx, ch)
		}
		for _, tx := range checkBlkResult.OldTxs {
			validator.SubmitVerifyTask(tx, ch)
		}
		for i := 0; i < lenStateFul; i++ {
			resp := <-ch
			processedTxs = append(processedTxs, &tc.VerifyTxResult{
				Height:  resp.Height,
				Tx:      txs[resp.Hash],
				ErrCode: resp.ErrCode,
			})
			if resp.ErrCode != errors.ErrNoError {
				sender.Tell(&tc.VerifyBlockRsp{TxnPool: processedTxs})
				return
			}
		}
	}

	processedTxs = append(processedTxs, checkBlkResult.VerifiedTxs...)
	sender.Tell(&tc.VerifyBlockRsp{TxnPool: processedTxs})
}

// handles the verified response from the validator and if
// the tx is valid, add it to the tx pool, or remove it from the pending
// list
func (server *TXPoolServer) handleRsp(rsp *types.CheckResponse) {
	pt := server.GetPendingTx(rsp.Hash)
	if pt == nil {
		return
	}

	if rsp.ErrCode != errors.ErrNoError {
		//Verify fail
		log.Debugf("handleRsp: validator %d transaction %x invalid: %s", rsp.Type, rsp.Hash, rsp.ErrCode.Error())
		server.removePendingTx(rsp.Hash, rsp.ErrCode)
		return
	}

	if rsp.Type == types.Stateful && rsp.Height < server.getHeight() {
		// If validator's height is less than the required one, re-validate it.
		server.stateful.SubmitVerifyTask(rsp.Tx, server.rspCh)
		return
	}

	switch rsp.Type {
	case types.Stateful:
		pt.checkingStatus.SetStateful(rsp.Height)
	case types.Stateless:
		pt.checkingStatus.SetStateless()
	}

	if pt.checkingStatus.GetStateless() && pt.checkingStatus.GetStateful() {
		txEntry := &tc.VerifiedTx{
			Tx:             pt.tx,
			VerifiedHeight: pt.checkingStatus.CheckHeight,
		}
		server.addTxList(txEntry)
		server.removePendingTx(pt.tx.Hash(), errors.ErrNoError)
	}
}
