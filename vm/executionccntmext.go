package vm

import (
)

type ScriptCcntmext struct {
	Script []byte
	OpReader * VmReader
	//BreakPoints
}

func NewScriptCcntmext(script []byte) *ScriptCcntmext {
	var stackCcntmext ScriptCcntmext
	stackCcntmext.Script = script
	stackCcntmext.OpReader = NewVmReader( script )
	return &stackCcntmext
}
/*
func (sc *ScriptCcntmext) Dispose() {
	sc.OpReader.Dispose();

}
*/