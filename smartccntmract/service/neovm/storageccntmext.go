package service

import (
	"github.com/Ontology/common"
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

