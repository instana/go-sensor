package main

import (
	"fmt"
	"net/http"
	"time"

	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
)

func main() {
	opts := instana.Options{LogLevel: instana.Debug}
	recorder := instana.NewRecorder()
	tracer := instana.NewTracerWithEverything(&opts, recorder)

	fmt.Println("Hello.  Sleeping for 10 to allow announce.")
	time.Sleep(10 * time.Second)

	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://gameface.in/", nil)

	if err != nil {
		fmt.Println(err)
	}

	for i := 1; i <= 100; i++ {
		sp := tracer.StartSpan("multi_request")
		sp.SetBaggageItem("foo", "bar")
		for i := 1; i <= 2; i++ {

			headersCarrier := ot.HTTPHeadersCarrier(req.Header)
			if err = tracer.Inject(sp.Context(), ot.HTTPHeaders, headersCarrier); err != nil {
				fmt.Println(err)
			}

			httpSpan := tracer.StartSpan("net-http", ot.ChildOf(sp.Context()))
			fmt.Println("Making request to Gameface...")
			resp, err := client.Do(req)

			if err != nil {
				fmt.Println(err)
			} else {
				httpSpan.SetTag("http.status_code", resp.StatusCode)
			}

			fmt.Println("Done.  Code & sleeping for 5.  StatusCode: ", resp.StatusCode)
			time.Sleep(5 * time.Second)

			httpSpan.Finish()
		}
		sp.Finish()
	}
}
