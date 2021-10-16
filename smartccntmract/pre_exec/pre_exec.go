package pre_exec

import (
	"github.com/Ontology/smartccntmract/service"
	"github.com/Ontology/vm/neovm"
	"github.com/Ontology/vm/neovm/interfaces"
	"github.com/Ontology/smartccntmract/types"
	"github.com/Ontology/core/store/ChainStore"
	"github.com/Ontology/smartccntmract/common"
	"github.com/Ontology/core/ledger"
	"github.com/Ontology/core/store/statestore"
	. "github.com/Ontology/common"
)

//var DefaultEventStore ChainStore.IEventStore

func PreExec(code []byte, ccntmainer interfaces.ICodeCcntmainer) ([]interface{}, error) {
	var (
		crypto interfaces.ICrypto
		err error
	)
	crypto = new(neovm.ECDsaCrypto)

	stateStore := ChainStore.NewStateStore(statestore.NewMemDatabase(), ledger.DefaultLedger.Store.(*ChainStore.ChainStore), Uint256{})
	stateMachine := service.NewStateMachine(stateStore, types.Application, nil)
	se := neovm.NewExecutionEngine(ccntmainer, crypto, ChainStore.NewCacheCodeTable(stateStore), stateMachine)
	se.LoadCode(code, false)
	err = se.Execute()
	if err != nil {
		return nil, err
	}
	if se.GetEvaluationStackCount() == 0 {
		return nil, err
	}
	if neovm.Peek(se).GetStackItem() == nil {
		return nil, err
	}
	return common.ConvertReturnTypes(neovm.Peek(se).GetStackItem()), nil
}
