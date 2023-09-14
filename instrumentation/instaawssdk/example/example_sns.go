package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instaawssdk"
)

var (
	//Provide values based on the SNS topic
	topicArn               = ""
	messageGroupId         = ""
	messageDeduplicationId = ""
)

func main() {
	mux := http.NewServeMux()

	tr := instana.InitCollector(instana.DefaultOptions())

	snsClient, err := getSNSClientV1(tr)
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

func getSNSClientV1(tr instana.TracerLogger) (*sns.SNS, error) {
	var client *sns.SNS

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-central-1")},
	)

	if err != nil {
		return nil, fmt.Errorf("error while creating session. Details: %s", err.Error())
	}

	instaawssdk.InstrumentSession(sess, tr)

	client = sns.New(sess)

	return client, nil
}

func handleSNSSendMessage(client *sns.SNS, msg string) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {

		_, err := client.Publish(&sns.PublishInput{
			Message:                &msg,
			MessageDeduplicationId: &messageDeduplicationId,
			MessageGroupId:         &messageGroupId,
			TopicArn:               &topicArn,
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
