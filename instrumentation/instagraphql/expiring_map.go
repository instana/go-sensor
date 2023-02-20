// (c) Copyright IBM Corp. 2023

package instagraphql

import (
	"sync"
	"time"

	ot "github.com/opentracing/opentracing-go"
)

type spanWithTimer struct {
	sp ot.Span
	t  *time.Timer
}

// ExpiringMap holds a map of spans that are automatically removed from this map after the provided duration expires.
// The expiration time is renewed if the map is set with the same key before the original time exipres.
type ExpiringMap struct {
	sync.RWMutex
	m map[string]spanWithTimer
}

// Set will set the span which will expire after d is reached.
// If a span is set with the same key before the original one expires, the time will be renewed.
func (em *ExpiringMap) Set(k string, v ot.Span, d time.Duration) {
	if em.m == nil {
		em.m = make(map[string]spanWithTimer)
	}

	if _, ok := em.m[k]; !ok {
		em.Lock()
		newWrapper := spanWithTimer{
			sp: v,
			t:  time.NewTimer(d),
		}

		em.m[k] = newWrapper

		go func(k string) {
			<-newWrapper.t.C
			em.Lock()
			delete(em.m, k)
			em.Unlock()
		}(k)

		em.Unlock()
		return
	}

	em.Lock()
	t := em.m[k].t
	em.m[k] = spanWithTimer{v, t}
	t.Reset(d)
	em.Unlock()
}

// Get returns the span for the given k or nil if not found.
func (em *ExpiringMap) Get(k string) ot.Span {
	if em.m == nil {
		em.m = make(map[string]spanWithTimer)
	}

	em.RLock()
	defer em.RUnlock()
	iw, ok := em.m[k]

	if !ok {
		return nil
	}

	return iw.sp
}
