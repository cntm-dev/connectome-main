package actor

import (
	"fmt"
	"github.com/Ontology/eventbus/actor"
	"github.com/Ontology/eventbus/eventhub"
)

var completeBlockPid *actor.PID
var smartcodePid *actor.PID

func SubscribeEvent(topic string, handle func(v interface{})) {
	eh := eventhub.GlobalEventHub
	subprops := actor.FromFunc(func(ccntmext actor.Ccntmext) {
		switch msg := ccntmext.Message().(type) {

		case interface{}:
			handle(msg)
			fmt.Println(ccntmext.Self().Id + " get message ")
			//ccntmext.Sender().Request(ResponseMessage{"response message from "+ccntmext.Self().Id },ccntmext.Self())
		default:
			//ledger.DefaultLedger.Blockchain.BCEvents.Subscribe(events.EventBlockPersistCompleted, SendBlock2WSclient)
			//ledger.DefaultLedger.Blockchain.BCEvents.Subscribe(events.EventSmartCode, PushSmartCodeEvent)
		}
	})
	sub1, _ := actor.SpawnNamed(subprops, "sub1")
	eh.Subscribe(topic, sub1)
}
