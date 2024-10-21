// (c) Copyright IBM Corp. 2023

package main

import (
	"fmt"
	"log"

	"github.com/valyala/fasthttp"
)

func sampleEndpointHandler(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusOK)
	fmt.Fprintf(ctx, "This is the first part of body!\n")
}

// request handler in fasthttp style, i.e. just plain function.
func fastHTTPHandler(ctx *fasthttp.RequestCtx) {
	fmt.Fprintf(ctx, "Hi there! RequestURI is %q\n", ctx.RequestURI())
	switch string(ctx.Path()) {
	case "/endpoint":
		sampleEndpointHandler(ctx)
	default:
		ctx.Error("Unsupported path", fasthttp.StatusNotFound)
	}
}

func main() {
	// col := instana.InitCollector(&instana.Options{
	// 	Service:           "Nithin Basic Usage",
	// 	EnableAutoProfile: true,
	// })

	// http.HandleFunc("/endpoint", instana.TracingHandlerFunc(col, "/endpoint", func(w http.ResponseWriter, r *http.Request) {
	// 	w.WriteHeader(http.StatusOK)
	// }))

	log.Fatal(fasthttp.ListenAndServe(":7070", fastHTTPHandler))

	// log.Fatal(http.ListenAndServe(":8080", nil))
}
