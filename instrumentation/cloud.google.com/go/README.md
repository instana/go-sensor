Instana instrumentation for Google Cloud Platform
=================================================

This module contains instrumentation code for Google Cloud Platform clients that use `cloud.google.com/go` libraries.

[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=blue)][godoc]

Installation
------------

Unlike the Instana Go sensor, this instrumentation module requires Go v1.11+ which is the minimal version for `cloud.google.com/go`.

To add the module to your `go.mod` file run the following command in your project directory:

```bash
$ go get github.com/instana/go-sensor/instrumentation/cloud.google.com/go
```

Usage
-----

The instrumentation is split by Google Cloud Platform services and follows the structure of [the original library][cloud.google.com/go].

The general approach of this module is to replace `cloud.google.com/go` package in your imports section with the corresponding package of
`github.com/instana/go-sensor/instrumentation/cloud.google.com/go`. However, if your code references value types, such as
`cloud.google.com/go/storage.ObjectAttrs` or `cloud.google.com/go/iam.Policy` you might need to add a named import of the original library
as well.

Instrumented services
---------------------

| GCP service          | Instrumentation package                                                                 | Support              |
|----------------------|-----------------------------------------------------------------------------------------|----------------------|
| Google Cloud Storage | [`github.com/instana/go-sensor/instrumentation/cloud.google.com/go/storage`](./storage) | Fully instrumented   |
| Google Cloud IAM     | [`github.com/instana/go-sensor/instrumentation/cloud.google.com/go/iam`](./iam)         | GCS buckets IAM only |

[godoc]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/cloud.google.com/go
[cloud.google.com/go]: https://pkg.go.dev/cloud.google.com/go/?tab=doc
