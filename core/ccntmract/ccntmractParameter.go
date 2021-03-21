package ccntmract

//parameter defined type.
type CcntmractParameterType byte

const (
	Signature CcntmractParameterType = iota
	Integer
	Hash160
	Hash256
	ByteArray
)

