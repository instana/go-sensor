package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instaawssdk"
)

var (
	// Provide SQS Queue url here
	queueUrl = "sqs-queue-url"
)

func main() {
	mux := http.NewServeMux()

	tr := instana.InitCollector(instana.DefaultOptions())

	sqsClient, err := getSQSClientV1(tr)
	if err != nil {
		log.Fatal("Unable to create the sqs client. Check the config. Details: ", err.Error())
	}

	msg := "this is a test message for amazon sqs at " + time.Now().String()
	mux.HandleFunc("/testsqsmessage",
		instana.TracingHandlerFunc(tr, "/testsqsmessage",
			handleSQSSendMessage(sqsClient, msg)))

	log.Println("Starting service for testing sqs messages ...")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func getSQSClientV1(tr instana.TracerLogger) (*sqs.SQS, error) {
	var client *sqs.SQS

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-central-1")},
	)

	if err != nil {
		return nil, fmt.Errorf("error while creating session. Details: %s", err.Error())
	}

	instaawssdk.InstrumentSession(sess, tr)

	client = sqs.New(sess)

	return client, nil
}

func handleSQSSendMessage(client *sqs.SQS, msg string) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		_, err := client.SendMessageWithContext(request.Context(), &sqs.SendMessageInput{
			MessageBody: &msg,
			QueueUrl:    &queueUrl,
		})

		var errMsg string
		if err != nil {
			errMsg = fmt.Sprintf("Unable to send message to the sqs queue: %s. Error: %s",
				queueUrl, err.Error())
		} else {
			errMsg = fmt.Sprintf("Successfully sent the message to the queue: %s", queueUrl)
		}
		writer.Write([]byte(errMsg))
	}
}
