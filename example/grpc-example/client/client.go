// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2016

package client

import (
	"context"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instagrpc"
	"github.com/opentracing/opentracing-go/ext"
	"google.golang.org/grpc"
	"grpc-example/pb"
	"log"
)

type Client struct {
	sensor *instana.Sensor
}

func NewClient(sensorName string) *Client {
	return &Client{
		// Initialize client tracer
		sensor: instana.NewSensor(sensorName),
	}
}

// Call send pb.EchoRequest request with "message" to the "address" using generated grpc pb.EchoServiceClient client
// and returns the message from the response
func (c *Client) Call(ctx context.Context, address, message string) string {
	conn, err := grpc.Dial(
		address,
		grpc.WithInsecure(),
		grpc.WithBlock(),
		// To instrument client calls add instagrpc.UnaryClientInterceptor(sensor) and
		// instagrpc.StringClientInterceptor(sensor) to the DialOption list while dialing
		// the GRPC server.
		grpc.WithUnaryInterceptor(instagrpc.UnaryClientInterceptor(c.sensor)),
		grpc.WithStreamInterceptor(instagrpc.StreamClientInterceptor(c.sensor)),
	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	defer conn.Close()

	client := pb.NewEchoServiceClient(conn)

	// The call should always start with an entry span (https://www.instana.com/docs/tracing/custom-best-practices/#start-new-traces-with-entry-spans)
	// Normally this would be your HTTP/GRPC/message queue request span, but here we need to
	// create it explicitly.
	sp := c.sensor.Tracer().
		StartSpan("client-call").
		SetTag(string(ext.SpanKind), "entry")

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
