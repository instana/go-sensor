// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package pubsub_test

import (
	"context"
	"sort"
	"testing"
	"time"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/cloud.google.com/go/pubsub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"
	pb "google.golang.org/genproto/googleapis/pubsub/v1"
)

func TestTopic_Publish(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	defer instana.ShutdownSensor()

	srv, conn, teardown, err := setupMockServer()
	require.NoError(t, err)
	defer teardown()

	_, err = srv.GServer.CreateTopic(context.Background(), &pb.Topic{
		Name: "projects/test-project/topics/test-topic",
	})
	require.NoError(t, err)

	client, err := pubsub.NewClient(
		context.Background(),
		"test-project",
		instana.NewSensorWithTracer(tracer),
		option.WithGRPCConn(conn),
	)
	require.NoError(t, err)

	parent := tracer.StartSpan("testing")

	ctx := instana.ContextWithSpan(context.Background(), parent)
	res := client.Topic("test-topic").Publish(ctx, &pubsub.Message{
		Data: []byte("message data"),
		Attributes: map[string]string{
			"key1": "value1",
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	msgID, err := res.Get(ctx)
	require.NoError(t, ctx.Err())
	require.NoError(t, err)

	parent.Finish()

	require.Eventually(t, func() bool {
		return recorder.QueuedSpansCount() == 2
	}, 250*time.Millisecond, 25*time.Millisecond)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 2)

	sort.Slice(spans, func(i, j int) bool {
		return spans[i].Name < spans[j].Name
	})
	testSpan, gcpsSpan := spans[1], spans[0]

	// trace continuation
	assert.Equal(t, testSpan.TraceID, gcpsSpan.TraceID)
	assert.Equal(t, testSpan.SpanID, gcpsSpan.ParentID)
	assert.NotEqual(t, gcpsSpan.ParentID, gcpsSpan.SpanID)

	// span tags
	assert.Equal(t, "gcps", gcpsSpan.Name)
	assert.EqualValues(t, instana.ExitSpanKind, gcpsSpan.Kind)
	assert.Equal(t, 0, gcpsSpan.Ec)

	require.IsType(t, instana.GCPPubSubSpanData{}, gcpsSpan.Data)

	data := gcpsSpan.Data.(instana.GCPPubSubSpanData)
	assert.Equal(t, instana.GCPPubSubSpanTags{
		Operation: "PUBLISH",
		ProjectID: "test-project",
		Topic:     "test-topic",
		MessageID: msgID,
	}, data.Tags)

	// trace context propagation
	msg := srv.Message(msgID)
	require.NotNil(t, msg)

	assert.Equal(t, map[string]string{
		"x-instana-t": instana.FormatID(gcpsSpan.TraceID),
		"x-instana-s": instana.FormatID(gcpsSpan.SpanID),
		"x-instana-l": "1",
		"key1":        "value1",
	}, msg.Attributes)
}

func TestTopic_Publish_NoTrace(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(instana.DefaultOptions(), recorder)
	defer instana.ShutdownSensor()

	srv, conn, teardown, err := setupMockServer()
	require.NoError(t, err)
	defer teardown()

	srv.GServer.CreateTopic(context.Background(), &pb.Topic{
		Name: "projects/test-project/topics/test-topic",
	})

	client, err := pubsub.NewClient(
		context.Background(),
		"test-project",
		instana.NewSensorWithTracer(tracer),
		option.WithGRPCConn(conn),
	)
	require.NoError(t, err)

	res := client.Topic("test-topic").Publish(context.Background(), &pubsub.Message{
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

	assert.Equal(t, map[string]string{
		"key1": "value1",
	}, msg.Attributes)
}
