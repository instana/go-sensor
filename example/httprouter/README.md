An Instrumented HTTP Server Example
===================================

An example of usage of Instana tracer to instrument an HTTP server written with [github.com/julienschmidt/httprouter](https://github.com/julienschmidt/httprouter)

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
