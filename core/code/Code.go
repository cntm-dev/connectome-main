package code

import (
	. "DNA/common"
	. "DNA/core/ccntmract"
)
//ICode is the abstract interface of smart ccntmract code.
type ICode interface {

	GetCode() []byte

	GetParameterTypes() []CcntmractParameterType

	GetReturnTypes() []CcntmractParameterType

	CodeHash() Uint160

}

