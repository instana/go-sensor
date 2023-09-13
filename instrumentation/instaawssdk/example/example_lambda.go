package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instaawssdk"
)

var (
	// Provide lambda function name here
	functionName = ""
)

func main() {
	mux := http.NewServeMux()

	tr := instana.InitCollector(instana.DefaultOptions())

	lambdaClient, err := getLambdaClient(tr)
	if err != nil {
		log.Fatal("Unable to create the lambda client. Check the config. Details: ", err.Error())
	}

	mux.HandleFunc("/testlambdainvoke",
		instana.TracingHandlerFunc(tr, "/testlambdainvoke",
			handleLambdaInvoke(lambdaClient)))

	log.Println("Starting service for testing lambda invocations ...")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func getLambdaClient(tr instana.TracerLogger) (*lambda.Lambda, error) {
	var client *lambda.Lambda

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-central-1")},
	)

	if err != nil {
		return nil, fmt.Errorf("error while creating session. Details: %s", err.Error())
	}

	instaawssdk.InstrumentSession(sess, tr)

	client = lambda.New(sess)

	return client, nil
}

func handleLambdaInvoke(client *lambda.Lambda) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {

		_, err := client.InvokeWithContext(request.Context(), &lambda.InvokeInput{
			FunctionName:   &functionName,
			InvocationType: aws.String("RequestResponse"),
			Payload:        []byte("{}"),
		})

		var errMsg string
		if err != nil {
			errMsg = fmt.Sprintf("Unable to invoke the lambda: %s. Error: %s",
				functionName, err.Error())
		} else {
			errMsg = fmt.Sprintf("Successfully invoked the lambda: %s", functionName)
		}
		writer.Write([]byte(errMsg))
	}
}
