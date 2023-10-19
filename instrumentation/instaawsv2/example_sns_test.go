// (c) Copyright IBM Corp. 2023

package instaawsv2_test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instaawsv2"
)

var ()

func Example_sns() {
	mux := http.NewServeMux()

	tr := instana.InitCollector(instana.DefaultOptions())

	snsClient, err := getSNSClientV2(tr)
	if err != nil {
		log.Fatal("Unable to create the sns client. Check the config. Details: ", err.Error())
	}

	msg := "this is a test message for amazon sns at " + time.Now().String()
	mux.HandleFunc("/testsnsmessage",
		instana.TracingHandlerFunc(tr, "/testsnsmessage",
			handleSNSSendMessage(snsClient, msg)))

	log.Println("Starting service for testing sns messages ...")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func getSNSClientV2(tr instana.TracerLogger) (*sns.Client, error) {
	var client *sns.Client

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("error while loading aws config: %s", err.Error())
	}

	instaawsv2.Instrument(tr, &cfg)

	client = sns.NewFromConfig(cfg)

	return client, nil
}

func handleSNSSendMessage(client *sns.Client, msg string) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		//Provide values based on the SNS topic
		topicArn := ""

		_, err := client.Publish(request.Context(), &sns.PublishInput{
			Message:  &msg,
			TopicArn: &topicArn,
		})

		var errMsg string
		if err != nil {
			errMsg = fmt.Sprintf("Unable to send message to the sns. Error: %s",
				err.Error())
		} else {
			errMsg = fmt.Sprintf("Successfully sent the message to the sns topic")
		}
		writer.Write([]byte(errMsg))
	}
}
