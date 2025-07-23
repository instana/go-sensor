// (c) Copyright IBM Corp. 2023

//go:build go1.18
// +build go1.18

package main

import (
	"fmt"
	"time"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instagorm"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func agentReady() chan bool {
	ch := make(chan bool)

	go func() {
		for {
			if instana.Ready() {
				ch <- true
				break
			}
			time.Sleep(time.Millisecond)
		}
	}()

	return ch
}

func main() {
	hold := make(chan bool)

	s := instana.InitCollector(&instana.Options{
		Service: "gorm-postgres",
		Tracer:  instana.DefaultTracerOptions(),
	})

	<-agentReady()
	fmt.Println("agent ready")

	dsn := "host=localhost user=postgres password=mysecretpassword dbname=postgres port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		panic(err)
	}

	instagorm.Instrument(db, s, dsn)

	var stud student

	db.First(&stud)

	fmt.Println(">>>", stud.StudentName)

	fmt.Println("holding process up")
	<-hold
}

type student struct {
	StudentName string `gorm:"column:studentname"`
	StudentID   uint   `gorm:"primarykey,column:studentid"`
}

// implementing the schema.Tabler interface
func (student) TableName() string {
	return "student"
}
