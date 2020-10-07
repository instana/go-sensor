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

	return &Client{
		Client:    c,
		projectID: projectID,
	}, err
}
