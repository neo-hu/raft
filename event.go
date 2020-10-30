package raft

import "sync"

type eventListener struct {
	sync.RWMutex
	listeners []EventHandles
}

func newEventListener() *eventListener {
	return &eventListener{}
}

func (e *eventListener) dispatchEvent(event interface{}) {
	e.RLock()
	defer e.RUnlock()
	switch t := event.(type) {
	case *leaderChangeEvent:
		for _, listener := range e.listeners {
			listener.OnLeaderChange(t.leader, t.prevLeader)
		}
	}
}
func (e *eventListener) OnEvent(h EventHandles) {
	e.listeners = append(e.listeners, h)
}

type leaderChangeEvent struct {
	leader, prevLeader string
}

type EventHandles interface {
	OnLeaderChange(leader, prevLeader string)
}
