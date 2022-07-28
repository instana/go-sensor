Greeter Server
==============

Greeter is an example app that demonstrates how to use `github.com/instana/go-sensor` to instrument
an HTTP server that makes requests to a database with Instana. In this case we use PostgreSQL via
[`github.com/lib/pq`](https://github.com/lib/pq), but the same approach can be applied to any driver
that is compatible with `database/sql`.

Usage
-----

To start a server listening on `localhost:8081` that connects to the database using `postgres://postgres:mysecretpassword@localhost:5432/postgres?sslmode=disable` connection string, run:

```bash
docker run --name http-database-greeter-postgres -e POSTGRES_PASSWORD=mysecretpassword -p 5432:5432 -d postgres
go run . -l localhost:8081 -db "postgres://postgres:mysecretpassword@localhost:5432/postgres?sslmode=disable"
```

In your browser navigate to [localhost:8081/alice](http://localhost:8081/alice) to see the greeting message for Alice.

Trace details
-------------

The request above will generate the following trace in your Instana dashboard:

![Trace details](./trace.png)

The HTTP instrumentation has started a new entry span, potentially picking up the trace initiated by
another service monitored by Instana or the End-User Monitoring script. This span is than used as a parent
to trace the request processing, which is `time.Sleep()` in our case, and finally there is a database call
that happens after the request handling is done.
