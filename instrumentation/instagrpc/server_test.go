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
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	grpctest "google.golang.org/grpc/test/grpc_testing"
)

func TestUnaryServerInterceptor(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{}, recorder),
	)

	addr, teardown, err := startTestServer(
		&testServer{},
		grpc.UnaryInterceptor(instagrpc.UnaryServerInterceptor(sensor)),
	)
	require.NoError(t, err)
	defer teardown()

	client, err := newTestServiceClient(addr, time.Second)
	require.NoError(t, err)

	_, err = client.EmptyCall(context.Background(), &grpctest.Empty{})
	require.NoError(t, err)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	span, err := extractAgentSpan(spans[0])
	require.NoError(t, err)

	assert.False(t, span.Error)
	assert.Equal(t, 0, span.Ec)

	assert.Equal(t, "rpc-server", span.Data.SDK.Name)
	assert.Equal(t, "entry", span.Data.SDK.Type)

	host, port, err := net.SplitHostPort(addr)
	require.NoError(t, err)

	assert.Equal(t, ot.Tags{
		"span.kind":     string(ext.SpanKindRPCServerEnum),
		"rpc.host":      host,
		"rpc.port":      port,
		"rpc.flavor":    "grpc",
		"rpc.call_type": "unary",
		"rpc.call":      "/grpc.testing.TestService/EmptyCall",
	}, span.Data.SDK.Custom.Tags)
}

func TestUnaryServerInterceptor_WithClientTraceID(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{}, recorder),
	)

	addr, teardown, err := startTestServer(
		&testServer{},
		grpc.UnaryInterceptor(instagrpc.UnaryServerInterceptor(sensor)),
	)
	require.NoError(t, err)
	defer teardown()

	client, err := newTestServiceClient(addr, time.Second)
	require.NoError(t, err)

	md := metadata.New(map[string]string{
		instana.FieldT:            instana.FormatID(1234567890),
		instana.FieldS:            instana.FormatID(1),
		instana.FieldB + "custom": "banana",
	})

	_, err = client.EmptyCall(
		metadata.NewOutgoingContext(context.Background(), md),
		&grpctest.Empty{},
	)
	require.NoError(t, err)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	span, err := extractAgentSpan(spans[0])
	require.NoError(t, err)

	assert.Equal(t, int64(1234567890), span.TraceID)
	assert.Equal(t, int64(1), span.ParentID)
	assert.Equal(t, "banana", span.Data.SDK.Custom.Baggage["custom"])
}

func TestUnaryServerInterceptor_ErrorHandling(t *testing.T) {
	serverErr := status.Error(codes.Internal, "something went wrong")

	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{}, recorder),
	)

	addr, teardown, err := startTestServer(
		&testServer{Error: serverErr},
		grpc.UnaryInterceptor(instagrpc.UnaryServerInterceptor(sensor)),
	)
	require.NoError(t, err)
	defer teardown()

	client, err := newTestServiceClient(addr, time.Second)
	require.NoError(t, err)

	_, err = client.EmptyCall(context.Background(), &grpctest.Empty{})
	assert.Error(t, err)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	span, err := extractAgentSpan(spans[0])
	require.NoError(t, err)

	assert.True(t, span.Error)
	assert.Equal(t, 1, span.Ec)

	assert.Equal(t, "rpc-server", span.Data.SDK.Name)
	assert.Equal(t, "entry", span.Data.SDK.Type)

	assert.Equal(t, serverErr.Error(), span.Data.SDK.Custom.Tags["rpc.error"])

	var logRecords []map[string]interface{}
	for _, v := range span.Data.SDK.Custom.Logs {
		logRecords = append(logRecords, v)
	}

	require.Len(t, logRecords, 1)
	assert.Equal(t, map[string]interface{}{
		"code":    float64(codes.Internal),
		"message": "something went wrong",
	}, logRecords[0]["error"])
}

func TestUnaryServerInterceptor_PanicHandling(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{}, recorder),
	)

	addr, teardown, err := startTestServer(
		&panickingTestServer{},
		suppressUnaryHandlerPanics(instagrpc.UnaryServerInterceptor(sensor)),
	)
	require.NoError(t, err)
	defer teardown()

	client, err := newTestServiceClient(addr, time.Second)
	require.NoError(t, err)

	_, err = client.EmptyCall(context.Background(), &grpctest.Empty{})
	require.NoError(t, err)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	span, err := extractAgentSpan(spans[0])
	require.NoError(t, err)

	assert.True(t, span.Error)
	assert.Equal(t, 1, span.Ec)

	assert.Equal(t, "rpc-server", span.Data.SDK.Name)
	assert.Equal(t, "entry", span.Data.SDK.Type)

	assert.Equal(t, "something went wrong", span.Data.SDK.Custom.Tags["rpc.error"])

	var logRecords []map[string]interface{}
	for _, v := range span.Data.SDK.Custom.Logs {
		logRecords = append(logRecords, v)
	}

	require.Len(t, logRecords, 1)
	assert.Equal(t, "something went wrong", logRecords[0]["error"])
}

