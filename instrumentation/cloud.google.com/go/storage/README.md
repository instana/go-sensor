Instana instrumentation for Google Cloud Storage
================================================

This module contains instrumentation code for [Google Cloud Storage][gcs] clients that use `cloud.google.com/go/storage` library starting from `v1.7.0` and above.

[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=blue)][godoc]

Installation
------------

To add the module to your `go.mod` file run the following command in your project directory:

```bash
$ go get github.com/instana/go-sensor/instrumentation/cloud.google.com/go/storage
```

Usage
-----

This module is a drop-in replacement for `cloud.google.com/go/storage`. However if your code references any value types,
e.g. `cloud.google.com/go/storage.ObjectAttrs`, you might need to add a named import for the original library as well.

The instrumentation is implemented as a thin wrapper around service object methods and does not change their behavior. Thus
any limitations/usage patterns/recommendations for the original method also apply to the wrapped one.

In most cases changing the import path of `cloud.google.com/go/storage` should be enough:

```go
package main

import (
	"google.golang.org/api/iterator"
	"github.com/instana/go-sensor/instrumentation/cloud.google.com/go/storage" // replaces "cloud.google.com/go/storage"
	gstorage "cloud.google.com/go/storage" // in case your code references value types
)

func main() {
	c, _ := storage.NewClient(context.Background()) // creates an instrumented GCS client

	// use wrapped client as usual
	it := c.Buckets(context.Background(), "my-gcp-project")
	for {
	    bucket, err := buckets.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
		}

		// process bucket
	}
}
```

[godoc]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/cloud.google.com/go/storage
[gcs]: https://cloud.google.com/storage
