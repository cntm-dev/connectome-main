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

	"github.com/Ontology/common"
	"github.com/Ontology/core/genesis"
	scommon "github.com/Ontology/core/store/common"
	"github.com/Ontology/core/types"
	"github.com/Ontology/errors"
	"github.com/Ontology/smartccntmract/ccntmext"
	"github.com/Ontology/smartccntmract/event"
	"github.com/Ontology/smartccntmract/service/native/states"
	"github.com/Ontology/smartccntmract/storage"
	vmtypes "github.com/Ontology/vm/types"
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

type NativeService struct {
	CloneCache    *storage.CloneCache
	ServiceMap    map[string]Handler
	Notifications []*event.NotifyEventInfo
	Input         []byte
	Tx            *types.Transaction
	Height        uint32
	CcntmextRef    ccntmext.CcntmextRef
}

func NewNativeService(dbCache scommon.StateStore, height uint32, tx *types.Transaction, ctxRef ccntmext.CcntmextRef) *NativeService {
	var nativeService NativeService
	nativeService.CloneCache = storage.NewCloneCache(dbCache)
	nativeService.Tx = tx
	nativeService.Height = height
	nativeService.CcntmextRef = ctxRef
	nativeService.ServiceMap = make(map[string]Handler)
	return &nativeService
}

func (native *NativeService) Register(methodName string, handler Handler) {
	native.ServiceMap[methodName] = handler
}

func (native *NativeService) Invoke() error {
	ctx := native.CcntmextRef.CurrentCcntmext()
	if ctx == nil {
		return errors.NewErr("[Invoke] Native service current ccntmext doesn't exist!")
	}
	bf := bytes.NewBuffer(ctx.Code.Code)
	ccntmract := new(states.Ccntmract)
	if err := ccntmract.Deserialize(bf); err != nil {
		return err
	}
	services, ok := Ccntmracts[ccntmract.Address]
	if !ok {
		return fmt.Errorf("Native ccntmract address %x haven't been registered.", ccntmract.Address)
	}
	services(native)
	service, ok := native.ServiceMap[ccntmract.Method]
	if !ok {
		return fmt.Errorf("Native ccntmract %x doesn't support this function %s.", ccntmract.Address, ccntmract.Method)
	}
	native.CcntmextRef.PushCcntmext(&ccntmext.Ccntmext{CcntmractAddress: ccntmract.Address})
	native.Input = ccntmract.Args
	if err := service(native); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Invoke] Native serivce function execute error!")
	}
	native.CcntmextRef.PopCcntmext()
	native.CcntmextRef.PushNotifications(native.Notifications)
	native.CloneCache.Commit()
	return nil
}

func (native *NativeService) AppCall(address common.Address, method string, args []byte) error {
	bf := new(bytes.Buffer)
	ccntmract := &states.Ccntmract{
		Address: address,
		Method:  method,
		Args:    args,
	}

	if err := ccntmract.Serialize(bf); err != nil {
		return err
	}
	code := vmtypes.VmCode{
		VmType: vmtypes.Native,
		Code:   bf.Bytes(),
	}
	native.CcntmextRef.PushCcntmext(&ccntmext.Ccntmext{
		Code:            code,
		CcntmractAddress: code.AddressFromVmCode(),
	})
	if err := native.CcntmextRef.Execute(); err != nil {
		return err
	}
	native.CcntmextRef.PopCcntmext()
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
