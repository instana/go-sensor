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
	otlog "github.com/opentracing/opentracing-go/log"
)

// Topic is an instrumented wrapper for cloud.google.com/go/pubsub.Topic that traces Publish() calls
// and ensures Instana trace propagation across the Pub/Sub producers and consumers.
//
// See https://pkg.go.dev/cloud.google.com/go/pubsub?tab=doc#Topic for further details on wrapped type.
type Topic struct {
	*pubsub.Topic

	projectID string

	sensor *instana.Sensor
}

// Publish adds the trace context found in ctx to the message and publishes it to the wrapped topic.
// The exit span created for this operation will be finished only after the message was submitted to
// the server.
//
// See https://pkg.go.dev/cloud.google.com/go/pubsub?tab=doc#Topic.Publish for further details on wrapped method.
func (top *Topic) Publish(ctx context.Context, msg *pubsub.Message) *pubsub.PublishResult {
	parent, ok := instana.SpanFromContext(ctx)
	if !ok {
		return top.Topic.Publish(ctx, msg)
	}

	sp := parent.Tracer().StartSpan("gcps",
		ext.SpanKindProducer,
		opentracing.ChildOf(parent.Context()),
		opentracing.Tags{
			tags.GcpsOp:     "PUBLISH",
			tags.GcpsProjid: top.projectID,
			tags.GcpsTop:    top.ID(),
		},
	)

	if msg.Attributes == nil {
		msg.Attributes = make(map[string]string)
	}
	sp.Tracer().Inject(sp.Context(), opentracing.TextMap, opentracing.TextMapCarrier(msg.Attributes))

	res := top.Topic.Publish(ctx, msg)
	go func() {
		id, err := res.Get(context.Background())
		if err != nil {
			sp.LogFields(otlog.Error(err))
		} else {
			sp.SetTag("gcps.msgid", id)
		}

		sp.Finish()
	}()

	return res
}

// Subscriptions calls Subscriptions() of underlying Topic and wraps the returned subscription iterator.
//
// See https://pkg.go.dev/cloud.google.com/go/pubsub?tab=doc#Topic.Subscriptions for further details on wrapped method.
func (top *Topic) Subscriptions(ctx context.Context) *SubscriptionIterator {
	return &SubscriptionIterator{top.Topic.Subscriptions(ctx), top.projectID, top.sensor}
}

// TopicIterator is a wrapper for cloud.google.com/go/pubsub.TopicIterator that retrieves and instruments
// topics in a project.
//
// See https://pkg.go.dev/cloud.google.com/go/pubsub?tab=doc#TopicIterator for further details on wrapped type.
type TopicIterator struct {
	*pubsub.TopicIterator

	projectID string

	sensor *instana.Sensor
}

// Next fetches the next topic in project via the wrapped TopicIterator and returns its wrapped version.
//
// See https://pkg.go.dev/cloud.google.com/go/pubsub?tab=doc#TopicIterator.Next for further details on wrapped method.
func (it *TopicIterator) Next() (*Topic, error) {
	top, err := it.TopicIterator.Next()
	return &Topic{top, it.projectID, it.sensor}, err
}
