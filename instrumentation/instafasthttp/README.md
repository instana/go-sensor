instafasthttp - Instana instrumentation for fasthttp
=====================================

This package provides Instana instrumentation for the [`fasthttp`](https://pkg.go.dev/github.com/valyala/fasthttp) package.

[![PkgGoDev](https://pkg.go.dev/badge/github.com/instana/go-sensor/instrumentation/instafasthttp)](https://pkg.go.dev/badge/github.com/instana/go-sensor/instrumentation/instafasthttp)

Installation
------------

To add the module to your `go.mod` file run the following command in your project directory:

```bash
$ go get github.com/instana/go-sensor/instrumentation/instafasthttp
```

Usage
-----

### Server
The `instafasthttp.TraceHandler` returns an instrumented `fasthttp.RequestHandler`, which can be used to add instrumentation to calls in a fasthttp server. Please refer to the details below for more information.

```go
// Create a collector  for instana instrumentation
c = instana.InitCollector(&instana.Options{
	Service:  "fasthttp-example",
	LogLevel: instana.Debug,
	Tracer:  instana.DefaultTracerOptions(),
})

// request handler
func fastHTTPHandler(ctx *fasthttp.RequestCtx) {
	fmt.Fprintf(ctx, "Hi there! RequestURI is %q\n", ctx.RequestURI())
	switch string(ctx.Path()) {
	case "/greet":
        // Use the instafasthttp.TraceHandler for instrumenting the greet handler
		instafasthttp.TraceHandler(c, "greet", "/greet", func(ctx *fasthttp.RequestCtx) {
			ctx.SetStatusCode(fasthttp.StatusOK)
			fmt.Fprintf(ctx, "Hello brother!\n")

		})(ctx)
	default:
		ctx.Error("Unsupported path", fasthttp.StatusNotFound)
	}
}

log.Fatal(fasthttp.ListenAndServe(":7070", fastHTTPHandler))

```

#### Trace propagation

Trace propagation is achieved by correctly using the context. In an instrumented handler, if you need to perform additional operations such as a database call and want the trace propagation to ensure that spans fall under the HTTP span, you must use the `instafasthttp.UserContext` function. This function provides the appropriate context containing the parent span information, which should then be passed to the subsequent database operation to get the parent span. Refer to the example code below for further clarity.

```go
func greetEndpointHandler(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusOK)
	fmt.Fprintf(ctx, "This is the first part of body!\n")

	var stud student

	// This context is required for span propagation.
	// It will be set by instafasthttp, ensuring it contains the parent span info.
	uCtx := instafasthttp.UserContext(ctx)
	db.WithContext(uCtx).First(&stud)

	fmt.Fprintf(ctx, "Hello "+stud.StudentName+"!\n")
}
```

### HostClient

The `instafasthttp.RoundTripper` provides an implementation of the `fasthttp.RoundTripper` interface. It can be used to instrument client calls with the help of `instafasthttp.HostClient`. Refer to the details below for more information.

```go
// Create a collector for instana instrumentation
c = instana.InitCollector(&instana.Options{
	Service:  "fasthttp-example",
	LogLevel: instana.Debug,
	Tracer:  instana.DefaultTracerOptions(),
})

// request handler
func fastHTTPHandler(ctx *fasthttp.RequestCtx) {
	fmt.Fprintf(ctx, "Hi there! RequestURI is %q\n", ctx.RequestURI())
	switch string(ctx.Path()) {
	case "/round-trip":
		instafasthttp.TraceHandler(c, "round-trip", "/round-trip", func(ctx *fasthttp.RequestCtx) {
			// user context
			uCtx := instafasthttp.UserContext(ctx)

			hc := &fasthttp.HostClient{
				Transport: instafasthttp.RoundTripper(uCtx, c, nil),
				Addr:      "localhost:7070",
			}

			url := fasthttp.AcquireURI()
			url.Parse(nil, []byte("http://localhost:7070/greet"))
			req := fasthttp.AcquireRequest()
			defer fasthttp.ReleaseRequest(req)
			req.SetURI(url)
			fasthttp.ReleaseURI(url) // now you may release the URI
			req.Header.SetMethod(fasthttp.MethodGet)

			resp := fasthttp.AcquireResponse()
			defer fasthttp.ReleaseResponse(resp)

			// Make the request
			err := hc.Do(req, resp)
			if err != nil {
				log.Fatalf("failed to GET http://localhost:7070/greet: %s", err)
			}

			// getting response body
			bs := string(resp.Body())
			fmt.Println(bs)

			// Respond with a 200 status code and include the body as well
			ctx.SetStatusCode(fasthttp.StatusOK)
			fmt.Fprintf(ctx, bs)
		})(ctx)
	default:
		ctx.Error("Unsupported path", fasthttp.StatusNotFound)
	}
}

log.Fatal(fasthttp.ListenAndServe(":7070", fastHTTPHandler))

```
### Client 

The `client.Do` and related methods can be traced using Instana. However, the usage differs slightly from that of the standard HostClient. Below are the steps to use an Instana instrumented client.

- To enable tracing, you must create an instrumented client using the `instafasthttp.GetInstrumentedClient` method as shown below:

```go
	// fasthttp client
	client := &fasthttp.Client{
		ReadTimeout:                   readTimeout,
		WriteTimeout:                  writeTimeout,
		MaxIdleConnDuration:           maxIdleConnDuration,
		NoDefaultUserAgentHeader:      true, // Don't send: User-Agent: fasthttp
		DisableHeaderNamesNormalizing: true, // If you set the case on your headers correctly you can enable this
		DisablePathNormalizing:        true,
		// increase DNS cache time to an hour instead of default minute
		Dial: (&fasthttp.TCPDialer{
			Concurrency:      4096,
			DNSCacheDuration: time.Hour,
		}).Dial,
	}

	// create instana instrumented client
	ic := instafasthttp.GetInstrumentedClient(collector, client)
```
- Use the instrumented client(ic) for all requests instead of the original client.
- Tracing is supported for the following methods, where an additional `context.Context` parameter is required as the first argument. Ensure the context is set properly for span correlation:
1. Do
2. DoTimeout
3. DoDeadline
4. DoRedirects

```go
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	req.SetURI(url)
	fasthttp.ReleaseURI(url) // now you may release the URI
	req.Header.SetMethod(fasthttp.MethodGet)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// Make the request
	err := ic.Do(uCtx, req, resp)
	if err != nil {
		log.Fatalf("failed to GET http://localhost:7070/greet: %s", err)
	}
```

- For methods other than the four mentioned above, use the standard method signatures without passing a context. These methods do not support tracing. For example, `client.Get` and `client.Post` do not currently support Instana tracing. If you wish to trace the `GET` and `POST` requests, please use `client.Do` method instead.
- Use the `Unwrap()` method if you require the original fasthttp.Client instance. However, avoid using the unwrapped instance directly for the above four methods, as Instana tracing will not be applied in such cases.
