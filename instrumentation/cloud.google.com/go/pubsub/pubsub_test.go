// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

//go:build go1.19
// +build go1.19

package pubsub_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	gpubsub "cloud.google.com/go/pubsub"
	"cloud.google.com/go/pubsub/pstest"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/cloud.google.com/go/pubsub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	pb "google.golang.org/genproto/googleapis/pubsub/v1"
	"google.golang.org/grpc"
)

func TestClient_Topic(t *testing.T) {
	srv, conn, teardown, err := setupMockServer()
	require.NoError(t, err)
	defer teardown()

	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)

	sensor := instana.NewSensorWithTracer(tracer)
	defer instana.ShutdownSensor()

	pSpan := tracer.StartSpan("parent-span")
	ctx := context.Background()
	if pSpan != nil {
		ctx = instana.ContextWithSpan(ctx, pSpan)
	}

	_, err = srv.GServer.CreateTopic(ctx, &pb.Topic{
		Name: "projects/test-project/topics/test-topic",
	})
	require.NoError(t, err)

	examples := map[string]func(*testing.T, *pubsub.Message) *pubsub.PublishResult{
		"ClientProject": func(t *testing.T, msg *pubsub.Message) *pubsub.PublishResult {
			client, err := pubsub.NewClient(ctx, "test-project", sensor, option.WithGRPCConn(conn))
			require.NoError(t, err)

			top := client.Topic("test-topic")

			return top.Publish(ctx, msg)
		},
		"OtherProject": func(t *testing.T, msg *pubsub.Message) *pubsub.PublishResult {
			client, err := pubsub.NewClient(ctx, "other-project", sensor, option.WithGRPCConn(conn))
			require.NoError(t, err)

			top := client.TopicInProject("test-topic", "test-project")

			return top.Publish(ctx, msg)
		},
		"CreateTopic": func(t *testing.T, msg *pubsub.Message) *pubsub.PublishResult {
			client, err := pubsub.NewClient(ctx, "test-project", sensor, option.WithGRPCConn(conn))
			require.NoError(t, err)

			top, err := client.CreateTopic(ctx, "new-test-topic")
			require.NoError(t, err)

			return top.Publish(ctx, msg)
		},
		"CreateTopicWithConfig": func(t *testing.T, msg *pubsub.Message) *pubsub.PublishResult {
			client, err := pubsub.NewClient(ctx, "test-project", sensor, option.WithGRPCConn(conn))
			require.NoError(t, err)

			conf := &gpubsub.TopicConfig{
				MessageStoragePolicy: gpubsub.MessageStoragePolicy{
					AllowedPersistenceRegions: []string{"us-east1"},
				},
			}

			top, err := client.CreateTopicWithConfig(ctx, "new-test-topic-with-config", conf)
			require.NoError(t, err)

			createdConf, err := top.Config(ctx)
			require.NoError(t, err)
			assert.Equal(t, conf.Labels, createdConf.Labels)
			assert.Equal(t, conf.MessageStoragePolicy, createdConf.MessageStoragePolicy)
			assert.Equal(t, conf.KMSKeyName, createdConf.KMSKeyName)
			assert.Equal(t, conf.SchemaSettings, createdConf.SchemaSettings)

			// name cannot be tested because in new versions of pubsub this new name attribute is not replicated to the original conf. only top.Config() has it
			// assert.Equal(t, conf.name, createdConf.name)

			return top.Publish(ctx, msg)
		},
	}

	for name, publish := range examples {
		t.Run(name, func(t *testing.T) {
			srv.ClearMessages()

			res := publish(t, &pubsub.Message{
				Data: []byte("message data"),
				Attributes: map[string]string{
					"key1": "value1",
				},
			})

			require.Eventually(t, func() bool {
				return recorder.QueuedSpansCount() == 1
			}, 250*time.Millisecond, 25*time.Millisecond)

			spans := recorder.GetQueuedSpans()
			require.Len(t, spans, 1)
			span := spans[0]

			ctx, cancel := context.WithTimeout(ctx, time.Second)
			defer cancel()

			msgID, err := res.Get(ctx)
			require.NoError(t, err)

			msg := srv.Message(msgID)
			require.NotNil(t, msg)

			assert.Equal(t, []byte("message data"), msg.Data)
			assert.Equal(t, map[string]string{
				"x-instana-t": instana.FormatID(span.TraceID),
				"x-instana-s": instana.FormatID(span.SpanID),
				"x-instana-l": "1",
				"key1":        "value1",
			}, msg.Attributes)
		})
	}
}

func TestClient_Topics(t *testing.T) {
	srv, conn, teardown, err := setupMockServer()
	require.NoError(t, err)
	defer teardown()

	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(
			instana.DefaultOptions(),
			instana.NewTestRecorder(),
		),
	)
	defer instana.ShutdownSensor()

	topicNames := []string{"first-topic", "second-topic"}

	for _, topicName := range topicNames {
		_, err = srv.GServer.CreateTopic(context.Background(), &pb.Topic{
			Name: "projects/test-project/topics/" + topicName,
		})
		require.NoError(t, err)
	}

	client, err := pubsub.NewClient(context.Background(), "test-project", sensor, option.WithGRPCConn(conn))
	require.NoError(t, err)

	var res []*pubsub.PublishResult

	it := client.Topics(context.Background())
	for {
		top, err := it.Next()
		if err == iterator.Done {
			break
		}

		res = append(res, top.Publish(context.Background(), &pubsub.Message{
			Data: []byte("message in " + top.ID()),
		}))
	}

	for _, res := range res {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		_, err := res.Get(ctx)
		require.NoError(t, err)
	}

	var msgs []string
	for _, msg := range srv.Messages() {
		msgs = append(msgs, string(msg.Data))
	}
	for _, topicName := range topicNames {
		assert.Contains(t, msgs, "message in "+topicName)
	}
}

