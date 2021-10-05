// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

package instamongo

import (
	"sync"

	"github.com/opentracing/opentracing-go"
)

type spanRegistry struct {
	mu    sync.Mutex
	spans map[int64]opentracing.Span
}

func newSpanRegistry() *spanRegistry {
	return &spanRegistry{
		spans: make(map[int64]opentracing.Span),
	}
}

// Add puts an opentracing.Span into registry with given key
func (r *spanRegistry) Add(key int64, span opentracing.Span) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.spans[key] = span
}

// Remove deletes and returns an opentracing.Span from registry using provided key. Returns
// false as a second value if the registry does not contain a span with such key.
func (r *spanRegistry) Remove(key int64) (opentracing.Span, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	sp, ok := r.spans[key]
	if ok {
		delete(r.spans, key)
	}

	return sp, ok
}
