package neovm

func opPushData(e *ExecutionEngine) (VMState, error) {
	data := getPushData(e)
	PushData(e, data)
	return NONE, nil
}

func getPushData(e *ExecutionEngine) interface{} {
	var data interface{}
	if e.opCode >= PUSHBYTES1 && e.opCode <= PUSHBYTES75 {
		data = e.ccntmext.OpReader.ReadBytes(int(e.opCode))
	}
	switch e.opCode {
	case PUSH0:
		data = []byte{}
	case PUSHDATA1:
		d, _ := e.ccntmext.OpReader.ReadByte()
		data = e.ccntmext.OpReader.ReadBytes(int(d))
	case PUSHDATA2:
		data = e.ccntmext.OpReader.ReadBytes(int(e.ccntmext.OpReader.ReadUint16()))
	case PUSHDATA4:
		i := int(e.ccntmext.OpReader.ReadInt32())
		data = e.ccntmext.OpReader.ReadBytes(i)
	case PUSHM1, PUSH1, PUSH2, PUSH3, PUSH4, PUSH5, PUSH6, PUSH7, PUSH8, PUSH9, PUSH10, PUSH11, PUSH12, PUSH13, PUSH14, PUSH15, PUSH16:
		data = int8(e.opCode - PUSH1 + 1)
	}

	return data
}
