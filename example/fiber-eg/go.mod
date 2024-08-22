module fiber-example

go 1.21.3

require (
	github.com/gofiber/fiber/v2 v2.52.5
	github.com/instana/go-sensor v1.63.1
	github.com/instana/go-sensor/instrumentation/instafiber v0.14.0
	github.com/instana/go-sensor/instrumentation/instagorm v1.12.1
	gorm.io/driver/postgres v1.5.9
	gorm.io/gorm v1.25.11
)

require (
	github.com/andybalholm/brotli v1.0.5 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgx/v5 v5.5.5 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/klauspost/compress v1.17.0 // indirect
	github.com/looplab/fsm v1.0.1 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.51.0 // indirect
	github.com/valyala/tcplisten v1.0.0 // indirect
	golang.org/x/crypto v0.17.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
	golang.org/x/text v0.14.0 // indirect
)

replace (
	github.com/instana/go-sensor => ../../
	github.com/instana/go-sensor/instrumentation/instafiber => ../../instrumentation/instafiber
)
