package ccntmract

import (
	. "GoOnchain/common"
	"GoOnchain/vm"
)

//Ccntmract address is the hash of ccntmract program .
//which be used to ccntmrol asset or indicate the smart ccntmract address �?


//Ccntmract include the program codes with parameters which can be executed on specific evnrioment
type Ccntmract struct {

	//the ccntmract program code,which will be run on VM or specific envrionment
	Code []byte

	//the Ccntmract Parameter type list
	// describe the number of ccntmract program parameters and the parameter type
	Parameters []CcntmractParameterType

	//The program hash as ccntmract address
	ProgramHash Uint160

	//owner's pubkey hash indicate the owner of ccntmract
	OwnerPubkeyHash Uint160

}

func (c *Ccntmract) IsStandard() bool {
	if len(c.Code) != 35 {
		return false
	}
	if c.Code[0] != 33 || c.Code[34] != byte(vm.OP_CHECKSIG) {
		return false
	}
	return true
}

func (c *Ccntmract) IsMultiSigCcntmract() bool {
	var m int16 = 0
	var n int16 = 0
	i := 0

	if len(c.Code) < 37 {return false}
	if c.Code[i] > byte(vm.OP_16) {return false}
	if c.Code[i] < byte(vm.OP_1) && c.Code[i] != 1 && c.Code[i] != 2 {
		return false
	}

	switch c.Code[i] {
	case 1:
		i++
		m = int16(c.Code[i])
		i++
		break
	case 2:
		i++
		m = BytesToInt16(c.Code[i:])
		i += 2
		break
	default:
		m = int16(c.Code[i]) - 80
		i++
		break
	}

	if m < 1 || m > 1024 {return false}

	for c.Code[i] == 33 {
		i += 34
		if len(c.Code) <= i {return false}
		n++
	}
	if n < m || n > 1024 {return false}

	switch c.Code[i] {
	case 1:
		i++
		if n != int16(c.Code[i]) {return false}
		i++
		break
	case 2:
		i++
		if n != BytesToInt16(c.Code[i:]) {return false}
		i += 2
		break
	default:
		if n != (int16(c.Code[i]) - 80) {return false}
		i++
		break
	}

	if c.Code[i] != byte(vm.OP_CHECKMULTISIG) {return false}
	i++
	if len(c.Code) != i {return false}

	return true
}

func (c *Ccntmract) GetType() CcntmractType{
	if c.IsStandard() {
		return SignatureCcntmract
	}
	if c.IsMultiSigCcntmract() {
		return MultiSigCcntmract
	}
	return CustomCcntmract
}

