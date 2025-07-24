// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2016

//go:build go1.17
// +build go1.17

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

	// create an instana collector
	collector := instana.InitCollector(&instana.Options{
		Service: "gin-sensor",
		Tracer:  instana.DefaultTracerOptions(),
	})

	// add middleware to the gin handlers
	instagin.AddMiddleware(collector, engine)

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
