// (c) Copyright IBM Corp. 2023

//go:build go1.18
// +build go1.18

package instaredis_test

import (
	"context"
	"time"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instaredis/v2"
	"github.com/redis/go-redis/v9"
)

func Example() {

	// Important note: Instaredis v2 needs go-redis major version v9 and is not backward compatible with older versions of go-redis.

	// Initialize Instana sensor
	sensor := instana.NewSensor("redis-client")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6382",
	})

	// Use instaredis.WrapClient() after the Redis client instance is created in order to instrument the client.
	// The same instaredis.WrapClient() can be used when creating a Redis client with redis.NewFailoverClient.
	// When creating an instance of Redis to communicate with a cluster, instaredis.WrapClusterClient() should be used
	// instead. These cases apply when a client is being created via redis.NewClusterClient() or
	// redis.NewFailoverClusterClient()
	instaredis.WrapClient(rdb, sensor)

	// Use the API normally.
	rdb.Do(ctx, "incr", "counter")
	rdb.Get(ctx, "counter").Bytes()

	pipe := rdb.Pipeline()
	pipe.Set(ctx, "name", "Instana", time.Minute)
	pipe.Incr(ctx, "some_counter")
	pipe.Exec(ctx)

	txPipe := rdb.TxPipeline()
	txPipe.Set(ctx, "email", "info@instana.com", time.Minute)
	txPipe.Incr(ctx, "some_counter")
	txPipe.Exec(ctx)
}
