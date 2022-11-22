// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instagrpc

import (
	"context"
	"net"

	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// UnaryServerInterceptor returns a tracing interceptor to be used in grpc.NewServer() calls.
// This interceptor is responsible for extracting the Instana OpenTracing headers from incoming requests
// and staring a new span that can later be accessed inside the handler:
//
//	if parent, ok := instana.SpanFromContext(ctx); ok {
//		sp := parent.Tracer().StartSpan("child-span")
//		defer sp.Finish()
//	}
//
// If the handler returns an error or panics, the error message is then attached to the span logs.
func UnaryServerInterceptor(sensor *instana.Sensor) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		sp := startServerSpan(ctx, info.FullMethod, "unary", sensor)
		defer sp.Finish()

		// log request in case handler panics
		defer func() {
			if err := recover(); err != nil {
				addRPCError(sp, err)
				// re-throw
				panic(err)
			}
		}()

		m, err := handler(instana.ContextWithSpan(ctx, sp), req)
		if err != nil {
			addRPCError(sp, err)
		}

		return m, err
	}
}

// StreamServerInterceptor returns a tracing interceptor to be used in grpc.NewServer() calls.
// This interceptor is responsible for extracting the Instana OpenTracing headers from incoming streaming
// requests and starting a new span that can later be accessed inside the handler:
//
//	if parent, ok := instana.SpanFromContext(srv.Context()); ok {
//		sp := parent.Tracer().StartSpan("child-span")
//		defer sp.Finish()
//	}
//
// If the handler returns an error or panics, the error message is then attached to the span logs.
func StreamServerInterceptor(sensor *instana.Sensor) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		sp := startServerSpan(ss.Context(), info.FullMethod, "stream", sensor)
		defer sp.Finish()

		// log request in case handler panics
		defer func() {
			if err := recover(); err != nil {
				addRPCError(sp, err)
				// re-throw
				panic(err)
			}
		}()

		if err := handler(srv, &wrappedServerStream{ss, sp}); err != nil {
			addRPCError(sp, err)

			return err
		}

		return nil
	}
}

func startServerSpan(ctx context.Context, method, callType string, sensor *instana.Sensor) ot.Span {
	tracer := sensor.Tracer()
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
		return tracer.StartSpan("rpc-server", opts...)
	}

	host, port := extractServerAddr(md)
	if host != "" {
		opts = append(opts, ot.Tags{
			"rpc.host": host,
			"rpc.port": port,
		})
	}

	switch wireContext, err := tracer.Extract(ot.HTTPHeaders, ot.HTTPHeadersCarrier(md)); err {
	case nil:
		opts = append(opts, ext.RPCServerOption(wireContext))
	case ot.ErrSpanContextNotFound:
		sensor.Logger().Debug("no tracing context found in request to ", method, ", starting a new trace")
	case ot.ErrUnsupportedFormat:
		sensor.Logger().Warn("unsupported grpc request context format for ", method)
	default:
		sensor.Logger().Error("failed to extract request context for ", method, ": ", err)
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
		// take our best guess and use :authority as a host if the net.SplitHostPort() fails to parse
		return authority[0], ""
	}

	return host, port
}

type wrappedServerStream struct {
	grpc.ServerStream
	Span ot.Span
}

func (ss wrappedServerStream) Context() context.Context {
	return instana.ContextWithSpan(ss.ServerStream.Context(), ss.Span)
}
