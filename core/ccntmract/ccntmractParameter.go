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
	PublicKey
	String
	Array = 0x10
	InteropInterface = 0xf0
	Void = 0xff
)
