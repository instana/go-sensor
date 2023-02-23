Instana instrumentation for pgx
=============================================

This module contains the middleware to instrument services written with [`github.com/jackc/pgx/v4`](https://github.com/jackc/pgx/v4).

[![PkgGoDev](https://pkg.go.dev/badge/github.com/instana/go-sensor/instrumentation/instapgx)](https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instapgx)

The API of this instrumentation might change in the future.

The current version contains the instrumentation for most of the methods defined for `pgx.Tx` and `*pg.Conn`.

### Known limitation:

- LargeObjects are not supported by this version.
- For methods `BeginTxFunc` and `BeginFunc` time for `Begin` statement is not measured precisely.
- Span for `CopyFrom` statement does not contain detailed information about the parameters.
- Function `QueryRow` requires the result row to be `Scan`. This operation will close the corresponding span and prevent resource leaks. Span's duration will be a sum of query and scan duration.
- For a `SendBatch` result, when methods either `Exec`, `Query`, `QueryRow` or `QueryFunc` are called, the correspondent span will have the wrong duration because these methods are only reading the result of the execution that happened before.

Installation
------------

To add the module to your `go.mod` file, run the following command in your project directory:

```bash
$ go get github.com/instana/go-sensor/instrumentation/instapgx
```

Usage
-----

```go
// Create a sensor
sensor := instana.NewSensor("pgx-sensort")

// Parse config
conf, err := pgx.ParseConfig("postgres://postgres:mysecretpassword@localhost/postgres")
...

// Instrument connection 
conn, err := instapgx.ConnectConfig(context.Background(), sensor, conf)
```

For a `SendBatch` method, to have more information about statements in the span, please enable detailed mode.

```go
// Enable detailed batch mode globally
instapgx.EnableDetailedBatchMode()
...
// Create batch
b := &pgx.Batch{}
...
// Send batch
br := conn.SendBatch(ctx, b)
```


Examples
---

[Connection examples](https://github.com/instana/go-sensor/blob/master/instrumentation/instapgx/example_conn_test.go)

[Transaction examples](https://github.com/instana/go-sensor/blob/master/instrumentation/instapgx/example_tx_test.go)

Testing
---
To run integration tests, a PostgreSQL database is required in the environment. It can be started with:

```bash
docker run -e POSTGRES_PASSWORD=mysecretpassword -p 5432:5432 -d postgres
```
