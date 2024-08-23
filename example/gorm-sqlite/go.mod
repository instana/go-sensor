module github.com/instana/go-sensor/gorm-sqlite

go 1.22

require (
	github.com/instana/go-sensor v1.64.0
	github.com/instana/go-sensor/instrumentation/instagorm v1.3.0
	gorm.io/driver/sqlite v1.5.4
	gorm.io/gorm v1.25.11
)

require (
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/looplab/fsm v1.0.1 // indirect
	github.com/mattn/go-sqlite3 v1.14.17 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	golang.org/x/text v0.14.0 // indirect
)

replace (
	github.com/instana/go-sensor v1.57.0 => ../../
	github.com/instana/go-sensor/instrumentation/instagorm v1.3.0 => ../../instrumentation/instagorm
)
