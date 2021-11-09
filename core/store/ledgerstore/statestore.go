package ledgerstore

import (
	"bytes"
	"fmt"
	"github.com/Ontology/common"
	"github.com/Ontology/common/serialization"
	. "github.com/Ontology/core/states"
	. "github.com/Ontology/core/store/common"
	"github.com/Ontology/core/store/leveldbstore"
	. "github.com/Ontology/core/store/statestore"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/Ontology/core/payload"
)

var (
	CurrentStateRoot = []byte("Current-State-Root")
	BookerKeeper     = []byte("Booker-Keeper")
)

type StateStore struct {
	dbDir string
	store IStore
}

func NewStateStore(dbDir string) (*StateStore, error) {
	var err error
	store, err := leveldbstore.NewLevelDBStore(dbDir)
	if err != nil {
		return nil, err
	}
	return &StateStore{
		dbDir: dbDir,
		store: store,
	}, nil
}

func (this *StateStore) NewBatch() error {
	err := this.store.NewBatch()
	if err != nil {
		return fmt.Errorf("NewBatch error %s", err)
	}
	return nil
}

func (this *StateStore) NewStateBatch(stateRoot common.Uint256) (*StateBatch, error) {
	return NewStateStoreBatch(NewMemDatabase(), this.store, stateRoot)
}

func (this *StateStore) CommitTo() error {
	return this.store.BatchCommit()
}

func (this *StateStore) GetCurrentStateRoot() (common.Uint256, error) {
	key, err := this.getCurrentStateRootKey()
	if err != nil {
		return common.Uint256{}, err
	}
	value, err := this.store.Get(key)
	if err != nil {
		if err == leveldb.ErrNotFound{
			return common.Uint256{}, nil
		}
		return common.Uint256{}, err
	}
	stateRoot, err := common.Uint256ParseFromBytes(value)
	if err != nil {
		return common.Uint256{}, err
	}
	return stateRoot, nil
}


func (this *StateStore) GetCcntmractState(ccntmractHash common.Uint160) (*payload.DeployCode, error) {
	key, err := this.getCcntmractStateKey(ccntmractHash)
	if err != nil {
		return nil, err
	}

	value, err := this.store.Get(key)
	if err != nil {
		if err == leveldb.ErrNotFound{
			return nil, nil
		}
		return nil, err
	}
	reader := bytes.NewReader(value)
	ccntmractState := new(payload.DeployCode)
	err = ccntmractState.Deserialize(reader)
	if err != nil {
		return nil, err
	}
	return ccntmractState, nil
}

func (this *StateStore) SaveCurrentStateRoot(stateRoot common.Uint256) error {
	key, err := this.getCurrentStateRootKey()
	if err != nil {
		return err
	}

	return this.store.BatchPut(key, stateRoot.ToArray())
}

func (this *StateStore) GetBookKeeperState() (*BookKeeperState, error) {
	key, err := this.getBookKeeperKey()
	if err != nil {
		return nil, err
	}

	value, err := this.store.Get(key)
	if err != nil {
		if err == leveldb.ErrNotFound{
			return nil, nil
		}
		return nil, err
	}
	reader := bytes.NewReader(value)
	bookKeeperState := new(BookKeeperState)
	err = bookKeeperState.Deserialize(reader)
	if err != nil {
		return nil, err
	}
	return bookKeeperState, nil
}

func (this *StateStore) SaveBookKeeperState(bookKeeperState *BookKeeperState) error {
	key, err := this.getBookKeeperKey()
	if err != nil {
		return err
	}
	value := bytes.NewBuffer(nil)
	err = bookKeeperState.Serialize(value)
	if err != nil {
		return err
	}

	return this.store.Put(key, value.Bytes())
}

func (this *StateStore) GetStorageState(key *StorageKey) (*StorageItem, error) {
	storeKey, err := this.getStorageKey(key)
	if err != nil {
		return nil, err
	}

	data, err := this.store.Get(storeKey)
	if err != nil {
		if err == leveldb.ErrNotFound{
			return nil,nil
		}
		return nil, err
	}
	reader := bytes.NewReader(data)
	storageState := new(StorageItem)
	err = storageState.Deserialize(reader)
	if err != nil {
		return nil, err
	}
	return storageState, nil
}

func (this *StateStore) GetVoteStates() (map[common.Uint160]*VoteState, error) {
	votes := make(map[common.Uint160]*VoteState)
	iter := this.store.NewIterator([]byte{byte(ST_Vote)})
	for iter.Next() {
		rk := bytes.NewReader(iter.Key())
		// read prefix
		_, err := serialization.ReadBytes(rk, 1)
		if err != nil {
			return nil, fmt.Errorf("ReadBytes error %s", err)
		}
		var programHash common.Uint160
		if err := programHash.Deserialize(rk); err != nil {
			return nil, err
		}
		vote := new(VoteState)
		r := bytes.NewReader(iter.Value())
		if err := vote.Deserialize(r); err != nil {
			return nil, err
		}
		votes[programHash] = vote
	}
	return votes, nil
}

func (this *StateStore) getCurrentStateRootKey() ([]byte, error) {
	key := make([]byte, 1+len(CurrentStateRoot))
	key[0] = byte(SYS_CurrentStateRoot)
	copy(key[1:], []byte(CurrentStateRoot))
	return key, nil
}

func (this *StateStore) getBookKeeperKey() ([]byte, error) {
	key := make([]byte, 1+len(BookerKeeper))
	key[0] = byte(ST_BookKeeper)
	copy(key[1:], []byte(BookerKeeper))
	return key, nil
}

func (this *StateStore) getCcntmractStateKey(ccntmractHash common.Uint160) ([]byte, error) {
	data := ccntmractHash.ToArray()
	key := make([]byte, 1+len(data))
	key[0] = byte(ST_Ccntmract)
	copy(key[1:], []byte(data))
	return key, nil
}

func (this *StateStore) getStorageKey(key *StorageKey) ([]byte, error) {
	data := key.ToArray()
	storeKey := make([]byte, 1+len(data))
	storeKey[0] = byte(ST_Storage)
	copy(storeKey[1:], []byte(data))
	return storeKey, nil
}

func (this *StateStore) ClearAll() error {
	err := this.store.NewBatch()
	if err != nil {
		return err
	}
	iter := this.store.NewIterator(nil)
	for iter.Next() {
		err = this.store.BatchDelete(iter.Key())
		if err != nil {
			return fmt.Errorf("BatchDelete error %s", err)
		}
	}
	iter.Release()
	return this.store.BatchCommit()
}

func (this *StateStore) Close() error {
	return this.store.Close()
}