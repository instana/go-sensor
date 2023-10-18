// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

//go:build go1.17
// +build go1.17

package instagrpc_test

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instagrpc"
	"github.com/opentracing/opentracing-go/ext"
	"google.golang.org/grpc"
	grpctest "google.golang.org/grpc/interop/grpc_testing"
)

// EchoServer is an implementation of GRPC server
type TestServiceServer struct {
	grpctest.UnimplementedTestServiceServer
}

// UnaryCall responds with a static greeting from server
func (s TestServiceServer) UnaryCall(ctx context.Context, req *grpctest.SimpleRequest) (*grpctest.SimpleResponse, error) {
	// Extract the parent span and use its tracer to initialize any child spans to trace the calls
	// inside the handler, e.g. database queries, 3rd-party API requests, etc.
	if parent, ok := instana.SpanFromContext(ctx); ok {
		sp := parent.Tracer().StartSpan("unary-call")
		defer sp.Finish()
	}

	time.Sleep(100 * time.Microsecond)

	return &grpctest.SimpleResponse{
		Payload: &grpctest.Payload{
			Body: []byte("hello from server"),
		},
	}, nil
}

// setupServer starts a new instrumented GRPC server instance and returns
// the server address
func setupServer() (net.Addr, error) {
	// Initialize server sensor to instrument request handlers
	sensor := instana.NewSensor("grpc-server")

	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, fmt.Errorf("failed to start listener: %s", err)
	}

	// To instrument server calls add instagrpc.UnaryServerInterceptor(sensor) and
	// instagrpc.StreamServerInterceptor(sensor) to the list of server options when
	// initializing the server
	srv := grpc.NewServer(
		grpc.UnaryInterceptor(instagrpc.UnaryServerInterceptor(sensor)),
		grpc.StreamInterceptor(instagrpc.StreamServerInterceptor(sensor)),
	)

	grpctest.RegisterTestServiceServer(srv, &TestServiceServer{})
	go func() {
		if err := srv.Serve(ln); err != nil {
			log.Fatalf("failed to start server: %s", err)
		}
	}()

	return ln.Addr(), nil
}

func Example() {
	serverAddr, err := setupServer()
	if err != nil {
		log.Fatalf("failed to setup a server: %s", err)
	}

	// Initialize client tracer
	sensor := instana.NewSensor("grpc-client")

	// To instrument client calls add instagrpc.UnaryClientInterceptor(sensor) and
	// instagrpc.StringClientInterceptor(sensor) to the DialOption list while dialing
	// the GRPC server.
	conn, err := grpc.Dial(
		serverAddr.String(),
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(instagrpc.UnaryClientInterceptor(sensor)),
		grpc.WithStreamInterceptor(instagrpc.StreamClientInterceptor(sensor)),
	)
	if err != nil {
		log.Fatalf("failed to dial server on %s: %s", serverAddr.String(), err)
	}
	defer conn.Close()

	c := grpctest.NewTestServiceClient(conn)

	// The call should always start with an entry span (https://www.instana.com/docs/tracing/custom-best-practices/#start-new-traces-with-entry-spans)
	// Normally this would be your HTTP/GRPC/message queue request span, but here we need to
	// create it explicitly.
	sp := sensor.Tracer().StartSpan("client-call")
	sp.SetTag(string(ext.SpanKind), "entry")

	// Create a context that holds the parent entry span and pass it to the GRPC call
	resp, err := c.UnaryCall(
		instana.ContextWithSpan(context.Background(), sp),
		&grpctest.SimpleRequest{
			Payload: &grpctest.Payload{
				Body: []byte("hello from client"),
			},
		},
	)
	if err != nil {
		log.Fatalf("server responded with an error: %s", err)
	}
	fmt.Println(string(resp.GetPayload().GetBody()))
	// Output: hello from server
}
