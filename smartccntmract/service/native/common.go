package native

import (
	scommon "github.com/Ontology/core/store/common"
	"github.com/Ontology/errors"
	"math/big"
	cstates "github.com/Ontology/core/states"
	"github.com/Ontology/common"
	"github.com/Ontology/smartccntmract/service/native/states"
	"github.com/Ontology/vm/types"
)

var (
	addressHeight = []byte("addressHeight")
)

func checkWitness(native *NativeService, u160 common.Address) bool {
	addresses := native.Tx.GetSignatureAddresses()
	for _, v := range addresses {
		if v == u160 {
			return true
		}
	}
	return false
}

func getAddressHeightKey(ccntmract, address common.Address) []byte {
	temp := append(addressHeight, address[:]...)
	return append(ccntmract[:], temp...)
}

func getHeightStorageItem(height uint32) *cstates.StorageItem {
	return &cstates.StorageItem{Value: big.NewInt(int64(height)).Bytes()}
}

func getAmountStorageItem(value *big.Int) *cstates.StorageItem {
	return &cstates.StorageItem{Value: value.Bytes()}
}

func getToAmountStorageItem(toBalance, value *big.Int) *cstates.StorageItem {
	return &cstates.StorageItem{Value: new(big.Int).Add(toBalance, value).Bytes()}
}

func getTotalSupplyKey(ccntmract common.Address) []byte {
	return append(ccntmract[:], totalSupplyName...)
}

func getTransferKey(ccntmract, from common.Address) []byte {
	return append(ccntmract[:], from[:]...)
}

func getApproveKey(ccntmract common.Address, state *states.State) []byte {
	temp := append(ccntmract[:], state.From[:]...)
	return append(temp, state.To[:]...)
}

func getTransferFromKey(ccntmract common.Address, state *states.TransferFrom) []byte {
	temp := append(ccntmract[:], state.From[:]...)
	return append(temp, state.Sender[:]...)
}

func isTransferValid(native *NativeService, state *states.State) error {
	if state.Value.Sign() < 0 {
		return errors.NewErr("Transfer amount invalid!")
	}

	if !checkWitness(native, state.From) {
		return errors.NewErr("Authentication failed!")
	}
	return nil
}

func transfer(native *NativeService, ccntmract common.Address, state *states.State) error {
	if err := isTransferValid(native, state); err != nil {
		return err
	}

	err := fromTransfer(native, getTransferKey(ccntmract, state.From), state.Value); if err != nil {
		return err
	}

	if err := toTransfer(native, getTransferKey(ccntmract, state.To), state.Value); err != nil {
		return err
	}
	return nil
}

func transferFrom(native *NativeService, currentCcntmract common.Address, state *states.TransferFrom) error {
	if err := isTransferFromValid(native, state); err != nil {
		return err
	}

	if err := fromApprove(native, getTransferFromKey(currentCcntmract, state), state.Value); err != nil {
		return err
	}

	err := fromTransfer(native, getTransferKey(currentCcntmract, state.From), state.Value); if err != nil {
		return err
	}

	if err := toTransfer(native, getTransferKey(currentCcntmract, state.To), state.Value); err != nil {
		return err
	}
	return nil
}

func isTransferFromValid(native *NativeService, state *states.TransferFrom) error {
	if state.Value.Sign() < 0 {
		return errors.NewErr("TransferFrom amount invalid!")
	}
	if err := isSenderValid(native, state.Sender); err != nil {
		return err
	}
	return nil
}

func isApproveValid(native *NativeService, state *states.State) error {
	if state.Value.Sign() < 0 {
		return errors.NewErr("Approve amount invalid!")
	}
	if err := isSenderValid(native, state.From); err != nil {
		return err
	}
	return nil
}

func isSenderValid(native *NativeService, sender common.Address) error {
	vmType := sender[0]
	if  vmType == byte(types.Native) || vmType == byte(types.NEOVM) || vmType == byte(types.WASMVM) {
		callCcntmext := native.CcntmextRef.CallingCcntmext()
		if callCcntmext != nil {
			return errors.NewErr("[Sender] CallingCcntmext nil, Authentication failed!")
		}
		if sender == callCcntmext.CcntmractAddress {
			return errors.NewErr("[Sender] CallingCcntmext invalid, Authentication failed!")
		}
	} else {
		if !checkWitness(native, sender) {
			return errors.NewErr("[Sender] Authentication failed!")
		}
	}
	return nil
}

func fromApprove(native *NativeService, fromApproveKey []byte, value *big.Int) error {
	approveValue, err := getStorageBigInt(native, fromApproveKey); if err != nil {
		return err
	}
	approveBalance := new(big.Int).Sub(approveValue,value)
	sign := approveBalance.Sign()
	if sign < 0 {
		return errors.NewErr("[TransferFrom] approve balance insufficient!")
	} else if sign == 0 {
		native.CloneCache.Delete(scommon.ST_Storage, fromApproveKey)
	} else {
		native.CloneCache.Add(scommon.ST_Storage, fromApproveKey, getAmountStorageItem(approveBalance))
	}
	return nil
}

func fromTransfer(native *NativeService, fromKey []byte, value *big.Int) error {
	fromBalance, err := getStorageBigInt(native, fromKey); if err != nil {
		return err
	}
	balance := new(big.Int).Sub(fromBalance, value)
	sign := balance.Sign()
	if sign < 0 {
		return errors.NewErr("[Transfer] balance insufficient!")
	} else if sign == 0 {
		native.CloneCache.Delete(scommon.ST_Storage, fromKey)
	} else {
		native.CloneCache.Add(scommon.ST_Storage, fromKey, getAmountStorageItem(balance))
	}
	return nil
}

func toTransfer(native *NativeService, toKey []byte, value *big.Int) error {
	toBalance, err := getStorageBigInt(native, toKey); if err != nil {
		return err
	}
	native.CloneCache.Add(scommon.ST_Storage, toKey, getToAmountStorageItem(toBalance, value))
	return nil
}

func getStartHeight(native *NativeService, ccntmract, from common.Address) (uint32, error) {
	startHeight, err := getStorageBigInt(native, getAddressHeightKey(ccntmract, from)); if err != nil {
		return 0, err
	}
	return uint32(startHeight.Int64()), nil
}

func getStorageBigInt(native *NativeService, key []byte) (*big.Int, error) {
	balance, err := native.CloneCache.Get(scommon.ST_Storage, key)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[getBalance] storage error!")
	}
	if balance == nil {
		return big.NewInt(0), nil
	}
	item, ok := balance.(*cstates.StorageItem); if !ok {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[getBalance] get amount error!")
	}
	return new(big.Int).SetBytes(item.Value), nil
}

