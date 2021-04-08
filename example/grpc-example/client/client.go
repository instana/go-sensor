// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2016

package client

import (
	"context"
	"google.golang.org/grpc"
	"grpc-example/pb"
	"log"
)

// Call send pb.EchoRequest request with "message" to the "address" using generated grpc pb.EchoServiceClient client
// and returns the message from the response
func Call(ctx context.Context, address, message string) string {
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	defer conn.Close()

	client := pb.NewEchoServiceClient(conn)

	r, err := client.Echo(ctx, &pb.EchoRequest{Message: message})
	if err != nil {
		log.Fatalf("error while echoing: %v", err)
	}

	return r.Message
}
