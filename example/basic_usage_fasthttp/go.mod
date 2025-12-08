module basic_usage

go 1.24.0

require (
	github.com/instana/go-sensor v1.72.0
	github.com/instana/go-sensor/instrumentation/instafasthttp v0.18.0
	github.com/instana/go-sensor/instrumentation/instagorm v1.29.0
	github.com/valyala/fasthttp v1.68.0
	gorm.io/driver/postgres v1.6.0
	gorm.io/gorm v1.30.0
)

require (
	github.com/andybalholm/brotli v1.2.0 // indirect
	github.com/google/pprof v0.0.0-20250630185457-6e76a2b096b5 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.7.5 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/klauspost/compress v1.18.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/looplab/fsm v1.0.3 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	golang.org/x/crypto v0.45.0 // indirect
	golang.org/x/sync v0.18.0 // indirect
	golang.org/x/text v0.31.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/instana/go-sensor => ../../

replace github.com/instana/go-sensor/instrumentation/instafasthttp => ../../instrumentation/instafasthttp
