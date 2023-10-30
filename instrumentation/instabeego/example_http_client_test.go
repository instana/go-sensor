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

func Example_http_client_instrument() {
	sensor := instana.NewSensor("my-http-client")

	sp := sensor.Tracer().StartSpan("client-call")
	sp.SetTag(string(ext.SpanKind), "entry")

	defer sp.Finish()

	builder := &instabeego.FilterChainBuilder{
		Sensor: sensor,
	}

	ctx := instana.ContextWithSpan(context.Background(), sp)

	req := httplib.NewBeegoRequestWithCtx(ctx, "https://www.instana.com", http.MethodGet)
	req.AddFilters(builder.FilterChain)

	_, err := req.Response()
	if err != nil {
		log.Fatalf("failed to GET https://www.instana.com: %s", err)
	}
}
