package genesis

import (
	"errors"
	"time"

	"github.com/Ontology/common"
	"github.com/Ontology/common/config"
	"github.com/Ontology/core/code"
	"github.com/Ontology/core/ccntmract/program"
	"github.com/Ontology/core/types"
	"github.com/Ontology/core/utils"
	"github.com/Ontology/crypto"
	vm "github.com/Ontology/vm/neovm"
	vmtypes "github.com/Ontology/vm/types"
)

const (
	BlockVersion      uint32 = 0
	GenesisNonce      uint64 = 2083236893
	DecrementInterval uint32 = 2000000

	OntRegisterAmount = 1000000000
	OngRegisterAmount = 1000000000
)

var (
	GenerationAmount = [17]uint32{80, 70, 60, 50, 40, 30, 20, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10}

	cntmToken   = NewGoverningToken()
	cntmToken   = NewUtilityToken()
	cntmTokenID = cntmToken.Hash()
	cntmTokenID = cntmToken.Hash()
)

var GenBlockTime = (config.DEFAULTGENBLOCKTIME * time.Second)

var GenesisBookKeepers []*crypto.PubKey

func GenesisBlockInit(defaultBookKeeper []*crypto.PubKey) (*types.Block, error) {
	//getBookKeeper
	GenesisBookKeepers = defaultBookKeeper
	nextBookKeeper, err := utils.AddressFromBookKeepers(defaultBookKeeper)
	if err != nil {
		return nil, errors.New("[Block],GenesisBlockInit err with GetBookKeeperAddress")
	}
	//blockdata
	genesisHeader := &types.Header{
		Version:          BlockVersion,
		PrevBlockHash:    common.Uint256{},
		TransactionsRoot: common.Uint256{},
		Timestamp:        uint32(uint32(time.Date(2017, time.February, 23, 0, 0, 0, 0, time.UTC).Unix())),
		Height:           uint32(0),
		ConsensusData:    GenesisNonce,
		NextBookKeeper:   nextBookKeeper,
		Program: &program.Program{
			Code:      []byte{},
			Parameter: []byte{byte(vm.PUSHT)},
		},
	}

	//block
	cntm := NewGoverningToken()
	cntm := NewUtilityToken()

	genesisBlock := &types.Block{
		Header: genesisHeader,
		Transactions: []*types.Transaction{
			cntm,
			cntm,
		},
	}
	return genesisBlock, nil
}

func NewGoverningToken() *types.Transaction {
	fnCode := code.FunctionCode{
		Code: []byte("cntm Token"),
	}

	tx := utils.NewDeployTransaction(&fnCode, "cntm", "0.1.0",
		"Ontology", "", "Ontology Network cntm Token", vmtypes.NativeVM, true)
	return tx
}

func NewUtilityToken() *types.Transaction {
	fnCode := code.FunctionCode{
		Code: []byte("cntm Token"),
	}

	tx := utils.NewDeployTransaction(&fnCode, "cntm", "0.1.0",
		"Ontology", "", "Ontology Network cntm Token", vmtypes.NativeVM, true)
	return tx
}

func NewInitSystemTokenTransaction() *types.Transaction {
	// invoke transaction to init cntm/cntm token
	return nil
}
