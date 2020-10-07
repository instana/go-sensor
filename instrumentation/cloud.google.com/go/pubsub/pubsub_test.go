package pubsub_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	gpubsub "cloud.google.com/go/pubsub"
	"cloud.google.com/go/pubsub/pstest"
	"github.com/instana/go-sensor/instrumentation/cloud.google.com/go/pubsub"
	"github.com/instana/testify/assert"
	"github.com/instana/testify/require"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	pb "google.golang.org/genproto/googleapis/pubsub/v1"
	"google.golang.org/grpc"
)

func TestClient_Topic(t *testing.T) {
	srv, conn, teardown, err := setupMockServer()
	require.NoError(t, err)
	defer teardown()

	_, err = srv.GServer.CreateTopic(context.Background(), &pb.Topic{
		Name: "projects/test-project/topics/test-topic",
	})
	require.NoError(t, err)

	examples := map[string]func(*testing.T, *gpubsub.Message) *gpubsub.PublishResult{
		"ClientProject": func(t *testing.T, msg *gpubsub.Message) *gpubsub.PublishResult {
			client, err := pubsub.NewClient(context.Background(), "test-project", option.WithGRPCConn(conn))
			require.NoError(t, err)

			top := client.Topic("test-topic")

			return top.Publish(context.Background(), msg)
		},
		"OtherProject": func(t *testing.T, msg *gpubsub.Message) *gpubsub.PublishResult {
			client, err := pubsub.NewClient(context.Background(), "other-project", option.WithGRPCConn(conn))
			require.NoError(t, err)

			top := client.TopicInProject("test-topic", "test-project")

			return top.Publish(context.Background(), msg)
		},
		"CreateTopic": func(t *testing.T, msg *gpubsub.Message) *gpubsub.PublishResult {
			client, err := pubsub.NewClient(context.Background(), "test-project", option.WithGRPCConn(conn))
			require.NoError(t, err)

			top, err := client.CreateTopic(context.Background(), "new-test-topic")
			require.NoError(t, err)

			return top.Publish(context.Background(), msg)
		},
		"CreateTopicWithConfig": func(t *testing.T, msg *gpubsub.Message) *gpubsub.PublishResult {
			client, err := pubsub.NewClient(context.Background(), "test-project", option.WithGRPCConn(conn))
			require.NoError(t, err)

			conf := &gpubsub.TopicConfig{
				MessageStoragePolicy: gpubsub.MessageStoragePolicy{
					AllowedPersistenceRegions: []string{"us-east1"},
				},
			}

			top, err := client.CreateTopicWithConfig(context.Background(), "new-test-topic-with-config", conf)
			require.NoError(t, err)

			createdConf, err := top.Config(context.Background())
			require.NoError(t, err)
			assert.Equal(t, *conf, createdConf)

			return top.Publish(context.Background(), msg)
		},
	}

	for name, publish := range examples {
		t.Run(name, func(t *testing.T) {
			srv.ClearMessages()

			res := publish(t, &gpubsub.Message{
				Data: []byte("message data"),
				Attributes: map[string]string{
					"key1": "value1",
				},
			})

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			msgID, err := res.Get(ctx)
			require.NoError(t, err)

			msg := srv.Message(msgID)
			require.NotNil(t, msg)

			assert.Equal(t, []byte("message data"), msg.Data)
			assert.Equal(t, map[string]string{
				"key1": "value1",
			}, msg.Attributes)
		})
	}
}

func TestClient_Topics(t *testing.T) {
	srv, conn, teardown, err := setupMockServer()
	require.NoError(t, err)
	defer teardown()

	topicNames := []string{"first-topic", "second-topic"}

	for _, topicName := range topicNames {
		_, err = srv.GServer.CreateTopic(context.Background(), &pb.Topic{
			Name: "projects/test-project/topics/" + topicName,
		})
		require.NoError(t, err)
	}

	client, err := pubsub.NewClient(context.Background(), "test-project", option.WithGRPCConn(conn))
	require.NoError(t, err)

	var res []*gpubsub.PublishResult

	it := client.Topics(context.Background())
	for {
		top, err := it.Next()
		if err == iterator.Done {
			break
		}

		res = append(res, top.Publish(context.Background(), &gpubsub.Message{
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
