// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2016

// +build go1.10

package main

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/instana/go-sensor/example/grpc-example/pb"
	"github.com/instana/go-sensor/example/grpc-example/server"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instagrpc"
	"google.golang.org/grpc"
)

const (
	port        = ":43210"
	address     = "localhost"
	testMessage = "Hi!"
)

func main() {

	// Initialize server sensor to instrument request handlers
	sensor := instana.NewSensor("grpc-server")

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// To instrument server calls add instagrpc.UnaryServerInterceptor(sensor) and
	// instagrpc.StreamServerInterceptor(sensor) to the list of server options when
	// initializing the server
	srv := grpc.NewServer(
		grpc.UnaryInterceptor(instagrpc.UnaryServerInterceptor(sensor)),
		grpc.StreamInterceptor(instagrpc.StreamServerInterceptor(sensor)),
	)

	pb.RegisterEchoServiceServer(srv, &server.Service{})

	go func() {
		log.Println("Starting server...")
		if err := srv.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		c := NewClient("grpc-client")
		for {
			select {
			case <-ticker.C:
				log.Println("Call server...")
				response := c.Call(context.Background(), address+port, testMessage)
				log.Printf("Response << %s", response)
			}
		}
	}()

	select {}
}
