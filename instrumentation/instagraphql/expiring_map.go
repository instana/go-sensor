// (c) Copyright IBM Corp. 2023

package instagraphql

import (
	"sync"
	"time"

	ot "github.com/opentracing/opentracing-go"
)

type internalWrapper struct {
	Obj   ot.Span
	Timer *time.Timer
}

type ExpiringMap struct {
	sync.RWMutex
	list map[string]internalWrapper
}

func (m *ExpiringMap) Set(k string, v ot.Span, d time.Duration) {
	if m.list == nil {
		m.list = make(map[string]internalWrapper)
	}

	if _, ok := m.list[k]; !ok {
		m.Lock()
		newWrapper := internalWrapper{
			Obj:   v,
			Timer: time.NewTimer(d),
		}

		m.list[k] = newWrapper

		go func(k string) {
			<-newWrapper.Timer.C
			m.Lock()
			delete(m.list, k)
			m.Unlock()
		}(k)

		m.Unlock()
	} else {
		m.Lock()
		t := m.list[k].Timer
		m.list[k] = internalWrapper{v, t}
		t.Reset(d)
		m.Unlock()
	}
}

func (m *ExpiringMap) Get(k string) ot.Span {
	m.RLock()
	defer m.RUnlock()
	iw, ok := m.list[k]

	if !ok {
		return nil
	}

	return iw.Obj
}
