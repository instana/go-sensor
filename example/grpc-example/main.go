// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2016

// +build go1.10

package main

import (
	"context"
	"flag"
	"log"
	"net"

	"github.com/opentracing/opentracing-go/ext"

	"github.com/instana/go-sensor/example/grpc-example/pb"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instagrpc"
	"google.golang.org/grpc"
)

const (
	defaultPort    = ":43210"
	defaultAddress = "localhost"
	testMessage    = "Hi!"
)

func main() {
	// `port` will be used by an example
	port := flag.String("defaultPort", defaultPort, "defaultPort to use by an example")
	// `address` will be used by an example
	address := flag.String("address", defaultAddress, "address to use by an example")
	// how many requests to send
	nRequestsToSend := flag.Int("n", 10, "how many requests to send")

	flag.Parse()

	// initialize server sensor to instrument request handlers
	sensor := instana.NewSensor("grpc-server")

	// to instrument server calls add instagrpc.UnaryServerInterceptor(sensor) and
	// instagrpc.StreamServerInterceptor(sensor) to the list of server options when
	// initializing the server
	srv := grpc.NewServer(
		grpc.UnaryInterceptor(instagrpc.UnaryServerInterceptor(sensor)),
		grpc.StreamInterceptor(instagrpc.StreamServerInterceptor(sensor)),
	)

	// register an implementation of the service
	pb.RegisterEchoServiceServer(srv, &Service{})

	// start the server and listen to the incoming requests
	go startServer(srv, *address+*port)

	// create a channel to signal when all requests are send
	done := make(chan struct{})

	// start goroutine
	go func() {
		// create a new sensor for a client
		sensor := instana.NewSensor("grpc-client")

		for i := 1; i <= *nRequestsToSend; i++ {
			log.Println("\n", "Send request #", i)
			// init the client with Instana gRPC instrumentation and call the service
			response := Call(context.Background(), sensor, *address+*port, testMessage)
			log.Printf("Response << %s", response)
		}

		done <- struct{}{}
	}()

	<-done
}

// Call send pb.EchoRequest request with "message" to the "address" using generated grpc pb.EchoServiceClient client
// and returns the message from the response
func Call(ctx context.Context, sensor *instana.Sensor, address, message string) string {
	conn, err := grpc.Dial(
		address,
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

	// create a new service client
	client := pb.NewEchoServiceClient(conn)

	// The call should always start with an entry span (https://www.instana.com/docs/tracing/custom-best-practices/#start-new-traces-with-entry-spans)
	// Normally this would be your HTTP/GRPC/message queue request span, but here we need to
	// create it explicitly.
	sp := sensor.Tracer().
		StartSpan("client-call").
		SetTag(string(ext.SpanKind), "entry")

	// call the service
	r, err := client.Echo(
		// Create a context that holds the parent entry span and pass it to the GRPC call
		instana.ContextWithSpan(ctx, sp),
		&pb.EchoRequest{Message: message},
	)

	if err != nil {
		log.Fatalf("error while echoing: %v", err)
	}

	return r.Message
}

// Service is used to implement pb.EchoService
type Service struct {
	pb.UnimplementedEchoServiceServer
}

// Echo implements pb.EchoService
func (s *Service) Echo(ctx context.Context, in *pb.EchoRequest) (*pb.EchoReply, error) {
	log.Printf("Request << %s", in.GetMessage())

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
