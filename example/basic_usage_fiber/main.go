// (c) Copyright IBM Corp. 2026

package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v3"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instafiber/v2"
	"github.com/instana/go-sensor/instrumentation/instagorm"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func agentReady() chan bool {
	ch := make(chan bool)

	go func() {
		for {
			if instana.Ready() {
				ch <- true
			}
		}
	}()

	return ch
}

func main() {

	opts := instana.DefaultTracerOptions()

	// add any other headers you would like to collect
	opts.CollectableHTTPHeaders = []string{"Host", "Connection"}

	col := instana.InitCollector(&instana.Options{
		Service:           "Fiber Basic Usage Nithin3",
		EnableAutoProfile: true,
		Tracer:            opts,
	})

	<-agentReady()

	fmt.Println("agent ready")

	dsn := "data.db"

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}

	instagorm.Instrument(db, col, dsn)

	if err = db.AutoMigrate(&student{}); err != nil {
		panic("failed to migrate the schema" + err.Error())
	}

	// fiber handler

	app := fiber.New()

	handler := instafiber.TraceHandler(col, "/", "/", func(c fiber.Ctx) error {

		db.Statement.Context = c.Context()
		db.Create(&student{Name: "Alex", RollNumber: 32})
		fmt.Println("Student added to DB!")

		// Send a string response to the client
		return c.SendString("Hello, World!")

	})

	app.Get("/", handler)

	log.Fatal(app.Listen(":3000"))
}

type student struct {
	gorm.Model
	Name       string
	RollNumber uint
}
