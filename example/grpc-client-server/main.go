// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2016

package main

import (
	"context"
	"flag"
	"log"
	"net"
	"os"

	"github.com/opentracing/opentracing-go/ext"

	"github.com/instana/go-sensor/example/grpc-client-server/pb"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instagrpc"
	"google.golang.org/grpc"
)

const (
	testMessage = "Hi!"
)

var listenAddr string

func main() {
	flag.StringVar(&listenAddr, "l", os.Getenv("LISTEN_ADDR"), "Server listen address")
	flag.Parse()

	if listenAddr == "" {
		flag.Usage()
		os.Exit(2)
	}

	// initialize server sensor to instrument request handlers
	sensor := instana.NewSensor("grpc-client-server")

	// to instrument server calls add instagrpc.UnaryServerInterceptor(sensor) and
	// instagrpc.StreamServerInterceptor(sensor) to the list of server options when
	// initializing the server
	srv := grpc.NewServer(
		grpc.UnaryInterceptor(instagrpc.UnaryServerInterceptor(sensor)),
		grpc.StreamInterceptor(instagrpc.StreamServerInterceptor(sensor)),
	)

	pb.RegisterEchoServiceServer(srv, &Service{})

	go startServer(srv, listenAddr)

	conn, err := grpc.Dial(
		listenAddr,
		grpc.WithInsecure(),
		grpc.WithBlock(),
		// To instrument client calls add instagrpc.UnaryClientInterceptor(sensor) and
		// instagrpc.StringClientInterceptor(sensor) to the DialOption list while dialing
		// the GRPC server.
		grpc.WithUnaryInterceptor(instagrpc.UnaryClientInterceptor(sensor)),
		grpc.WithStreamInterceptor(instagrpc.StreamClientInterceptor(sensor)),
	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	defer conn.Close()

	client := pb.NewEchoServiceClient(conn)

	// The call should always start with an entry span (https://www.instana.com/docs/tracing/custom-best-practices/#start-new-traces-with-entry-spans)
	// Normally this would be your HTTP/GRPC/message queue request span, but here we need to
	// create it explicitly.
	sp := sensor.Tracer().
		StartSpan("client-call").
		SetTag(string(ext.SpanKind), "entry")

	log.Println("Send request")

	// call the service using context with a span
	r, err := client.Echo(
		// Create a context that holds the parent entry span and pass it to the GRPC call
		instana.ContextWithSpan(context.Background(), sp),
		&pb.EchoRequest{Message: testMessage},
	)

	if err != nil {
		log.Fatalf("error while echoing: %v", err)
	}

	log.Printf("Response << %s", r.GetMessage())
}

// Service is used to implement pb.EchoService
type Service struct {
	pb.UnimplementedEchoServiceServer
}

// Echo implements pb.EchoService
func (s *Service) Echo(ctx context.Context, in *pb.EchoRequest) (*pb.EchoReply, error) {
	log.Printf("Request << %s", in.GetMessage())

	// This step is not required for a minimal instrumentation.
	//
	// Extract the parent span and use its tracer to initialize any child spans to trace the calls
	// inside the handler, e.g. database queries, 3rd-party API requests, etc.
	if parent, ok := instana.SpanFromContext(ctx); ok {
		sp := parent.Tracer().StartSpan("echo-call")
		defer sp.Finish()
	}

	return &pb.EchoReply{Message: in.GetMessage()}, nil
}

func startServer(srv *grpc.Server, address string) {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Println("Starting server...")
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
