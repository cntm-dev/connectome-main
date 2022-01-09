/*
 * Copyright (C) 2018 The cntmology Authors
 * This file is part of The cntmology library.
 *
 * The cntmology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The cntmology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * alcntm with The cntmology.  If not, see <http://www.gnu.org/licenses/>.
 */

package ccntmract

import (
	"errors"
	"math/big"
	"sort"

	"github.com/Ontology/common"
	"github.com/Ontology/common/log"
	pg "github.com/Ontology/core/ccntmract/program"
	sig "github.com/Ontology/core/signature"
	_ "github.com/Ontology/errors"
	"github.com/cntmio/cntmology-crypto/keypair"
)

type CcntmractCcntmext struct {
	Data          sig.SignableData
	ProgramHashes []common.Address
	Codes         [][]byte
	Parameters    [][][]byte

	MultiPubkeyPara [][]PubkeyParameter

	//temp index for multi sig
	tempParaIndex int
}

func NewCcntmractCcntmext(data sig.SignableData) *CcntmractCcntmext {
	programHashes, _ := data.GetProgramHashes() //TODO: check error
	log.Debug("programHashes= ", programHashes)
	log.Debug("hashLen := len(programHashes) ", len(programHashes))
	hashLen := len(programHashes)
	return &CcntmractCcntmext{
		Data:            data,
		ProgramHashes:   programHashes,
		Codes:           make([][]byte, hashLen),
		Parameters:      make([][][]byte, hashLen),
		MultiPubkeyPara: make([][]PubkeyParameter, hashLen),
		tempParaIndex:   0,
	}
}

func (cxt *CcntmractCcntmext) Add(ccntmract *Ccntmract, index int, parameter []byte) error {
	log.Debug()
	i := cxt.GetIndex(ccntmract.ProgramHash)
	if i < 0 {
		log.Warn("Program Hash is not exist, using 0 by default")
		i = 0
	}
	if cxt.Codes[i] == nil {
		cxt.Codes[i] = ccntmract.Code
	}
	if cxt.Parameters[i] == nil {
		cxt.Parameters[i] = make([][]byte, len(ccntmract.Parameters))
	}
	cxt.Parameters[i][index] = parameter
	return nil
}

func (cxt *CcntmractCcntmext) AddCcntmract(ccntmract *Ccntmract, pubkey keypair.PublicKey, parameter []byte) error {
	log.Debug()
	if ccntmract.GetType() == MultiSigCcntmract {
		log.Debug()
		// add multi sig ccntmract

		log.Debug("Multi Sig: ccntmract.ProgramHash:", ccntmract.ProgramHash)
		log.Debug("Multi Sig: cxt.ProgramHashes:", cxt.ProgramHashes)

		index := cxt.GetIndex(ccntmract.ProgramHash)

		log.Debug("Multi Sig: GetIndex:", index)

		if index < 0 {
			log.Error("The program hash is not exist.")
			return errors.New("The program hash is not exist.")
		}

		log.Debug("Multi Sig: ccntmract.Code:", cxt.Codes[index])

		if cxt.Codes[index] == nil {
			cxt.Codes[index] = ccntmract.Code
		}
		log.Debug("Multi Sig: cxt.Codes[index]:", cxt.Codes[index])

		if cxt.Parameters[index] == nil {
			cxt.Parameters[index] = make([][]byte, len(ccntmract.Parameters))
		}
		log.Debug("Multi Sig: cxt.Parameters[index]:", cxt.Parameters[index])

		if err := cxt.Add(ccntmract, cxt.tempParaIndex, parameter); err != nil {
			return err
		}

		cxt.tempParaIndex++

		//all paramenters added, sort the parameters
		if cxt.tempParaIndex == len(ccntmract.Parameters) {
			cxt.tempParaIndex = 0
		}

		//TODO: Sort the parameter according ccntmract's PK list sequence
		//if err := cxt.AddSignatureToMultiList(index,ccntmract,pubkey,parameter); err != nil {
		//	return err
		//}
		//
		//if(cxt.tempParaIndex == len(ccntmract.Parameters)){
		//	//all multi sigs added, sort the sigs and add to ccntmext
		//	if err := cxt.AddMultiSignatures(index,ccntmract,pubkey,parameter);err != nil {
		//		return err
		//	}
		//}

	} else {
		//add non multi sig ccntmract
		log.Debug()
		index := -1
		for i := 0; i < len(ccntmract.Parameters); i++ {
			if ccntmract.Parameters[i] == Signature {
				if index >= 0 {
					return errors.New("Ccntmract Parameters are not supported.")
				} else {
					index = i
				}
			}
		}
		return cxt.Add(ccntmract, index, parameter)
	}
	return nil
}

