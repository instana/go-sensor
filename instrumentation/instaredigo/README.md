Instana instrumentation for Redigo
=====================================

This module contains instrumentation code for Redis clients written with [`Redigo`](https://pkg.go.dev/github.com/gomodule/redigo)

[![GoDoc](https://pkg.go.dev/badge/github.com/instana/go-sensor/instrumentation/instaredigo)](https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instaredigo)

Installation
------------

To add the module to your `go.mod` file, run the following command in your project directory:

```bash
$ go get github.com/instana/go-sensor/instrumentation/instaredigo
```

Usage
-----
`instaredigo` offers function wrappers for [redis.DialContext()](https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instaredigo#DialContext), [redis.Dial()](https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instaredigo#Dial), [redis.DialURL()](https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instaredigo#DialURL), [redis.DialURLContext()](https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instaredigo#DialURLContext), [redis.NewConn()](https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instaredigo#NewConn) and for [redis.Conn](https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instaredigo#Conn)  that instrument an instance of `redis.Conn` using the provided `instana.Sensor` to trace Redis calls made with this instance.

```go
conn, err := instaredigo.Dial(sensor, "tcp", ":6379")
if err != nil {
    //handle error
}
reply, err := conn.Do("SET", "greetings", "helloworld")
```

See the [`instaredigo` package documentation](https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instaredigo) for detailed examples.
