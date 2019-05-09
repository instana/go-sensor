package main

import (
	"log"
	"net/http"
	"time"

	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"golang.org/x/net/context"
)

const (
	Service = "go-microservice-14c"
	Entry   = "http://localhost:9060/golang/entry"
	Exit    = "http://localhost:9060/golang/exit"
)

func request(ctx context.Context, url string, op string) (*http.Client, *http.Request) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "text/plain")
	client := &http.Client{Timeout: 5 * time.Second}

	return client, req
}

func requestEntry(ctx context.Context) {
	client, req := request(ctx, Entry, "entry")
	client.Do(req)
}

//TODO: handle erroneous requests
func requestExit(span ot.Span) {
	client, req := request(context.Background(), Exit, "exit")
	ot.GlobalTracer().Inject(span.Context(), ot.HTTPHeaders, ot.HTTPHeadersCarrier(req.Header))
	resp, _ := client.Do(req)
	span.SetTag(string(ext.SpanKind), string(ext.SpanKindRPCClientEnum))
	span.SetTag(string(ext.PeerHostname), req.Host)
	span.SetTag(string(ext.HTTPUrl), Exit)
	span.SetTag(string(ext.HTTPMethod), req.Method)
	span.SetTag(string(ext.HTTPStatusCode), resp.StatusCode)
}

func server() {
	http.HandleFunc("/golang/entry", func(w http.ResponseWriter, req *http.Request) {
		wireContext, _ := ot.GlobalTracer().Extract(ot.HTTPHeaders, ot.HTTPHeadersCarrier(req.Header))
		parentSpan := ot.GlobalTracer().StartSpan("server", ext.RPCServerOption(wireContext))
		parentSpan.SetTag(string(ext.SpanKind), string(ext.SpanKindRPCServerEnum))
		parentSpan.SetTag(string(ext.PeerHostname), req.Host)
		parentSpan.SetTag(string(ext.HTTPUrl), req.URL.Path)
		parentSpan.SetTag(string(ext.HTTPMethod), req.Method)
		parentSpan.SetTag(string(ext.HTTPStatusCode), 200)

		childSpan := ot.StartSpan("client", ot.ChildOf(parentSpan.Context()))

		requestExit(childSpan)

		time.Sleep(450 * time.Millisecond)

		childSpan.Finish()

		time.Sleep(550 * time.Millisecond)

		ot.GlobalTracer().Inject(parentSpan.Context(), ot.HTTPHeaders, ot.HTTPHeadersCarrier(w.Header()))

		parentSpan.Finish()
	})

	http.HandleFunc("/golang/exit", func(w http.ResponseWriter, req *http.Request) {
		time.Sleep(450 * time.Millisecond)
	})

	log.Fatal(http.ListenAndServe(":9060", nil))
}

func main() {
	ot.InitGlobalTracer(instana.NewTracerWithOptions(&instana.Options{
		Service:  Service,
		LogLevel: instana.Info}))

	go server()

	go forever()
	select {}
}

func forever() {
	for {
		requestEntry(context.Background())
	}
}