func (cxt *CcntmractCcntmext) AddSignatureToMultiList(ccntmractIndex int, ccntmract *Ccntmract, pubkey keypair.PublicKey, parameter []byte) error {
	if cxt.MultiPubkeyPara[ccntmractIndex] == nil {
		cxt.MultiPubkeyPara[ccntmractIndex] = make([]PubkeyParameter, len(ccntmract.Parameters))
	}
	pk := keypair.SerializePublicKey(pubkey)

	pubkeyPara := PubkeyParameter{
		PubKey:    common.ToHexString(pk),
		Parameter: common.ToHexString(parameter),
	}
	cxt.MultiPubkeyPara[ccntmractIndex] = append(cxt.MultiPubkeyPara[ccntmractIndex], pubkeyPara)

	return nil
}

func (cxt *CcntmractCcntmext) AddMultiSignatures(index int, ccntmract *Ccntmract, pubkey keypair.PublicKey, parameter []byte) error {
	pkIndexs, err := cxt.ParseCcntmractPubKeys(ccntmract)
	if err != nil {
		return errors.New("Ccntmract Parameters are not supported.")
	}

	paraIndexs := []ParameterIndex{}
	for _, pubkeyPara := range cxt.MultiPubkeyPara[index] {
		pubKeyBytes, err := common.HexToBytes(pubkeyPara.Parameter)
		if err != nil {
			return errors.New("Ccntmract AddCcntmract pubKeyBytes HexToBytes failed.")
		}

		paraIndex := ParameterIndex{
			Parameter: pubKeyBytes,
			Index:     pkIndexs[pubkeyPara.PubKey],
		}
		paraIndexs = append(paraIndexs, paraIndex)
	}

	//sort parameter by Index
	sort.Sort(sort.Reverse(ParameterIndexSlice(paraIndexs)))

	//generate sorted parameter list
	for i, paraIndex := range paraIndexs {
		if err := cxt.Add(ccntmract, i, paraIndex.Parameter); err != nil {
			return err
		}
	}

	cxt.MultiPubkeyPara[index] = nil

	return nil
}

func (cxt *CcntmractCcntmext) ParseCcntmractPubKeys(ccntmract *Ccntmract) (map[string]int, error) {

	pubkeyIndex := make(map[string]int)

	Index := 0
	//parse ccntmract's pubkeys
	i := 0
	switch ccntmract.Code[i] {
	case 1:
		i += 2
		break
	case 2:
		i += 3
		break
	}
	for ccntmract.Code[i] == 33 {
		i++

		//add to parameter index
		pubkeyIndex[common.ToHexString(ccntmract.Code[i:33])] = Index

		i += 33
		Index++
	}

	return pubkeyIndex, nil
}

func (cxt *CcntmractCcntmext) GetIndex(programHash common.Address) int {
	for i := 0; i < len(cxt.ProgramHashes); i++ {
		if cxt.ProgramHashes[i] == programHash {
			return i
		}
	}
	return -1
}

func (cxt *CcntmractCcntmext) GetPrograms() []*pg.Program {
	log.Debug()
	//log.Debug("!cxt.IsCompleted()=",!cxt.IsCompleted())
	//log.Debug(cxt.Codes)
	//log.Debug(cxt.Parameters)
	if !cxt.IsCompleted() {
		return nil
	}
	programs := make([]*pg.Program, len(cxt.Parameters))

	log.Debug(" len(cxt.Codes)", len(cxt.Codes))

	for i := 0; i < len(cxt.Codes); i++ {
		sb := pg.NewProgramBuilder()

		for _, parameter := range cxt.Parameters[i] {
			if len(parameter) <= 2 {
				sb.PushNumber(new(big.Int).SetBytes(parameter))
			} else {
				sb.PushData(parameter)
			}
		}
		//log.Debug(" cxt.Codes[i])", cxt.Codes[i])
		//log.Debug(" sb.ToArray()", sb.ToArray())
		programs[i] = &pg.Program{
			Code:      cxt.Codes[i],
			Parameter: sb.ToArray(),
		}
	}
	return programs
}

func (cxt *CcntmractCcntmext) IsCompleted() bool {
	for _, p := range cxt.Parameters {
		if p == nil {
			return false
		}

		for _, pp := range p {
			if pp == nil {
				return false
			}
		}
	}
	return true
}
