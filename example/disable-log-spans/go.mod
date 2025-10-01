module github.com/instana/go-sensor/example/disable_log_spans

go 1.23.0

replace github.com/instana/go-sensor => ../../../go-sensor

require (
	github.com/instana/go-sensor v1.70.0
	github.com/instana/go-sensor/instrumentation/instagorm v1.32.0
	github.com/instana/go-sensor/instrumentation/instalogrus v1.34.0
	github.com/sirupsen/logrus v1.9.3
	gorm.io/driver/sqlite v1.6.0
	gorm.io/gorm v1.31.0
)

require (
	github.com/google/pprof v0.0.0-20250630185457-6e76a2b096b5 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/looplab/fsm v1.0.3 // indirect
	github.com/mattn/go-sqlite3 v1.14.28 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/text v0.27.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
