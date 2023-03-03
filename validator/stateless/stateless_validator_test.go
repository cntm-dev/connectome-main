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
package stateless

import (
	"testing"
	"time"

	"github.com/conntectome/cntm-crypto/keypair"
	"github.com/conntectome/cntm-eventbus/actor"
	"github.com/conntectome/cntm/account"
	"github.com/conntectome/cntm/common/log"
	"github.com/conntectome/cntm/core/payload"
	"github.com/conntectome/cntm/core/signature"
	ctypes "github.com/conntectome/cntm/core/types"
	"github.com/conntectome/cntm/core/utils"
	"github.com/conntectome/cntm/errors"
	types2 "github.com/conntectome/cntm/validator/types"
	"github.com/stretchr/testify/assert"
)

func signTransaction(signer *account.Account, tx *ctypes.MutableTransaction) error {
	hash := tx.Hash()
	sign, _ := signature.Sign(signer, hash[:])
	tx.Sigs = append(tx.Sigs, ctypes.Sig{
		PubKeys: []keypair.PublicKey{signer.PublicKey},
		M:       1,
		SigData: [][]byte{sign},
	})
	return nil
}

func TestStatelessValidator(t *testing.T) {
	log.InitLog(log.InfoLog, log.Stdout)
	acc := account.NewAccount("")

	code := []byte{1, 2, 3}

	mutable, err := utils.NewDeployTransaction(code, "test", "1", "author", "author@123.com", "test desp", payload.CNTMVM_TYPE)
	assert.Nil(t, err)
	mutable.Payer = acc.Address

	signTransaction(acc, mutable)

	tx, err := mutable.IntoImmutable()
	assert.Nil(t, err)

	validator := &validator{id: "test"}
	props := actor.FromProducer(func() actor.Actor {
		return validator
	})

	pid, err := actor.SpawnNamed(props, validator.id)
	assert.Nil(t, err)

	msg := &types2.CheckTx{WorkerId: 1, Tx: tx}
	fut := pid.RequestFuture(msg, time.Second)

	res, err := fut.Result()
	assert.Nil(t, err)

	result := res.(*types2.CheckResponse)
	assert.Equal(t, result.ErrCode, errors.ErrNoError)
	assert.Equal(t, mutable.Hash(), result.Hash)
}
