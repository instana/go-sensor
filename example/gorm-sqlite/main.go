// (c) Copyright IBM Corp. 2023

//go:build go1.18
// +build go1.18

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

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
	s := instana.InitCollector(&instana.Options{
		Service: "gorm-sqlite",
	})

	<-agentReady()
	fmt.Println("agent ready")

	dsn := "data.db"

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}

	// Channel to listen for interrupt or terminate signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	instagorm.Instrument(db, s, dsn)

	if err = db.AutoMigrate(&student{}); err != nil {
		panic("failed to migrate the schema" + err.Error())
	}

	ctx := context.Background()
	db.Statement.Context = ctx

	db.Create(&student{Name: "Alex", RollNumber: 32})
	fmt.Println("Student added to DB. Type ctrl+c to exit")

	<-stop

	// close db
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.Close()
	}

	err = os.Remove(dsn)
	if err != nil {
		fmt.Println("unable to delete the database file: ", dsn, ": ", err.Error())
	}
}

type student struct {
	gorm.Model
	Name       string
	RollNumber uint
}
