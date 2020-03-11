package instagrpc

import (
	"context"
	"io"
	"net"
	"net/http"

	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// UnaryClientInterceptor returns a tracing interceptor to be used in grpc.NewClient() call.
// It is responsible for injecting the trace context into outgoing requests
func UnaryClientInterceptor(sensor *instana.Sensor) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, callOpts ...grpc.CallOption) error {
		sp := startClientSpan(ctx, cc.Target(), method, "unary", sensor.Tracer())
		defer sp.Finish()

		if err := invoker(outgoingTracingContext(ctx, sp), method, req, reply, cc, callOpts...); err != nil {
			sp.SetTag("rpc.error", err.Error())
			sp.LogFields(otlog.Error(err))

			return err
		}

		return nil
	}
}

// StreamClientInterceptor returns a tracing interceptor to be used in grpc.NewClient() call.
// It is responsible for injecting the trace context into outgoing requests
func StreamClientInterceptor(sensor *instana.Sensor) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {

		sp := startClientSpan(ctx, cc.Target(), method, "stream", sensor.Tracer())
		stream, err := streamer(outgoingTracingContext(ctx, sp), desc, cc, method, opts...)
		if err != nil {
			sp.SetTag("rpc.error", err.Error())
			sp.LogFields(otlog.Error(err))
			sp.Finish()

			return nil, err
		}

		return wrappedClientStream{
			ClientStream:  stream,
			Span:          sp,
			ServerStreams: desc.ServerStreams,
		}, nil
	}
}

func startClientSpan(ctx context.Context, target, method, callType string, tracer ot.Tracer) ot.Span {
	host, port, err := net.SplitHostPort(target)
	if err != nil {
		// TODO: log this error using the sensor logger
		// log.Printf("INFO: failed to extract server host and port from request metadata: %s", err)

		// take our best guess and use :authority as a host if the net.SplitHostPort() fails to parse
		host, port = target, ""
	}

	opts := []ot.StartSpanOption{
		ext.SpanKindRPCClient,
		ot.Tags{
			"rpc.flavor":    "grpc",
			"rpc.call":      method,
			"rpc.call_type": callType,
			"rpc.host":      host,
			"rpc.port":      port,
		},
	}

	if parentSpan, ok := instana.SpanFromContext(ctx); ok {
		tracer = parentSpan.Tracer() // use the same tracer as the parent span does
		opts = append(opts, ot.ChildOf(parentSpan.Context()))
	}

	return tracer.StartSpan("rpc-client", opts...)
}

func outgoingTracingContext(ctx context.Context, span ot.Span) context.Context {
	// gather opentracing headers and inject them into request metadata omitting empty values
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}

	headers := http.Header{}
	span.Tracer().Inject(span.Context(), ot.HTTPHeaders, ot.HTTPHeadersCarrier(headers))

	for k := range headers {
		if v := headers.Get(k); v != "" {
			md.Set(k, v)
		}
	}

	return metadata.NewOutgoingContext(ctx, md)
}

type wrappedClientStream struct {
	grpc.ClientStream
	Span          ot.Span
	ServerStreams bool
}

func (cs wrappedClientStream) RecvMsg(m interface{}) error {
	err := cs.ClientStream.RecvMsg(m)
	if err != nil {
		if err != io.EOF {
			cs.Span.SetTag("rpc.error", err.Error())
			cs.Span.LogFields(otlog.Error(err))
		}
	}

	if err != nil || !cs.ServerStreams {
		cs.Span.Finish()
	}

	return err
}
