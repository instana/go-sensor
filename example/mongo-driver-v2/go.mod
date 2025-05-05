module github.com/instana/go-sensor/example/mongo-driver-v2

go 1.23.0

require (
	github.com/instana/go-sensor v1.67.3
	github.com/instana/go-sensor/instrumentation/instamongo/v2 v2.0.0
	go.mongodb.org/mongo-driver/v2 v2.2.0
)

require (
	github.com/golang/snappy v1.0.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/looplab/fsm v1.0.2 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	golang.org/x/crypto v0.37.0 // indirect
	golang.org/x/sync v0.13.0 // indirect
	golang.org/x/text v0.24.0 // indirect
)

replace (
	github.com/instana/go-sensor v1.67.1 => ../..
	github.com/instana/go-sensor/instrumentation/instamongo/v2 v2.0.0 => ../../instrumentation/instamongo/v2
)
