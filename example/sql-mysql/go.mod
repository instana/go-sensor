module sql-mysql.com

go 1.23.0

require (
	github.com/go-sql-driver/mysql v1.9.3
	github.com/instana/go-sensor v1.68.0
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/google/pprof v0.0.0-20250630185457-6e76a2b096b5 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/looplab/fsm v1.0.3 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/stretchr/testify v1.10.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/instana/go-sensor => ../..
