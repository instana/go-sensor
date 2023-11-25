# Instana instrumentation of Couchbase SDK v2 (gocb) for Go

This module contains the code for instrumenting the Couchbase SDK, based on the [`gocb`](https://github.com/couchbase/gocb) library for Go.

The following services are currently instrumented:

1. Cluster
   1. Query
   2. SearchQuery
   3. AnalyticsQuery
2. Bucket
3. Bucket Manager
   1. Get/Create/Update/Drop/Flush Bucket
4. Collection
   1. Bulk Operations
   2. Data operation (Insert, Upsert, Get etc)
   3. Sub-doc operations
      1. LookupIn & MutateIn
5. Scope
   1. Query
   2. AnalyticsQuery
6. Binary Collection
7. Collection Manager
8. Collection Data Structures
   1. List
   2. Map
   3. Queue
   4. Set

> [!NOTE]
> While you can call methods such as QueryIndexManager or SearchIndexManager from the provided instagocb cluster interface, it's important to note that tracing support for these methods is currently not implemented. If you find the instrumentation of any unlisted features necessary, please feel free to raise an issue.


Installation
------------

To add the module to your `go.mod` file run the following command in your project directory:

```bash
$ go get github.com/instana/go-sensor/instrumentation/instagocb
```

Usage
------

- Instead of using `gocb.Connect`, use `instagocb.Connect` to connect to the Couchbase server. 
  - The function definition seems identical, except you need to pass an extra argument of `instana.TraceLogger` to instagocb.Connect.
- For each instrumented service, you will find an interface in `instagocb`. Use this interface instead of using the direct instances from `gocb`.
  - For example, instead of `*gocb.Cluster`, use `instagocb.Cluster` interface.
  - `*gocb.Collection` becomes `instagocb.Collection`.
  - This applies to all instrumented services.
- If you use `instagocb.Connect`, the returned cluster will be able to provide all the instrumented functionalities. For example, if you use `cluster.Buckets()`, it will return an instrumented `instagocb.BucketManager` interface instead of `*gocb.BucketManager`.
- Set the `ParentSpan` property of the `options` function argument using `instagocb.GetParentSpanFromContext(ctx)` if your Couchbase call is part of some HTTP request or something. Otherwise, the parent-child relationship of the spans won't be tracked (see the example for a full demo).
- There is an `Unwrap()` method in all instagocb provided interfaces; it will return the underlying gocb instance. For example, `cluster.Unwrap()` will return an instance of `*gocb.Cluster`.
  
> [!IMPORTANT]
> Use `Unwrap()` if you need the original instance other than the instrumented one. It is not advisable to use this directly, as Instana tracing will not be enabled if you directly utilize this instance.


Sample Usage
------------
 ```go

    var collector instana.TracerLogger
    collector = instana.InitCollector(&instana.Options{
		Service:           "sample-app-couchbase",
		EnableAutoProfile: true,
	}) 

    // connect to database
    // this will returns an instance of instagocb.Cluster, 
    // which is capable of enabling instana tracing for Couchbase calls.
	cluster, err := instagocb.Connect(collector, connectionString, gocb.ClusterOptions{
		Authenticator: gocb.PasswordAuthenticator{
			Username: username,
			Password: password,
		},
	})
	if err != nil {
		// Handle error
	}

	bucket := cluster.Bucket(bucketName)
	err = bucket.WaitUntilReady(5*time.Second, nil)
	if err != nil {
		// Handle error
	}

	collection := bucket.Scope("tenant_agent_00").Collection("users")

	type User struct {
		Name      string   `json:"name"`
		Email     string   `json:"email"`
		Interests []string `json:"interests"`
	}

	// Create and store a Document
	_, err = col.Upsert("u:jade",
		User{
			Name:      "Jade",
			Email:     "jade@test-email.com",
			Interests: []string{"Swimming", "Rowing"},
		}, &gocb.UpsertOptions{
            // If you are using couchbase call as part of some http request or something,
            // you need to set this parentSpan property using `instagocb.GetParentSpanFromContext` method,
            // Else the parent-child span relationship wont be tracked.
            // You can keep this as nil, otherwise.
            ParentSpan: instagocb.GetParentSpanFromContext(ctx)
        })
	if err != nil {
		// Handle error
	}

```
[Full example][fullExample]

[fullExample]: ../../example/couchbase/main.go

<!---
Mandatory comment section for CI/CD !!
target-pkg-url: github.com/couchbase/gocb/v2
current-version: v2.7.0
--->
