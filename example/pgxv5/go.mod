module pgxsample

go 1.23.0

require (
	github.com/gorilla/mux v1.8.1
	github.com/instana/go-sensor v1.70.0
	github.com/instana/go-sensor/instrumentation/instapgx/v2 v2.17.0
	github.com/jackc/pgx/v5 v5.7.5
)

require (
	github.com/google/pprof v0.0.0-20250630185457-6e76a2b096b5 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/looplab/fsm v1.0.3 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	golang.org/x/crypto v0.40.0 // indirect
	golang.org/x/text v0.27.0 // indirect
)

replace github.com/instana/go-sensor/instrumentation/instapgx/v2 => ../../instrumentation/instapgx/v2
