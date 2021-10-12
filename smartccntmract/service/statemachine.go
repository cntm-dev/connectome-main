package service

import (
	"bytes"
	"fmt"
	"github.com/Ontology/common"
	"github.com/Ontology/common/log"
	"github.com/Ontology/core/asset"
	"github.com/Ontology/core/code"
	"github.com/Ontology/core/ccntmract"
	"github.com/Ontology/core/ledger"
	"github.com/Ontology/core/states"
	"github.com/Ontology/core/store"
	"github.com/Ontology/core/transaction"
	"github.com/Ontology/crypto"
	"github.com/Ontology/errors"
	. "github.com/Ontology/smartccntmract/errors"
	"github.com/Ontology/smartccntmract/storage"
	"github.com/Ontology/smartccntmract/types"
	vm "github.com/Ontology/vm/neovm"
	"math"
)

type StateMachine struct {
	*StateReader
	CloneCache *storage.CloneCache
	trigger    types.TriggerType
	block      *ledger.Block
}

func NewStateMachine(dbCache store.IStateStore, trigger types.TriggerType, block *ledger.Block) *StateMachine {
	var stateMachine StateMachine
	stateMachine.CloneCache = storage.NewCloneCache(dbCache)
	stateMachine.StateReader = NewStateReader(trigger)
	stateMachine.trigger = trigger
	stateMachine.block = block

	stateMachine.StateReader.Register("Neo.Runtime.GetTrigger", stateMachine.RuntimeGetTrigger)
	stateMachine.StateReader.Register("Neo.Runtime.GetTime", stateMachine.RuntimeGetTime)

	stateMachine.StateReader.Register("Neo.Asset.Create", stateMachine.CreateAsset)
	stateMachine.StateReader.Register("Neo.Asset.Renew", stateMachine.AssetRenew)

	stateMachine.StateReader.Register("Neo.Ccntmract.Create", stateMachine.CcntmractCreate)
	stateMachine.StateReader.Register("Neo.Ccntmract.Migrate", stateMachine.CcntmractMigrate)
	stateMachine.StateReader.Register("Neo.Ccntmract.GetStorageCcntmext", stateMachine.GetStorageCcntmext)
	stateMachine.StateReader.Register("Neo.Ccntmract.GetScript", stateMachine.CcntmractGetCode)
	stateMachine.StateReader.Register("Neo.Ccntmract.Destroy", stateMachine.CcntmractDestory)

	stateMachine.StateReader.Register("Neo.Storage.Get", stateMachine.StorageGet)
	stateMachine.StateReader.Register("Neo.Storage.Put", stateMachine.StoragePut)
	stateMachine.StateReader.Register("Neo.Storage.Delete", stateMachine.StorageDelete)
	return &stateMachine
}

func (s *StateMachine) RuntimeGetTrigger(engine *vm.ExecutionEngine) (bool, error) {
	vm.PushData(engine, int(s.trigger))
	return true, nil
}

func (s *StateMachine) RuntimeGetTime(engine *vm.ExecutionEngine) (bool, error) {
	if s.block == nil {
		return false, errors.NewErr("[RuntimeGetTime] Block is fail!")
	}
	vm.PushData(engine, s.block.Header.Timestamp)
	return true, nil
}

