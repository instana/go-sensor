Instana instrumentation for go-redis
==========================================

This module contains instrumentation code for Redis clients written with [`go-redis`](https://pkg.go.dev/github.com/go-redis/redis/v8).

[![GoDoc](https://pkg.go.dev/badge/github.com/instana/go-sensor/instrumentation/instaredis)][godoc]

Installation
------------

To add the module to your `go.mod` file run the following command in your project directory:

```bash
$ go get github.com/instana/go-sensor/instrumentation/instaredis
```

Usage
-----

`instaredis` offers function wrappers for [`redis.NewClient()`][instaredis.WrapClient], [`redis.NewFailoverClient()`][instaredis.WrapClient],
[`redis.NewClusterClient()`][instaredis.WrapClusterClient] and [`redis.NewFailoverClusterClient()`][instaredis.WrapClusterClient]
that instrument an instance of `redis.Client` or `redis.ClusterClient` by adding hooks to the redis client. These hooks then
use the provided `instana.Sensor` to trace Redis calls made with this client instance:

```go
rdb := redis.NewClient(&redis.Options{
	Addr: "localhost:6382",
})

instaredis.WrapClient(rdb, sensor)
```


See the [`instaredis` package documentation][godoc] for detailed examples.


[godoc]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instaredis
[instaredis.WrapClient]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instaredis#WrapClient
[instaredis.WrapClusterClient]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instaredis#WrapClusterClient
