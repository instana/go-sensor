// (c) Copyright IBM Corp. 2023

package main

import (
	"errors"
	"fmt"
	"log"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instafasthttp"
	"github.com/instana/go-sensor/instrumentation/instagorm"
	"github.com/valyala/fasthttp"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	sensor instana.TracerLogger
	db     *gorm.DB
)

func sampleEndpointHandler(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusOK)
	fmt.Fprintf(ctx, "This is the first part of body!\n")

	var stud student

	uCtx := instafasthttp.UserContext(ctx)

	db.WithContext(uCtx).First(&stud)

	fmt.Fprintf(ctx, "Hello "+stud.StudentName+"!\n")

}

// request handler in fasthttp style, i.e. just plain function.
func fastHTTPHandler(ctx *fasthttp.RequestCtx) {
	fmt.Fprintf(ctx, "Hi there! RequestURI is %q\n", ctx.RequestURI())
	switch string(ctx.Path()) {
	case "/greet":
		instafasthttp.TraceHandler(sensor, "/greet", sampleEndpointHandler)(ctx)
	case "/error-handler":
		instafasthttp.TraceHandler(sensor, "/error-handler", func(ctx *fasthttp.RequestCtx) {
			ctx.SetStatusCode(fasthttp.StatusBadRequest)
			fmt.Fprintf(ctx, "This is an error!\n")
		})(ctx)
	case "/panic-handler":
		instafasthttp.TraceHandler(sensor, "/panic-handler", func(ctx *fasthttp.RequestCtx) {
			fmt.Fprintf(ctx, "This is a panic!\n")
			panic(errors.New("Panic nithin"))
		})(ctx)
	default:
		ctx.Error("Unsupported path", fasthttp.StatusNotFound)
	}
}

func init() {
	// Create a sensor for instana instrumentation
	sensor = instana.InitCollector(&instana.Options{
		Service:  "nithin-fasthttp-example",
		LogLevel: instana.Debug,
	})
}

func main() {
	// col := instana.InitCollector(&instana.Options{
	// 	Service:           "Nithin Basic Usage",
	// 	EnableAutoProfile: true,
	// })

	// http.HandleFunc("/endpoint", instana.TracingHandlerFunc(col, "/endpoint", func(w http.ResponseWriter, r *http.Request) {
	// 	w.WriteHeader(http.StatusOK)
	// }))

	var err error
	dsn := "host=localhost user=postgres password=mysecretpassword dbname=postgres port=5432 sslmode=disable"
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		panic(err)
	}

	instagorm.Instrument(db, sensor, dsn)

	log.Fatal(fasthttp.ListenAndServe(":7070", fastHTTPHandler))

	// log.Fatal(http.ListenAndServe(":8080", nil))
}

type student struct {
	StudentName string `gorm:"column:studentname"`
	StudentID   uint   `gorm:"primarykey,column:studentid"`
}

// implementing the schema.Tabler interface
func (student) TableName() string {
	return "student"
}
