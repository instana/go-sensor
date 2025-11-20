// (c) Copyright IBM Corp. 2024

package main

import (
	"github.com/gin-gonic/gin"
	_ "github.com/godror/godror"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instagin"
)

var s instana.TracerLogger

func init() {
	s = instana.InitCollector(&instana.Options{
		Service: "oracle-service",
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
	// TNS connection string format
	// user/password@(DESCRIPTION=(ADDRESS=(PROTOCOL=TCP)(HOST=localhost)(PORT=1521))(CONNECT_DATA=(SERVICE_NAME=XEPDB1)))
	connStr := "system/oracle@(DESCRIPTION=(ADDRESS=(PROTOCOL=TCP)(HOST=localhost)(PORT=1521))(CONNECT_DATA=(SERVICE_NAME=XEPDB1)))"

	db, err := instana.SQLInstrumentAndOpen(s, "godror", connStr)
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}
	defer db.Close()

	r, err := db.QueryContext(c.Request.Context(), "SELECT 'Current date is ' || TO_CHAR(SYSDATE, 'YYYY-MM-DD') FROM DUAL")

	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}
	defer r.Close()

	var res string

	for r.Next() {
		if err := r.Scan(&res); err != nil {
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
			return
		}
	}

	c.JSON(200, gin.H{
		"message": res + " - hello from Oracle",
	})
}

func main() {
	<-agentReady()

	router := gin.Default()
	instagin.AddMiddleware(s, router)
	router.GET("/oracle-test", handler)

	router.Run("localhost:8086")
}
