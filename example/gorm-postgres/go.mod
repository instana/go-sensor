module github.com/instana/go-sensor/gorm-postgres

go 1.23.0

require (
	github.com/instana/go-sensor v1.67.4
	github.com/instana/go-sensor/instrumentation/instagorm v1.23.0
	gorm.io/driver/postgres v1.5.11
	gorm.io/gorm v1.26.1
)

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.7.4 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/looplab/fsm v1.0.3 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	golang.org/x/crypto v0.38.0 // indirect
	golang.org/x/sync v0.14.0 // indirect
	golang.org/x/text v0.25.0 // indirect
)

replace (
	github.com/instana/go-sensor v1.67.1 => ../../
	github.com/instana/go-sensor/instrumentation/instagorm v1.19.0 => ../../instrumentation/instagorm
)
