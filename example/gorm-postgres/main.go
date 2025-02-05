// (c) Copyright IBM Corp. 2023

//go:build go1.18
// +build go1.18

package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instagorm"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	dsn string
	db  *gorm.DB
)

func main() {
	//hold := make(chan bool)

	col := instana.InitCollector(&instana.Options{
		Service: "gorm-postgres",
	})

	//<-agentReady()

	if err := initDb(col); err != nil {
		panic(err)
	}

	initDb(col)

	fmt.Println("agent ready")

	mux := http.NewServeMux()
	mux.HandleFunc("/home", instana.TracingHandlerFunc(col, "/home", handleDbOps))

	if err := http.ListenAndServe(":8080", mux); err != nil {
		panic(err)
	}

	fmt.Println("holding process up")
	//<-hold
}

func initDb(col instana.TracerLogger) error {
	var err error

	dsn := "host=localhost user=postgres password=mysecretpassword dbname=postgres port=5432 sslmode=disable"
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	instagorm.Instrument(db, col, dsn)

	return nil
}

func handleDbOps(w http.ResponseWriter, r *http.Request) {
	fmt.Println("handleDbOps called")

	databaseOperation(r.Context())

}

func databaseOperation(ctx context.Context) {
	var stud student

	db.WithContext(ctx).First(&stud)

	fmt.Println(">>>", stud.StudentName)
}

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

type student struct {
	StudentName string `gorm:"column:studentname"`
	StudentID   uint   `gorm:"primarykey,column:studentid"`
}

// implementing the schema.Tabler interface
func (student) TableName() string {
	return "student"
}
