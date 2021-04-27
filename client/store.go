package client

import (
	ct "GoOnchain/core/ccntmract"
	. "GoOnchain/common"
)

type IClientStore interface {
	BuildDatabase(path string)

	SaveStoredData(name string,value []byte)

	LoadStoredData(name string) []byte

	LoadAccount()  map[Uint160]*Account

	LoadCcntmracts() map[Uint160]*ct.Ccntmract
}
