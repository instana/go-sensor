// (c) Copyright IBM Corp. 2023

package instasarama

import (
	"sync"

	"github.com/IBM/sarama"
	ot "github.com/opentracing/opentracing-go"
)

type spanKeyType uint8

const (
	producerSpanKeyType spanKeyType = iota + 1
	consumerSpanKeyType
)

type spanKey struct {
	Type      spanKeyType
	Topic     string
	Partition int32
	Offset    int64
}

func producerSpanKey(msg *sarama.ProducerMessage) spanKey {
	return spanKey{
		Type:      producerSpanKeyType,
		Topic:     msg.Topic,
		Partition: msg.Partition,
		Offset:    msg.Offset,
	}
}

func consumerSpanKey(msg *sarama.ConsumerMessage) spanKey {
	return spanKey{
		Type:      consumerSpanKeyType,
		Topic:     msg.Topic,
		Partition: msg.Partition,
		Offset:    msg.Offset,
	}
}

// spanRegistry is a thread-safe storage for spans associated with Kafka messages
type spanRegistry struct {
	mu    sync.Mutex
	spans map[spanKey]ot.Span
}

func newSpanRegistry() *spanRegistry {
	return &spanRegistry{spans: make(map[spanKey]ot.Span)}
}

// Add puts an active span to the registry
func (r *spanRegistry) Add(key spanKey, sp ot.Span) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.spans[key] = sp
}

// Remove retrieves and removes an active span from registry
func (r *spanRegistry) Remove(key spanKey) (ot.Span, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	sp, ok := r.spans[key]
	if !ok {
		return nil, false
	}
	delete(r.spans, key)

	return sp, true
}
