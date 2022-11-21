// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package pubsub_test

import (
	"context"
	"testing"
	"time"

	"cloud.google.com/go/pubsub/pstest"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/cloud.google.com/go/pubsub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"
	pb "google.golang.org/genproto/googleapis/pubsub/v1"
)

func TestSubscription_Receive(t *testing.T) {
	pstest.SetMinAckDeadline(1 * time.Second)
	defer pstest.ResetMinAckDeadline()

	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	defer instana.ShutdownSensor()

	srv, conn, teardown, err := setupMockServer()
	require.NoError(t, err)
	defer teardown()

	top, err := srv.GServer.CreateTopic(context.Background(), &pb.Topic{
		Name: "projects/test-project/topics/test-topic",
	})
	require.NoError(t, err)

	_, err = srv.GServer.CreateSubscription(context.Background(), &pb.Subscription{
		Topic:              top.Name,
		Name:               "projects/test-project/subscriptions/test-subscription",
		AckDeadlineSeconds: 10,
	})
	require.NoError(t, err)

	msgID := srv.Publish(top.Name, []byte("test message"), map[string]string{
		"x-instana-t": "0000000000001234",
		"x-instana-s": "0000000000005678",
		"x-instana-l": "1",
	})

	client, err := pubsub.NewClient(
		context.Background(),
		"test-project",
		instana.NewSensorWithTracer(tracer),
		option.WithGRPCConn(conn),
	)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	client.Subscription("test-subscription").Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		assert.Equal(t, msgID, msg.ID)
		msg.Ack()
		cancel()
	})

	assert.Equal(t, context.Canceled, ctx.Err())

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	gcpsSpan := spans[0]

	// trace continuation
	assert.EqualValues(t, 0x1234, gcpsSpan.TraceID)
	assert.EqualValues(t, 0x5678, gcpsSpan.ParentID)
	assert.NotEqual(t, gcpsSpan.ParentID, gcpsSpan.SpanID)

	// span tags
	assert.Equal(t, "gcps", gcpsSpan.Name)
	assert.EqualValues(t, instana.EntrySpanKind, gcpsSpan.Kind)
	assert.Equal(t, 0, gcpsSpan.Ec)

	require.IsType(t, instana.GCPPubSubSpanData{}, gcpsSpan.Data)

	data := gcpsSpan.Data.(instana.GCPPubSubSpanData)
	assert.Equal(t, instana.GCPPubSubSpanTags{
		Operation:    "CONSUME",
		ProjectID:    "test-project",
		Subscription: "test-subscription",
		MessageID:    msgID,
	}, data.Tags)
}

func TestSubscription_Receive_NoTrace(t *testing.T) {
	pstest.SetMinAckDeadline(1 * time.Second)
	defer pstest.ResetMinAckDeadline()

	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	defer instana.ShutdownSensor()

	srv, conn, teardown, err := setupMockServer()
	require.NoError(t, err)
	defer teardown()

	top, err := srv.GServer.CreateTopic(context.Background(), &pb.Topic{
		Name: "projects/test-project/topics/test-topic",
	})
	require.NoError(t, err)

	_, err = srv.GServer.CreateSubscription(context.Background(), &pb.Subscription{
		Topic:              top.Name,
		Name:               "projects/test-project/subscriptions/test-subscription",
		AckDeadlineSeconds: 10,
	})
	require.NoError(t, err)

	msgID := srv.Publish(top.Name, []byte("test message"), nil)

	client, err := pubsub.NewClient(
		context.Background(),
		"test-project",
		instana.NewSensorWithTracer(tracer),
		option.WithGRPCConn(conn),
	)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	client.Subscription("test-subscription").Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		assert.Equal(t, msgID, msg.ID)
		msg.Ack()
		cancel()
	})

	assert.Equal(t, context.Canceled, ctx.Err())

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	gcpsSpan := spans[0]

	// new trace started
	assert.NotEmpty(t, gcpsSpan.TraceID)
	assert.Empty(t, gcpsSpan.ParentID)
	assert.NotEmpty(t, gcpsSpan.SpanID)

	// span tags
	assert.Equal(t, "gcps", gcpsSpan.Name)
	assert.EqualValues(t, instana.EntrySpanKind, gcpsSpan.Kind)
	assert.Equal(t, 0, gcpsSpan.Ec)

	require.IsType(t, instana.GCPPubSubSpanData{}, gcpsSpan.Data)

	data := gcpsSpan.Data.(instana.GCPPubSubSpanData)
	assert.Equal(t, instana.GCPPubSubSpanTags{
		Operation:    "CONSUME",
		ProjectID:    "test-project",
		Subscription: "test-subscription",
		MessageID:    msgID,
	}, data.Tags)
}
