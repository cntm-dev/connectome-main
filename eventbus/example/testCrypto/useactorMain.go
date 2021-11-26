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
	"github.com/Ontology/eventbus/actor"
	"github.com/Ontology/eventbus/example/testCrypto/signtest"
	"time"
	"runtime"
)



func main()  {
	//runtime.GOMAXPROCS(runtime.NumCPU())
	runtime.GOMAXPROCS(runtime.NumCPU() * 1)
	runtime.GC()
	props := actor.FromProducer(func() actor.Actor { return &signtest.BusynessActor{Datas:make(map[string][]byte)} })

	bActor:=actor.Spawn(props)

	//var wg sync.WaitGroup
	//
	//wg.Add(1)
	//start := time.Now()
	bActor.Tell(&signtest.RunMsg{})
	//wg.Wait()
	//elapsed := time.Since(start)
	//fmt.Printf("Elapsed %s\n", elapsed)

	for{
		time.Sleep(1 * time.Second)
	}
}