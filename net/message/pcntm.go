package message

import (
	. "DNA/net/protocol"
)

type pcntm struct {
	msgHdr
	Nonce uint64
}

func NewPcntmMsg() ([]byte, error) {
	var msg pcntm
	var sum []byte
	sum = []byte{0x5d, 0xf6, 0xe0, 0xe2}
	msg.msgHdr.init("pcntm", sum, 0)

	buf, err := msg.Serialization()
	if err != nil {
		return nil, err
	}
	return buf, err
}

func (msg pcntm) Verify(buf []byte) error {
	err := msg.msgHdr.Verify(buf)
	// TODO verify the message Ccntment
	return err
}

func (msg pcntm) Handle(node Noder) error {
	node.SetLastCcntmact()
	return nil
}
