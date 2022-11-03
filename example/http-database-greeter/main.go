// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	instana "github.com/instana/go-sensor"
	"github.com/opentracing/opentracing-go"

	_ "github.com/lib/pq"
)

var (
	db   *sql.DB
	args struct {
		DBConnStr  string
		ListenAddr string
	}
)

func main() {
	flag.StringVar(&args.DBConnStr, "db", os.Getenv("POSTGRES"), "PostgreSQL connection string")
	flag.StringVar(&args.ListenAddr, "l", os.Getenv("LISTEN_ADDR"), "Server listen address")
	flag.Parse()

	if args.DBConnStr == "" || args.ListenAddr == "" {
		flag.Usage()
		os.Exit(2)
	}

	// First we create an instance of instana.Sensor, a container that will be used to inject
	// tracer into all instrumented methods.
	sensor := instana.NewSensor("greeter-server")

	// Create a new DB connection using instana.SQLOpen(). This is a drop-in replacement for
	// database/sql.Open() that makes sure that the instrumented version of a driver is used.
	var err error
	db, err = instana.SQLInstrumentAndOpen(sensor, "postgres", args.DBConnStr)
	if err != nil {
		log.Fatalln(err)
	}

	// Here we instrument an HTTP handler by wrapping it with instana.TracingHandlerFunc() middleware.
	// The `/{name}` parameter here is used as a path template, so requests with different parameters
	// could be grouped together in the dashboard.
	http.HandleFunc("/", instana.TracingHandlerFunc(sensor, "/{name}", handle))

	log.Printf("greeter server is listening on %s...", args.ListenAddr)
	if err := http.ListenAndServe(args.ListenAddr, nil); err != nil {
		log.Panicln(err)
	}
}

// An HTTP handler that responds with a greeting message, using the name provided
// in request path and then updates the time a user was last seen in background.
// We also simulate work here by sleeping before sending a response.
func handle(w http.ResponseWriter, req *http.Request) {
	var span opentracing.Span

	// The HTTP instrumentation injects an entry span into request context, so it can be used
	// as a parent for any operation resulted from an incoming HTTP request. Here we're checking
	// whether the parent span present in request context, which means that our handler has been
	// instrumented.
	if parent, ok := instana.SpanFromContext(req.Context()); ok {
		// Since our handler does some substantial "work", we'd like to have more visibility
		// on how much time it takes to process a request. For this we're starting an _intermediate_
		// span that will be finished as soon as handling is finished.
		span = parent.Tracer().StartSpan("handle", opentracing.ChildOf(parent.Context()))
		defer span.Finish()
	}

	// Extract request parameter from the path, for simplicity we consider
	// anything that comes after the leading slash as a name
	name := strings.TrimPrefix(req.URL.Path, "/")
	if name == "" {
		name = "stranger"
	}

	time.Sleep(100 * time.Millisecond) // simulate work

	// Respond with a greeting message
	fmt.Fprintln(w, "Hello, "+strings.Title(name)+"!")

	// We don't want the database query to be cancelled after the request handling is done, so we need to
	// create a new context and inject the handler span there to make logging a part of this trace.
	ctx := context.Background()
	if span != nil {
		ctx = instana.ContextWithSpan(ctx, span)
	}

	// Update last greet time in background
	go logLastGreetTime(ctx, name)
}

func logLastGreetTime(ctx context.Context, name string) {
	// This is just a dummy request that simutates an update in a PostgreSQL database. In order for
	// the SQL driver instrumentation to pick up the trace and create an exit span for the database
	// call, we need to pass the context that holds an active span
	_, err := db.ExecContext(ctx, "SELECT $1 || ' greeted at ' || NOW();", name)
	if err != nil {
		log.Printf("failed to record last greet time: %s", err)
	}
}
