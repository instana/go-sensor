Instana instrumentation for github.com/sirupsen/logrus
======================================================

This module contains instrumentation code for [`github.com/sirupsen/logrus`](https://github.com/sirupsen/logrus) logger.

[![PkgGoDev](https://pkg.go.dev/badge/github.com/instana/go-sensor/instrumentation/instalogrus)][godoc]

Installation
------------

To add the module to your `go.mod` file run the following command in your project directory:

```bash
$ go get github.com/instana/go-sensor/instrumentation/instalogrus
```

Usage
-----

The `instalogrus.NewHook()` collects any warning or errors logged with `logrus.Logger`, associates them with the current span
and sends to Instana.

```go
// Create a collector
collector := instana.InitCollector(&instana.Options{
	Service: "my-web-server",
	Tracer:  instana.DefaultTracerOptions(),
})

// Register the instalogrus hook
logrus.AddHook(instalogrus.NewHook(collector))

// ...

// Make sure that you provide context.Context while logging so that
// the hook could corellate log records to operations:
logrus.WithContext(ctx).
	Error("something went wrong")
```
[Full example][fullExample]



[godoc]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instalogrus
[fullExample]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instalogrus#example-package
