// (c) Copyright IBM Corp. 2023

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instaazurefunction"
)

type InvokeResponse struct {
	Outputs     map[string]interface{}
	Logs        []string
	ReturnValue interface{}
}

func handleHttpTrigger(w http.ResponseWriter, r *http.Request) {
	u := r.URL
	message := fmt.Sprintf("Hello,\nThis HTTP triggered function executed successfully. Route handled: %s", u)

	resData := make(map[string]interface{})
	resData["body"] = message

	outputs := make(map[string]interface{})
	outputs["res"] = resData

	invokeResponse := InvokeResponse{outputs, nil, nil}

	responseJson, _ := json.Marshal(invokeResponse)

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJson)
}

func handleQueueTrigger(_ http.ResponseWriter, r *http.Request) {
	u := r.URL
	message := fmt.Sprintf("Hello,\nThis Queue triggered function executed successfully. Route handled: %s", u)

	resData := make(map[string]interface{})
	resData["body"] = message

	outputs := make(map[string]interface{})
	outputs["res"] = resData

	invokeResponse := InvokeResponse{outputs, nil, nil}

	responseJson, _ := json.Marshal(invokeResponse)

	fmt.Println("Response data:", string(responseJson))
}

func main() {
	listenAddr := ":8080"
	if val, ok := os.LookupEnv("FUNCTIONS_CUSTOMHANDLER_PORT"); ok {
		listenAddr = ":" + val
	}

	sensor := instana.NewSensor("mysensor")
	http.HandleFunc("/greetings", instaazurefunction.WrapFunctionHandler(sensor, handleHttpTrigger))
	http.HandleFunc("/queue", instaazurefunction.WrapFunctionHandler(sensor, handleQueueTrigger))
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}
