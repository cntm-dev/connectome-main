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

package router

import (
	"github.com/Ontology/eventbus/actor"
)

type RouterType int

const (
	GroupRouterType RouterType = iota
	PoolRouterType
)

type RouterConfig interface {
	RouterType() RouterType
	OnStarted(ccntmext actor.Ccntmext, props *actor.Props, router Interface)
	CreateRouterState() Interface
}

type GroupRouter struct {
	Routees *actor.PIDSet
}

type PoolRouter struct {
	PoolSize int
}

func (config *GroupRouter) OnStarted(ccntmext actor.Ccntmext, props *actor.Props, router Interface) {
	config.Routees.ForEach(func(i int, pid actor.PID) {
		ccntmext.Watch(&pid)
	})
	router.SetRoutees(config.Routees)
}

func (config *GroupRouter) RouterType() RouterType {
	return GroupRouterType
}

func (config *PoolRouter) OnStarted(ccntmext actor.Ccntmext, props *actor.Props, router Interface) {
	var routees actor.PIDSet
	for i := 0; i < config.PoolSize; i++ {
		routees.Add(ccntmext.Spawn(props))
	}
	router.SetRoutees(&routees)
}

func (config *PoolRouter) RouterType() RouterType {
	return PoolRouterType
}

func spawner(config RouterConfig) actor.SpawnFunc {
	return func(id string, props *actor.Props, parent *actor.PID) (*actor.PID, error) {
		return spawn(id, config, props, parent)
	}
}
