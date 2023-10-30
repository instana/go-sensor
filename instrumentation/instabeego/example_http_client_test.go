// (c) Copyright IBM Corp. 2023

//go:build go1.18
// +build go1.18

package instabeego_test

import (
	"context"
	"log"
	"net/http"

	"github.com/beego/beego/v2/client/httplib"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instabeego"
	"github.com/opentracing/opentracing-go/ext"
)

// This example shows how to instrument beego httplib module (HTTP client) with Instana tracing
func Example_http_client_instrument() {
	sensor := instana.NewSensor("my-http-client")

	// Every call should start with an entry span (https://docs.instana.io/quick_start/custom_tracing/#always-start-new-traces-with-entry-spans)
	// Normally this would be your HTTP/GRPC/message queue request span, but here we need to create it explicitly, since an HTTP client call is
	// an exit span. And all exit spans must have a parent entry span.
	sp := sensor.Tracer().StartSpan("client-call")
	sp.SetTag(string(ext.SpanKind), "entry")

	defer sp.Finish()

	// As the filter function need sensor data, it should be injected to FilterChainBuilder
	builder := &instabeego.FilterChainBuilder{
		Sensor: sensor,
	}

	// Inject the parent span into request context
	ctx := instana.ContextWithSpan(context.Background(), sp)

	req := httplib.NewBeegoRequestWithCtx(ctx, "https://www.instana.com", http.MethodGet)
	// Adding filter function to wrap the beego BeegoHTTPRequest transport with instana.RoundTripper().
	req.AddFilters(builder.FilterChain)

	_, err := req.Response()
	if err != nil {
		log.Fatalf("failed to GET https://www.instana.com: %s", err)
	}
}
