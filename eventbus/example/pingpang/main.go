/*
 * Copyright (C) 2018 The cntmology Authors
 * This file is part of The cntmology library.
 *
 * The cntmology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The cntmology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * alcntm with The cntmology.  If not, see <http://www.gnu.org/licenses/>.
 */

package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/Ontology/eventbus/actor"
)

type ping struct{ val int }
type pingActor struct{}

var start, end int64

func (state *pingActor) Receive(ccntmext actor.Ccntmext) {
	switch msg := ccntmext.Message().(type) {
	case *actor.Started:
		fmt.Println("Started, initialize actor here")
	case *actor.Stopping:
		fmt.Println("Stopping, actor is about shut down")
	case *actor.Restarting:
		fmt.Println("Restarting, actor is about restart")
	case *ping:
		val := msg.val
		if val < 10000000 {
			ccntmext.Sender().Request(&ping{val: val + 1}, ccntmext.Self())
		} else {
			end = time.Now().UnixNano()
			fmt.Printf("%s end %d\n", ccntmext.Self().Id, end)
		}
	}
}
func main() {
	fmt.Printf("test pingpang")
	runtime.GOMAXPROCS(runtime.NumCPU())
	props := actor.FromProducer(func() actor.Actor { return &pingActor{} })
	actora := actor.Spawn(props)
	actorb := actor.Spawn(props)
	start = time.Now().UnixNano()
	fmt.Printf("begin time %d\n", start)
	actora.Request(&ping{val: 1}, actorb)

	time.Sleep(10 * time.Second)
	fmt.Println((end - start) / 1000000)
	actora.Stop()
	actorb.Stop()
}
