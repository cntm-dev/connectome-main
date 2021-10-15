package signtest

import (
	"fmt"
	"github.com/Ontology/eventbus/actor"
	"github.com/Ontology/crypto"
)



type VerifyActor struct{

}

func (s *VerifyActor) Receive(ccntmext actor.Ccntmext) {
	switch msg := ccntmext.Message().(type) {
	case *actor.Started:
		fmt.Println("Started, initialize actor here")
	case *actor.Stopping:
		fmt.Println("Stopping, actor is about shut down")
	case *actor.Restarting:
		fmt.Println("Restarting, actor is about restart")

	case *VerifyRequest:
		//fmt.Println(ccntmext.Self().Id, "is verifying...")
		err := crypto.Verify(msg.PublicKey,msg.Data,msg.Signature)
		//fmt.Println(ccntmext.Self().Id, "done verifying...")
		if err != nil{
			response:=&VerifyResponse{Seq:msg.Seq,Result:false,ErrorMsg:err.Error()}
			ccntmext.Sender().Tell(response)
		}else{
			response:=&VerifyResponse{Seq:msg.Seq,Result:true,ErrorMsg:""}
			ccntmext.Sender().Tell(response)
		}

	default:
		fmt.Printf("---unknown message%v\n",msg)
	}
}