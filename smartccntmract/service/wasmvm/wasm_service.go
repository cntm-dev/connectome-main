package wasmvm

import (
	"bytes"
	"encoding/binary"

	"github.com/cntmio/cntmology/core/store"
	"github.com/cntmio/cntmology/smartccntmract/storage"
	"github.com/cntmio/cntmology/smartccntmract/ccntmext"
	"github.com/cntmio/cntmology/smartccntmract/event"
	"github.com/cntmio/cntmology/core/types"
	scommon "github.com/cntmio/cntmology/core/store/common"
	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/smartccntmract/states"
	"github.com/cntmio/cntmology/errors"
	"github.com/cntmio/cntmology/vm/wasmvm/exec"
	"github.com/cntmio/cntmology/vm/wasmvm/util"
)

type WasmVmService struct {
	Store         store.LedgerStore
	CloneCache    *storage.CloneCache
	CcntmextRef    ccntmext.CcntmextRef
	Notifications []*event.NotifyEventInfo
	Tx            *types.Transaction
	Time          uint32
}

func NewWasmVmService(store store.LedgerStore,
						dbCache scommon.StateStore,
						tx *types.Transaction,
						time uint32,
						ctxRef ccntmext.CcntmextRef) *WasmVmService {
	var service WasmVmService
	service.Store = store
	service.CloneCache = storage.NewCloneCache(dbCache)
	service.Time = time
	service.Tx = tx
	service.CcntmextRef = ctxRef
	return &service
}

func (this *WasmVmService)Invoke() ([]byte,error){
	stateMachine := NewWasmStateMachine(this.Store, this.CloneCache,  this.Time)
	//register the "CallCcntmract" function
	stateMachine.Register("CallCcntmract",this.callCcntmract)

	ctx := this.CcntmextRef.CurrentCcntmext()
	engine := exec.NewExecutionEngine(
		this.Tx,
		new(util.ECDsaCrypto),
		stateMachine,
	)

	ccntmract := &states.Ccntmract{}
	ccntmract.Deserialize(bytes.NewBuffer(ctx.Code.Code))
	addr := ccntmract.Address

	if ccntmract.Code == nil{
		dpcode, err := this.GetCcntmractCodeFromAddress(addr)
		if err != nil {
			return nil,errors.NewErr("get ccntmract  error")
		}
		ccntmract.Code = dpcode
	}

	var caller common.Address
	if this.CcntmextRef.CallingCcntmext() == nil {
		caller = common.Address{}
	} else {
		caller = this.CcntmextRef.CallingCcntmext().CcntmractAddress
	}
	res, err := engine.Call(caller, ccntmract.Code, ccntmract.Method, ccntmract.Args, ccntmract.Version)

	if err != nil {
		return nil,err
	}

	//get the return message
	result, err := engine.GetVM().GetPointerMemory(uint64(binary.LittleEndian.Uint32(res)))
	if err != nil {
		return nil,err
	}

	this.CloneCache.Commit()
	this.CcntmextRef.PushNotifications(stateMachine.Notifications)
	return result,nil
}

func(this *WasmVmService)callCcntmract(engine *exec.ExecutionEngine)(bool,error){

	vm := engine.GetVM()
	envCall := vm.GetEnvCall()
	params := envCall.GetParams()
	if len(params) != 3 {
		return false, errors.NewErr("[callCcntmract]parameter count error while call readMessage")
	}
	ccntmractAddressIdx := params[0]
	addr, err := vm.GetPointerMemory(ccntmractAddressIdx)
	if err != nil {
		return false, errors.NewErr("[callCcntmract]get Ccntmract address failed:" + err.Error())
	}
	//the ccntmract codes
	//ccntmractBytes, err := this.getCcntmractFromAddr(addr)
	//if err != nil {
	//	return false, err
	//}

	addrbytes, err := common.HexToBytes(util.TrimBuffToString(addr))
	if err != nil {
		return false, errors.NewErr("[callCcntmract]get ccntmract address error:" + err.Error())
	}
	ccntmractAddress, err := common.AddressParseFromBytes(addrbytes)
	if err != nil {
		return false, errors.NewErr("[callCcntmract]get ccntmract address error:" + err.Error())
	}

	//vmcode := stype.VmCode{VmType:stype.WASMVM,Code:ccntmractBytes}
	//ccntmractAddress := vmcode.AddressFromVmCode()

	methodName, err := vm.GetPointerMemory(params[1])
	if err != nil {
		return false, errors.NewErr("[callCcntmract]get Ccntmract methodName failed:" + err.Error())
	}

	arg, err := vm.GetPointerMemory(params[2])
	if err != nil {
		return false, errors.NewErr("[callCcntmract]get Ccntmract arg failed:" + err.Error())
	}
	//todo get result from AppCall
	//res := 0
	result ,err := this.CcntmextRef.AppCall(ccntmractAddress,util.TrimBuffToString(methodName),nil,arg)
	if err != nil {
		return false, errors.NewErr("[callCcntmract]AppCall failed:" + err.Error())
	}

	vm.RestoreCtx()
	if envCall.GetReturns() {
		idx,err := vm.SetPointerMemory(result)
		if err != nil {
			return false, errors.NewErr("[callCcntmract]SetPointerMemory failed:" + err.Error())
		}
		vm.PushResult(uint64(idx))
	}

	return true ,nil
}

func (this *WasmVmService) GetCcntmractCodeFromAddress(address common.Address) ([]byte, error) {

	dcode, err := this.Store.GetCcntmractState(address)
	if err != nil {
		return nil, err
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
