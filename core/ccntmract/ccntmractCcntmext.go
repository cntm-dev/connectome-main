package ccntmract

import (
	. "DNA/common"
	"DNA/common/log"
	pg "DNA/core/ccntmract/program"
	sig "DNA/core/signature"
	"DNA/crypto"
	_ "DNA/errors"
	"errors"
	"math/big"
	"sort"
)

type CcntmractCcntmext struct {
	Data          sig.SignableData
	ProgramHashes []Uint160
	Codes         [][]byte
	Parameters    [][][]byte

	MultiPubkeyPara [][]PubkeyParameter

	//temp index for multi sig
	tempParaIndex int
}

func NewCcntmractCcntmext(data sig.SignableData) *CcntmractCcntmext {
	log.Trace()
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
	log.Trace()
	i := cxt.GetIndex(ccntmract.ProgramHash)
	if i < 0 {
		return errors.New("Program Hash is not exist.")
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

func (cxt *CcntmractCcntmext) AddCcntmract(ccntmract *Ccntmract, pubkey *crypto.PubKey, parameter []byte) error {
	log.Trace()
	if ccntmract.GetType() == MultiSigCcntmract {
		log.Trace()
		// add multi sig ccntmract

		log.Debug("Multi Sig: ccntmract.ProgramHash:", ccntmract.ProgramHash)
		log.Debug("Multi Sig: cxt.ProgramHashes:", cxt.ProgramHashes)

		index := cxt.GetIndex(ccntmract.ProgramHash)

		log.Debug("Multi Sig: GetIndex:", index)

		if index < 0 {
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
		log.Trace()
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

func (cxt *CcntmractCcntmext) AddSignatureToMultiList(ccntmractIndex int, ccntmract *Ccntmract, pubkey *crypto.PubKey, parameter []byte) error {
	if cxt.MultiPubkeyPara[ccntmractIndex] == nil {
		cxt.MultiPubkeyPara[ccntmractIndex] = make([]PubkeyParameter, len(ccntmract.Parameters))
	}
	pk, err := pubkey.EncodePoint(true)
	if err != nil {
		return err
	}

	pubkeyPara := PubkeyParameter{
		PubKey:    ToHexString(pk),
		Parameter: ToHexString(parameter),
	}
	cxt.MultiPubkeyPara[ccntmractIndex] = append(cxt.MultiPubkeyPara[ccntmractIndex], pubkeyPara)

	return nil
}

func (cxt *CcntmractCcntmext) AddMultiSignatures(index int, ccntmract *Ccntmract, pubkey *crypto.PubKey, parameter []byte) error {
	pkIndexs, err := cxt.ParseCcntmractPubKeys(ccntmract)
	if err != nil {
		return errors.New("Ccntmract Parameters are not supported.")
	}

	paraIndexs := []ParameterIndex{}
	for _, pubkeyPara := range cxt.MultiPubkeyPara[index] {
		pubKeyBytes, err := HexToBytes(pubkeyPara.Parameter)
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
		//pubkey, err := crypto.DecodePoint(ccntmract.Code[i:33])
		//if err != nil {
		//	return nil, errors.New("[Ccntmract],AddCcntmract DecodePoint failed.")
		//}

		//add to parameter index
		pubkeyIndex[ToHexString(ccntmract.Code[i:33])] = Index

		i += 33
		Index++
	}

	return pubkeyIndex, nil
}

func (cxt *CcntmractCcntmext) GetIndex(programHash Uint160) int {
	for i := 0; i < len(cxt.ProgramHashes); i++ {
		if cxt.ProgramHashes[i] == programHash {
			return i
		}
	}
	return -1
}

func (cxt *CcntmractCcntmext) GetPrograms() []*pg.Program {
	log.Trace()
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
