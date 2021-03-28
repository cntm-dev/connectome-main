package ccntmract

import (
	"GoOnchain/common"
	sig "GoOnchain/core/signature"
	"GoOnchain/crypto"
	"GoOnchain/core/ccntmract/program"
)

type CcntmractCcntmext struct {
	//TODO: define CcntmractCcntmext。
	Data sig.SignableData
	ProgramHashes []common.Uint160
	Programs [][]byte
	Parameters [][][]byte
}


func NewCcntmractCcntmext(data sig.SignableData) *CcntmractCcntmext {

	programHashes,_ := data.GetProgramHashes() //TODO: check error
	hashLen := len(programHashes)

	return &CcntmractCcntmext{
		Data: data,
		ProgramHashes: programHashes,
		Programs: make([][]byte,hashLen),
		Parameters: make([][][]byte,hashLen),
	}
}

func (cxt *CcntmractCcntmext) AddCcntmract(ccntmract *Ccntmract, pubkey *crypto.PubKey,paramenter []byte ) error {
	//TODO: implement AddCcntmract()

	//TODO: check ccntmract type for diff building
	return  nil
}


func (cxt *CcntmractCcntmext) GetPrograms() ([]*program.Program) {
	//TODO: implement GetProgram()

	return  []*program.Program{}

}
