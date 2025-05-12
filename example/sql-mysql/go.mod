module sql-mysql.com

go 1.23.0

require (
	github.com/go-sql-driver/mysql v1.7.1
	github.com/instana/go-sensor v1.67.1
)

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/looplab/fsm v1.0.2 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
)

replace github.com/instana/go-sensor => ../..
