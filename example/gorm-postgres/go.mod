module github.com/instana/go-sensor/gorm-postgres

go 1.22

require (
	github.com/instana/go-sensor v1.66.2
	github.com/instana/go-sensor/instrumentation/instagorm v1.3.0
	gorm.io/driver/postgres v1.5.3
	gorm.io/gorm v1.25.12
)

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgx/v5 v5.5.4 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/looplab/fsm v1.0.1 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/sync v0.10.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	gorm.io/driver/sqlite v1.5.4 // indirect
)

replace (
	github.com/instana/go-sensor v1.57.0 => ../../
	github.com/instana/go-sensor/instrumentation/instagorm v1.3.0 => ../../instrumentation/instagorm
)
