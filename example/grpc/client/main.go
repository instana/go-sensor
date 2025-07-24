// (c) Copyright IBM Corp. 2024

//go:build go1.23
// +build go1.23

package main

import (
	"context"
	"io"
	"log"
	"time"

	instana "github.com/instana/go-sensor"
	pb "github.com/instana/go-sensor/example/grpc/hellopb"
	"github.com/instana/go-sensor/instrumentation/instagrpc"
	"github.com/opentracing/opentracing-go/ext"
	"google.golang.org/grpc"
)

func main() {

	collector := instana.InitCollector(&instana.Options{
		Service: "grpc-client",
		Tracer:  instana.DefaultTracerOptions(),
	})

	// Connect to the server.
	conn, err := grpc.Dial("localhost:50051",
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(instagrpc.UnaryClientInterceptor(collector)),
		grpc.WithStreamInterceptor(instagrpc.StreamClientInterceptor(collector)))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewGreeterClient(conn)

	// Unary call
	doUnaryCall(collector, client)

	// Server-side streaming call
	doStreamingCall(collector, client)

	// Make a call to an unknown service/method.
	doUnknownServiceCall(collector, conn)

	time.Sleep(10 * time.Minute)
}

func doUnknownServiceCall(collector instana.TracerLogger, client *grpc.ClientConn) {

	sp := collector.Tracer().
		StartSpan("grpc-unknown-service-call").
		SetTag(string(ext.SpanKind), "entry")

	sp.Finish()

	ctx := instana.ContextWithSpan(context.Background(), sp)

	// Invoke a non-existent method (this will trigger the UnknownServiceHandler).
	err := client.Invoke(ctx, "/UnknownService/UnknownMethod", nil, nil)
	if err != nil {
		log.Printf("Error from server: %v", err)
	} else {
		log.Println("Call succeeded (unexpected).")
	}
}

func doUnaryCall(collector instana.TracerLogger, client pb.GreeterClient) {

	sp := collector.Tracer().
		StartSpan("grpc-unary-client-call").
		SetTag(string(ext.SpanKind), "entry")

	sp.Finish()

	ctx := instana.ContextWithSpan(context.Background(), sp)

	log.Println("Starting Unary Call...")
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	resp, err := client.SayHello(ctx, &pb.HelloRequest{Name: "World"})
	if err != nil {
		log.Fatalf("Unary call failed: %v", err)
	}
	log.Printf("Unary Response: %s", resp.GetMessage())
}

func doStreamingCall(collector instana.TracerLogger, client pb.GreeterClient) {

	sp := collector.Tracer().
		StartSpan("grpc-stream-client-call").
		SetTag(string(ext.SpanKind), "entry")

	sp.Finish()

	ctx := instana.ContextWithSpan(context.Background(), sp)

	log.Println("Starting Streaming Call...")
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	stream, err := client.SayHelloStream(ctx, &pb.HelloRequest{Name: "World"})
	if err != nil {
		log.Fatalf("Streaming call failed: %v", err)
	}

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			log.Println("Streaming completed.")
			break
		}
		if err != nil {
			log.Fatalf("Error receiving stream: %v", err)
		}
		log.Printf("Streaming Response: %s", resp.GetMessage())
	}
}