func TestStreamServerInterceptor(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{}, recorder),
	)

	addr, teardown, err := startTestServer(
		&testServer{},
		grpc.StreamInterceptor(instagrpc.StreamServerInterceptor(sensor)),
	)
	require.NoError(t, err)
	defer teardown()

	client, err := newTestServiceClient(addr, time.Second)
	require.NoError(t, err)

	stream, err := client.FullDuplexCall(context.Background())
	require.NoError(t, err)

	for i := 0; i < 2; i++ {
		require.NoError(t, stream.Send(&grpctest.StreamingOutputCallRequest{}))
	}
	require.NoError(t, stream.CloseSend())

	_, err = stream.Recv()
	require.NoError(t, err)

	require.Eventually(t, func() bool { return recorder.QueuedSpansCount() == 1 }, 100*time.Millisecond, 50*time.Millisecond)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	span, err := extractAgentSpan(spans[0])
	require.NoError(t, err)

	assert.False(t, span.Error)
	assert.Equal(t, 0, span.Ec)

	assert.Equal(t, "rpc-server", span.Data.SDK.Name)
	assert.Equal(t, "entry", span.Data.SDK.Type)

	host, port, err := net.SplitHostPort(addr)
	require.NoError(t, err)

	assert.Equal(t, ot.Tags{
		"span.kind":     string(ext.SpanKindRPCServerEnum),
		"rpc.host":      host,
		"rpc.port":      port,
		"rpc.flavor":    "grpc",
		"rpc.call_type": "stream",
		"rpc.call":      "/grpc.testing.TestService/FullDuplexCall",
	}, span.Data.SDK.Custom.Tags)
}

func TestStreamServerInterceptor_WithClientTraceID(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{}, recorder),
	)

	addr, teardown, err := startTestServer(
		&testServer{},
		grpc.StreamInterceptor(instagrpc.StreamServerInterceptor(sensor)),
	)
	require.NoError(t, err)
	defer teardown()

	client, err := newTestServiceClient(addr, time.Second)
	require.NoError(t, err)

	md := metadata.New(map[string]string{
		instana.FieldT:            instana.FormatID(1234567890),
		instana.FieldS:            instana.FormatID(1),
		instana.FieldB + "custom": "banana",
	})

	stream, err := client.FullDuplexCall(metadata.NewOutgoingContext(context.Background(), md))
	require.NoError(t, err)

	require.NoError(t, stream.Send(&grpctest.StreamingOutputCallRequest{}))
	require.NoError(t, stream.CloseSend())

	_, err = stream.Recv()
	require.NoError(t, err)

	require.Eventually(t, func() bool { return recorder.QueuedSpansCount() == 1 }, 100*time.Millisecond, 50*time.Millisecond)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	span, err := extractAgentSpan(spans[0])
	require.NoError(t, err)

	assert.Equal(t, int64(1234567890), span.TraceID)
	assert.Equal(t, int64(1), span.ParentID)
	assert.Equal(t, "banana", span.Data.SDK.Custom.Baggage["custom"])
}

func TestStreamServerInterceptor_ErrorHandling(t *testing.T) {
	serverErr := status.Error(codes.Internal, "something went wrong")

	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{}, recorder),
	)

	addr, teardown, err := startTestServer(
		&testServer{Error: serverErr},
		grpc.StreamInterceptor(instagrpc.StreamServerInterceptor(sensor)),
	)
	require.NoError(t, err)
	defer teardown()

	client, err := newTestServiceClient(addr, time.Second)
	require.NoError(t, err)

	stream, err := client.FullDuplexCall(context.Background())
	require.NoError(t, err)

	require.NoError(t, stream.Send(&grpctest.StreamingOutputCallRequest{}))
	require.NoError(t, stream.CloseSend())

	_, err = stream.Recv()
	assert.Error(t, err)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	span, err := extractAgentSpan(spans[0])
	require.NoError(t, err)

	assert.True(t, span.Error)
	assert.Equal(t, 1, span.Ec)

	assert.Equal(t, "rpc-server", span.Data.SDK.Name)
	assert.Equal(t, "entry", span.Data.SDK.Type)

	assert.Equal(t, serverErr.Error(), span.Data.SDK.Custom.Tags["rpc.error"])

	var logRecords []map[string]interface{}
	for _, v := range span.Data.SDK.Custom.Logs {
		logRecords = append(logRecords, v)
	}

	require.Len(t, logRecords, 1)
	assert.Equal(t, map[string]interface{}{
		"code":    float64(codes.Internal),
		"message": "something went wrong",
	}, logRecords[0]["error"])
}

