module github.com/instana/go-sensor/example/beego

go 1.23

require (
	github.com/beego/beego/v2 v2.3.5
	github.com/instana/go-sensor v1.67.1
	github.com/instana/go-sensor/instrumentation/instabeego v0.1.0
	github.com/opentracing/opentracing-go v1.2.0
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/looplab/fsm v1.0.1 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/prometheus/client_golang v1.19.0 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.48.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/shiena/ansicolor v0.0.0-20200904210342-c7312218db18 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/net v0.33.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/instana/go-sensor => ../../../go-sensor
	github.com/instana/go-sensor/instrumentation/instabeego => ../../instrumentation/instabeego
)
