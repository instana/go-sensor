// (c) Copyright IBM Corp. 2023

package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	instana "github.com/instana/go-sensor"

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
	db, err := instana.SQLInstrumentAndOpen(s, "redis", ":redispw@localhost:6379")
	// db, err := instana.SQLInstrumentAndOpen(s, "redis", "localhost:6379")

	if err != nil {
		panic("instrument error:" + err.Error())
	}
	defer db.Close()

	_, err = db.ExecContext(ctx, "SET name Instana EX 5")
	if err != nil {
		panic("ExecContext error:" + err.Error())
	}

	rows, err := db.QueryContext(ctx, "GET name")

	if err != nil {
		fmt.Println("error on query context")
	}

	if err = rows.Err(); err != nil {
		panic(err)
	}
	defer rows.Close()

	if ok := rows.Next(); ok {
		var res string
		rows.Scan(&res)
		return res
	}

	return ""
}

func main() {
	s := instana.InitCollector(&instana.Options{
		Service: "Redis with SQL instrumentation",
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
