// (c) Copyright IBM Corp. 2023

package instaawsv2_test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instaawsv2"
)

func Example_sqs() {
	mux := http.NewServeMux()

	tr := instana.InitCollector(instana.DefaultOptions())

	sqsClient, err := getSQSClientV2(tr)
	if err != nil {
		log.Fatal("Unable to create the sqs client. Check the config. Details: ", err.Error())
	}

	msg := "this is a test message for amazon sqs at " + time.Now().String()
	mux.HandleFunc("/testsqsmessage",
		instana.TracingHandlerFunc(tr, "/testsqsmessage",
			handleSQSSendMessage(sqsClient, msg)))

	log.Println("Starting service v2 for testing sqs messages ...")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func getSQSClientV2(tr instana.TracerLogger) (*sqs.Client, error) {
	var client *sqs.Client

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("error while loading aws config: %s", err.Error())
	}

	instaawsv2.Instrument(tr, &cfg)

	client = sqs.NewFromConfig(cfg)

	return client, nil
}

func handleSQSSendMessage(client *sqs.Client, msg string) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		// Provide SQS Queue url here
		queueUrl := ""

		_, err := client.SendMessage(request.Context(), &sqs.SendMessageInput{
			MessageBody:  &msg,
			QueueUrl:     &queueUrl,
			DelaySeconds: 0,
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
