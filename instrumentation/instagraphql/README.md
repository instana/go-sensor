Instana instrumentation for graphql
=============================================

This module contains instrumentation code for the [`graphql API`](https://pkg.go.dev/github.com/graphql-go/graphql).

[![GoDoc](https://pkg.go.dev/badge/github.com/instana/go-sensor/instrumentation/instagraphql)][godoc]

Installation
------------

To add the module to your `go.mod` file run the following command in your project directory:

```bash
$ go get github.com/instana/go-sensor/instrumentation/instagraphql
```

Usage
-----

`instagraphql` offers a function wrapper for the [`graphql.Do()`][instagraphql.Do] method capable of collecting the relevant
data about the GraphQL query and sending them to the Instana backend.

See the usage example below:

```go
// create an instance of the Instana sensor
sensor := instana.NewSensor("go-graphql")

// setup GraphQL normally
...
schema, err := graphql.NewSchema(schemaConfig)
...

// Create a graphql.Params instance to prepare the query to be processed

query := `query myQuery {
  myfield
}`

params := graphql.Params{Schema: schema, RequestString: query}

// Call instagraphql.Do instead of the original graphql.Do.
// Make sure to provide a valid context (usually an HTTP req.Context()) if any.
r := instagraphql.Do(context.Background(), sensor, params)

fmt.Println("do something with the result", r)
```

See the [`instagraphql` package documentation][godoc] for detailed examples.


[godoc]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instagraphql
[instagraphql.Do]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instagraphql#Do
