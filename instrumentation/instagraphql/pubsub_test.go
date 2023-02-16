// (c) Copyright IBM Corp. 2023

package instagraphql_test

import "sync"

// pubsub is a basic Pub Sub implementation that helps test GraphQL's subscription.
// its existence is not relevant for the instrumentation, but helpful for tests and mocks
type pubsub struct {
	sync.Mutex
	pool map[string][]chan interface{}
}

func (ps *pubsub) pub(name string, val interface{}) {
	ps.Lock()
	defer ps.Unlock()

	if ps.pool == nil {
		ps.pool = make(map[string][]chan interface{})
	}

	for idx, ch := range ps.pool[name] {
		select {
		case ch <- val:
		default:
			ps.pool[name] = append(ps.pool[name][:idx], ps.pool[name][idx+1:]...)
		}
	}
}

func (ps *pubsub) sub(name string, ch chan interface{}) {
	ps.Lock()
	defer ps.Unlock()

	if ps.pool == nil {
		ps.pool = make(map[string][]chan interface{})
	}

	ps.pool[name] = append(ps.pool[name], ch)
}
