package server

import (
	"context"
	instana "github.com/instana/go-sensor"
	"grpc-example/pb"
	"log"
)

// Service is used to implement pb.EchoService
type Service struct {
	pb.UnimplementedEchoServiceServer
}

// Service implements pb.EchoService
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