func TestStreamServerInterceptor_PanicHandling(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{}, recorder),
	)

	addr, teardown, err := startTestServer(
		&panickingTestServer{},
		suppressStreamHandlerPanics(instagrpc.StreamServerInterceptor(sensor)),
	)
	require.NoError(t, err)
	defer teardown()

	client, err := newTestServiceClient(addr, time.Second)
	require.NoError(t, err)

	stream, err := client.FullDuplexCall(context.Background())
	require.NoError(t, err)

	for i := 0; i < 2; i++ {
		stream.Send(&grpctest.StreamingOutputCallRequest{})
	}
	require.NoError(t, stream.CloseSend())

	require.Eventually(t, func() bool { return recorder.QueuedSpansCount() == 1 }, 100*time.Millisecond, 50*time.Millisecond)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	span, err := extractAgentSpan(spans[0])
	require.NoError(t, err)

	assert.True(t, span.Error)
	assert.Equal(t, 1, span.Ec)

	assert.Equal(t, "rpc-server", span.Data.SDK.Name)
	assert.Equal(t, "entry", span.Data.SDK.Type)

	assert.Equal(t, "something went wrong", span.Data.SDK.Custom.Tags["rpc.error"])

	var logRecords []map[string]interface{}
	for _, v := range span.Data.SDK.Custom.Logs {
		logRecords = append(logRecords, v)
	}

	require.Len(t, logRecords, 1)
	assert.Equal(t, "something went wrong", logRecords[0]["error"])
}

func startTestServer(ts grpctest.TestServiceServer, opts ...grpc.ServerOption) (string, func(), error) {
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		return "", nil, fmt.Errorf("failed to listen: %s", err)
	}

	srv := grpc.NewServer(opts...)
	grpctest.RegisterTestServiceServer(srv, ts)

	go srv.Serve(ln)

	return ln.Addr().String(), srv.Stop, nil
}

func suppressUnaryHandlerPanics(next grpc.UnaryServerInterceptor) grpc.ServerOption {
	return grpc.UnaryInterceptor(
		func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			defer func() {
				// suppress server panic
				recover()
			}()

			return next(ctx, req, info, handler)
		},
	)
}

func suppressStreamHandlerPanics(next grpc.StreamServerInterceptor) grpc.ServerOption {
	return grpc.StreamInterceptor(
		func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
			defer func() {
				// suppress server panic
				recover()
			}()

			return next(srv, ss, info, handler)
		},
	)
}

// basic implementation of grpctest.TestServiceServer with all handlers returning "Unimplemented" error
type unimplementedTestServer struct{}

func (ts unimplementedTestServer) EmptyCall(ctx context.Context, req *grpctest.Empty) (*grpctest.Empty, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (ts unimplementedTestServer) UnaryCall(context.Context, *grpctest.SimpleRequest) (*grpctest.SimpleResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (ts unimplementedTestServer) StreamingOutputCall(*grpctest.StreamingOutputCallRequest, grpctest.TestService_StreamingOutputCallServer) error {
	return status.Error(codes.Unimplemented, "not implemented")
}

func (ts unimplementedTestServer) StreamingInputCall(grpctest.TestService_StreamingInputCallServer) error {
	return status.Error(codes.Unimplemented, "not implemented")
}

func (ts unimplementedTestServer) FullDuplexCall(s grpctest.TestService_FullDuplexCallServer) error {
	return status.Error(codes.Unimplemented, "not implemented")
}

func (ts unimplementedTestServer) HalfDuplexCall(grpctest.TestService_HalfDuplexCallServer) error {
	return status.Error(codes.Unimplemented, "not implemented")
}

// a test server that optionally returns an error on EmptyCall and FullDuplexCall requests
type testServer struct {
	unimplementedTestServer
	Error error
}

func (ts *testServer) EmptyCall(ctx context.Context, req *grpctest.Empty) (*grpctest.Empty, error) {
	return &grpctest.Empty{}, ts.Error
}

func (ts *testServer) FullDuplexCall(s grpctest.TestService_FullDuplexCallServer) error {
	for {
		_, err := s.Recv()
		if err == io.EOF {
			break
		}
	}

	if ts.Error == nil {
		s.Send(&grpctest.StreamingOutputCallResponse{})
	}

	return ts.Error
}

// a test server that throws panics on EmptyCall requests
type panickingTestServer struct {
	unimplementedTestServer
}

func (ts *panickingTestServer) EmptyCall(ctx context.Context, req *grpctest.Empty) (*grpctest.Empty, error) {
	panic("something went wrong")
}

func (ts *panickingTestServer) FullDuplexCall(s grpctest.TestService_FullDuplexCallServer) error {
	panic("something went wrong")
}

type metadataCapturer struct {
	MD metadata.MD
}

func (s *metadataCapturer) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			s.MD = md
		}

		return handler(ctx, req)
	}
}

func (s *metadataCapturer) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if md, ok := metadata.FromIncomingContext(ss.Context()); ok {
			s.MD = md
		}

		return handler(srv, ss)
	}
}
