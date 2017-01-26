package main

import (
	"log"
	"net/http"
	"time"

	"github.com/instana/golang-sensor"
	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	golog "github.com/opentracing/opentracing-go/log"
	"golang.org/x/net/context"
)

const (
	Service = "golang-http"
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
	span.LogFields(
		golog.String("type", instana.HTTPClient),
		golog.Object("data", &instana.Data{
			HTTP: &instana.HTTPData{
				Host:   req.Host,
				URL:    Exit,
				Status: resp.StatusCode,
				Method: req.Method}}))
}

func server() {
	http.HandleFunc("/golang/entry", func(w http.ResponseWriter, req *http.Request) {
		wireContext, _ := ot.GlobalTracer().Extract(ot.HTTPHeaders, ot.HTTPHeadersCarrier(req.Header))
		parentSpan := ot.GlobalTracer().StartSpan("server", ext.RPCServerOption(wireContext))
		parentSpan.LogFields(
			golog.String("type", instana.HTTPServer),
			golog.Object("data", &instana.Data{
				HTTP: &instana.HTTPData{
					Host:   req.Host,
					URL:    req.URL.Path,
					Status: 200,
					Method: req.Method}}))

		childSpan := ot.StartSpan("client", ot.ChildOf(parentSpan.Context()))

		requestExit(childSpan)

		time.Sleep(450 * time.Millisecond)

		childSpan.Finish()

		time.Sleep(550 * time.Millisecond)

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
		LogLevel: instana.Debug}))

	go server()

	go forever()
	select {}
}

func forever() {
	for {
		requestEntry(context.Background())
	}
}
