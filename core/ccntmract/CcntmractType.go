package ccntmract

type CcntmractType byte

const (
	SignatureCcntmract CcntmractType = iota
	MultiSigCcntmract
	CustomCcntmract
)