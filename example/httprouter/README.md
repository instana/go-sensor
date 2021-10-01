An Instrumented github.com/julienschmidt/httprouter Server Example
==================================================================

An example of instrumenting a `github.com/julienschmidt/httprouter` HTTP service with Instana using [`github.com/instana/go-sensor/tree/master/instrumentation/instahttprouter`](https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instahttprouter).

An example of usage of Instana tracer to instrument an HTTP server written with [`github.com/julienschmidt/httprouter`](https://github.com/julienschmidt/httprouter)

Usage
-----

To start an example locally on `localhost:8081` run:

```bash
go run main.go -l localhost:8081
```

In case when the port is already in use, please choose another one.

```
  -l string
        Server listen address
```

Available endpoints:
```
GET    /foo
POST   /foo/:id
DELETE /foo/:id
```
