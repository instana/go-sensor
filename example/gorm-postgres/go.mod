module github.com/instana/go-sensor/gorm-postgres

go 1.23.0

require (
	github.com/instana/go-sensor v1.67.3
	github.com/instana/go-sensor/instrumentation/instagorm v1.19.0
	gorm.io/driver/postgres v1.5.11
	gorm.io/gorm v1.26.0
)

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgx/v5 v5.5.5 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/looplab/fsm v1.0.2 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	golang.org/x/crypto v0.35.0 // indirect
	golang.org/x/sync v0.13.0 // indirect
	golang.org/x/text v0.24.0 // indirect
)

replace (
	github.com/instana/go-sensor v1.67.1 => ../../
	github.com/instana/go-sensor/instrumentation/instagorm v1.19.0 => ../../instrumentation/instagorm
)
