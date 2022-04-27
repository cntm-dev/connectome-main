package wasmvm

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"math/big"

	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/core/store"
	"github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology/errors"
	sccommon "github.com/cntmio/cntmology/smartccntmract/common"
	"github.com/cntmio/cntmology/smartccntmract/ccntmext"
	"github.com/cntmio/cntmology/smartccntmract/event"
	nstates "github.com/cntmio/cntmology/smartccntmract/service/native/states"
	"github.com/cntmio/cntmology/smartccntmract/states"
	"github.com/cntmio/cntmology/smartccntmract/storage"
	vmtypes "github.com/cntmio/cntmology/smartccntmract/types"
	"github.com/cntmio/cntmology/vm/wasmvm/exec"
	"github.com/cntmio/cntmology/vm/wasmvm/util"
)

type WasmVmService struct {
	Store         store.LedgerStore
	CloneCache    *storage.CloneCache
	CcntmextRef    ccntmext.CcntmextRef
	Notifications []*event.NotifyEventInfo
	Code          []byte
	Tx            *types.Transaction
	Time          uint32
}

func (this *WasmVmService) Invoke() (interface{}, error) {
	stateMachine := NewWasmStateMachine(this.Store, this.CloneCache, this.Time)
	//register the "CallCcntmract" function
	stateMachine.Register("cntm_CallCcntmract", this.callCcntmract)
	stateMachine.Register("cntm_MarshalNativeParams", this.marshalNativeParams)
	//runtime
	stateMachine.Register("cntm_Runtime_CheckWitness", this.runtimeCheckWitness)
	stateMachine.Register("cntm_Runtime_Notify", this.runtimeNotify)
	stateMachine.Register("cntm_Runtime_CheckSig", this.runtimeCheckSig)
	stateMachine.Register("cntm_Runtime_GetTime", this.runtimeGetTime)
	stateMachine.Register("cntm_Runtime_Log", this.runtimeLog)
	//attribute
	stateMachine.Register("cntm_Attribute_GetUsage", this.attributeGetUsage)
	stateMachine.Register("cntm_Attribute_GetData", this.attributeGetData)
	//block
	stateMachine.Register("cntm_Block_GetCurrentHeaderHash", this.blockGetCurrentHeaderHash)
	stateMachine.Register("cntm_Block_GetCurrentHeaderHeight", this.blockGetCurrentHeaderHeight)
	stateMachine.Register("cntm_Block_GetCurrentBlockHash", this.blockGetCurrentBlockHash)
	stateMachine.Register("cntm_Block_GetCurrentBlockHeight", this.blockGetCurrentBlockHeight)
	stateMachine.Register("cntm_Block_GetTransactionByHash", this.blockGetTransactionByHash)
	stateMachine.Register("cntm_Block_GetTransactionCount", this.blockGetTransactionCount)
	stateMachine.Register("cntm_Block_GetTransactions", this.blockGetTransactions)

	//blockchain
	stateMachine.Register("cntm_BlockChain_GetHeight", this.blockChainGetHeight)
	stateMachine.Register("cntm_BlockChain_GetHeaderByHeight", this.blockChainGetHeaderByHeight)
	stateMachine.Register("cntm_BlockChain_GetHeaderByHash", this.blockChainGetHeaderByHash)
	stateMachine.Register("cntm_BlockChain_GetBlockByHeight", this.blockChainGetBlockByHeight)
	stateMachine.Register("cntm_BlockChain_GetBlockByHash", this.blockChainGetBlockByHash)
	stateMachine.Register("cntm_BlockChain_GetCcntmract", this.blockChainGetCcntmract)

	//header
	stateMachine.Register("cntm_Header_GetHash", this.headerGetHash)
	stateMachine.Register("cntm_Header_GetVersion", this.headerGetVersion)
	stateMachine.Register("cntm_Header_GetPrevHash", this.headerGetPrevHash)
	stateMachine.Register("cntm_Header_GetMerkleRoot", this.headerGetMerkleRoot)
	stateMachine.Register("cntm_Header_GetIndex", this.headerGetIndex)
	stateMachine.Register("cntm_Header_GetTimestamp", this.headerGetTimestamp)
	stateMachine.Register("cntm_Header_GetConsensusData", this.headerGetConsensusData)
	stateMachine.Register("cntm_Header_GetNextConsensus", this.headerGetNextConsensus)

	//storage
	stateMachine.Register("cntm_Storage_Put", this.putstore)
	stateMachine.Register("cntm_Storage_Get", this.getstore)
	stateMachine.Register("cntm_Storage_Delete", this.deletestore)

	//transaction
	stateMachine.Register("cntm_Transaction_GetHash", this.transactionGetHash)
	stateMachine.Register("cntm_Transaction_GetType", this.transactionGetType)
	stateMachine.Register("cntm_Transaction_GetAttributes", this.transactionGetAttributes)

	engine := exec.NewExecutionEngine(
		this.Tx,
		new(util.ECDsaCrypto),
		stateMachine,
	)

	ccntmract := &states.Ccntmract{}
	ccntmract.Deserialize(bytes.NewBuffer(this.Code))
	addr := ccntmract.Address
	if ccntmract.Code == nil {
		dpcode, err := this.GetCcntmractCodeFromAddress(addr)
		if err != nil {
			return nil, errors.NewErr("get ccntmract  error")
		}
		ccntmract.Code = dpcode
	}

	var caller common.Address
	if this.CcntmextRef.CallingCcntmext() == nil {
		caller = common.Address{}
	} else {
		caller = this.CcntmextRef.CallingCcntmext().CcntmractAddress
	}
	this.CcntmextRef.PushCcntmext(&ccntmext.Ccntmext{CcntmractAddress: ccntmract.Address})
	res, err := engine.Call(caller, ccntmract.Code, ccntmract.Method, ccntmract.Args, ccntmract.Version)

	if err != nil {
		return nil, err
	}

	//get the return message
	result, err := engine.GetVM().GetPointerMemory(uint64(binary.LittleEndian.Uint32(res)))
	if err != nil {
		return nil, err
	}

	this.CcntmextRef.PopCcntmext()
	this.CcntmextRef.PushNotifications(this.Notifications)
	return result, nil
}

