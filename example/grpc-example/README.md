A gRPC Server/Client Example 
==========================

An example of usage of Instana tracer in the app instrumented with gRPC. It demonstrates how to use `https://github.com/instana/go-sensor/tree/master/instrumentation/instagrpc`.

Usage
-----

To start an example locally on `localhost:43210` and send 10 requests, run:

```bash
go run main.go
```

In case when the port is already in use, please choose another one. Also, some other options can be specified.

```bash
-address string
    address to use by an example (default "localhost")
-defaultPort string
    defaultPort to use by an example (default ":43210")
```