func (s *StateMachine) CreateAsset(engine *vm.ExecutionEngine) (bool, error) {
	ccntmainer := engine.GetCodeCcntmainer()
	if ccntmainer == nil {
		log.Error("[CreateAsset] Get ccntmainer fail!")
		return false, errors.NewErr("[CreateAsset] Get ccntmainer fail!")
	}
	tran, ok := ccntmainer.(*transaction.Transaction)
	if !ok {
		log.Error("[CreateAsset] Ccntmainer not transaction!")
		return false, errors.NewErr("[CreateAsset] Ccntmainer not transaction!")
	}
	assetId := tran.Hash()
	if vm.EvaluationStackCount(engine) < 7 {
		log.Error("[CreateAsset] Too few input parameters ")
		return false, errors.NewErr("[CreateAsset] Too few input parameters ")
	}
	assertType := asset.AssetType(vm.PopInt(engine))
	name := vm.PopByteArray(engine)
	if len(name) > 1024 {
		log.Error("[CreateAsset] Asset name invalid, too lcntm")
		return false, ErrAssetNameInvalid
	}
	amount := vm.PopBigInt(engine)
	if amount.Sign() == 0 {
		log.Error("[CreateAsset] Asset amount invalid")
		return false, ErrAssetAmountInvalid
	}
	precision := vm.PopBigInt(engine)
	if precision.Int64() > 8 {
		log.Error("[CreateAsset] Asset precision invalid")
		return false, ErrAssetPrecisionInvalid
	}
	if amount.Int64()%int64(math.Pow(10, 8-float64(precision.Int64()))) != 0 {
		log.Error("[CreateAsset] Asset precision invalid")
		return false, ErrAssetAmountInvalid
	}
	ownerByte := vm.PopByteArray(engine)
	owner, err := crypto.DecodePoint(ownerByte)
	if err != nil {
		log.Error("[CreateAsset] Decode publickey fail!")
		return false, err
	}
	if result, _ := s.StateReader.CheckWitnessPublicKey(engine, owner); !result {
		log.Error("[CreateAsset] Check publickey fail!")
		return result, ErrAssetCheckOwnerInvalid
	}
	adminByte := vm.PopByteArray(engine)
	admin, err := common.Uint160ParseFromBytes(adminByte)
	if err != nil {
		log.Error("[CreateAsset] Convert admin fail!")
		return false, err
	}
	assetState := &states.AssetState{
		AssetId:    assetId,
		AssetType:  asset.AssetType(assertType),
		Name:       string(name),
		Amount:     common.Fixed64(amount.Int64()),
		Precision:  byte(precision.Int64()),
		Owner:      owner,
		Admin:      admin,
		Issuer:     admin,
		Expiration: ledger.DefaultLedger.Store.GetHeight() + 1 + 2000000,
		IsFrozen:   false,
	}
	state, err := s.CloneCache.GetOrAdd(store.ST_Asset, assetId.ToArray(), assetState)
	if err != nil {
		log.Error("[CreateAsset] GetOrAdd asset fail!")
		return false, errors.NewDetailErr(err, errors.ErrNoCode, "[CreateAsset] GetOrAdd error!")
	}
	vm.PushData(engine, state)
	return true, nil
}

func (s *StateMachine) CcntmractCreate(engine *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(engine) < 8 {
		return false, errors.NewErr("[CcntmractCreate] Too few input parameters")
	}
	codeByte := vm.PopByteArray(engine)
	if len(codeByte) > 1024*1024 {
		return false, errors.NewErr("[CcntmractCreate] Code too lcntm!")
	}
	parameters := vm.PopByteArray(engine)
	if len(parameters) > 252 {
		return false, errors.NewErr("[CcntmractCreate] Parameters too lcntm!")
	}
	parameterList := make([]ccntmract.CcntmractParameterType, 0)
	for _, v := range parameters {
		parameterList = append(parameterList, ccntmract.CcntmractParameterType(v))
	}
	returnType := vm.PopInt(engine)
	nameByte := vm.PopByteArray(engine)
	if len(nameByte) > 252 {
		return false, errors.NewErr("[CcntmractCreate] Name too lcntm!")
	}
	versionByte := vm.PopByteArray(engine)
	if len(versionByte) > 252 {
		return false, errors.NewErr("[CcntmractCreate] Version too lcntm!")
	}
	authorByte := vm.PopByteArray(engine)
	if len(authorByte) > 252 {
		return false, errors.NewErr("[CcntmractCreate] Author too lcntm!")
	}
	emailByte := vm.PopByteArray(engine)
	if len(emailByte) > 252 {
		return false, errors.NewErr("[CcntmractCreate] Email too lcntm!")
	}
	descByte := vm.PopByteArray(engine)
	if len(descByte) > 65536 {
		return false, errors.NewErr("[CcntmractCreate] Desc too lcntm!")
	}
	funcCode := &code.FunctionCode{
		Code:           codeByte,
		ParameterTypes: parameterList,
		ReturnType:     ccntmract.CcntmractParameterType(returnType),
	}
	ccntmractState := &states.CcntmractState{
		Code:        funcCode,
		Name:        string(nameByte),
		Version:     string(versionByte),
		Author:      string(authorByte),
		Email:       string(emailByte),
		Description: string(descByte),
	}
	codeHash := common.ToCodeHash(codeByte)
	state, err := s.CloneCache.GetOrAdd(store.ST_Ccntmract, codeHash.ToArray(), ccntmractState)
	if err != nil {
		return false, errors.NewDetailErr(err, errors.ErrNoCode, "[CcntmractCreate] GetOrAdd error!")
	}
	vm.PushData(engine, state)
	return true, nil
}

