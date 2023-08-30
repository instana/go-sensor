package main

import (
	"fmt"
	"strconv"
	"time"

	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

var s instana.TracerLogger

func init() {
	s = instana.InitCollector(&instana.Options{
		Service: "Custom Entry Span",
	})
}

func agentReady() chan bool {
	ch := make(chan bool)

	go func() {
		for {
			time.Sleep(time.Millisecond * 2)
			if instana.Ready() {
				ch <- true
				break
			}
		}
	}()

	return ch
}

func entrySpan() ot.Span {
	entrySp := s.StartSpan("my-entry-span", []ot.StartSpanOption{
		ext.SpanKindRPCServer,
	}...)

	return entrySp
}

func exitSpan(ps ot.Span) ot.Span {
	tags := ot.Tags{}

	// cm := map[string]int{"cmkey": 555}
	// go func() {
	// 	var counter int
	// 	for {
	// 		time.Sleep(time.Millisecond * 1)
	// 		cm["rnd"+strconv.Itoa(counter)] = counter

	// 		counter++
	// 	}
	// }()
	// tags["crazy_map"] = cm

	opts := []ot.StartSpanOption{
		ext.SpanKindRPCClient,
		ot.ChildOf(ps.Context()),
		tags,
	}

	exitSp := ps.Tracer().StartSpan("my-exit-span", opts...)

	go func() {
		var counter int
		for {
			time.Sleep(time.Millisecond * 1)
			exitSp.SetTag("rnd"+strconv.Itoa(counter), counter)

			counter++
		}
	}()

	return exitSp
}

func trace() {
	entry := entrySpan()
	defer entry.Finish()

	time.Sleep(time.Millisecond * 340)

	exit := exitSpan(entry)
	defer exit.Finish()
}

func main() {
	<-agentReady()
	fmt.Println("Agent ready")
	hold := make(chan struct{})

	go func() {
		t := time.NewTicker(time.Millisecond * 100)

		for range t.C {
			go trace()
		}
	}()

	<-hold
}
