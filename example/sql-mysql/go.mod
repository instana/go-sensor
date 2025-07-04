module sql-mysql.com

go 1.23

require (
	github.com/go-sql-driver/mysql v1.9.3
	github.com/instana/go-sensor v1.67.7
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/looplab/fsm v1.0.3 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
)

replace github.com/instana/go-sensor => ../..
