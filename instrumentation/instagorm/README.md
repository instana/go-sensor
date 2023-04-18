Instana instrumentation for gorm
=============================================

This module provides the instrumentation for database handling using [`gorm`](https://github.com/go-gorm/gorm) library.

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
sensor := instana.NewSensor("gin-sensor")

dsn := "<relevant DSN information for the database>"

// example here uses sqlite driver
db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})

// instrument the GORM database handle
instagorm.Instrument(sensor, db, dsn)


...
```



[godoc]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instagorm
