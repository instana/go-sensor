// (c) Copyright IBM Corp. 2024

//go:build go1.23
// +build go1.23

package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	instana "github.com/instana/go-sensor"
	pb "github.com/instana/go-sensor/example/grpc/hellopb"
	"github.com/instana/go-sensor/instrumentation/instagrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server is used to implement the Greeter service.
type Server struct {
	pb.UnimplementedGreeterServer
}

// SayHello implements the unary RPC.
func (s *Server) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Unary call received: %v", req.GetName())
	return &pb.HelloReply{Message: "Hello " + req.GetName()}, nil
}

// SayHelloStream implements the server-side streaming RPC.
func (s *Server) SayHelloStream(req *pb.HelloRequest, stream pb.Greeter_SayHelloStreamServer) error {
	log.Printf("Streaming call received: %v", req.GetName())

	for i := 1; i <= 5; i++ {
		msg := &pb.HelloReply{Message: req.GetName() + ", this is message #" + fmt.Sprintf("%d", i)}
		if err := stream.Send(msg); err != nil {
			return err
		}
		time.Sleep(1 * time.Second)
	}
	return nil
}

func main() {

	collector := instana.InitCollector(&instana.Options{
		Service: "grpc-server",
		Tracer:  instana.DefaultTracerOptions(),
	})

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Define the UnknownServiceHandler.
	unknownHandler := func(srv interface{}, stream grpc.ServerStream) error {
		methodName, _ := grpc.MethodFromServerStream(stream)
		log.Printf("Received call to unknown service/method: %s", methodName)

		return status.Errorf(codes.Unimplemented, "Service/method %s is not implemented on this server", methodName)
	}

	grpcServer := grpc.NewServer(
		grpc.UnknownServiceHandler(unknownHandler),
		grpc.UnaryInterceptor(instagrpc.UnaryServerInterceptor(collector)),
		grpc.StreamInterceptor(instagrpc.StreamServerInterceptor(collector)),
	)

	pb.RegisterGreeterServer(grpcServer, &Server{})

	log.Println("Server is listening on port 50051...")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
