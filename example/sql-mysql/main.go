// (c) Copyright IBM Corp. 2023

package main

import (
	"io"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	instana "github.com/instana/go-sensor"
)

var s instana.TracerLogger

func init() {
	s = instana.InitCollector(&instana.Options{
		Service: "MySQL app",
		Tracer:  instana.DefaultTracerOptions(),
	})
}

func agentReady() chan bool {
	ch := make(chan bool)

	go func() {
		for {
			if instana.Ready() {
				ch <- true
			}
		}
	}()

	return ch
}

func handler(w http.ResponseWriter, req *http.Request) {
	db, err := instana.SQLInstrumentAndOpen(s, "mysql", "go:gopw@tcp(localhost:3306)/godb")
	if err != nil {
		panic(err)
	}

	r, err := db.QueryContext(req.Context(), "SELECT 'Current date is' || CURDATE();")

	if err != nil {
		panic(err)
	}

	var buf, res string

	for r.Next() {
		r.Scan(&buf)
		res += " " + buf
	}

	io.WriteString(w, res+" - hello\n")

}

func main() {
	<-agentReady()
	http.HandleFunc("/mysql", instana.TracingHandlerFunc(s, "/mysql", handler))
	http.ListenAndServe(":9090", nil)
}
