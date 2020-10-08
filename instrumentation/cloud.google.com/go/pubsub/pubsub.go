package pubsub

import (
	"context"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/option"
)

// Client is an instrumented wrapper for cloud.google.com/go/pubsub.Client that traces message reads and
// writes to and from Google Cloud Pub/Sub topics. It also  and ensures Instana trace propagation across
// the Pub/Sub producers and consumers.
//
// See https://pkg.go.dev/cloud.google.com/go/pubsub?tab=doc#Client for furter details on wrapped type.
type Client struct {
	*pubsub.Client
	projectID string
}

// NewClient returns a new wrapped cloud.google.com/go/pubsub.Client.
//
// See https://pkg.go.dev/cloud.google.com/go/pubsub?tab=doc#NewClient for furter details on wrapped method.
func NewClient(ctx context.Context, projectID string, opts ...option.ClientOption) (*Client, error) {
	c, err := pubsub.NewClient(ctx, projectID, opts...)
	return &Client{c, projectID}, err
}

// CreateTopic calls CreateTopic() of underlying Client and wraps the returned topic.
//
// See https://pkg.go.dev/cloud.google.com/go/pubsub?tab=doc#Client.CreateTopic for furter details on wrapped method.
func (c *Client) CreateTopic(ctx context.Context, id string) (*Topic, error) {
	top, err := c.Client.CreateTopic(ctx, id)
	return &Topic{top, c.projectID}, err
}

// CreateTopicWithConfig calls CreateTopicWithConfig() of underlying Client and wraps returned topic.
//
// See https://pkg.go.dev/cloud.google.com/go/pubsub?tab=doc#Client.CreateTopicWithConfig for furter details on wrapped method.
func (c *Client) CreateTopicWithConfig(ctx context.Context, topicID string, tc *pubsub.TopicConfig) (*Topic, error) {
	top, err := c.Client.CreateTopicWithConfig(ctx, topicID, tc)
	return &Topic{top, c.projectID}, err
}

// Topic calls Topic() of underlying Client and wraps the returned topic.
//
// See https://pkg.go.dev/cloud.google.com/go/pubsub?tab=doc#Client.Topic for furter details on wrapped method.
func (c *Client) Topic(id string) *Topic {
	return &Topic{c.Client.Topic(id), c.projectID}
}

// TopicInProject calls TopicInProject() of underlying Client and wraps the returned topic.
//
// See https://pkg.go.dev/cloud.google.com/go/pubsub?tab=doc#Client.TopicInProject for furter details on wrapped method.
func (c *Client) TopicInProject(id, projectID string) *Topic {
	return &Topic{c.Client.TopicInProject(id, projectID), projectID}
}

// Topics calls Topics() of underlying Client and wraps the returned topic iterator.
//
// See https://pkg.go.dev/cloud.google.com/go/pubsub?tab=doc#Client.Topics for furter details on wrapped method.
func (c *Client) Topics(ctx context.Context) *TopicIterator {
	return &TopicIterator{c.Client.Topics(ctx), c.projectID}
}

// CreateSubscription calls CreateSubscription() of underlying Client and wraps the returned subscription.
//
// See https://pkg.go.dev/cloud.google.com/go/pubsub?tab=doc#Client.CreateSubscription for furter details on wrapped method.
func (c *Client) CreateSubscription(ctx context.Context, id string, cfg pubsub.SubscriptionConfig) (*Subscription, error) {
	sub, err := c.Client.CreateSubscription(ctx, id, cfg)

	var topicID string
	if cfg.Topic != nil {
		topicID = cfg.Topic.ID()
	}

	return &Subscription{sub, c.projectID, topicID}, err
}

// Subscription calls Subscription() of underlying Client and wraps the returned subscription.
//
// See https://pkg.go.dev/cloud.google.com/go/pubsub?tab=doc#Client.Subscription for furter details on wrapped method.
func (c *Client) Subscription(id string) *Subscription {
	return &Subscription{c.Client.Subscription(id), c.projectID, ""}
}

// SubscriptionInProject calls SubscriptionInProject() of underlying Client and wraps the returned subscription.
//
// See https://pkg.go.dev/cloud.google.com/go/pubsub?tab=doc#Client.SubscriptionInProject for furter details on wrapped method.
func (c *Client) SubscriptionInProject(id, projectID string) *Subscription {
	return &Subscription{c.Client.SubscriptionInProject(id, projectID), projectID, ""}
}

// Subscriptions calls Subscriptions() of underlying Client and wraps the returned subscription iterator.
//
// See https://pkg.go.dev/cloud.google.com/go/pubsub?tab=doc#Client.Subscriptions for furter details on wrapped method.
func (c *Client) Subscriptions(ctx context.Context) *SubscriptionIterator {
	return &SubscriptionIterator{c.Client.Subscriptions(ctx), c.projectID, ""}
}
