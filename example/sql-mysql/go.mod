module sql-mysql.com

go 1.18

require (
	github.com/go-sql-driver/mysql v1.7.1
	github.com/instana/go-sensor v1.59.0
)

require (
	github.com/looplab/fsm v1.0.1 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
)

replace github.com/instana/go-sensor => ../..
