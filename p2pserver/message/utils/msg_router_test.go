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

package utils

import (
	"github.com/cntmio/cntmology-crypto/keypair"
	"github.com/cntmio/cntmology/common/log"
	msgCommon "github.com/cntmio/cntmology/p2pserver/common"
	"github.com/cntmio/cntmology/p2pserver/net/netserver"
	"github.com/stretchr/testify/assert"
	"testing"
)

func testHandler(data *msgCommon.MsgPayload, args ...interface{}) error {
	log.Info("Test handler")
	return nil
}

// TestMsgRouter tests a basic function of a message router
func TestMsgRouter(t *testing.T) {
	_, pub, _ := keypair.GenerateKeyPair(keypair.PK_ECDSA, keypair.P256)
	network := netserver.NewNetServer(pub)
	msgRouter := NewMsgRouter(network)
	assert.NotNil(t, msgRouter)

	msgRouter.RegisterMsgHandler("test", testHandler)
	msgRouter.UnRegisterMsgHandler("test")
	msgRouter.Start()
	msgRouter.Stop()
}
