Instana instrumentation for Beego framework
=============================================

This module contains middleware to instrument HTTP services written with [`https://github.com/beego/beego`](https://github.com/beego/beego).

[![PkgGoDev](https://pkg.go.dev/badge/github.com/instana/go-sensor/instrumentation/instabeego)][godoc]

Installation
------------

To add the module to your `go.mod` file run the following command in your project directory:

```bash
$ go get github.com/instana/go-sensor/instrumentation/instabeego
```

Usage
-----

## Web server instrumentation

```go
// create an Instana collector
c := instana.InitCollector(&instana.Options{
    Service:           "beego-server",
    EnableAutoProfile: true,
    Tracer:  instana.DefaultTracerOptions(),
})

// instrument the web server
instabeego.InstrumentWebServer(c)

// define API 
beego.Get("/foo", func(ctx *beecontext.Context) {/* ... */})

// Run beego application
beego.Run() 
// ...
```
[Full example][serverInstrumentationExample]

## HTTP client instrumentation

```go
// create an Instana collector 
c := instana.InitCollector(&instana.Options{
    Service:           "my-http-client",
    EnableAutoProfile: true,
    Tracer:  instana.DefaultTracerOptions(),
})

// get the parent span and inject into the request context
ctx := instana.ContextWithSpan(context.Background(), /* parent span */)

// create a new http request using beego
req := httplib.NewBeegoRequestWithCtx(ctx, "https://www.instana.com", http.MethodGet)

// instrument the client request
instabeego.InstrumentRequest(c, req)

// execute the client request and get the response
_, err := req.Response()
// ...
```

[Full example][clientInstrumentationExample]




[godoc]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instabeego
[serverInstrumentationExample]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instabeego#example-package-ServerInstrumentation
[clientInstrumentationExample]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instabeego#example-package-HttpClientInstrumentation

