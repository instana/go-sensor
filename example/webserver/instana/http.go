package main

import (
	"github.com/instana/golang-sensor"
	"net/http"
	"time"
)

const (
	Service = "go-microservice-14c"
	Entry   = "http://localhost:9060/golang/entry"
	Exit    = "http://localhost:9060/golang/exit"
)

var sensor = instana.NewSensor(Service)

func request(url string) *http.Request {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "text/plain")
	return req
}

func requestEntry() {
	client := &http.Client{Timeout: 5 * time.Second}
	req := request(Entry)
	client.Do(req)
}

func requestExit(parent *http.Request) (*http.Response, error) {
	client := http.Client{Timeout: 5 * time.Second}
	req := request(Exit)
	return sensor.TracingHttpRequest("exit", parent, req, client)
}

func server() {
	http.HandleFunc("/golang/entry",
		sensor.TracingHandler("/golang/entry", func(writer http.ResponseWriter, req *http.Request) {
			requestExit(req)
		}),
	)

	http.HandleFunc("/golang/exit",
		sensor.TracingHandler("/golang/exit", func(w http.ResponseWriter, req *http.Request) {
			time.Sleep(450 * time.Millisecond)
		}),
	)
}

func main() {
	go server()
	go forever()
	select {}
}

func forever() {
	for {
		requestEntry()
	}
}
