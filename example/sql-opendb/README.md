User Database Service
-------------
This application demonstrates the usage of `Instana Go Tracer SDK` for instrumenting database operations. In the provided example, one can understand how to use the Instana SDK to wrap the `sql.Connector` and use the `sql.OpenDB` API to make database calls.

## Running the application
- An Instana Host agent must be running to collect the traces.
- Compile the example by issuing `go build -o server .` and run the application using `./server`

## Querying the server
The available routes are,
- localhost:8080/adduser

Sample query: curl -X GET http://localhost:8080/adduser

## Output
After querying the above-mentioned route, you will be able to see the call traces in the Instana dashboard.

