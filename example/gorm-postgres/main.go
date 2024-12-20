// (c) Copyright IBM Corp. 2023

//go:build go1.18
// +build go1.18

package main

import (
	"fmt"
	"log"
	"net/http"
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

var (
	collector instana.TracerLogger
)

func init() {
	collector = instana.InitCollector(&instana.Options{
		Service: "gorm-sample-app-reproduce",
	})
}

func main() {
	http.HandleFunc("/gorm-sample", instana.TracingHandlerFunc(collector, "/gorm-sample", handler))
	log.Fatal(http.ListenAndServe("localhost:9990", nil))
}

func handler(w http.ResponseWriter, r *http.Request) {

	<-agentReady()
	fmt.Println("agent ready")

	ctx := r.Context()

	// dsn := "host=localhost user=postgres password=mysecretpassword dbname=postgres port=5432 sslmode=disable"
	dsn := "postgresql://postgres@localhost:5432/postgres?sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		panic(err)
	}

	db = db.WithContext(ctx)
	instagorm.Instrument(db, collector, dsn)

	var stud student

	db.First(&stud)

	fmt.Println(">>>", stud.StudentName)

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("{message:Status OK! Check terminal for full log!}"))
}

type student struct {
	StudentName string `gorm:"column:studentname"`
	StudentID   uint   `gorm:"primarykey,column:studentid"`
}

// implementing the schema.Tabler interface
func (student) TableName() string {
	return "student"
}
