// Copyright (C) 2021 The Ontology Authors
package runtime

import (
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/holiman/uint256"
	"github.com/cntmio/cntmology/common/log"
)

type Ccntmract struct {
	Abi        abi.ABI
	Cfg        *Config
	Address    common.Address
	AutoCommit bool
}

func Ensure(err error) {
	if err != nil {
		panic(err)
	}
}

func Create2Ccntmract(cfg *Config, jsonABI, hexCode string, salt uint64, params ...interface{}) *Ccntmract {
	ccntmractBin, err := hexutil.Decode(hexCode)
	Ensure(err)

	cabi, err := abi.JSON(strings.NewReader(jsonABI))
	Ensure(err)
	var p []byte
	if len(params) != 0 {
		p, err = cabi.Pack("", params...)
		Ensure(err)
	}
	deploy := append(ccntmractBin, p...)

	_, ctAddr, leftGas, err := Create2(deploy, cfg, uint256.NewInt().SetUint64(salt))
	Ensure(err)

	log.Infof("deploy code at: %s, used gas: %d", ctAddr.String(), cfg.GasLimit-leftGas)

	return &Ccntmract{
		Abi:     cabi,
		Cfg:     cfg,
		Address: ctAddr,
	}
}

func CreateCcntmract(cfg *Config, jsonABI string, hexCode string, params ...interface{}) *Ccntmract {
	ccntmractBin, err := hexutil.Decode(hexCode)
	Ensure(err)

	cabi, err := abi.JSON(strings.NewReader(jsonABI))
	Ensure(err)
	var p []byte
	if len(params) != 0 {
		p, err = cabi.Pack("", params...)
		Ensure(err)
	}
	deploy := append(ccntmractBin, p...)

	_, ctAddr, leftGas, err := Create(deploy, cfg)
	Ensure(err)

	log.Infof("deploy code at: %s, used gas: %d", ctAddr.String(), cfg.GasLimit-leftGas)

	return &Ccntmract{
		Abi:     cabi,
		Cfg:     cfg,
		Address: ctAddr,
	}
}

func NewCcntmract(cfg *Config, jsonABI string, addr common.Address) *Ccntmract {
	cabi, err := abi.JSON(strings.NewReader(jsonABI))
	Ensure(err)

	return &Ccntmract{
		Abi:     cabi,
		Cfg:     cfg,
		Address: addr,
	}
}

func (self *Ccntmract) Call(method string, params ...interface{}) ([]byte, uint64, error) {
	input, err := self.Abi.Pack(method, params...)
	Ensure(err)

	ret, gas, err := Call(self.Address, input, self.Cfg)
	if self.AutoCommit {
		err := self.Cfg.State.Commit()
		Ensure(err)
	}
	return ret, self.Cfg.GasLimit - gas, err
}

func (self *Ccntmract) Balance() *big.Int {
	return self.Cfg.State.GetBalance(self.Address)
}

func (self *Ccntmract) BalanceOf(addr common.Address) *big.Int {
	return self.Cfg.State.GetBalance(addr)
}
