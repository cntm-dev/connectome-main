package neovm

import (
	vm "github.com/cntmio/cntmology/vm/neovm"
	"github.com/cntmio/cntmology/errors"
	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/core/types"
	"github.com/cntmio/cntmology-crypto/keypair"
	"github.com/cntmio/cntmology/smartccntmract/event"
	scommon "github.com/cntmio/cntmology/smartccntmract/common"
)

func RuntimeGetTime(service *NeoVmService, engine *vm.ExecutionEngine) error {
	vm.PushData(engine, int(service.Time))
	return nil
}

func RuntimeCheckWitness(service *NeoVmService, engine *vm.ExecutionEngine) error {
	if vm.EvaluationStackCount(engine) < 1 {
		return errors.NewErr("[RuntimeCheckWitness] Too few input parameters ")
	}
	data := vm.PopByteArray(engine)
	var result bool
	if len(data) == 20 {
		address, err := common.AddressParseFromBytes(data)
		if err != nil {
			return err
		}
		result = checkWitnessAddress(service, address)
	} else {
		pk, err := keypair.DeserializePublicKey(data); if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[RuntimeCheckWitness] data invalid.")
		}
		result = checkWitnessPublicKey(service, pk)
	}

	vm.PushData(engine, result)
	return nil
}

func RuntimeNotify(service *NeoVmService, engine *vm.ExecutionEngine) error {
	item := vm.PopStackItem(engine)
	ccntmext := service.CcntmextRef.CurrentCcntmext()
	service.Notifications = append(service.Notifications, &event.NotifyEventInfo{TxHash: service.Tx.Hash(), CcntmractAddress: ccntmext.CcntmractAddress, States: scommon.ConvertReturnTypes(item)})
	return nil
}

func RuntimeLog(service *NeoVmService, engine *vm.ExecutionEngine) error {
	item := vm.PopByteArray(engine)
	ccntmext := service.CcntmextRef.CurrentCcntmext()
	txHash := service.Tx.Hash()
	event.PushSmartCodeEvent(txHash, 0, "InvokeTransaction", &event.LogEventArgs{TxHash:txHash, CcntmractAddress: ccntmext.CcntmractAddress, Message: string(item)})
	return nil
}

func checkWitnessAddress(service *NeoVmService, address common.Address) bool {
	return service.CcntmextRef.CheckWitness(address)
}

func checkWitnessPublicKey(service *NeoVmService, publicKey keypair.PublicKey) bool {
	return checkWitnessAddress(service, types.AddressFromPubKey(publicKey))
}



