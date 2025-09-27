// (c) Copyright IBM Corp. 2025

package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instagorm"
	"github.com/instana/go-sensor/instrumentation/instalogrus"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	c    instana.TracerLogger
	db   *gorm.DB
	once sync.Once
)

func init() {
	c = instana.InitCollector(&instana.Options{
		Service: "disable-log-example",
	})
	connectDB()
}

type student struct {
	gorm.Model
	Name       string
	RollNumber uint
}

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

func connectDB() {
	var err error
	dsn := "data.db"
	once.Do(func() {
		db, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})
		if err != nil {
			panic("failed to connect database: " + err.Error())
		}
	})

	instagorm.Instrument(db, c, dsn)
}

// Handler for /start
func startHandler(w http.ResponseWriter, r *http.Request) {
	logrus.WithContext(r.Context()).Info("Start API is called")

	db.Statement.Context = r.Context()

	if err := db.AutoMigrate(&student{}); err != nil {
		logrus.WithContext(r.Context()).Error("failed to migrate the schema" + err.Error())
	}

	db.Create(&student{Name: "Alex", RollNumber: 32})
	fmt.Println("Student added to DB.")

	w.Write([]byte("Response from API 1"))
}

// Handler for /error
func errorHandler(w http.ResponseWriter, r *http.Request) {
	logrus.WithContext(r.Context()).Info("Error API is called")

	db.Statement.Context = r.Context()

	// Try inserting into a table that doesn't exist to cause an error
	type unknownTable struct {
		ID   int
		Name string
	}

	err := db.Table("non_existent_table").Create(&unknownTable{ID: 1, Name: "Error"}).Error
	if err != nil {
		logrus.WithContext(r.Context()).Errorf("Expected DB error occurred: %v\n", err)
		http.Error(w, "Database error occurred", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Should not reach here"))
}

// Function that makes 2 calls to each API
func makeRequest() {
	urls := []string{
		"http://localhost:8080/start",
		"http://localhost:8080/start",
		"http://localhost:8080/error",
		"http://localhost:8080/error",
	}

	var wg sync.WaitGroup
	wg.Add(len(urls))

	for i, url := range urls {
		go func(i int, url string) {
			defer wg.Done()

			resp, err := http.Get(url)
			if err != nil {
				log.Printf("Request %d to %s failed: %v\n", i+1, url, err)
				return
			}
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			log.Printf("Request %d to %s got response: %s\n", i+1, url, string(body))
		}(i, url)

		time.Sleep(time.Second)
	}

	wg.Wait()
}

func main() {

	<-agentReady()

	logrus.SetLevel(logrus.InfoLevel)
	logrus.AddHook(instalogrus.NewHook(c))

	// Register handlers
	http.HandleFunc("/start", instana.TracingHandlerFunc(c, "/start", startHandler))
	http.HandleFunc("/error", instana.TracingHandlerFunc(c, "/error", errorHandler))

	// Start server in a goroutine
	go func() {
		log.Println("Starting server on :8080")
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	// Wait a bit to ensure the server has started
	time.Sleep(1 * time.Second)

	// Call the makeRequest function
	makeRequest()

	// close db
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.Close()
	}

	err := os.Remove("data.db")
	if err != nil {
		fmt.Println("unable to delete the database file: ", "data.db", ": ", err.Error())
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	fmt.Println("type CTRL+C to exit")
	// wait for Ctrl+C to exit
	<-stop
	fmt.Println("\nShutting down server...")

}
