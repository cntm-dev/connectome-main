package client

import (
	ct "GoOnchain/core/ccntmract"
	. "GoOnchain/common"
)

type ClientStore interface {

	SaveStoredData(name string,value []byte)

	LoadStoredData(name string) []byte

	LoadAccount()  map[Uint160]*Account

	LoadCcntmracts() map[Uint160]*ct.Ccntmract
}