// marshalNativeParams
// make paramter bytes for call native ccntmract
func (this *WasmVmService) marshalNativeParams(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 1 {
		return false, errors.NewErr("[callCcntmract]parameter count error while call marshalNativeParams")
	}

	transferbytes, err := vm.GetPointerMemory(params[0])
	if err != nil {
		return false, err
	}
	//transferbytes is a nested struct with states.Transfer
	//type Transfers struct {
	//	Version byte               -------->i32  4 bytes
	//	States  []*State		   -------->i32 pointer 4 bytes
	//}
	if len(transferbytes) != 8 {
		return false, errors.NewErr("[callCcntmract]parameter format error while call marshalNativeParams")
	}
	transfer := &nstates.Transfers{}
	tver := binary.LittleEndian.Uint32(transferbytes[:4])
	transfer.Version = byte(tver)

	statesAddr := binary.LittleEndian.Uint32(transferbytes[4:])
	statesbytes, err := vm.GetPointerMemory(uint64(statesAddr))
	if err != nil {
		return false, err
	}

	//statesbytes is slice of struct with states.
	//type State struct {
	//	Version byte            -------->i32 4 bytes
	//	From    common.Address  -------->i32 pointer 4 bytes
	//	To      common.Address  -------->i32 pointer 4 bytes extra padding 4 bytes
	//	Value   *big.Int        -------->i64 8 bytes
	//}
	//total is 4 + 4 + 4 + 4(dummy) + 8 = 24 bytes
	statecnt := len(statesbytes) / 24
	states := make([]*nstates.State, statecnt)

	for i := 0; i < statecnt; i++ {
		tmpbytes := statesbytes[i * 24 : (i + 1) * 24]
		state := &nstates.State{}
		state.Version = byte(binary.LittleEndian.Uint32(tmpbytes[:4]))
		fromAddessBytes, err := vm.GetPointerMemory(uint64(binary.LittleEndian.Uint32(tmpbytes[4:8])))
		if err != nil {
			return false, err
		}
		fromAddress, err := common.AddressFromBase58(util.TrimBuffToString(fromAddessBytes))
		if err != nil {
			return false, err
		}
		state.From = fromAddress

		toAddressBytes, err := vm.GetPointerMemory(uint64(binary.LittleEndian.Uint32(tmpbytes[8:12])))
		if err != nil {
			return false, err
		}
		toAddress, err := common.AddressFromBase58(util.TrimBuffToString(toAddressBytes))
		state.To = toAddress
		//tmpbytes[12:16] is padding
		amount := binary.LittleEndian.Uint64(tmpbytes[16:])
		state.Value = big.NewInt(int64(amount))
		states[i] = state

	}

	transfer.States = states
	tbytes := new(bytes.Buffer)
	transfer.Serialize(tbytes)

	result, err := vm.SetPointerMemory(tbytes.Bytes())
	if err != nil {
		return false, err
	}
	vm.RestoreCtx()
	vm.PushResult(uint64(result))
	return true, nil
}


