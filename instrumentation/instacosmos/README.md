# Instana instrumentation of Azure Cosmos DB SDK v2 for Go

This module contains the code for instrumenting the Azure Cosmos DB SDK based on the [`azcosmos`](https://github.com/Azure/azure-sdk-for-go/tree/main/sdk/data/azcosmos) library for Go.

The following operations are currently instrumented:

* Container Operations
   * CreateItem
   * DeleteItem
   * ExecuteTransactionalBatch
   * ID
   * NewQueryItemsPager
   * NewTransactionalBatch
   * PatchItem
   * Read
   * ReadItem
   * ReadThroughput
   * Replace
   * ReplaceItem
   * ReplaceThroughput
   * UpsertItem

Installation
------------

To add the module to your `go.mod` file run the following command in your project directory:

```bash
$ go get github.com/instana/go-sensor/instrumentation/instacosmos
```

Usage
------

- Utilize the following functions provided in the `instacosmos` module to obtain the Azure Cosmos DB client. These functions serve as wrappers for the identically named functions in the `azcosmos` library. Notably, in contrast to the original functions, `instana.TraceLogger` needs to be passed additionally.
    - NewClient
    - NewClientFromConnectionString
    - NewClientWithKey
- The aforementioned functions return an interface that offers the functionalities of `azcosmos.Client`. Utilizing this interface, call the `NewContainer` method to acquire the `ContainerClient` interface. This interface encompasses all the operations supported by `azcosmos.ContainerClient`.   

Sample Usage
------------
 ```go

    var collector instana.TracerLogger
	collector = instana.InitCollector(&instana.Options{
		Service: "cosmos-example",
		Tracer:  instana.DefaultTracerOptions(),
	})

    // creates an KeyCredential containing the account's primary or secondary key.
	cred, err := instacosmos.NewKeyCredential(key)
	if err != nil {
		// handle error
	}

    // creates an instance of instrumented *azcosmos.Client
	client, err := instacosmos.NewClientWithKey(collector, endpoint, cred, &azcosmos.ClientOptions{})
	if err != nil {
		// handle error
	}

	// creates an instance of instrumented *azcosmos.Client
	containerClient, err := client.NewContainer(dbName, containerName)
	if err != nil {
		// handle error
	}

	pk := azcosmos.NewPartitionKeyString("newPartitionKey")

	item := map[string]string{
		"id":             "anId",
		"value":          "2",
		"myPartitionKey": "newPartitionKey",
	}

	marshalled, err := json.Marshal(item)
	if err != nil {
		// handle error
	}

	// NOTE: All Cosmos DB operations requires a parent context to be passed in. 
	// Otherwise, the trace will not occur, unless the user explicitly allows opt-in exit spans without an entry span.
	itemResponse, err := containerClient.CreateItem(context.Background(), pk, marshalled, nil)
	if err != nil {
        // handle error
	}

	fmt.Printf("Item created. ActivityId %s consuming %v RU", itemResponse.ActivityID, itemResponse.RequestCharge) 
```
[Full example][fullExample]

[fullExample]: ../../example/cosmos/main.go
