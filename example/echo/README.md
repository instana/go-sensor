An Instrumented github.com/labstack/echo Server Example
=======================================================

An example of instrumenting a `github.com/labstack/echo` HTTP service with Instana using [`github.com/instana/go-sensor/tree/master/instrumentation/instaecho`](https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instaecho).

Usage
-----

To start an example locally on `localhost:8881` run:

```bash
go run main.go -l localhost:8881
```

In case when the port is already in use, please choose another one.

```
  -l string
        Server listen address
```

There will be two endpoints exposed:

```
GET    /myendpoint
```
