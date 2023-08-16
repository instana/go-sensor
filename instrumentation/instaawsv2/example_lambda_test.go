// (c) Copyright IBM Corp. 2023

package instaawsv2_test

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instaawsv2"
)

func Example_lambda() {
	mux := http.NewServeMux()

	tr := instana.InitCollector(instana.DefaultOptions())

	lambdaClient, err := getLambdaClientV2(tr)
	if err != nil {
		log.Fatal("Unable to create the lambda client. Check the config. Details: ", err.Error())
	}

	mux.HandleFunc("/testlambdainvoke",
		instana.TracingHandlerFunc(tr, "/testlambdainvoke",
			handleLambdaInvoke(lambdaClient)))

	log.Println("Starting service for testing lambda invocations ...")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func getLambdaClientV2(tr instana.TracerLogger) (*lambda.Client, error) {
	var client *lambda.Client

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("error while loading aws config: %s", err.Error())
	}

	instaawsv2.Instrument(tr, &cfg)
	client = lambda.NewFromConfig(cfg)

	return client, nil
}

func handleLambdaInvoke(client *lambda.Client) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		// Provide lambda function name here
		functionName := ""

		_, err := client.Invoke(request.Context(), &lambda.InvokeInput{
			FunctionName:   &functionName,
			InvocationType: types.InvocationTypeRequestResponse,
			Payload:        []byte("{}"),
		})

		var errMsg string
		if err != nil {
			errMsg = fmt.Sprintf("Unable to invoke the lambda: %s. Error: %s",
				functionName, err.Error())
		} else {
			errMsg = fmt.Sprintf("Successfully invoked the lambda: %s", functionName)
		}
		_, err = writer.Write([]byte(errMsg))
		if err != nil {
			fmt.Println("Error while writing the response: ", err.Error())
			return
		}
	}
}
