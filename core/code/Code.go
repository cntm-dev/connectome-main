package code

import (
	."github.com/DNAProject/DNA/common"
	."github.com/DNAProject/DNA/core/ccntmract"
)
//ICode is the abstract interface of smart ccntmract code.
type ICode interface {

	GetCode() []byte

	GetParameterTypes() []CcntmractParameterType

	GetReturnTypes() []CcntmractParameterType

	CodeHash() Uint160

}