func (s *StateMachine) CcntmractMigrate(engine *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(engine) < 8 {
		return false, errors.NewErr("[CcntmractMigrate] Too few input parameters ")
	}
	codeByte := vm.PopByteArray(engine)
	if len(codeByte) > 1024*1024 {
		return false, errors.NewErr("[CcntmractMigrate] Code too lcntm!")
	}
	codeHash := common.ToCodeHash(codeByte)
	item, err := s.CloneCache.Get(store.ST_Ccntmract, codeHash.ToArray())
	if err != nil {
		return false, errors.NewErr("[CcntmractMigrate] Get Ccntmract error!")
	}
	if item != nil {
		return false, errors.NewErr("[CcntmractMigrate] Migrate Ccntmract has exist!")
	}

	parameters := vm.PopByteArray(engine)
	if len(parameters) > 252 {
		return false, errors.NewErr("[CcntmractMigrate] Parameters too lcntm!")
	}
	parameterList := make([]ccntmract.CcntmractParameterType, 0)
	for _, v := range parameters {
		parameterList = append(parameterList, ccntmract.CcntmractParameterType(v))
	}
	returnType := vm.PopInt(engine)
	nameByte := vm.PopByteArray(engine)
	if len(nameByte) > 252 {
		return false, errors.NewErr("[CcntmractMigrate] Name too lcntm!")
	}
	versionByte := vm.PopByteArray(engine)
	if len(versionByte) > 252 {
		return false, errors.NewErr("[CcntmractMigrate] Version too lcntm!")
	}
	authorByte := vm.PopByteArray(engine)
	if len(authorByte) > 252 {
		return false, errors.NewErr("[CcntmractMigrate] Author too lcntm!")
	}
	emailByte := vm.PopByteArray(engine)
	if len(emailByte) > 252 {
		return false, errors.NewErr("[CcntmractMigrate] Email too lcntm!")
	}
	descByte := vm.PopByteArray(engine)
	if len(descByte) > 65536 {
		return false, errors.NewErr("[CcntmractMigrate] Desc too lcntm!")
	}
	funcCode := &code.FunctionCode{
		Code:           codeByte,
		ParameterTypes: parameterList,
		ReturnType:     ccntmract.CcntmractParameterType(returnType),
	}
	ccntmractState := &states.CcntmractState{
		Code:        funcCode,
		Name:        string(nameByte),
		Version:     string(versionByte),
		Author:      string(authorByte),
		Email:       string(emailByte),
		Description: string(descByte),
	}
	s.CloneCache.Add(store.ST_Ccntmract, codeHash.ToArray(), ccntmractState)
	stateValues, err := s.CloneCache.Store.Find(store.ST_Ccntmract, codeHash.ToArray())
	if err != nil {
		return false, errors.NewDetailErr(err, errors.ErrNoCode, "[CcntmractMigrate] Find error!")
	}
	for _, v := range stateValues {
		key := new(states.StorageKey)
		bf := bytes.NewBuffer([]byte(v.Key))
		if err := key.Deserialize(bf); err != nil {
			return false, errors.NewErr("[CcntmractMigrate] Key deserialize error!")
		}
		key = &states.StorageKey{CodeHash: codeHash, Key: key.Key}
		b := new(bytes.Buffer)
		if _, err := key.Serialize(b); err != nil {
			return false, errors.NewErr("[CcntmractMigrate] Key Serialize error!")
		}
		s.CloneCache.Add(store.ST_Storage, key.ToArray(), v.Value)
	}
	vm.PushData(engine, ccntmractState)
	return s.CcntmractDestory(engine)
}

func (s *StateMachine) AssetRenew(engine *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(engine) < 2 {
		return false, errors.NewErr("[AssetRenew] Too few input parameters ")
	}
	data := vm.PopInteropInterface(engine)
	if data == nil {
		return false, errors.NewErr("[AssetRenew] Get Asset nil!")
	}
	years := vm.PopInt(engine)
	assetState := data.(*states.AssetState)
	if assetState.Expiration < ledger.DefaultLedger.Store.GetHeight()+1 {
		assetState.Expiration = ledger.DefaultLedger.Store.GetHeight() + 1
	}
	assetState.Expiration += uint32(years) * 2000000
	vm.PushData(engine, assetState.Expiration)
	return true, nil
}

func (s *StateMachine) CcntmractDestory(engine *vm.ExecutionEngine) (bool, error) {
	ccntmext, err := engine.CurrentCcntmext()
	if err != nil {
		return false, err
	}
	hash, err := ccntmext.GetCodeHash()
	if err != nil {
		return false, nil
	}
	item, err := s.CloneCache.Store.TryGet(store.ST_Ccntmract, hash.ToArray())
	if err != nil {
		return false, err
	}
	if item == nil {
		return false, nil
	}
	s.CloneCache.Delete(store.ST_Ccntmract, hash.ToArray())
	stateValues, err := s.CloneCache.Store.Find(store.ST_Ccntmract, hash.ToArray())
	if err != nil {
		return false, errors.NewDetailErr(err, errors.ErrNoCode, "[CcntmractDestory] Find error!")
	}
	for _, v := range stateValues {
		s.CloneCache.Delete(store.ST_Storage, []byte(v.Key))
	}
	return true, nil
}

