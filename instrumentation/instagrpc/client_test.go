// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instagrpc_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/autoprofile"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instagrpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	grpctest "google.golang.org/grpc/test/grpc_testing"
)

func TestUnaryClientInterceptor(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	defer instana.ShutdownSensor()

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

	span, err := extractAgentSpan(spans[0])
	require.NoError(t, err)

	assert.Equal(t, "rpc-client", span.Name)
	assert.EqualValues(t, instana.ExitSpanKind, span.Kind)
	assert.Equal(t, 0, span.Ec)

	host, port, err := net.SplitHostPort(addr)
	require.NoError(t, err)

	assert.Equal(t, agentRPCSpanData{
		Host:     host,
		Port:     port,
		Flavor:   "grpc",
		CallType: "unary",
		Call:     "/grpc.testing.TestService/EmptyCall",
	}, span.Data.RPC)

	// ensure that trace context has been propagated
	require.NotNil(t, mdRec.MD)

	traceIDHeader := mdRec.MD.Get(instana.FieldT)
	require.Len(t, traceIDHeader, 1)

	assert.Equal(t, span.TraceID, traceIDHeader[0])

	spanIDHeader := mdRec.MD.Get(instana.FieldS)
	require.Len(t, spanIDHeader, 1)

	assert.Equal(t, span.SpanID, spanIDHeader[0])

	assert.Equal(t, []string{"banana"}, mdRec.MD.Get(instana.FieldB+"custom"))
}

func TestUnaryClientInterceptor_ErrorHandling(t *testing.T) {
	serverErr := status.Error(codes.Internal, "something went wrong")

	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder),
	)
	defer instana.ShutdownSensor()

	addr, teardown, err := startTestServer(&testServer{Error: serverErr})
	require.NoError(t, err)
	defer teardown()

	client, err := newTestServiceClient(
		addr,
		time.Second,
		grpc.WithUnaryInterceptor(instagrpc.UnaryClientInterceptor(sensor)),
	)
	require.NoError(t, err)

	sp := sensor.Tracer().StartSpan("test-span")

	_, err = client.EmptyCall(instana.ContextWithSpan(context.Background(), sp), &grpctest.Empty{})
	assert.Error(t, err)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 2)

	span, err := extractAgentSpan(spans[0])
	require.NoError(t, err)

	assert.Equal(t, "rpc-client", span.Name)
	assert.EqualValues(t, instana.ExitSpanKind, span.Kind)
	assert.Equal(t, 1, span.Ec)

	assert.Equal(t, serverErr.Error(), span.Data.RPC.Error)
}

func TestUnaryClientInterceptor_NoParentSpan(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	defer instana.ShutdownSensor()

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

	_, err = client.EmptyCall(context.Background(), &grpctest.Empty{})
	require.NoError(t, err)

	// check recorded spans
	assert.Empty(t, recorder.GetQueuedSpans())
}

func TestStreamClientInterceptor(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	defer instana.ShutdownSensor()

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

	span, err := extractAgentSpan(spans[0])
	require.NoError(t, err)

	assert.Equal(t, "rpc-client", span.Name)
	assert.EqualValues(t, instana.ExitSpanKind, span.Kind)
	assert.Equal(t, 0, span.Ec)

	host, port, err := net.SplitHostPort(addr)
	require.NoError(t, err)

	assert.Equal(t, agentRPCSpanData{
		Host:     host,
		Port:     port,
		Flavor:   "grpc",
		CallType: "stream",
		Call:     "/grpc.testing.TestService/FullDuplexCall",
	}, span.Data.RPC)

	// ensure that trace context has been propagated
	require.NotNil(t, mdRec.MD)

	traceIDHeader := mdRec.MD.Get(instana.FieldT)
	require.Len(t, traceIDHeader, 1)

	assert.Equal(t, span.TraceID, traceIDHeader[0])

	spanIDHeader := mdRec.MD.Get(instana.FieldS)
	require.Len(t, spanIDHeader, 1)

	assert.Equal(t, span.SpanID, spanIDHeader[0])

	assert.Equal(t, []string{"banana"}, mdRec.MD.Get(instana.FieldB+"custom"))
}

func TestStreamClientInterceptor_ErrorHandling(t *testing.T) {
	serverErr := status.Error(codes.Internal, "something went wrong")

	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder),
	)
	defer instana.ShutdownSensor()

	addr, teardown, err := startTestServer(&testServer{Error: serverErr})
	require.NoError(t, err)
	defer teardown()

	client, err := newTestServiceClient(
		addr,
		time.Second,
		grpc.WithStreamInterceptor(instagrpc.StreamClientInterceptor(sensor)),
	)
	require.NoError(t, err)

	sp := sensor.Tracer().StartSpan("test-span")

	stream, err := client.FullDuplexCall(instana.ContextWithSpan(context.Background(), sp))
	require.NoError(t, err)

	require.NoError(t, stream.Send(&grpctest.StreamingOutputCallRequest{}))
	require.NoError(t, stream.CloseSend())

	_, err = stream.Recv()
	assert.Error(t, err)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 2)

	span, err := extractAgentSpan(spans[0])
	require.NoError(t, err)

	assert.Equal(t, "rpc-client", span.Name)
	assert.EqualValues(t, instana.ExitSpanKind, span.Kind)
	assert.Equal(t, 1, span.Ec)

	assert.Equal(t, serverErr.Error(), span.Data.RPC.Error)
}

func TestStreamClientInterceptor_NoParentSpan(t *testing.T) {
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	defer instana.ShutdownSensor()

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

	stream, err := client.FullDuplexCall(context.Background())
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

	assert.Empty(t, recorder.GetQueuedSpans())
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

type alwaysReadyClient struct{}

func (alwaysReadyClient) Ready() bool                                       { return true }
func (alwaysReadyClient) SendMetrics(data acceptor.Metrics) error           { return nil }
func (alwaysReadyClient) SendEvent(event *instana.EventData) error          { return nil }
func (alwaysReadyClient) SendSpans(spans []instana.Span) error              { return nil }
func (alwaysReadyClient) SendProfiles(profiles []autoprofile.Profile) error { return nil }
func (alwaysReadyClient) Flush(context.Context) error                       { return nil }
