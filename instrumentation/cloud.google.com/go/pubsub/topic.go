package pubsub

import (
	"context"

	"cloud.google.com/go/pubsub"
)

// Topic is an instrumented wrapper for cloud.google.com/go/pubsub.Topic that traces Publish() calls
// and ensures Instana trace propagation across the Pub/Sub producers and consumers.
//
// See https://pkg.go.dev/cloud.google.com/go/pubsub?tab=doc#Topic for furter details on wrapped type.
type Topic struct {
	*pubsub.Topic
	projectID string
}

// Publish adds the trace context found in ctx to the message and publishes it to the wrapped topic.
//
// See https://pkg.go.dev/cloud.google.com/go/pubsub?tab=doc#Topic.Publish for furter details on wrapped method.
//
// TODO: instrument
func (top *Topic) Publish(ctx context.Context, msg *pubsub.Message) *pubsub.PublishResult {
	return top.Publish(ctx, msg)
}

// TopicIterator is a wrapper for cloud.google.com/go/pubsub.TopicIterator that retrieves and instruments
// topics in a project.
//
// See https://pkg.go.dev/cloud.google.com/go/pubsub?tab=doc#TopicIterator for furter details on wrapped type.
type TopicIterator struct {
	*pubsub.TopicIterator
	projectID string
}

// Next fetches the next topic in project via the wrapped TopicIterator and returns its wrapped version.
//
// See https://pkg.go.dev/cloud.google.com/go/pubsub?tab=doc#TopicIterator.Next for furter details on wrapped method.
func (it *TopicIterator) Next() (*Topic, error) {
	top, err := it.TopicIterator.Next()
	return &Topic{top, it.projectID}, err
}
