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
	"bytes"
	"io"
	"math/big"
	"sort"

	"github.com/cntmio/cntmology-crypto/keypair"
	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/serialization"
	"github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/errors"
	scommon "github.com/cntmio/cntmology/smartccntmract/common"
	"github.com/cntmio/cntmology/smartccntmract/event"
	vm "github.com/cntmio/cntmology/vm/neovm"
	vmtypes "github.com/cntmio/cntmology/vm/neovm/types"
)

// HeaderGetNextConsensus put current block time to vm stack
func RuntimeGetTime(service *NeoVmService, engine *vm.ExecutionEngine) error {
	vm.PushData(engine, int(service.Time))
	return nil
}

// RuntimeCheckWitness provide check permissions service
// If param address isn't exist in authorization list, check fail
func RuntimeCheckWitness(service *NeoVmService, engine *vm.ExecutionEngine) error {
	data, err := vm.PopByteArray(engine)
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

	vm.PushData(engine, result)
	return nil
}

func RuntimeSerialize(service *NeoVmService, engine *vm.ExecutionEngine) error {
	item := vm.PopStackItem(engine)
	bf := new(bytes.Buffer)
	err := SerializeStackItem(item, bf)
	if err != nil {
		return err
	}
	vm.PushData(engine, bf.Bytes())
	return nil
}

func RuntimeDeserialize(service *NeoVmService, engine *vm.ExecutionEngine) error {
	data, err := vm.PopByteArray(engine)
	if err != nil {
		return err
	}
	bf := bytes.NewBuffer(data)
	item, err := DeserializeStackItem(bf)
	if err != nil {
		return err
	}

	if item == nil {
		return nil
	}
	vm.PushData(engine, item)
	return nil
}

// RuntimeNotify put smart ccntmract execute event notify to notifications
func RuntimeNotify(service *NeoVmService, engine *vm.ExecutionEngine) error {
	item := vm.PopStackItem(engine)
	ccntmext := service.CcntmextRef.CurrentCcntmext()
	service.Notifications = append(service.Notifications, &event.NotifyEventInfo{CcntmractAddress: ccntmext.CcntmractAddress, States: scommon.ConvertNeoVmTypeHexString(item)})
	return nil
}

// RuntimeLog push smart ccntmract execute event log to client
func RuntimeLog(service *NeoVmService, engine *vm.ExecutionEngine) error {
	item, err := vm.PopByteArray(engine)
	if err != nil {
		return err
	}
	ccntmext := service.CcntmextRef.CurrentCcntmext()
	txHash := service.Tx.Hash()
	event.PushSmartCodeEvent(txHash, 0, event.EVENT_LOG, &event.LogEventArgs{TxHash: txHash, CcntmractAddress: ccntmext.CcntmractAddress, Message: string(item)})
	return nil
}

func RuntimeGetTrigger(service *NeoVmService, engine *vm.ExecutionEngine) error {
	vm.PushData(engine, 0)
	return nil
}

