module example1.com

go 1.19

require (
	github.com/instana/go-sensor v1.52.0
	github.com/mattn/go-sqlite3 v1.14.15
	gorm.io/driver/sqlite v1.4.4
	gorm.io/gorm v1.24.6
)

require (
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/looplab/fsm v0.1.0 // indirect
	github.com/opentracing/opentracing-go v1.1.0 // indirect
)


replace github.com/instana/go-sensor => ../../../
