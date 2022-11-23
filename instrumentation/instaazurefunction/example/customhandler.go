// (c) Copyright IBM Corp. 2022
// (c) Copyright Instana Inc. 2022

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	instana "github.com/instana/go-sensor"
	azf "github.com/instana/go-sensor/instrumentation/instaazurefunction"
)

func handleGreetings(w http.ResponseWriter, r *http.Request) {
	message := "This HTTP triggered function executed successfully. Route handled: /api/greetings\n"
	_, err := fmt.Fprint(w, message)
	if err != nil {
		return
	}
}

func main() {
	listenAddr := ":8080"
	if val, ok := os.LookupEnv("FUNCTIONS_CUSTOMHANDLER_PORT"); ok {
		listenAddr = ":" + val
	}
	sensor := instana.NewSensor("mysensor")
	http.HandleFunc("/api/greetings", azf.WrapFunctionHandler(sensor, handleGreetings))
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}
