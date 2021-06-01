// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2016

package main

import (
	"flag"

	"github.com/gin-gonic/gin"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instagin"
)

const (
	defaultPort    = "8881"
	defaultAddress = "localhost"
)

func main() {
	port := flag.String("port", defaultPort, "port to use by an example")
	address := flag.String("address", defaultAddress, "address to use by an example")

	flag.Parse()
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

	{
		v1.GET("/myendpoint", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "pong",
			})
		})
	}

	engine.Run(*address + ":" + *port)
}
