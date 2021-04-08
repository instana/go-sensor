Instana instrumentation for Google Cloud IAM
============================================

This module contains instrumentation code for Google Cloud IAM clients that use `cloud.google.com/go/iam` library.

[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=blue)][godoc]

Installation
------------

To add the module to your `go.mod` file run the following command in your project directory:

```bash
$ go get github.com/instana/go-sensor/instrumentation/cloud.google.com/go/iam
```

Usage
-----

Currently, this module is meant to be used together with [the Google Cloud Storage instrumentation](../storage) and
limited to the Buckets IAM only.

[godoc]: https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/cloud.google.com/go/iam