func SerializeStackItem(item vmtypes.StackItems, w io.Writer) error {
	switch item.(type) {
	case *vmtypes.ByteArray:
		if err := serialization.WriteByte(w, vmtypes.ByteArrayType); err != nil {
			return errors.NewErr("Serialize ByteArray stackItems error: " + err.Error())
		}
		ba, _ := item.GetByteArray()
		if err := serialization.WriteVarBytes(w, ba); err != nil {
			return errors.NewErr("Serialize ByteArray stackItems error: " + err.Error())
		}

	case *vmtypes.Boolean:
		if err := serialization.WriteByte(w, vmtypes.BooleanType); err != nil {
			return errors.NewErr("Serialize Boolean StackItems error: " + err.Error())
		}
		b, _ := item.GetBoolean()
		if err := serialization.WriteBool(w, b); err != nil {
			return errors.NewErr("Serialize Boolean stackItems error: " + err.Error())
		}

	case *vmtypes.Integer:
		if err := serialization.WriteByte(w, vmtypes.IntegerType); err != nil {
			return errors.NewErr("Serialize Integer stackItems error: " + err.Error())
		}
		i, _ := item.GetByteArray()
		if err := serialization.WriteVarBytes(w, i); err != nil {
			return errors.NewErr("Serialize Integer stackItems error: " + err.Error())
		}

	case *vmtypes.Array:
		if err := serialization.WriteByte(w, vmtypes.ArrayType); err != nil {
			return errors.NewErr("Serialize Array stackItems error: " + err.Error())
		}
		a, _ := item.GetArray()
		if err := serialization.WriteVarUint(w, uint64(len(a))); err != nil {
			return errors.NewErr("Serialize Array stackItems error: " + err.Error())
		}

		for _, v := range a {
			SerializeStackItem(v, w)
		}

	case *vmtypes.Struct:
		if err := serialization.WriteByte(w, vmtypes.StructType); err != nil {
			return errors.NewErr("Serialize Struct stackItems error: " + err.Error())
		}
		s, _ := item.GetStruct()
		if err := serialization.WriteVarUint(w, uint64(len(s))); err != nil {
			return errors.NewErr("Serialize Struct stackItems error: " + err.Error())
		}

		for _, v := range s {
			SerializeStackItem(v, w)
		}

	case *vmtypes.Map:
		var unsortKey []string
		mp, _ := item.GetMap()
		keyMap := make(map[string]vmtypes.StackItems, 0)

		if err := serialization.WriteByte(w, vmtypes.MapType); err != nil {
			return errors.NewErr("Serialize Map stackItems error: " + err.Error())
		}
		if err := serialization.WriteVarUint(w, uint64(len(mp))); err != nil {
			return errors.NewErr("Serialize Map stackItems error: " + err.Error())
		}

		for k, _ := range mp {
			switch k.(type) {
			case *vmtypes.ByteArray, *vmtypes.Integer:
				ba, _ := k.GetByteArray()
				key := string(ba)
				if key == "" {
					return errors.NewErr("Serialize Map error: invalid key type")
				}
				unsortKey = append(unsortKey, key)
				keyMap[key] = k

			default:
				return errors.NewErr("Unsupport map key type.")
			}
		}

		sort.Strings(unsortKey)
		for _, v := range unsortKey {
			key := keyMap[v]
			SerializeStackItem(key, w)
			SerializeStackItem(mp[key], w)
		}

	default:
		return errors.NewErr("unknown type")
	}

	return nil
}

func DeserializeStackItem(r io.Reader) (items vmtypes.StackItems, err error) {
	t, err := serialization.ReadByte(r)
	if err != nil {
		return nil, errors.NewErr("Deserialize error: " + err.Error())
	}

	switch t {
	case vmtypes.ByteArrayType:
		b, err := serialization.ReadVarBytes(r)
		if err != nil {
			return nil, errors.NewErr("Deserialize stackItems ByteArray error: " + err.Error())
		}
		return vmtypes.NewByteArray(b), nil

	case vmtypes.BooleanType:
		b, err := serialization.ReadBool(r)
		if err != nil {
			return nil, errors.NewErr("Deserialize stackItems Boolean error: " + err.Error())
		}
		return vmtypes.NewBoolean(b), nil

	case vmtypes.IntegerType:
		b, err := serialization.ReadVarBytes(r)
		if err != nil {
			return nil, errors.NewErr("Deserialize stackItems Integer error: " + err.Error())
		}
		return vmtypes.NewInteger(new(big.Int).SetBytes(b)), nil

	case vmtypes.ArrayType, vmtypes.StructType:
		count, err := serialization.ReadVarUint(r, 0)
		if err != nil {
			return nil, errors.NewErr("Deserialize stackItems error: " + err.Error())
		}

		var arr []vmtypes.StackItems
		for count > 0 {
			item, err := DeserializeStackItem(r)
			if err != nil {
				return nil, err
			}
			arr = append(arr, item)
			count--
		}

		if t == vmtypes.StructType {
			return vmtypes.NewStruct(arr), nil
		}

		return vmtypes.NewArray(arr), nil

	case vmtypes.MapType:
		count, err := serialization.ReadVarUint(r, 0)
		if err != nil {
			return nil, errors.NewErr("Deserialize stackItems map error: " + err.Error())
		}

		mp := vmtypes.NewMap()
		for count > 0 {
			key, err := DeserializeStackItem(r)
			if err != nil {
				return nil, err
			}

			value, err := DeserializeStackItem(r)
			if err != nil {
				return nil, err
			}
			m, _ := mp.GetMap()
			m[key] = value
			count--
		}
		return mp, nil

	default:
		return nil, errors.NewErr("unknown type")
	}

	return nil, nil
}
