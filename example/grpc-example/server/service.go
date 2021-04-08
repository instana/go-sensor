package server

import (
	"context"
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

	return &pb.EchoReply{Message: in.GetMessage()}, nil
}
