## Kafka Instrumentation Example

This example demonstrates how the Instana Go Tracer can instrument Kakfa messages with the sarama API.

### Steps

1. Make sure to have an active Instana Agent running and properly configured (details are not scope of this README).
1. Make sure that Kafka is running. For convenience, you can run `docker-compose up kafka` from the `example` directory.
1. Run `go run .` in order to call the example application.
1. After a few seconds, the trace can be seen in the Instana UI.
