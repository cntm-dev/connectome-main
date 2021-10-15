package serviceB

import (
	"fmt"

	"github.com/Ontology/eventbus/actor"
	. "github.com/Ontology/eventbus/example/services/messages"
)

type ServiceB struct {
}

func (this *ServiceB) Receive(ccntmext actor.Ccntmext) {
	switch msg := ccntmext.Message().(type) {

	case *ServiceBRequest:
		fmt.Println("Receive ServiceBRequest:", msg.Message)
		ccntmext.Sender().Request(&ServiceBResponse{"response from serviceB"}, ccntmext.Self())

	case *ServiceAResponse:
		fmt.Println("Receive ServiceAResonse:", msg.Message)

	default:
		//fmt.Println("unknown message")
	}
}
