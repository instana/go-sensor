Instana instrumentation for pgx/v5
=============================================

This package provides Instana instrumentation for the [`pgx/v5`](https://github.com/jackc/pgx) package.

[![PkgGoDev](https://pkg.go.dev/badge/github.com/instana/go-sensor/instrumentation/instapgx/v2)](https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instapgx/v2)

Installation
---

To add the module to your `go.mod` file, run the following command in your project directory:

```bash
$ go get github.com/instana/go-sensor/instrumentation/instapgx/v2
```

Usage
---
```go
// Create an Instana collector
c := instana.InitCollector(&instana.Options{
    Service: "pgx-v5-service",
    Tracer:  instana.DefaultTracerOptions(),
})

// Parse Config
cfg, err := pgx.ParseConfig("postgres://username:password@localhost/database")

// Assign the tracer interface with Instana tracer
cfg.Tracer = instapgx.InstanaTracer(cfg, c)

// Create the connection using the cfg with Instana tracer
conn, err := pgx.ConnectConfig(ctx, cfg)
defer conn.Close(ctx)

```
