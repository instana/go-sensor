A gRPC Server/Client Example 
==========================

An example of usage of Instana tracer in the app instrumented with gRPC. It demonstrates how to use `https://github.com/instana/go-sensor/tree/master/instrumentation/instagrpc`.

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