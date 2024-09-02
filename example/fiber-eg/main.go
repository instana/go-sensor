// (c) Copyright IBM Corp. 2023

//go:build go1.18
// +build go1.18

package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instafiber"
	"github.com/instana/go-sensor/instrumentation/instagorm"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func main() {

	os.Setenv("INSTANA_TIMEOUT", "2000")

	// Create a sensor for instana instrumentation
	sensor := instana.InitCollector(&instana.Options{
		Service:  "nithin-fb-example-with-host-agent",
		LogLevel: instana.Debug,
	})

	var err error
	dsn := "host=localhost user=postgres password=mysecretpassword dbname=postgres port=5432 sslmode=disable"
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		panic(err)
	}

	instagorm.Instrument(db, sensor, dsn)

	app := fiber.New()

	// Use the instafiber.TraceHandler for instrumenting the handler
	app.Get("/greet", instafiber.TraceHandler(sensor, "greet", "/greet", hello))

	// Start server
	log.Fatal(app.Listen(":3000"))

	// test()
}

func hello(c *fiber.Ctx) error {

	var stud student

	db.WithContext(c.UserContext()).First(&stud)

	return c.SendString("Hello " + stud.StudentName + "!")
}

type student struct {
	StudentName string `gorm:"column:studentname"`
	StudentID   uint   `gorm:"primarykey,column:studentid"`
}

// implementing the schema.Tabler interface
func (student) TableName() string {
	return "student"
}