func TestClient_Subscription(t *testing.T) {
	srv, conn, teardown, err := setupMockServer()
	require.NoError(t, err)
	defer teardown()

	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(
			instana.DefaultOptions(),
			instana.NewTestRecorder(),
		),
	)
	defer instana.ShutdownSensor()

	top, err := srv.GServer.CreateTopic(context.Background(), &pb.Topic{
		Name: "projects/test-project/topics/test-topic",
	})
	require.NoError(t, err)

	examples := map[string]func(*testing.T, string) *pubsub.Subscription{
		"ClientProject": func(t *testing.T, topicName string) *pubsub.Subscription {
			pstest.SetMinAckDeadline(1 * time.Second)
			defer pstest.ResetMinAckDeadline()

			_, err = srv.GServer.CreateSubscription(context.Background(), &pb.Subscription{
				Topic:              "projects/test-project/topics/" + topicName,
				Name:               "projects/test-project/subscriptions/test-subscription",
				AckDeadlineSeconds: 10,
			})
			require.NoError(t, err)

			client, err := pubsub.NewClient(context.Background(), "test-project", sensor, option.WithGRPCConn(conn))
			require.NoError(t, err)

			return client.Subscription("test-subscription")
		},
		"OtherProject": func(t *testing.T, topicName string) *pubsub.Subscription {
			pstest.SetMinAckDeadline(1 * time.Second)
			defer pstest.ResetMinAckDeadline()

			_, err = srv.GServer.CreateSubscription(context.Background(), &pb.Subscription{
				Topic:              "projects/test-project/topics/" + topicName,
				Name:               "projects/test-project/subscriptions/test-subscription",
				AckDeadlineSeconds: 10,
			})
			require.NoError(t, err)

			client, err := pubsub.NewClient(context.Background(), "other-project", sensor, option.WithGRPCConn(conn))
			require.NoError(t, err)

			return client.SubscriptionInProject("test-subscription", "test-project")
		},

		"CreateSubscriptionWithConfig": func(t *testing.T, topicName string) *pubsub.Subscription {
			client, err := pubsub.NewClient(context.Background(), "test-project", sensor, option.WithGRPCConn(conn))
			require.NoError(t, err)

			sub, err := client.CreateSubscription(context.Background(), "test-subscription", gpubsub.SubscriptionConfig{
				Topic: client.Topic(topicName).Topic,
			})
			require.NoError(t, err)

			return sub
		},
	}

	for name, subscribe := range examples {
		t.Run(name, func(t *testing.T) {
			srv.ClearMessages()

			sub := subscribe(t, "test-topic")
			defer sub.Delete(context.Background())

			msgID := srv.Publish(top.Name, []byte("test message"), nil)

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			require.NoError(t, sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
				assert.Equal(t, msgID, msg.ID)
				msg.Ack()
				cancel()
			}))

			// ensure the context has been cancelled and not timed out
			assert.Equal(t, context.Canceled, ctx.Err())
		})
	}
}

func TestClient_Subscriptions(t *testing.T) {
	pstest.SetMinAckDeadline(1 * time.Second)
	defer pstest.ResetMinAckDeadline()

	srv, conn, teardown, err := setupMockServer()
	require.NoError(t, err)
	defer teardown()

	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(
			instana.DefaultOptions(),
			instana.NewTestRecorder(),
		),
	)
	defer instana.ShutdownSensor()

	top, err := srv.GServer.CreateTopic(context.Background(), &pb.Topic{
		Name: "projects/test-project/topics/test-topic",
	})
	require.NoError(t, err)

	subscriptionNames := []string{"first", "second"}
	for _, subName := range subscriptionNames {
		_, err = srv.GServer.CreateSubscription(context.Background(), &pb.Subscription{
			Topic:              top.Name,
			Name:               "projects/test-project/subscriptions/" + subName,
			AckDeadlineSeconds: 10,
		})
		require.NoError(t, err)
	}

	client, err := pubsub.NewClient(context.Background(), "test-project", sensor, option.WithGRPCConn(conn))
	require.NoError(t, err)

	var subs []string

	it := client.Subscriptions(context.Background())
	for {
		sub, err := it.Next()
		if err == iterator.Done {
			break
		}

		require.NoError(t, err)
		subs = append(subs, sub.ID())
	}

	assert.ElementsMatch(t, subscriptionNames, subs)
}

func setupMockServer() (*pstest.Server, *grpc.ClientConn, func(), error) {
	srv := pstest.NewServer()

	conn, err := grpc.Dial(srv.Addr, grpc.WithInsecure())
	if err != nil {
		srv.Close()
		return nil, nil, nil, fmt.Errorf("failed to start new pubsub server: %s", err)
	}

	return srv, conn, func() {
		conn.Close()
		srv.Close()
	}, nil
}
