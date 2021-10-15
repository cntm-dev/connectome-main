package main

import (
	"fmt"
	"runtime"
	"github.com/Ontology/common/log"
	"github.com/Ontology/eventbus/actor"
	"github.com/Ontology/eventbus/example/remotebenchmark/messages"
	"github.com/Ontology/eventbus/mailbox"
	"github.com/Ontology/eventbus/remote"
	"time"
)

func main() {
	log.Init()
	runtime.GOMAXPROCS(runtime.NumCPU() * 1)
	runtime.GC()

	remote.Start("127.0.0.1:8080")
	var sender *actor.PID
	props := actor.
		FromFunc(
			func(ccntmext actor.Ccntmext) {
				switch msg := ccntmext.Message().(type) {
				case *messages.StartRemote:
					fmt.Println("Starting")
					sender = msg.Sender
					ccntmext.Respond(&messages.Start{})
				case *messages.Ping:
					sender.Tell(&messages.Pcntm{})
				}
			}).
		WithMailbox(mailbox.Bounded(1000000))
	actor.SpawnNamed(props, "remote")
	for{
		time.Sleep(1 * time.Second)
	}
}
