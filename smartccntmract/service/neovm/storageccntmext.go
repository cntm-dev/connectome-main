package service

import (
	"github.com/Ontology/common"
)

type StorageCcntmext struct {
	codeHash common.Address
}

func NewStorageCcntmext(codeHash common.Address) *StorageCcntmext {
	var storageCcntmext StorageCcntmext
	storageCcntmext.codeHash = codeHash
	return &storageCcntmext
}

func (sc *StorageCcntmext) ToArray() []byte {
	return sc.codeHash.ToArray()
}

