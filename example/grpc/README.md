# An Example For gRPC Instrumentation
============================

An example of instrumenting a gRPC client and server with Instana using [`github.com/instana/go-sensor/tree/main/instrumentation/instagrpc`](https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instagrpc).


## Usage

Install the packages 

```bash
go mod tidy
```

Start the gRPC Server:

```bash
go run server/main.go
```

Run the gRPC client

```bash
go run client/main.go
```

## Output

The client makes a unary, stream and an unknown call to the gRPC server. You will be able to see those 3 call traces in the Instana dashboard.