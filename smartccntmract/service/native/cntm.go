package native

import (
	scommon "github.com/Ontology/core/store/common"
	"github.com/Ontology/errors"
	"github.com/Ontology/core/genesis"
	ctypes "github.com/Ontology/core/types"
	"math/big"
	"github.com/Ontology/smartccntmract/event"
	"github.com/Ontology/smartccntmract/service/native/states"
	cstates "github.com/Ontology/core/states"
	"bytes"
	"github.com/Ontology/account"
	"github.com/Ontology/common"
)

var (
	decrementInterval = uint32(2000000)
	generationAmount = [17]uint32{80, 70, 60, 50, 40, 30, 20, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10}
	gl = uint32(len(generationAmount))
)

func OntInit(native *NativeService) error {
	booKeepers := account.GetBookkeepers()

	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	amount, err := getStorageBigInt(native, getTotalSupplyKey(ccntmract))
	if err != nil {
		return err
	}

	if amount != nil && amount.Sign() != 0 {
		return errors.NewErr("Init cntm has been completed!")
	}

	ts := new(big.Int).Div(totalSupply, big.NewInt(int64(len(booKeepers))))
	for _, v := range booKeepers {
		address := ctypes.AddressFromPubKey(v)
		native.CloneCache.Add(scommon.ST_Storage, append(ccntmract[:], address[:]...), &cstates.StorageItem{Value: ts.Bytes()})
		native.CloneCache.Add(scommon.ST_Storage, getTotalSupplyKey(ccntmract), &cstates.StorageItem{Value: ts.Bytes()})
		native.Notifications = append(native.Notifications, &event.NotifyEventInfo{
			Ccntmainer: native.Tx.Hash(),
			CodeHash: genesis.OntCcntmractAddress,
			States: []interface{}{nil, address, ts},
		})
	}

	return nil
}

func OntTransfer(native *NativeService) error {
	transfers := new(states.Transfers)
	if err := transfers.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Transfer] Transfers deserialize error!")
	}
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	for _, v := range transfers.States {
		if err := transfer(native, ccntmract, v); err != nil {
			return err
		}

		startHeight, err := getStartHeight(native, ccntmract, v.From); if err != nil {
			return err
		}

		if err := grantOng(native, ccntmract, v, startHeight); err != nil {
			return err
		}

		native.Notifications = append(native.Notifications, &event.NotifyEventInfo{
			Ccntmainer: native.Tx.Hash(),
			CodeHash: native.CcntmextRef.CurrentCcntmext().CcntmractAddress,
			States: []interface{}{v.From, v.To, v.Value},
		})
	}
	return nil
}

func OntTransferFrom(native *NativeService) error {
	state := new(states.TransferFrom)
	if err := state.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[OntTransferFrom] State deserialize error!")
	}
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	if err := transferFrom(native, ccntmract, state); err != nil {
		return err
	}
	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			Ccntmainer: native.Tx.Hash(),
			CodeHash: ccntmract,
			States: []interface{}{state.From, state.To, state.Value},
		})
	return nil
}

func OntApprove(native *NativeService) error {
	state := new(states.State)
	if err := state.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[OngApprove] state deserialize error!")
	}
	if err := isApproveValid(native, state); err != nil {
		return err
	}
	ccntmract := native.CcntmextRef.CurrentCcntmext().CcntmractAddress
	native.CloneCache.Add(scommon.ST_Storage, getApproveKey(ccntmract, state), &cstates.StorageItem{Value: state.Value.Bytes()})
	return nil
}

func grantOng(native *NativeService, ccntmract common.Address, state *states.State, startHeight uint32) error {
	var amount uint32 = 0
	ustart := startHeight / decrementInterval
	if ustart < gl {
		istart := startHeight % decrementInterval
		uend := native.Height / decrementInterval
		iend := native.Height % decrementInterval
		if uend >= gl {
			uend = gl
			iend = 0
		}
		if iend == 0 {
			uend--
			iend = decrementInterval
		}
		for {
			if ustart >= uend {
				break
			}
			amount += (decrementInterval - istart) * generationAmount[ustart]
			ustart++
			istart = 0
		}
		amount += (iend - istart) * generationAmount[ustart]
	}

	args, err := getApproveArgs(native, ccntmract, state, amount); if err != nil {
		return err
	}

	if err := native.AppCall(genesis.OngCcntmractAddress, "approve", args); err != nil {
		return err
	}

	native.CloneCache.Add(scommon.ST_Storage, getAddressHeightKey(ccntmract, state.From), getHeightStorageItem(native.Height))
	return nil
}

func getApproveArgs(native *NativeService, ccntmract common.Address, state *states.State, amount uint32) ([]byte, error) {
	bf := new(bytes.Buffer)
	approve := &states.State {
		From: ccntmract,
		To: state.From,
		Value: big.NewInt(state.Value.Int64() / int64(genesis.OntRegisterAmount) * int64(amount)),
	}

	stateValue, err := getStorageBigInt(native, getApproveKey(ccntmract, state)); if err != nil {
		return nil, err
	}

	approve.Value = new(big.Int).Add(approve.Value, stateValue)

	if err := approve.Serialize(bf); err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}

