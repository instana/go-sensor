Instana instrumentation for gorm
=============================================

This module provides Instana instrumentation for database operations using [`gorm`](https://github.com/go-gorm/gorm) library.

[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=blue)][godoc]


Installation
------------

To add the module to your `go.mod` file run the following command in your project directory:

```bash
$ go get github.com/instana/go-sensor/instrumentation/instagorm
```

Usage
-----

```go
// create a collector
collector := instana.InitCollector(&instana.Options{
    Service: "gorm-app",
    Tracer:  instana.DefaultTracerOptions(),
})

dsn := "<relevant DSN information for the database>"

// example here uses sqlite driver
db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})

// instrument the GORM database handle
instagorm.Instrument(db, collector, dsn)

...
```



[godoc]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instagorm

