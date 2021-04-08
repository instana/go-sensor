// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2016

package main

import (
	"context"
	"google.golang.org/grpc"
	"grpc-example/client"
	"grpc-example/pb"
	"grpc-example/server"
	"log"
	"net"
	"time"
)

const (
	port        = ":43210"
	address     = "localhost"
	testMessage = "Hi!"
)

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	srv := grpc.NewServer()
	pb.RegisterEchoServiceServer(srv, &server.Service{})

	go func() {
		log.Println("Starting server...")
		if err := srv.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	go func(tck *time.Ticker) {
		for {
			select {
			case <-tck.C:
				log.Println("Call server...")
				response := client.Call(context.Background(), address+port, testMessage)
				log.Printf("Response << %s", response)
			}
		}
	}(ticker)

	select {}
}
