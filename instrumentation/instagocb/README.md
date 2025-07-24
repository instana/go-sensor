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
		Tracer:  instana.DefaultTracerOptions(),
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


Transactions
------------
- Get a full idea of using transactions in Couchbase using Go SDK from [here][gocb-transactions-example].
- Create a new transactions instance by calling `cluster.Transactions()`. Like all other instrumented features, this one also returns an instagocb provided interface (`instagocb.Transactions`) instead of the original one (`*gocb.Transactions`).
- You can use the same `transactions.Run()` method to start transactions.
```go
	// Starting transactions
	transactions := cluster.Transactions()
```
- Commit, Rollback, and all other transaction-specific things are handled by gocb only. The advantage is that you will be getting Instana tracing support on top of transactions. 
- There is one more thing you need to do in the transaction callback function. In the first line, call the below method to create an instagocb instrumented interface of `TransactionAttemptContext` and use it for the rest of the function to do all the operations like Insert, Replace, Remove, Get, and Query. 

```go
	// Create new TransactionAttemptContext from instagocb
	tacNew := cluster.WrapTransactionAttemptContext(tac, instagocb.GetParentSpanFromContext(ctx))
```
- The function signatures of the `tacNew.Replace` and `tacNew.Remove` have a small change from the original method; you need to pass the collection as an extra argument to these functions if you are using the instrumented TransactionAttemptContext. 
> [!IMPORTANT]
> If you need to use the scope or collection inside the transaction function, use the unwrapped one (`scope.Unwrap()`) instead of the instagocb interface.
  
Sample Usage
------------
 ```go
 // Starting transactions
	transactions := cluster.Transactions()
	_, err = transactions.Run(func(tac *gocb.TransactionAttemptContext) error {

		// Create new TransactionAttemptContext from instagocb
		tacNew := cluster.WrapTransactionAttemptContext(tac, instagocb.GetParentSpanFromContext(ctx))

		// Unwrapped collection is required to pass it to transaction operations
		collectionUnwrapped := collection.Unwrap()

		// Inserting a doc:
		_, err := tacNew.Insert(collectionUnwrapped, "doc-a", map[string]interface{}{})
		if err != nil {
			return err
		}

		// Getting documents:
		docA, err := tacNew.Get(collectionUnwrapped, "doc-a")
		// Use err != nil && !errors.Is(err, gocb.ErrDocumentNotFound) if the document may or may not exist
		if err != nil {
			return err
		}

		// Replacing a doc:
		var content map[string]interface{}
		err = docA.Content(&content)
		if err != nil {
			return err
		}
		content["transactions"] = "are awesome"
		_, err = tacNew.Replace(collectionUnwrapped, docA, content)
		if err != nil {
			return err
		}

		// Removing a doc:
		docA1, err := tacNew.Get(collectionUnwrapped, "doc-a")
		if err != nil {
			return err
		}
		err = tacNew.Remove(collectionUnwrapped, docA1)
		if err != nil {
			return err
		}

		// Performing a SELECT N1QL query against a scope:
		qr, err := tacNew.Query("SELECT * FROM hotel WHERE country = $1", &gocb.TransactionQueryOptions{
			PositionalParameters: []interface{}{"United Kingdom"},

			// Unwrapped scope is required here
			Scope: inventoryScope.Unwrap(),
		})
		if err != nil {
			return err
		}

		type hotel struct {
			Name string `json:"name"`
		}

		var hotels []hotel
		for qr.Next() {
			var h hotel
			err = qr.Row(&h)
			if err != nil {
				return err
			}

			hotels = append(hotels, h)
		}

		// Performing an UPDATE N1QL query on multiple documents, in the `inventory` scope:
		_, err = tacNew.Query("UPDATE route SET airlineid = $1 WHERE airline = $2", &gocb.TransactionQueryOptions{
			PositionalParameters: []interface{}{"airline_137", "AF"},
			// Unwrapped scope is required here
			Scope: inventoryScope.Unwrap(),
		})
		if err != nil {
			return err
		}

		// There is no commit call, by not returning an error the transaction will automatically commit
		return nil
	}, nil)
	var ambigErr gocb.TransactionCommitAmbiguousError
	if errors.As(err, &ambigErr) {
		log.Println("Transaction possibly committed")

		log.Printf("%+v", ambigErr)
		return nil
	}
	var failedErr gocb.TransactionFailedError
	if errors.As(err, &failedErr) {
		log.Println("Transaction did not reach commit point")

		log.Printf("%+v", failedErr)
		return nil
	}

	if err != nil {
		return err
	}
 ```
 [Full example][fullExample]

[gocb-transactions-example]: https://docs.couchbase.com/go-sdk/2.6/howtos/distributed-acid-transactions-from-the-sdk.html#creating-a-transaction

[fullExample]: ../../example/couchbase/main.go
