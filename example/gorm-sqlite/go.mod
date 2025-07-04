module github.com/instana/go-sensor/gorm-sqlite

go 1.23.0

require (
	github.com/instana/go-sensor v1.67.7
	github.com/instana/go-sensor/instrumentation/instagorm v1.28.0
	gorm.io/driver/sqlite v1.6.0
	gorm.io/gorm v1.30.0
)

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/looplab/fsm v1.0.3 // indirect
	github.com/mattn/go-sqlite3 v1.14.28 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	golang.org/x/text v0.26.0 // indirect
)

replace (
	github.com/instana/go-sensor v1.67.1 => ../../
	github.com/instana/go-sensor/instrumentation/instagorm v1.19.0 => ../../instrumentation/instagorm
)