// callCcntmract
// need 4 paramters
//0: ccntmract address
//1: ccntmract code
//2: method name
//3: args
func (this *WasmVmService) callCcntmract(engine *exec.ExecutionEngine) (bool, error) {
	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 4 {
		return false, errors.NewErr("[callCcntmract]parameter count error while call readMessage")
	}
	var ccntmractAddress common.Address
	var ccntmractBytes []byte
	//get ccntmract address
	ccntmractAddressIdx := params[0]
	addr, err := vm.GetPointerMemory(ccntmractAddressIdx)
	if err != nil {
		return false, errors.NewErr("[callCcntmract]get Ccntmract address failed:" + err.Error())
	}

	if addr != nil {
		addrbytes, err := common.HexToBytes(util.TrimBuffToString(addr))
		if err != nil {
			return false, errors.NewErr("[callCcntmract]get ccntmract address error:" + err.Error())
		}
		ccntmractAddress, err = common.AddressParseFromBytes(addrbytes)
		if err != nil {
			return false, errors.NewErr("[callCcntmract]get ccntmract address error:" + err.Error())
		}

	}

	//get ccntmract code
	codeIdx := params[1]

	offchainCcntmractCode, err := vm.GetPointerMemory(codeIdx)
	if err != nil {
		return false, errors.NewErr("[callCcntmract]get Ccntmract address failed:" + err.Error())
	}
	if offchainCcntmractCode != nil {
		ccntmractBytes, err = common.HexToBytes(util.TrimBuffToString(offchainCcntmractCode))
		if err != nil {
			return false, err

		}
		//compute the offchain code address
		codestring := util.TrimBuffToString(offchainCcntmractCode)
		ccntmractAddress = GetCcntmractAddress(codestring, vmtypes.WASMVM)
	}
	//get method
	methodName, err := vm.GetPointerMemory(params[2])
	if err != nil {
		return false, errors.NewErr("[callCcntmract]get Ccntmract methodName failed:" + err.Error())
	}
	//get args
	arg, err := vm.GetPointerMemory(params[3])

	if err != nil {
		return false, errors.NewErr("[callCcntmract]get Ccntmract arg failed:" + err.Error())
	}
	this.CcntmextRef.PushCcntmext(&ccntmext.Ccntmext{
		Code: vm.VMCode,
		CcntmractAddress: vm.CcntmractAddress})
	result, err := this.CcntmextRef.AppCall(ccntmractAddress, util.TrimBuffToString(methodName), ccntmractBytes, arg)
	this.CcntmextRef.PopCcntmext()
	if err != nil {
		return false, errors.NewErr("[callCcntmract]AppCall failed:" + err.Error())
	}
	vm.RestoreCtx()
	if envCall.GetReturns() {
		if ccntmractAddress[0] == byte(vmtypes.NEOVM) {
			result = sccommon.ConvertNeoVmReturnTypes(result)
		}
		if ccntmractAddress[0] == byte(vmtypes.Native) {
			bresult := result.(bool)
			if bresult == true {
				result = "true"
			} else {
				result = false
			}

		}
		if ccntmractAddress[0] == byte(vmtypes.WASMVM) {
			//reserve for further process
		}

		idx, err := vm.SetPointerMemory(result.(string))
		if err != nil {
			return false, errors.NewErr("[callCcntmract]SetPointerMemory failed:" + err.Error())
		}
		vm.PushResult(uint64(idx))
	}

	return true, nil
}

func (this *WasmVmService) GetCcntmractCodeFromAddress(address common.Address) ([]byte, error) {

	dcode, err := this.Store.GetCcntmractState(address)
	if err != nil {
		return nil, err
	}

	if dcode == nil {
		return nil, errors.NewErr("[GetCcntmractCodeFromAddress] deployed code is nil")
	}

	return dcode.Code.Code, nil

}

func (this *WasmVmService) getCcntmractFromAddr(addr []byte) ([]byte, error) {
	addrbytes, err := common.HexToBytes(util.TrimBuffToString(addr))
	if err != nil {
		return nil, errors.NewErr("get ccntmract address error")
	}
	ccntmactaddress, err := common.AddressParseFromBytes(addrbytes)
	if err != nil {
		return nil, errors.NewErr("get ccntmract address error")
	}
	dpcode, err := this.GetCcntmractCodeFromAddress(ccntmactaddress)
	if err != nil {
		return nil, errors.NewErr("get ccntmract  error")
	}
	return dpcode, nil
}

//GetCcntmractAddress return ccntmract address
func GetCcntmractAddress(code string, vmType vmtypes.VmType) common.Address {
	data, _ := hex.DecodeString(code)
	vmCode := &vmtypes.VmCode{
		VmType: vmType,
		Code:   data,
	}
	return vmCode.AddressFromVmCode()
}
