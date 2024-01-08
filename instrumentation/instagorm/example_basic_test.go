// (c) Copyright IBM Corp. 2023

//go:build go1.18
// +build go1.18

package instagorm_test

import (
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instagorm"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Example() {
	s := instana.NewSensor("go-sensor-gorm")

	dsn := "<DSN information for database>"

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	instagorm.Instrument(db, s, dsn)

	if err = db.AutoMigrate(&student{}); err != nil {
		panic("failed to migrate the schema")
	}

	db.Create(&student{Name: "Alex", RollNumber: 32})
}

type student struct {
	gorm.Model
	Name       string
	RollNumber uint
}
