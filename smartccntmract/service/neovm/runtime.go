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

package neovm

import (
	"fmt"

	"github.com/cntmio/cntmology-crypto/keypair"
	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/core/signature"
	"github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/errors"
	"github.com/cntmio/cntmology/smartccntmract/event"
	vm "github.com/cntmio/cntmology/vm/neovm"
	vmtypes "github.com/cntmio/cntmology/vm/neovm/types"
)

// HeaderGetNextConsensus put current block time to vm stack
func RuntimeGetTime(service *NeoVmService, engine *vm.Executor) error {
	return engine.EvalStack.PushInt64(int64(service.Time))
}

// RuntimeCheckWitness provide check permissions service
// If param address isn't exist in authorization list, check fail
func RuntimeCheckWitness(service *NeoVmService, engine *vm.Executor) error {
	data, err := engine.EvalStack.PopAsBytes()
	if err != nil {
		return err
	}
	var result bool
	if len(data) == 20 {
		address, err := common.AddressParseFromBytes(data)
		if err != nil {
			return err
		}
		result = service.CcntmextRef.CheckWitness(address)
	} else {
		pk, err := keypair.DeserializePublicKey(data)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[RuntimeCheckWitness] data invalid.")
		}
		result = service.CcntmextRef.CheckWitness(types.AddressFromPubKey(pk))
	}

	return engine.EvalStack.PushBool(result)
}

func RuntimeSerialize(service *NeoVmService, engine *vm.Executor) error {
	val, err := engine.EvalStack.Pop()
	if err != nil {
		return err
	}
	sink := new(common.ZeroCopySink)
	err = val.Serialize(sink)
	if err != nil {
		return err
	}
	return engine.EvalStack.PushBytes(sink.Bytes())
}

//TODO check consistency with original implementation
func RuntimeDeserialize(service *NeoVmService, engine *vm.Executor) error {
	data, err := engine.EvalStack.PopAsBytes()
	if err != nil {
		return fmt.Errorf("[RuntimeDeserialize] PopAsBytes error: %s", err)
	}
	source := common.NewZeroCopySource(data)
	vmValue := vmtypes.VmValue{}
	err = vmValue.Deserialize(source)
	if err != nil {
		return fmt.Errorf("[RuntimeDeserialize] Deserialize error: %s", err)
	}
	return engine.EvalStack.Push(vmValue)
}

func RuntimeVerifyMutiSig(service *NeoVmService, engine *vm.Executor) error {
	data, err := engine.EvalStack.PopAsBytes()
	if err != nil {
		return err
	}
	arr1, err := engine.EvalStack.PopAsArray()
	if err != nil {
		return err
	}
	pks := make([]keypair.PublicKey, 0, len(arr1.Data))
	for i := 0; i < len(arr1.Data); i++ {
		value, err := arr1.Data[i].AsBytes()
		if err != nil {
			return err
		}
		pk, err := keypair.DeserializePublicKey(value)
		if err != nil {
			return err
		}
		pks = append(pks, pk)
	}

	m, err := engine.EvalStack.PopAsInt64()
	if err != nil {
		return err
	}
	if m > int64(len(pks)) || m < 0 {
		return fmt.Errorf("runtime verify multisig error: wrcntm m %d", m)

	}
	arr2, err := engine.EvalStack.PopAsArray()
	if err != nil {
		return err
	}
	signs := make([][]byte, 0, len(arr2.Data))
	for i := 0; i < len(arr2.Data); i++ {
		value, err := arr2.Data[i].AsBytes()
		if err != nil {
			return err
		}
		signs = append(signs, value)
	}
	err = signature.VerifyMultiSignature(data, pks, int(m), signs)
	return engine.EvalStack.PushBool(err == nil)
}

// RuntimeNotify put smart ccntmract execute event notify to notifications
func RuntimeNotify(service *NeoVmService, engine *vm.Executor) error {
	item, err := engine.EvalStack.Pop()
	if err != nil {
		return err
	}

	ccntmext := service.CcntmextRef.CurrentCcntmext()
	states, err := item.ConvertNeoVmValueHexString()
	if err != nil {
		return err
	}
	service.Notifications = append(service.Notifications, &event.NotifyEventInfo{CcntmractAddress: ccntmext.CcntmractAddress, States: states})
	return nil
}

// RuntimeLog push smart ccntmract execute event log to client
func RuntimeLog(service *NeoVmService, engine *vm.Executor) error {
	item, err := engine.EvalStack.PopAsBytes()
	if err != nil {
		return err
	}
	ccntmext := service.CcntmextRef.CurrentCcntmext()
	txHash := service.Tx.Hash()
	event.PushSmartCodeEvent(txHash, 0, event.EVENT_LOG, &event.LogEventArgs{TxHash: txHash, CcntmractAddress: ccntmext.CcntmractAddress, Message: string(item)})
	return nil
}

func RuntimeGetTrigger(service *NeoVmService, engine *vm.Executor) error {
	return engine.EvalStack.PushInt64(int64(0))
}

func RuntimeBase58ToAddress(service *NeoVmService, engine *vm.Executor) error {
	item, err := engine.EvalStack.PopAsBytes()
	if err != nil {
		return err
	}
	address, err := common.AddressFromBase58(string(item))
	if err != nil {
		return err
	}
	return engine.EvalStack.PushBytes(address[:])
}

func RuntimeAddressToBase58(service *NeoVmService, engine *vm.Executor) error {
	item, err := engine.EvalStack.PopAsBytes()
	if err != nil {
		return err
	}
	address, err := common.AddressParseFromBytes(item)
	if err != nil {
		return err
	}
	return engine.EvalStack.PushBytes([]byte(address.ToBase58()))
}

func RuntimeGetCurrentBlockHash(service *NeoVmService, engine *vm.Executor) error {
	return engine.EvalStack.PushBytes(service.BlockHash.ToArray())
}
