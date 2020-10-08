package pubsub

import "cloud.google.com/go/pubsub"

// Subscription is an instrumented wrapper for cloud.google.com/go/pubsub.Subscription that traces Receive() calls
// and ensures Instana trace propagation across the Pub/Sub producers and consumers.
//
// See https://pkg.go.dev/cloud.google.com/go/pubsub?tab=doc#Subscription for furter details on wrapped type.
type Subscription struct {
	*pubsub.Subscription
	projectID string
	topicID   string
}

// SubscriptionIterator is a wrapper for cloud.google.com/go/pubsub.SubscriptionIterator that retrieves and instruments
// subscriptions in a project.
//
// See https://pkg.go.dev/cloud.google.com/go/pubsub?tab=doc#SubscriptionIterator for furter details on wrapped type.
type SubscriptionIterator struct {
	*pubsub.SubscriptionIterator
	projectID string
	topicID   string
}

// Next fetches the next subscription in project via the wrapped SubscriptionIterator and returns its wrapped version.
//
// See https://pkg.go.dev/cloud.google.com/go/pubsub?tab=doc#SubscriptionIterator.Next for furter details on wrapped method.
func (it *SubscriptionIterator) Next() (*Subscription, error) {
	sub, err := it.SubscriptionIterator.Next()
	return &Subscription{sub, it.projectID, it.topicID}, err
}
