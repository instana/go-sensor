A gRPC Server/Client Example
============================

An example of instrumenting a gRPC client and server with Instana using [`github.com/instana/go-sensor/tree/master/instrumentation/instagrpc`](https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instagrpc).

Usage
-----

To start an example locally on `localhost:43210` and send 10 requests, run:

```bash
go run main.go -l localhost:43210
```

In case when the port is already in use, please choose another one.

```bash
  -l string
        Server listen address
```
