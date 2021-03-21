package ccntmract

import (
	"GoOnchain/common"
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
	ProgramHash common.Uint160

	//owner's pubkey hash indicate the owner of ccntmract
	OwnerPubkeyHash common.Uint160

}


