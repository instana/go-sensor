// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

// Package pubsub provides Instana tracing instrumentation for
// Google Cloud Pub/Sub producers and consumers that use cloud.google.com/go/pubsub.
package pubsub

import (
	"context"

	"cloud.google.com/go/pubsub"
	instana "github.com/instana/go-sensor"
	"google.golang.org/api/option"
)

type (
	// Message is a type alias for cloud.google.com/go/pubsub.Message
	Message = pubsub.Message
	// PublishResult is a type alias for cloud.google.com/go/pubsub.PublishResult
	PublishResult = pubsub.PublishResult
	// Snapshot is a type alias for cloud.google.com/go/pubsub.Snapshot
	Snapshot = pubsub.Snapshot
)

// Client is an instrumented wrapper for cloud.google.com/go/pubsub.Client that traces message reads and
// writes to and from Google Cloud Pub/Sub topics. It also  and ensures Instana trace propagation across
// the Pub/Sub producers and consumers.
//
// See https://pkg.go.dev/cloud.google.com/go/pubsub?tab=doc#Client for further details on wrapped type.
type Client struct {
	*pubsub.Client
	projectID string

	sensor *instana.Sensor
}

// NewClient returns a new wrapped cloud.google.com/go/pubsub.Client that uses provided instana.Sensor to
// trace the publish/receive operations.
//
// See https://pkg.go.dev/cloud.google.com/go/pubsub?tab=doc#NewClient for further details on wrapped method.
func NewClient(ctx context.Context, projectID string, sensor *instana.Sensor, opts ...option.ClientOption) (*Client, error) {
	c, err := pubsub.NewClient(ctx, projectID, opts...)
	return &Client{c, projectID, sensor}, err
}

// CreateTopic calls CreateTopic() of underlying Client and wraps the returned topic.
//
// See https://pkg.go.dev/cloud.google.com/go/pubsub?tab=doc#Client.CreateTopic for further details on wrapped method.
func (c *Client) CreateTopic(ctx context.Context, id string) (*Topic, error) {
	top, err := c.Client.CreateTopic(ctx, id)
	return &Topic{top, c.projectID, c.sensor}, err
}

// CreateTopicWithConfig calls CreateTopicWithConfig() of underlying Client and wraps returned topic.
//
// See https://pkg.go.dev/cloud.google.com/go/pubsub?tab=doc#Client.CreateTopicWithConfig for further details on wrapped method.
func (c *Client) CreateTopicWithConfig(ctx context.Context, topicID string, tc *pubsub.TopicConfig) (*Topic, error) {
	top, err := c.Client.CreateTopicWithConfig(ctx, topicID, tc)
	return &Topic{top, c.projectID, c.sensor}, err
}

// Topic calls Topic() of underlying Client and wraps the returned topic.
//
// See https://pkg.go.dev/cloud.google.com/go/pubsub?tab=doc#Client.Topic for further details on wrapped method.
func (c *Client) Topic(id string) *Topic {
	return &Topic{c.Client.Topic(id), c.projectID, c.sensor}
}

// TopicInProject calls TopicInProject() of underlying Client and wraps the returned topic.
//
// See https://pkg.go.dev/cloud.google.com/go/pubsub?tab=doc#Client.TopicInProject for further details on wrapped method.
func (c *Client) TopicInProject(id, projectID string) *Topic {
	return &Topic{c.Client.TopicInProject(id, projectID), projectID, c.sensor}
}

// Topics calls Topics() of underlying Client and wraps the returned topic iterator.
//
// See https://pkg.go.dev/cloud.google.com/go/pubsub?tab=doc#Client.Topics for further details on wrapped method.
func (c *Client) Topics(ctx context.Context) *TopicIterator {
	return &TopicIterator{c.Client.Topics(ctx), c.projectID, c.sensor}
}

// CreateSubscription calls CreateSubscription() of underlying Client and wraps the returned subscription.
//
// See https://pkg.go.dev/cloud.google.com/go/pubsub?tab=doc#Client.CreateSubscription for further details on wrapped method.
func (c *Client) CreateSubscription(ctx context.Context, id string, cfg pubsub.SubscriptionConfig) (*Subscription, error) {
	sub, err := c.Client.CreateSubscription(ctx, id, cfg)
	return &Subscription{sub, c.projectID, c.sensor}, err
}

// Subscription calls Subscription() of underlying Client and wraps the returned subscription.
//
// See https://pkg.go.dev/cloud.google.com/go/pubsub?tab=doc#Client.Subscription for further details on wrapped method.
func (c *Client) Subscription(id string) *Subscription {
	return &Subscription{c.Client.Subscription(id), c.projectID, c.sensor}
}

// SubscriptionInProject calls SubscriptionInProject() of underlying Client and wraps the returned subscription.
//
// See https://pkg.go.dev/cloud.google.com/go/pubsub?tab=doc#Client.SubscriptionInProject for further details on wrapped method.
func (c *Client) SubscriptionInProject(id, projectID string) *Subscription {
	return &Subscription{c.Client.SubscriptionInProject(id, projectID), projectID, c.sensor}
}

// Subscriptions calls Subscriptions() of underlying Client and wraps the returned subscription iterator.
//
// See https://pkg.go.dev/cloud.google.com/go/pubsub?tab=doc#Client.Subscriptions for further details on wrapped method.
func (c *Client) Subscriptions(ctx context.Context) *SubscriptionIterator {
	return &SubscriptionIterator{c.Client.Subscriptions(ctx), c.projectID, c.sensor}
}
