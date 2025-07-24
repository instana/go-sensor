// (c) Copyright IBM Corp. 2023

package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instaredis/v2"
	"github.com/redis/go-redis/v9"

	_ "github.com/bonede/go-redis-driver"
)

func agentReady() chan bool {
	ch := make(chan bool)

	go func() {
		for {
			time.Sleep(time.Millisecond * 5)
			if instana.Ready() {
				ch <- true
			}
		}
	}()

	return ch
}

func nameFromRedis(ctx context.Context, s instana.TracerLogger) string {

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "redispw",
	})

	instaredis.WrapClient(rdb, s)

	v := "Instana"

	cmd := rdb.Set(ctx, "name", v, 0)

	if cmd.Err() != nil {
		panic("ExecContext error:" + cmd.Err().Error())
	}

	val, err := rdb.Get(ctx, "name").Result()

	if err != nil {
		fmt.Println("error on query context")
	}

	return val
}

func main() {
	s := instana.InitCollector(&instana.Options{
		Service: "go-redis-example",
		Tracer:  instana.DefaultTracerOptions(),
	})

	<-agentReady()
	fmt.Println("agent connected")

	http.HandleFunc("/redis", instana.TracingHandlerFunc(s, "/redis", func(w http.ResponseWriter, r *http.Request) {
		name := nameFromRedis(r.Context(), s)
		io.WriteString(w, name+"\n")
	}))

	http.ListenAndServe(":9090", nil)
}
