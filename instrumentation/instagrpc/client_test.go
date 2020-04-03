// +build go1.9

package instagrpc_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instagrpc"
	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	grpctest "google.golang.org/grpc/test/grpc_testing"
)

func TestUnaryClientInterceptor(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

	mdRec := &metadataCapturer{}
	addr, teardown, err := startTestServer(
		&testServer{},
		grpc.UnaryInterceptor(mdRec.UnaryServerInterceptor()),
	)
	require.NoError(t, err)
	defer teardown()

	client, err := newTestServiceClient(
		addr,
		time.Second,
		grpc.WithUnaryInterceptor(
			instagrpc.UnaryClientInterceptor(instana.NewSensorWithTracer(tracer)),
		),
	)
	require.NoError(t, err)

	sp := tracer.StartSpan("test-span")
	sp.SetBaggageItem("custom", "banana")

	_, err = client.EmptyCall(
		instana.ContextWithSpan(context.Background(), sp),
		&grpctest.Empty{},
	)
	require.NoError(t, err)

	// check recorded spans
	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	assert.False(t, span.Error)
	assert.Equal(t, 0, span.Ec)

	require.NotNil(t, span.Data)
	require.NotNil(t, span.Data.SDK)
	assert.Equal(t, "rpc-client", span.Data.SDK.Name)
	assert.Equal(t, "exit", span.Data.SDK.Type)

	host, port, err := net.SplitHostPort(addr)
	require.NoError(t, err)

	require.NotNil(t, span.Data.SDK.Custom)
	assert.Equal(t, ot.Tags{
		"span.kind":     ext.SpanKindRPCClientEnum,
		"rpc.host":      host,
		"rpc.port":      port,
		"rpc.flavor":    "grpc",
		"rpc.call_type": "unary",
		"rpc.call":      "/grpc.testing.TestService/EmptyCall",
	}, span.Data.SDK.Custom.Tags)

	// ensure that trace context has been propagated
	require.NotNil(t, mdRec.MD)

	traceIDHeader := mdRec.MD.Get(instana.FieldT)
	require.Len(t, traceIDHeader, 1)

	traceID, err := instana.ParseID(traceIDHeader[0])
	require.NoError(t, err)
	assert.Equal(t, span.TraceID, traceID)

	spanIDHeader := mdRec.MD.Get(instana.FieldS)
	require.Len(t, spanIDHeader, 1)

	spanID, err := instana.ParseID(spanIDHeader[0])
	require.NoError(t, err)
	assert.Equal(t, span.SpanID, spanID)

	assert.Equal(t, []string{"banana"}, mdRec.MD.Get(instana.FieldB+"custom"))
}

func TestUnaryClientInterceptor_ErrorHandling(t *testing.T) {
	serverErr := status.Error(codes.Internal, "something went wrong")

	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{}, recorder),
	)

	addr, teardown, err := startTestServer(&testServer{Error: serverErr})
	require.NoError(t, err)
	defer teardown()

	client, err := newTestServiceClient(
		addr,
		time.Second,
		grpc.WithUnaryInterceptor(instagrpc.UnaryClientInterceptor(sensor)),
	)
	require.NoError(t, err)

	_, err = client.EmptyCall(context.Background(), &grpctest.Empty{})
	assert.Error(t, err)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	assert.True(t, span.Error)
	assert.Equal(t, 1, span.Ec)

	require.NotNil(t, span.Data)
	require.NotNil(t, span.Data.SDK)
	assert.Equal(t, "rpc-client", span.Data.SDK.Name)
	assert.Equal(t, "exit", span.Data.SDK.Type)

	require.NotNil(t, span.Data.SDK.Custom)
	assert.Equal(t, serverErr.Error(), span.Data.SDK.Custom.Tags["rpc.error"])

	var logRecords []map[string]interface{}
	for _, v := range span.Data.SDK.Custom.Logs {
		logRecords = append(logRecords, v)
	}

	require.Len(t, logRecords, 1)
	assert.Equal(t, serverErr, logRecords[0]["error"])
}

