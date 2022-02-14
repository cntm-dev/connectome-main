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
	"bytes"
	"fmt"

	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/core/genesis"
	scommon "github.com/cntmio/cntmology/core/store/common"
	"github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/errors"
	"github.com/cntmio/cntmology/smartccntmract/ccntmext"
	"github.com/cntmio/cntmology/smartccntmract/event"
	"github.com/cntmio/cntmology/smartccntmract/storage"
	sstates "github.com/cntmio/cntmology/smartccntmract/states"
)

type (
	Handler         func(native *NativeService) error
	RegisterService func(native *NativeService)
)

var (
	Ccntmracts = map[common.Address]RegisterService{
		genesis.OntCcntmractAddress: RegisterOntCcntmract,
		genesis.OngCcntmractAddress: RegisterOngCcntmract,
	}
)

// Native service struct
// Invoke a native smart ccntmract, new a native service
type NativeService struct {
	CloneCache    *storage.CloneCache
	ServiceMap    map[string]Handler
	Notifications []*event.NotifyEventInfo
	Input         []byte
	Tx            *types.Transaction
	Height        uint32
	CcntmextRef    ccntmext.CcntmextRef
}

// New native service
func NewNativeService(dbCache scommon.StateStore, height uint32, tx *types.Transaction, ctxRef ccntmext.CcntmextRef) *NativeService {
	var nativeService NativeService
	nativeService.CloneCache = storage.NewCloneCache(dbCache)
	nativeService.Tx = tx
	nativeService.Height = height
	nativeService.CcntmextRef = ctxRef
	nativeService.ServiceMap = make(map[string]Handler)
	return &nativeService
}

func (this *NativeService) Register(methodName string, handler Handler) {
	this.ServiceMap[methodName] = handler
}

func (this *NativeService) Invoke() error {
	ctx := this.CcntmextRef.CurrentCcntmext()
	if ctx == nil {
		return errors.NewErr("[Invoke] Native service current ccntmext doesn't exist!")
	}
	bf := bytes.NewBuffer(ctx.Code.Code)
	ccntmract := new(sstates.Ccntmract)
	if err := ccntmract.Deserialize(bf); err != nil {
		return err
	}
	services, ok := Ccntmracts[ccntmract.Address]
	if !ok {
		return fmt.Errorf("Native ccntmract address %x haven't been registered.", ccntmract.Address)
	}
	services(this)
	service, ok := this.ServiceMap[ccntmract.Method]
	if !ok {
		return fmt.Errorf("Native ccntmract %x doesn't support this function %s.", ccntmract.Address, ccntmract.Method)
	}
	this.CcntmextRef.PushCcntmext(&ccntmext.Ccntmext{CcntmractAddress: ccntmract.Address})
	this.Input = ccntmract.Args
	if err := service(this); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Invoke] Native serivce function execute error!")
	}
	this.CcntmextRef.PopCcntmext()
	this.CcntmextRef.PushNotifications(this.Notifications)
	this.CloneCache.Commit()
	return nil
}

func RegisterOntCcntmract(native *NativeService) {
	native.Register("init", OntInit)
	native.Register("transfer", OntTransfer)
	native.Register("approve", OntApprove)
	native.Register("transferFrom", OntTransferFrom)
}

func RegisterOngCcntmract(native *NativeService) {
	native.Register("init", OngInit)
	native.Register("transfer", OngTransfer)
	native.Register("approve", OngApprove)
	native.Register("transferFrom", OngTransferFrom)
}
