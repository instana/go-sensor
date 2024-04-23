module pgxsample

go 1.21

require (
	github.com/gorilla/mux v1.8.1
	github.com/instana/go-sensor v1.63.0
	github.com/instana/go-sensor/instrumentation/instapgxv2 v0.0.0
	github.com/jackc/pgx/v5 v5.6.0
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20231201235250-de7065d80cb9 // indirect
	github.com/looplab/fsm v1.0.1 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	golang.org/x/crypto v0.24.0 // indirect
	golang.org/x/text v0.16.0 // indirect
)

replace github.com/instana/go-sensor/instrumentation/instapgxv2 => ../../instrumentation/instapgxv2
