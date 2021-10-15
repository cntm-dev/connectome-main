package router

import (
	"sync"

	"github.com/Ontology/eventbus/actor"
)

type groupRouterActor struct {
	props  *actor.Props
	config RouterConfig
	state  Interface
	wg     *sync.WaitGroup
}

func (a *groupRouterActor) Receive(ccntmext actor.Ccntmext) {
	switch m := ccntmext.Message().(type) {
	case *actor.Started:
		a.config.OnStarted(ccntmext, a.props, a.state)
		a.wg.Done()

	case *AddRoutee:
		r := a.state.GetRoutees()
		if r.Ccntmains(m.PID) {
			return
		}
		ccntmext.Watch(m.PID)
		r.Add(m.PID)
		a.state.SetRoutees(r)

	case *RemoveRoutee:
		r := a.state.GetRoutees()
		if !r.Ccntmains(m.PID) {
			return
		}

		ccntmext.Unwatch(m.PID)
		r.Remove(m.PID)
		a.state.SetRoutees(r)

	case *BroadcastMessage:
		msg := m.Message
		sender := ccntmext.Sender()
		a.state.GetRoutees().ForEach(func(i int, pid actor.PID) {
			pid.Request(msg, sender)
		})

	case *GetRoutees:
		r := a.state.GetRoutees()
		routees := make([]*actor.PID, r.Len())
		r.ForEach(func(i int, pid actor.PID) {
			routees[i] = &pid
		})

		ccntmext.Respond(&Routees{routees})
	}
}
