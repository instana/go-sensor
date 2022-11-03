// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package pubsub

import (
	"context"

	"cloud.google.com/go/pubsub"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/cloud.google.com/go/internal/tags"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// Subscription is an instrumented wrapper for cloud.google.com/go/pubsub.Subscription that traces Receive() calls
// and ensures Instana trace propagation across the Pub/Sub producers and consumers.
//
// See https://pkg.go.dev/cloud.google.com/go/pubsub?tab=doc#Subscription for further details on wrapped type.
type Subscription struct {
	*pubsub.Subscription

	projectID string

	sensor *instana.Sensor
}

// Receive wraps the Receive() call of the underlying cloud.google.com/go/pubsub.Subscription starting a new
// entry span using the provided instana.Sensor and ensuring the trace continuation.
//
// See https://pkg.go.dev/cloud.google.com/go/pubsub?tab=doc#Subscription.Receive for further details on wrapped method.
func (sub *Subscription) Receive(ctx context.Context, f func(context.Context, *pubsub.Message)) error {
	return sub.Subscription.Receive(ctx, func(mCtx context.Context, msg *pubsub.Message) {
		opts := []opentracing.StartSpanOption{
			ext.SpanKindConsumer,
			opentracing.Tags{
				tags.GcpsOp:     "CONSUME",
				tags.GcpsProjid: sub.projectID,
				tags.GcpsSub:    sub.ID(),
				tags.GcpsMsgid:  msg.ID,
			},
		}
		if spCtx, err := sub.sensor.Tracer().Extract(opentracing.TextMap, opentracing.TextMapCarrier(msg.Attributes)); err == nil {
			opts = append(opts, opentracing.ChildOf(spCtx))
		}

		sp := sub.sensor.Tracer().StartSpan("gcps", opts...)
		defer sp.Finish()

		f(instana.ContextWithSpan(mCtx, sp), msg)
	})
}

// SubscriptionIterator is a wrapper for cloud.google.com/go/pubsub.SubscriptionIterator that retrieves and instruments
// subscriptions in a project.
//
// See https://pkg.go.dev/cloud.google.com/go/pubsub?tab=doc#SubscriptionIterator for further details on wrapped type.
type SubscriptionIterator struct {
	*pubsub.SubscriptionIterator

	projectID string

	sensor *instana.Sensor
}

// Next fetches the next subscription in project via the wrapped SubscriptionIterator and returns its wrapped version.
//
// See https://pkg.go.dev/cloud.google.com/go/pubsub?tab=doc#SubscriptionIterator.Next for further details on wrapped method.
func (it *SubscriptionIterator) Next() (*Subscription, error) {
	sub, err := it.SubscriptionIterator.Next()
	return &Subscription{sub, it.projectID, it.sensor}, err
}
