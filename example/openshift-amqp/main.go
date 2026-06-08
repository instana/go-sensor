// (c) Copyright IBM Corp. 2025

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instaamqp091"
	"github.com/instana/go-sensor/instrumentation/instagin"
	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	sensor     instana.TracerLogger
	amqpConn   *amqp.Connection
	amqpURL    string
	queueName  = "instana-test-queue"
	exchange   = "instana-exchange"
	routingKey = "instana.test"
)

func init() {
	// Get service name from environment or use default
	serviceName := os.Getenv("INSTANA_SERVICE_NAME")
	if serviceName == "" {
		serviceName = "amqp-service"
	}

	// Initialize Instana sensor with environment-based configuration
	// The agent host and port are automatically picked up from environment variables:
	// - INSTANA_AGENT_HOST (set via fieldRef to status.hostIP in OpenShift)
	// - INSTANA_AGENT_PORT (default: 42699)
	sensor = instana.InitCollector(&instana.Options{
		Service: serviceName,
		Tracer:  instana.DefaultTracerOptions(),
	})

	// Get RabbitMQ URL from environment or use default
	amqpURL = os.Getenv("RABBITMQ_URL")
	if amqpURL == "" {
		amqpURL = "amqp://guest:guest@rabbitmq:5672/"
	}

	log.Printf("Initializing with service name: %s", serviceName)
	log.Printf("Instana agent host: %s", os.Getenv("INSTANA_AGENT_HOST"))
	log.Printf("Instana agent port: %s", os.Getenv("INSTANA_AGENT_PORT"))
}

func agentReady() chan bool {
	ch := make(chan bool)

	go func() {
		for {
			if instana.Ready() {
				ch <- true
				return
			}
			time.Sleep(1 * time.Second)
		}
	}()

	return ch
}

func connectRabbitMQ() error {
	var err error
	maxRetries := 30
	retryDelay := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		amqpConn, err = amqp.Dial(amqpURL)
		if err == nil {
			log.Println("Successfully connected to RabbitMQ")
			return nil
		}

		log.Printf("Failed to connect to RabbitMQ (attempt %d/%d): %v", i+1, maxRetries, err)
		time.Sleep(retryDelay)
	}

	return fmt.Errorf("failed to connect to RabbitMQ after %d attempts: %w", maxRetries, err)
}

func setupQueue() error {
	ch, err := amqpConn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	defer ch.Close()

	// Declare exchange
	err = ch.ExchangeDeclare(
		exchange, // name
		"topic",  // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declare queue
	_, err = ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue to exchange
	err = ch.QueueBind(
		queueName,  // queue name
		routingKey, // routing key
		exchange,   // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	log.Printf("Queue '%s' created and bound to exchange '%s' with routing key '%s'", queueName, exchange, routingKey)
	return nil
}

// publishHandler publishes a message to RabbitMQ
func publishHandler(c *gin.Context) {
	// Get the parent span from the request context
	parentSpan, ok := instana.SpanFromContext(c.Request.Context())
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get parent span from context",
		})
		return
	}

	ch, err := amqpConn.Channel()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to open channel: " + err.Error(),
		})
		return
	}
	defer ch.Close()

	// Wrap the channel with Instana instrumentation
	wrappedCh := instaamqp091.WrapChannel(sensor, ch, amqpURL)

	// Get message from request or use default
	message := c.DefaultQuery("message", "Hello from Instana AMQP example!")

	// Publish message with instrumentation
	err = wrappedCh.Publish(
		parentSpan,
		exchange,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
			Timestamp:   time.Now(),
		},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to publish message: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      "success",
		"message":     "Message published successfully",
		"content":     message,
		"queue":       queueName,
		"exchange":    exchange,
		"routing_key": routingKey,
	})
}

// consumeHandler triggers consumption of messages from RabbitMQ
func consumeHandler(c *gin.Context) {
	ch, err := amqpConn.Channel()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to open channel: " + err.Error(),
		})
		return
	}
	defer ch.Close()

	// Wrap the channel with Instana instrumentation
	wrappedCh := instaamqp091.WrapChannel(sensor, ch, amqpURL)

	// Consume messages
	msgs, err := wrappedCh.Consume(
		queueName, // queue
		"",        // consumer
		true,      // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to consume messages: " + err.Error(),
		})
		return
	}

	// Collect messages with timeout
	var messages []string
	timeout := time.After(2 * time.Second)
	messageCount := 0

	for {
		select {
		case msg, ok := <-msgs:
			if !ok {
				c.JSON(http.StatusOK, gin.H{
					"status":        "success",
					"message_count": messageCount,
					"messages":      messages,
				})
				return
			}
			messages = append(messages, string(msg.Body))
			messageCount++
			if messageCount >= 10 { // Limit to 10 messages per request
				c.JSON(http.StatusOK, gin.H{
					"status":        "success",
					"message_count": messageCount,
					"messages":      messages,
					"note":          "Limited to 10 messages",
				})
				return
			}
		case <-timeout:
			c.JSON(http.StatusOK, gin.H{
				"status":        "success",
				"message_count": messageCount,
				"messages":      messages,
			})
			return
		}
	}
}

// healthHandler checks the health of the application
func healthHandler(c *gin.Context) {
	// Check RabbitMQ connection
	if amqpConn == nil || amqpConn.IsClosed() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":   "unhealthy",
			"rabbitmq": "disconnected",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "healthy",
		"rabbitmq": "connected",
		"instana":  instana.Ready(),
	})
}

func main() {
	// Wait for Instana agent to be ready
	log.Println("Waiting for Instana agent to be ready...")
	<-agentReady()
	log.Println("Instana agent is ready")

	// Connect to RabbitMQ
	log.Println("Connecting to RabbitMQ...")
	if err := connectRabbitMQ(); err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer amqpConn.Close()

	// Setup queue
	if err := setupQueue(); err != nil {
		log.Fatalf("Failed to setup queue: %v", err)
	}

	// Setup Gin router
	router := gin.Default()
	instagin.AddMiddleware(sensor, router)

	// Define routes
	router.GET("/health", healthHandler)
	router.GET("/publish", publishHandler)
	router.GET("/consume", consumeHandler)
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "AMQP Service with Instana",
			"endpoints": map[string]string{
				"health":  "/health - Check service health",
				"publish": "/publish?message=your_message - Publish a message to RabbitMQ",
				"consume": "/consume - Consume messages from RabbitMQ",
			},
		})
	})

	// Start server
	log.Println("Starting server on :8085")
	if err := router.Run(":8085"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// Made with Bob
