module http-database-greeter

go 1.21

require (
	github.com/instana/go-sensor v1.59.0
	github.com/lib/pq v1.10.9
	github.com/opentracing/opentracing-go v1.2.0
)

replace github.com/instana/go-sensor => ../../
