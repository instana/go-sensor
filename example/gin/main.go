// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2016

package main

import (
	"flag"
	"os"

	"github.com/gin-gonic/gin"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instagin"
)

var listenAddr string

func main() {
	flag.StringVar(&listenAddr, "l", os.Getenv("LISTEN_ADDR"), "Server listen address")
	flag.Parse()

	if listenAddr == "" {
		flag.Usage()
		os.Exit(2)
	}

	engine := gin.Default()

	// create a sensor
	sensor := instana.NewSensor("gin-sensor")

	// add middleware to the gin handlers
	instagin.AddMiddleware(sensor, engine)

	engine.GET("/myendpoint", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// use group: v1
	v1 := engine.Group("/v1")

	v1.GET("/myendpoint", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	engine.Run(listenAddr)
}
