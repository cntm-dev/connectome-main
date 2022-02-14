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

package websocket

import (
	"bytes"
	"github.com/cntmio/cntmology/common"
	cfg "github.com/cntmio/cntmology/common/config"
	"github.com/cntmio/cntmology/common/log"
	"github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/events/message"
	bactor "github.com/cntmio/cntmology/http/base/actor"
	bcomn "github.com/cntmio/cntmology/http/base/common"
	Err "github.com/cntmio/cntmology/http/base/error"
	"github.com/cntmio/cntmology/http/base/rest"
	"github.com/cntmio/cntmology/http/websocket/websocket"
	"github.com/cntmio/cntmology/smartccntmract/event"
)

var ws *websocket.WsServer

func StartServer() {
	bactor.SubscribeEvent(message.TOPIC_SAVE_BLOCK_COMPLETE, sendBlock2WSclient)
	bactor.SubscribeEvent(message.TOPIC_SMART_CODE_EVENT, pushSmartCodeEvent)
	go func() {
		ws = websocket.InitWsServer()
		ws.Start()
	}()
}
func sendBlock2WSclient(v interface{}) {
	if cfg.Parameters.HttpWsPort != 0 {
		go func() {
			pushBlock(v)
			pushBlockTransactions(v)
		}()
	}
}
func Stop() {
	if ws == nil {
		return
	}
	ws.Stop()
}
func ReStartServer() {
	if ws == nil {
		ws = websocket.InitWsServer()
		ws.Start()
		return
	}
	ws.Restart()
}

func pushSmartCodeEvent(v interface{}) {
	if ws == nil {
		return
	}
	rs, ok := v.(types.SmartCodeEvent)
	if !ok {
		log.Errorf("[PushSmartCodeEvent]", "SmartCodeEvent err")
		return
	}
	go func() {
		switch object := rs.Result.(type) {
		case []*event.NotifyEventInfo:
			evts := []bcomn.NotifyEventInfo{}
			for _, v := range object {
				txhash := v.TxHash
				evts = append(evts, bcomn.NotifyEventInfo{common.ToHexString(txhash[:]), v.CcntmractAddress.ToHexString(), v.States})
			}
			pushEvent(rs.TxHash, rs.Error, rs.Action, evts)
		case *event.LogEventArgs:
			type logEventArgs struct {
				TxHash          string
				CcntmractAddress string
				Message         string
			}
			hash := object.TxHash
			pushEvent(rs.TxHash, rs.Error, rs.Action, logEventArgs{common.ToHexString(hash[:]), object.CcntmractAddress.ToHexString(), object.Message})
		default:
			pushEvent(rs.TxHash, rs.Error, rs.Action, rs.Result)
		}
	}()
}

func pushEvent(txHash string, errcode int64, action string, result interface{}) {
	if ws != nil {
		resp := rest.ResponsePack(Err.SUCCESS)
		resp["Result"] = result
		resp["Error"] = errcode
		resp["Action"] = action
		resp["Desc"] = Err.ErrMap[resp["Error"].(int64)]
		ws.PushTxResult(txHash, resp)
		ws.BroadcastToSubscribers(websocket.WSTOPIC_EVENT, resp)
	}
}

func pushBlock(v interface{}) {
	if ws == nil {
		return
	}
	resp := rest.ResponsePack(Err.SUCCESS)
	if block, ok := v.(types.Block); ok {
		resp["Action"] = "sendrawblock"
		w := bytes.NewBuffer(nil)
		block.Serialize(w)
		resp["Result"] = common.ToHexString(w.Bytes())
		ws.BroadcastToSubscribers(websocket.WSTOPIC_RAW_BLOCK, resp)

		resp["Action"] = "sendjsonblock"
		resp["Result"] = bcomn.GetBlockInfo(&block)
		ws.BroadcastToSubscribers(websocket.WSTOPIC_JSON_BLOCK, resp)
	}
}
func pushBlockTransactions(v interface{}) {
	if ws == nil {
		return
	}
	resp := rest.ResponsePack(Err.SUCCESS)
	if block, ok := v.(types.Block); ok {
		resp["Result"] = rest.GetBlockTransactions(&block)
		resp["Action"] = "sendblocktxhashs"
		ws.BroadcastToSubscribers(websocket.WSTOPIC_TXHASHS, resp)
	}
}