func TestStreamClientInterceptor(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{}, recorder)

	mdRec := &metadataCapturer{}
	addr, teardown, err := startTestServer(
		&testServer{},
		grpc.StreamInterceptor(mdRec.StreamServerInterceptor()),
	)
	require.NoError(t, err)
	defer teardown()

	client, err := newTestServiceClient(
		addr,
		time.Second,
		grpc.WithStreamInterceptor(
			instagrpc.StreamClientInterceptor(instana.NewSensorWithTracer(tracer)),
		),
	)
	require.NoError(t, err)

	sp := tracer.StartSpan("test-span")
	sp.SetBaggageItem("custom", "banana")

	stream, err := client.FullDuplexCall(instana.ContextWithSpan(context.Background(), sp))
	require.NoError(t, err)

	for i := 0; i < 2; i++ {
		require.NoError(t, stream.Send(&grpctest.StreamingOutputCallRequest{}))
	}
	require.NoError(t, stream.CloseSend())

	done := make(chan struct{})
	go func() {
		for {
			_, err := stream.Recv()
			if err == io.EOF {
				break
			}
		}

		close(done)
	}()

	timeout := time.After(time.Second)
	select {
	case <-timeout:
		require.FailNow(t, "the test server has never stopped streaming")
	case <-done:
		break
	}

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	assert.False(t, span.Error)
	assert.Equal(t, 0, span.Ec)

	require.NotNil(t, span.Data)
	require.NotNil(t, span.Data.SDK)
	assert.Equal(t, "rpc-client", span.Data.SDK.Name)
	assert.Equal(t, "exit", span.Data.SDK.Type)

	host, port, err := net.SplitHostPort(addr)
	require.NoError(t, err)

	require.NotNil(t, span.Data.SDK.Custom)
	assert.Equal(t, ot.Tags{
		"span.kind":     ext.SpanKindRPCClientEnum,
		"rpc.host":      host,
		"rpc.port":      port,
		"rpc.flavor":    "grpc",
		"rpc.call_type": "stream",
		"rpc.call":      "/grpc.testing.TestService/FullDuplexCall",
	}, span.Data.SDK.Custom.Tags)

	// ensure that trace context has been propagated
	require.NotNil(t, mdRec.MD)

	traceIDHeader := mdRec.MD.Get(instana.FieldT)
	require.Len(t, traceIDHeader, 1)

	traceID, err := instana.ParseID(traceIDHeader[0])
	require.NoError(t, err)
	assert.Equal(t, span.TraceID, traceID)

	spanIDHeader := mdRec.MD.Get(instana.FieldS)
	require.Len(t, spanIDHeader, 1)

	spanID, err := instana.ParseID(spanIDHeader[0])
	require.NoError(t, err)
	assert.Equal(t, span.SpanID, spanID)

	assert.Equal(t, []string{"banana"}, mdRec.MD.Get(instana.FieldB+"custom"))
}

func TestStreamClientInterceptor_ErrorHandling(t *testing.T) {
	serverErr := status.Error(codes.Internal, "something went wrong")

	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{}, recorder),
	)

	addr, teardown, err := startTestServer(&testServer{Error: serverErr})
	require.NoError(t, err)
	defer teardown()

	client, err := newTestServiceClient(
		addr,
		time.Second,
		grpc.WithStreamInterceptor(instagrpc.StreamClientInterceptor(sensor)),
	)
	require.NoError(t, err)

	stream, err := client.FullDuplexCall(context.Background())
	require.NoError(t, err)

	require.NoError(t, stream.Send(&grpctest.StreamingOutputCallRequest{}))
	require.NoError(t, stream.CloseSend())

	_, err = stream.Recv()
	assert.Error(t, err)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	assert.True(t, span.Error)
	assert.Equal(t, 1, span.Ec)

	require.NotNil(t, span.Data)
	require.NotNil(t, span.Data.SDK)
	assert.Equal(t, "rpc-client", span.Data.SDK.Name)
	assert.Equal(t, "exit", span.Data.SDK.Type)

	require.NotNil(t, span.Data.SDK.Custom)
	assert.Equal(t, serverErr.Error(), span.Data.SDK.Custom.Tags["rpc.error"])

	var logRecords []map[string]interface{}
	for _, v := range span.Data.SDK.Custom.Logs {
		logRecords = append(logRecords, v)
	}

	require.Len(t, logRecords, 1)
	assert.Equal(t, serverErr, logRecords[0]["error"])
}

func newTestServiceClient(addr string, timeout time.Duration, opts ...grpc.DialOption) (grpctest.TestServiceClient, error) {
	opts = append(opts, grpc.WithInsecure(), grpc.WithBlock())

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to establish connection with test server: %s", err)
	}

	return grpctest.NewTestServiceClient(conn), nil
}
