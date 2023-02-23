Instana instrumentation for graphql
=============================================

This module contains instrumentation code for the [`graphql API`](https://pkg.go.dev/github.com/graphql-go/graphql).

The instrumentation is compatible with `graphql v0.8.0`. If you need support for newer versions, file an issue on this repository.

[![GoDoc](https://pkg.go.dev/badge/github.com/instana/go-sensor/instrumentation/instagraphql)][godoc]

Installation
------------

To add the module to your `go.mod` file run the following command in your project directory:

```bash
$ go get github.com/instana/go-sensor/instrumentation/instagraphql
```

Usage
-----

A complete working example can be found [here](example).

`instagraphql` offers the following method wrappers able to collect GraphQL query data:

  * [graphql.Do()][instagraphql.Do]
  * [graphql.Subscribe()][instagraphql.Subscribe]
  * [handler.ResultCallbackFn()][instagraphql.ResultCallbackFn]

### `instagraphql.Do` usage example

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

### `instagraphql.Subscribe` usage example

```go
// create an instance of the Instana sensor
sensor := instana.NewSensor("go-graphql")

...

go func() {
  ctx := context.Background()

  subscribeParams := graphql.Params{
    Context:       ctx,
    RequestString: mySubscriptionQuery,
    Schema:        schema,
  }

  instagraphql.Subscribe(ctx, sensor, subscribeParams)
}()
```

### `instagraphql.ResultCallbackFn` usage example

The `instagraphql.ResultCallbackFn` depends on the [handler.ResultCallbackFn](https://pkg.go.dev/github.com/graphql-go/handler#ResultCallbackFn) function from the [handler API](https://github.com/graphql-go/handler), which provides a convenient HTTP handler, as well as a GraphQL playground and more.


```go
// create an instance of the Instana sensor
sensor := instana.NewSensor("go-graphql")

h := handler.New(&handler.Config{
  Schema:           &schema,
  Pretty:           true,
  GraphiQL:         false,
  Playground:       true,
  // The second argument is your previous provided callback function, or nil if you had none.
  ResultCallbackFn: instagraphql.ResultCallbackFn(sensor, nil),
})

http.Handle("/graphql", h)

if err := http.ListenAndServe("0.0.0.0:9191", nil); err != nil {
  log.Fatal(err)
}
```

See the [`instagraphql` package documentation][godoc] for detailed examples.

[godoc]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instagraphql
[instagraphql.Do]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instagraphql#Do
[instagraphql.Subscribe]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instagraphql#Subscribe
[instagraphql.ResultCallbackFn]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instagraphql#ResultCallbackFn
