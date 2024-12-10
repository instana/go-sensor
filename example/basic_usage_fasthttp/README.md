Example application for instafasthttp
=======================================

This application demonstrates the instrumentation capabilities of instafasthttp, an instrumentation package for the [`fasthttp`](https://pkg.go.dev/github.com/valyala/fasthttp) library. It also showcases how trace propagation works, from an instrumented HTTP handler to database calls. Additionally, the example highlights various features of the instafasthttp library, such as:

1. How to instrument a server.
2. Using RoundTripper for instrumenting client calls.
3. Leveraging context for trace propagation.

The example application includes four routes, all of which are self-explanatory. Comments have been added in the code to make it more user-friendly and informative.

Running the application
------------------------
- An Instana Host agent must be running to collect the traces (Or you can use agentless tracing)
- Run `docker-compose up` from the `go-sensor` root folder. This will bring the database service up.
- Compile the example by `go build -o server .` and run the application using `./server`

## Querying the server
The available routes are,
- [localhost:8080/insert](http://localhost:7070/greet)
- [localhost:8080/round-trip](http://localhost:7070/round-trip)
- [localhost:8080/error-handler](http://localhost:7070/error-handler)
- [localhost:8080/panic-handler](http://localhost:7070/panic-handler)

After issuing a couple of API requests, you will be able to see the call traces in the Instana dashboard.
