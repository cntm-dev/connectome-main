package server

import (
	"fmt"

	"github.com/Ontology/eventbus/actor"
	"github.com/Ontology/eventbus/example/serverclient/message"
)

type Server struct{}

func (server *Server) Receive(ccntmext actor.Ccntmext) {
	switch msg := ccntmext.Message().(type) {
	case *actor.Started:
		fmt.Println("Started, initialize server actor here")
	case *actor.Stopping:
		fmt.Println("Stopping, actor is about shut down")
	case *actor.Restarting:
		fmt.Println("Restarting, actor is about restart")
	case *message.Request:
		fmt.Println("Receive message", msg.Who)
		ccntmext.Sender().Request(&message.Response{Welcome: "Welcome!"}, ccntmext.Self())
	}
}

func (server *Server) Start() *actor.PID {
	props := actor.FromProducer(func() actor.Actor { return &Server{} })
	pid := actor.Spawn(props)
	return pid
}

func (server *Server) Stop(pid *actor.PID) {
	pid.Stop()
}
