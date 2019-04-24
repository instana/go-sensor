package main

import (
	"net/http"
	"time"

	instana "github.com/instana/go-sensor"
)

const (
	Service = "go-microservice-14c"
	Entry   = "http://localhost:9060/golang/entry"
	Exit1   = "http://localhost:9060/golang/exit"
	Exit2   = "http://localhost:9060/instana/exit"
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

func requestExit1(parent *http.Request) (*http.Response, error) {
	client := http.Client{Timeout: 5 * time.Second}
	req := request(Exit1)
	return sensor.TracingHttpRequest("exit", parent, req, client)
}

func requestExit2(parent *http.Request) (*http.Response, error) {
	client := http.Client{Timeout: 5 * time.Second}
	req := request(Exit2)
	return sensor.TracingHttpRequest("exit", parent, req, client)
}

func server() {
	// Wrap and register in one shot
	http.HandleFunc(
		sensor.TraceHandler("entry-handler", "/golang/entry",
			func(writer http.ResponseWriter, req *http.Request) {
				requestExit1(req)
				time.Sleep(time.Second)
				requestExit2(req)
			},
		),
	)

	// Wrap and register in two separate steps, depending on your preference
	http.HandleFunc("/golang/exit",
		sensor.TracingHandler("exit-handler", func(w http.ResponseWriter, req *http.Request) {
			time.Sleep(450 * time.Millisecond)
		}),
	)

	// Wrap and register in two separate steps, depending on your preference
	http.HandleFunc("/instana/exit",
		sensor.TracingHandler("exit-handler", func(w http.ResponseWriter, req *http.Request) {
			time.Sleep(450 * time.Millisecond)
		}),
	)

	if err := http.ListenAndServe(":9060", nil); err != nil {
		panic(err)
	}
}

func main() {
	go server()
	go forever()
	select {}
}

func forever() {
	for {
		requestEntry()
		time.Sleep(500 * time.Millisecond)
	}
}
