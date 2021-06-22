package ccntmract

//parameter defined type.
type CcntmractParameterType byte

const (
	Signature CcntmractParameterType = iota
	Boolean
	Integer
	Hash160
	Hash256
	ByteArray
)
