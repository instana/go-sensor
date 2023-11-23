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
// create a sensor
sensor := instana.NewSensor("gorm-sensor")

dsn := "<relevant DSN information for the database>"

// example here uses sqlite driver
db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})

// instrument the GORM database handle
instagorm.Instrument(db, sensor, dsn)

...
```



[godoc]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instagorm

<!---
Mandatory comment section for CI/CD !!
target-pkg-url: gorm.io/gorm
current-version: v1.25.0
--->
