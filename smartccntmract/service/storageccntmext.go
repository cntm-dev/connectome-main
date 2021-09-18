package service

import (
	"github.com/Ontology/common"
	"github.com/Ontology/vm/neovm/interfaces"
)

type StorageCcntmext struct {
	codeHash common.Uint160
}

func NewStorageCcntmext(codeHash common.Uint160) *StorageCcntmext {
	var storageCcntmext StorageCcntmext
	storageCcntmext.codeHash = codeHash
	return &storageCcntmext
}

func (sc *StorageCcntmext) ToArray() []byte {
	return sc.codeHash.ToArray()
}

func (sc *StorageCcntmext) Clone() interfaces.IInteropInterface {
	s := *sc
	return &s
}
