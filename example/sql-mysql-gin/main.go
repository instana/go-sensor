// (c) Copyright IBM Corp. 2024

package main

import (
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instagin"
)

var s instana.TracerLogger

func init() {
	s = instana.InitCollector(&instana.Options{
		Service: "mysql-service",
		Tracer:  instana.DefaultTracerOptions(),
	})
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

func handler(c *gin.Context) {
	db, err := instana.SQLInstrumentAndOpen(s, "mysql", "go:gopw@tcp(localhost:3306)/godb")
	if err != nil {
		panic(err)
	}

	r, err := db.QueryContext(c.Request.Context(), "SELECT 'Current date is' || CURDATE();")

	if err != nil {
		panic(err)
	}

	var buf, res string

	for r.Next() {
		r.Scan(&buf)
		res += " " + buf
	}

	c.JSON(200, gin.H{
		"message": res + " - hello",
	})

}

func main() {
	<-agentReady()

	router := gin.Default()
	instagin.AddMiddleware(s, router)
	router.GET("/gin-test", handler)

	router.Run("localhost:8085")
}
