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
-n int
    how many requests to send (default 10)
```

# How to modify a proto definitions for an example
Proto definitions are in `pb` folder. To generate corresponding go files, `protoc` is required. [Here](https://grpc.io/docs/protoc-installation) is how to install it. Then, in the `pb` folder run the following command:
```bash
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative serviceexample.proto
```