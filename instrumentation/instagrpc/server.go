package instagrpc

import (
	"context"
	"net"

	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// UnaryServerInterceptor returns a tracing interceptor to be used in grpc.NewServer() call.
// This interceptor is responsible for extracting the trace context from incoming requests
// and ensures trace continuation by injecting a child span into the handler context
func UnaryServerInterceptor(sensor *instana.Sensor) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		sp := startServerSpan(ctx, info.FullMethod, "unary", sensor.Tracer())
		defer sp.Finish()

		// log request in case handler panics
		defer func() {
			if err := recover(); err != nil {
				if e, ok := err.(error); ok {
					sp.SetTag("rpc.error", e.Error())
					sp.LogFields(otlog.Error(e))
				} else {
					sp.SetTag("rpc.error", err)
					sp.LogFields(otlog.Object("error", err))
				}

				// re-throw
				panic(err)
			}
		}()

		m, err := handler(instana.ContextWithSpan(ctx, sp), req)
		if err != nil {
			sp.SetTag("rpc.error", err.Error())
			sp.LogFields(otlog.Error(err))
		}

		return m, err
	}
}

// StreamServerInterceptor returns a tracing interceptor to be used in grpc.NewServer() call.
// This interceptor is responsible for extracting the trace context from incoming streams
// and ensures trace continuation by injecting a child span into the handler stream context
func StreamServerInterceptor(sensor *instana.Sensor) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		sp := startServerSpan(ss.Context(), info.FullMethod, "stream", sensor.Tracer())
		defer sp.Finish()

		// log request in case handler panics
		defer func() {
			if err := recover(); err != nil {
				if e, ok := err.(error); ok {
					sp.SetTag("rpc.error", e)
					sp.LogFields(otlog.Error(e))
				} else {
					sp.SetTag("rpc.error", err)
					sp.LogFields(otlog.Object("error", err))
				}

				// re-throw
				panic(err)
			}
		}()

		if err := handler(srv, &wrappedServerStream{ss, sp}); err != nil {
			sp.SetTag("rpc.error", err.Error())
			sp.LogFields(otlog.Error(err))

			return err
		}

		return nil
	}
}

func startServerSpan(ctx context.Context, method, callType string, tracer ot.Tracer) ot.Span {
	opts := []ot.StartSpanOption{
		ext.SpanKindRPCServer,
		ot.Tags{
			"rpc.flavor":    "grpc",
			"rpc.call":      method,
			"rpc.call_type": callType,
		},
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		// TODO: log this error using the sensor logger
		// log.Println("WARN: failed to extract request metadata")

		return tracer.StartSpan("rpc-server", opts...)
	}

	host, port := extractServerAddr(md)
	if host != "" {
		opts = append(opts, ot.Tags{
			"rpc.host": host,
			"rpc.port": port,
		})
	}

	if spanContext := remoteSpanContext(md, tracer); spanContext != nil {
		opts = append(opts, ext.RPCServerOption(spanContext))
	}

	return tracer.StartSpan("rpc-server", opts...)
}

func extractServerAddr(md metadata.MD) (string, string) {
	authority := md.Get(":authority")
	if len(authority) == 0 {
		return "", ""
	}

	host, port, err := net.SplitHostPort(authority[0])
	if err != nil {
		// TODO: log this using the sensor logger
		// log.Printf("INFO: failed to extract server host and port from request metadata: %s", err)

		// take our best guess and use :authority as a host if the net.SplitHostPort() fails to parse
		return authority[0], ""
	}

	return host, port
}

func remoteSpanContext(md metadata.MD, tracer ot.Tracer) ot.SpanContext {
	switch wireContext, err := tracer.Extract(ot.HTTPHeaders, ot.HTTPHeadersCarrier(md)); err {
	case nil:
		return wireContext
	case ot.ErrSpanContextNotFound:
		// TODO: log this using the sensor logger
		// the remote did not provide any OpenTracing headers, so we just start a new trace
	case ot.ErrUnsupportedFormat:
		// TODO: log this using the sensor logger
		// log.Printf("WARN: unsupported grpc request context format")
	default:
		// TODO: log this using the sensor logger
		// log.Printf("ERROR: failed to extract context")
	}

	return nil
}

type wrappedServerStream struct {
	grpc.ServerStream
	Span ot.Span
}

func (ss wrappedServerStream) Context() context.Context {
	return instana.ContextWithSpan(ss.ServerStream.Context(), ss.Span)
}
