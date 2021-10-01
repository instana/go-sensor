Doubler
=======

Doubler is an example app that demonstrates the use of [`github.com/instana/go-sensor/instrumentation/instasarama`](https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instasarama) to
instrument Kafka consumers and producers. It reads messages, containing integer values to the `-in` topic, multiplies
it by 2 and publishes the result to the topic defined by the `-out` flag.

Usage
-----

To start a processor consuming messages from `topic-a` and publishing the result into `topic-2a` of a Kafka cluster
with brokers listening on `localhost:9092` and `localhost:9093`, run:

```bash
go run . -in topic-a -out topic-2a localhost:9092 localhost:9093
```
