package account

import (
	. "github.com/Ontology/common"
	ct "github.com/Ontology/core/ccntmract"
)

type IClientStore interface {
	BuildDatabase(path string)

	SaveStoredData(name string, value []byte)

	LoadStoredData(name string) []byte

	LoadAccount() map[Uint160]*Account

	LoadCcntmracts() map[Uint160]*ct.Ccntmract
}
