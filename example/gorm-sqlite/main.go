// (c) Copyright IBM Corp. 2023

//go:build go1.18
// +build go1.18

package main

import (
	"fmt"

	instana "github.com/instana/go-sensor"
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
	hold := make(chan bool)
	s := instana.InitCollector(&instana.Options{
		Service: "gorm-sqlite",
	})

	<-agentReady()
	fmt.Println("agent ready")

	dsn := "data.db"

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	instagorm.Instrument(db, s, dsn)

	if err = db.AutoMigrate(&student{}); err != nil {
		panic("failed to migrate the schema")
	}

	db.Create(&student{Name: "Alex", RollNumber: 32})

	fmt.Println("holding process up")
	<-hold
}

type student struct {
	gorm.Model
	Name       string
	RollNumber uint
}
