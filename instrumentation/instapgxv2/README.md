Instana instrumentation for pgx/v5
=============================================

This package provides Instana instrumentation for the [`pgx/v5`](https://github.com/jackc/pgx/v5) package.

[![PkgGoDev](https://pkg.go.dev/badge/github.com/instana/go-sensor/instrumentation/instapgxv2)](https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instapgxv2)

Installation
---

To add the module to your `go.mod` file, run the following command in your project directory:

```bash
$ go get github.com/instana/go-sensor/instrumentation/instapgxv2
```

Usage
---
```go
// Create a sensor
sensor := instana.NewSensor("pgx-v5-service")

// Parse Config
cfg, err := pgx.ParseConfig("postgres://username:password@localhost/database")

// Assign the tracer interface with Instana tracer
cfg.Tracer = instapgxv2.InstanaTracer(cfg, sensor)

// Create the connection using the cfg with Instana tracer
conn, err := pgx.ConnectConfig(ctx, cfg)
defer conn.Close(ctx)

```
