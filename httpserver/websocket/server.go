/*
 * Copyright (C) 2018 The cntm Authors
 * This file is part of The cntm library.
 *
 * The cntm is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The cntm is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The cntm.  If not, see <http://www.gnu.org/licenses/>.
 */

// Package websocket privides a function to start websocket server
package websocket

import (
	"github.com/conntectome/cntm/common"
	cfg "github.com/conntectome/cntm/common/config"
	"github.com/conntectome/cntm/common/log"
	"github.com/conntectome/cntm/core/types"
	"github.com/conntectome/cntm/events/message"
	bactor "github.com/conntectome/cntm/http/base/actor"
	bcomn "github.com/conntectome/cntm/http/base/common"
	Err "github.com/conntectome/cntm/http/base/error"
	"github.com/conntectome/cntm/http/base/rest"
	"github.com/conntectome/cntm/http/websocket/websocket"
	"github.com/conntectome/cntm/smartcontract/event"
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
	if cfg.DefConfig.Ws.HttpWsPort != 0 {
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
		case *event.LogEventArgs:
			contractAddrs, evts := bcomn.GetLogEvent(object)
			pushEvent(contractAddrs, rs.TxHash.ToHexString(), rs.Error, rs.Action, evts)
		case *event.ExecuteNotify:
			contractAddrs, notify := bcomn.GetExecuteNotify(object)
			pushEvent(contractAddrs, rs.TxHash.ToHexString(), rs.Error, rs.Action, notify)
		default:
		}
	}()
}

func pushEvent(contractAddrs map[string]bool, txHash string, errcode int64, action string, result interface{}) {
	if ws != nil {
		resp := rest.ResponsePack(Err.SUCCESS)
		resp["Result"] = result
		resp["Error"] = errcode
		resp["Action"] = action
		resp["Desc"] = Err.ErrMap[resp["Error"].(int64)]
		ws.PushTxResult(contractAddrs, txHash, resp)
		ws.BroadcastToSubscribers(contractAddrs, websocket.WSTOPIC_EVENT, resp)
	}
}

func pushBlock(v interface{}) {
	if ws == nil {
		return
	}
	resp := rest.ResponsePack(Err.SUCCESS)
	if block, ok := v.(types.Block); ok {
		resp["Action"] = "sendrawblock"
		resp["Result"] = common.ToHexString(block.ToArray())
		ws.BroadcastToSubscribers(nil, websocket.WSTOPIC_RAW_BLOCK, resp)

		resp["Action"] = "sendjsonblock"
		resp["Result"] = bcomn.GetBlockInfo(&block)
		ws.BroadcastToSubscribers(nil, websocket.WSTOPIC_JSON_BLOCK, resp)
	}
}
func pushBlockTransactions(v interface{}) {
	if ws == nil {
		return
	}
	resp := rest.ResponsePack(Err.SUCCESS)
	if block, ok := v.(types.Block); ok {
		resp["Result"] = bcomn.GetBlockTransactions(&block)
		resp["Action"] = "sendblocktxhashs"
		ws.BroadcastToSubscribers(nil, websocket.WSTOPIC_TXHASHS, resp)
	}
}
