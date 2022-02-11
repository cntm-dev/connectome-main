package neovm

import (
	"bytes"

	vm "github.com/cntmio/cntmology/vm/neovm"
	"github.com/cntmio/cntmology/errors"
	stypes "github.com/cntmio/cntmology/smartccntmract/types"
	scommon "github.com/cntmio/cntmology/core/store/common"
	"github.com/cntmio/cntmology/core/payload"
	"github.com/cntmio/cntmology/core/states"
	"github.com/cntmio/cntmology/common"
)

// create a new ccntmract
func CcntmractCreate(service *NeoVmService, engine *vm.ExecutionEngine) error {
	ccntmract, err := isCcntmractParamValid(engine); if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[CcntmractCreate] ccntmract parameters invalid!")
	}
	ccntmractAddress := ccntmract.Code.AddressFromVmCode()
	state, err := service.CloneCache.GetOrAdd(scommon.ST_CcntmRACT, ccntmractAddress[:], ccntmract)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[CcntmractCreate] GetOrAdd error!")
	}
	vm.PushData(engine, state)
	return nil
}

// migrate older ccntmract to new ccntmract
func CcntmractMigrate(service *NeoVmService, engine *vm.ExecutionEngine) error {
	ccntmract, err := isCcntmractParamValid(engine); if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[CcntmractMigrate] ccntmract parameters invalid!")
	}
	ccntmractAddress := ccntmract.Code.AddressFromVmCode()

	if err := isCcntmractExist(service, ccntmractAddress); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[CcntmractMigrate] ccntmract invalid!")
	}

	service.CloneCache.Add(scommon.ST_CcntmRACT, ccntmractAddress[:], ccntmract)
	if err := storeMigration(service, ccntmractAddress); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[CcntmractMigrate] ccntmract store migration error!")
	}
	vm.PushData(engine, ccntmract)
	return CcntmractDestory(service, engine)
}

// destory a ccntmract
func CcntmractDestory(service *NeoVmService, engine *vm.ExecutionEngine) error {
	ccntmext := service.CcntmextRef.CurrentCcntmext(); if ccntmext == nil {
		return errors.NewErr("[CcntmractDestory] current ccntmract ccntmext invalid!")
	}
	item, err := service.CloneCache.Store.TryGet(scommon.ST_CcntmRACT, ccntmext.CcntmractAddress[:])

	if err != nil || item == nil {
		return errors.NewErr("[CcntmractDestory] get current ccntmract fail!")
	}

	service.CloneCache.Delete(scommon.ST_CcntmRACT, ccntmext.CcntmractAddress[:])
	stateValues, err := service.CloneCache.Store.Find(scommon.ST_CcntmRACT, ccntmext.CcntmractAddress[:]); if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[CcntmractDestory] find error!")
	}
	for _, v := range stateValues {
		service.CloneCache.Delete(scommon.ST_STORAGE, []byte(v.Key))
	}
	return nil
}

// get ccntmract storage ccntmext
func CcntmractGetStorageCcntmext(service *NeoVmService, engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 1 {
		return errors.NewErr("[GetStorageCcntmext] Too few input parameter!")
	}
	opInterface := vm.PopInteropInterface(engine); if opInterface == nil {
		return errors.NewErr("[GetStorageCcntmext] Pop data nil!")
	}
	ccntmractState, ok := opInterface.(*payload.DeployCode); if !ok {
		return errors.NewErr("[GetStorageCcntmext] Pop data not ccntmract!")
	}
	address := ccntmractState.Code.AddressFromVmCode()
	item, err := service.CloneCache.Store.TryGet(scommon.ST_CcntmRACT, address[:])
	if err != nil || item == nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[GetStorageCcntmext] Get StorageCcntmext nil")
	}
	if address != service.CcntmextRef.CurrentCcntmext().CcntmractAddress {
		return errors.NewErr("[GetStorageCcntmext] CodeHash not equal!")
	}
	vm.PushData(engine, &StorageCcntmext{address: address})
	return nil
}

// get ccntmract code
func CcntmractGetCode(service *NeoVmService, engine *vm.ExecutionEngine) error {
	vm.PushData(engine, vm.PopInteropInterface(engine).(*payload.DeployCode).Code)
	return nil
}

func isCcntmractParamValid(engine *vm.ExecutionEngine) (*payload.DeployCode, error) {
	if vm.EvaluationStackCount(engine) < 7 {
		return nil, errors.NewErr("[Ccntmract] Too few input parameters")
	}
	code := vm.PopByteArray(engine); if len(code) > 1024 * 1024 {
		return nil, errors.NewErr("[Ccntmract] Code too lcntm!")
	}
	needStorage := vm.PopBoolean(engine)
	name := vm.PopByteArray(engine); if len(name) > 252 {
		return nil, errors.NewErr("[Ccntmract] Name too lcntm!")
	}
	version := vm.PopByteArray(engine); if len(version) > 252 {
		return nil, errors.NewErr("[Ccntmract] Version too lcntm!")
	}
	author := vm.PopByteArray(engine); if len(author) > 252 {
		return nil, errors.NewErr("[Ccntmract] Author too lcntm!")
	}
	email := vm.PopByteArray(engine); if len(email) > 252 {
		return nil, errors.NewErr("[Ccntmract] Email too lcntm!")
	}
	desc := vm.PopByteArray(engine); if len(desc) > 65536 {
		return nil, errors.NewErr("[Ccntmract] Desc too lcntm!")
	}
	ccntmract := &payload.DeployCode{
		Code:        stypes.VmCode{VmType:stypes.NEOVM, Code: code},
		NeedStorage: needStorage,
		Name:        string(name),
		Version:     string(version),
		Author:      string(author),
		Email:       string(email),
		Description: string(desc),
	}
	return ccntmract, nil
}

func isCcntmractExist(service *NeoVmService, ccntmractAddress common.Address) error {
	item, err := service.CloneCache.Get(scommon.ST_CcntmRACT, ccntmractAddress[:])

	if err != nil || item != nil {
		return errors.NewErr("[Ccntmract] Get ccntmract error or ccntmract exist!")
	}
	return nil
}

func storeMigration(service *NeoVmService, ccntmractAddress common.Address) error {
	stateValues, err := service.CloneCache.Store.Find(scommon.ST_CcntmRACT, ccntmractAddress[:]); if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Ccntmract] Find error!")
	}
	for _, v := range stateValues {
		key := new(states.StorageKey)
		bf := bytes.NewBuffer([]byte(v.Key))
		if err := key.Deserialize(bf); err != nil {
			return errors.NewErr("[Ccntmract] Key deserialize error!")
		}
		key = &states.StorageKey{CodeHash: ccntmractAddress, Key: key.Key}
		b := new(bytes.Buffer)
		if _, err := key.Serialize(b); err != nil {
			return errors.NewErr("[Ccntmract] Key Serialize error!")
		}
		service.CloneCache.Add(scommon.ST_STORAGE, key.ToArray(), v.Value)
	}
	return nil
}


