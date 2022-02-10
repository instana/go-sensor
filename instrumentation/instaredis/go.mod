module github.com/instana/go-sensor/instrumentation/instaredis

go 1.17

require (
	github.com/go-redis/redis/v8 v8.11.4
	github.com/instana/go-sensor v1.40.0
	github.com/instana/testify v1.6.2-0.20200721153833-94b1851f4d65
	github.com/opentracing/opentracing-go v1.1.0
)

require (
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/looplab/fsm v0.1.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c // indirect
)

replace github.com/instana/go-sensor => ../..
// replace github.com/go-redis/redis/v8 => ../../../../go/pkg/mod/github.com/go-redis/redis/v8@v8.11.4