func (s *StateMachine) CheckStorageCcntmext(ccntmext *StorageCcntmext) (bool, error) {
	item, err := s.CloneCache.Get(store.ST_Ccntmract, ccntmext.codeHash.ToArray())
	if err != nil {
		return false, err
	}
	if item == nil {
		return false, errors.NewErr(fmt.Sprintf("get ccntmract by codehash=%v nil", ccntmext.codeHash))
	}
	return true, nil
}

func (s *StateMachine) StoragePut(engine *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(engine) < 3 {
		return false, errors.NewErr("[StoragePut] Too few input parameters ")
	}
	opInterface := vm.PopInteropInterface(engine)
	if opInterface == nil {
		return false, errors.NewErr("[StoragePut] Get StorageCcntmext nil")
	}
	ccntmext := opInterface.(*StorageCcntmext)
	key := vm.PopByteArray(engine)
	if len(key) > 1024 {
		return false, errors.NewErr("[StoragePut] Get Storage key to lcntm")
	}
	value := vm.PopByteArray(engine)
	k, err := serializeStorageKey(ccntmext.codeHash, key)
	if err != nil {
		return false, err
	}
	s.CloneCache.Add(store.ST_Storage, k, &states.StorageItem{Value: value})
	return true, nil
}

func (s *StateMachine) StorageDelete(engine *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(engine) < 2 {
		return false, errors.NewErr("[StorageDelete] Too few input parameters ")
	}
	opInterface := vm.PopInteropInterface(engine)
	if opInterface == nil {
		return false, errors.NewErr("[StorageDelete] Get StorageCcntmext nil")
	}
	ccntmext := opInterface.(*StorageCcntmext)
	key := vm.PopByteArray(engine)
	k, err := serializeStorageKey(ccntmext.codeHash, key)
	if err != nil {
		return false, err
	}
	s.CloneCache.Delete(store.ST_Storage, k)
	return true, nil
}

func (s *StateMachine) StorageGet(engine *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(engine) < 2 {
		return false, errors.NewErr("[StorageGet] Too few input parameters ")
	}
	opInterface := vm.PopInteropInterface(engine)
	if opInterface == nil {
		return false, errors.NewErr("[StorageGet] Get StorageCcntmext error!")
	}
	ccntmext := opInterface.(*StorageCcntmext)
	if exist, err := s.CheckStorageCcntmext(ccntmext); !exist {
		return false, err
	}
	key := vm.PopByteArray(engine)
	k, err := serializeStorageKey(ccntmext.codeHash, key)
	if err != nil {
		return false, err
	}
	item, err := s.CloneCache.Get(store.ST_Storage, k)
	if err != nil {
		return false, err
	}
	if item == nil {
		vm.PushData(engine, []byte{})
	} else {
		vm.PushData(engine, item.(*states.StorageItem).Value)
	}
	return true, nil
}

func (s *StateMachine) GetStorageCcntmext(engine *vm.ExecutionEngine) (bool, error) {
	if vm.EvaluationStackCount(engine) < 1 {
		return false, errors.NewErr("[GetStorageCcntmext] Too few input parameters ")
	}
	opInterface := vm.PopInteropInterface(engine)
	if opInterface == nil {
		return false, errors.NewErr("[GetStorageCcntmext] Get StorageCcntmext nil")
	}
	ccntmractState := opInterface.(*states.CcntmractState)
	codeHash := ccntmractState.Code.CodeHash()
	item, err := s.CloneCache.Store.TryGet(store.ST_Ccntmract, codeHash.ToArray())
	if err != nil {
		return false, errors.NewDetailErr(err, errors.ErrNoCode, "[GetStorageCcntmext] Get StorageCcntmext nil")
	}
	ccntmext, err := engine.CurrentCcntmext()
	if err != nil {
		return false, err
	}
	if item == nil {
		return false, errors.NewErr(fmt.Sprintf("[GetStorageCcntmext] Get ccntmract by codehash:%v nil", codeHash))
	}
	currentHash, err := ccntmext.GetCodeHash()
	if err != nil {
		return false, err
	}
	if codeHash.CompareTo(currentHash) != 0 {
		return false, errors.NewErr("[GetStorageCcntmext] CodeHash not equal!")
	}
	vm.PushData(engine, &StorageCcntmext{codeHash: codeHash})
	return true, nil
}

func ccntmains(programHashes []common.Uint160, programHash common.Uint160) bool {
	for _, v := range programHashes {
		if v.CompareTo(programHash) == 0 {
			return true
		}
	}
	return false
}

func serializeStorageKey(codeHash common.Uint160, key []byte) ([]byte, error) {
	bf := new(bytes.Buffer)
	storageKey := &states.StorageKey{CodeHash: codeHash, Key: key}
	if _, err := storageKey.Serialize(bf); err != nil {
		return []byte{}, errors.NewErr("[serializeStorageKey] StorageKey serialize error!")
	}
	return bf.Bytes(), nil
}
