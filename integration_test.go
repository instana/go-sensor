// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

//go:build integration
// +build integration

package instana_test

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
)

type serverlessAgentPluginPayload struct {
	EntityID string
	Data     map[string]interface{}
}

type serverlessAgentRequest struct {
	Header http.Header
	Body   []byte
}

type serverlessAgent struct {
	Bundles []serverlessAgentRequest

	ln           net.Listener
	restoreEnvFn func()
}

func setupServerlessAgent() (*serverlessAgent, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize the serverless agent listener: %s", err)
	}

	srv := &serverlessAgent{
		ln:           ln,
		restoreEnvFn: restoreEnvVarFunc("INSTANA_ENDPOINT_URL"),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/bundle", srv.HandleBundle)

	go http.Serve(ln, mux)

	os.Setenv("INSTANA_ENDPOINT_URL", "http://"+ln.Addr().String())

	return srv, nil
}

func (srv *serverlessAgent) HandleBundle(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		log.Printf("ERROR: failed to read serverless agent spans request body: %s", err)
		body = nil
	}

	var root Root
	err = json.Unmarshal(body, &root)
	if err != nil {
		log.Printf("ERROR: failed to unmarshal serverless agent spans request body: %s", err.Error())
	} else {
		if len(root.Spans) > 0 && (root.Spans[0].Data.SDKCustom.Tags.ReturnError == "true" ||
			root.Spans[0].Data.Lambda.ReturnError == "true") {

			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	srv.Bundles = append(srv.Bundles, serverlessAgentRequest{
		Header: req.Header,
		Body:   body,
	})

	w.WriteHeader(http.StatusNoContent)
}

func (srv *serverlessAgent) Reset() {
	srv.Bundles = nil
}

func (srv *serverlessAgent) Teardown() {
	srv.restoreEnvFn()
	srv.ln.Close()
}

func restoreEnvVarFunc(key string) func() {
	if oldValue, ok := os.LookupEnv(key); ok {
		return func() { os.Setenv(key, oldValue) }
	}

	return func() { os.Unsetenv(key) }
}

type Data struct {
	SDKCustom struct {
		Tags struct {
			ReturnError string `json:"returnError"`
		} `json:"tags"`
	} `json:"sdk.custom"`
	Lambda LambdaData `json:"lambda"`
}

type LambdaData struct {
	ReturnError string `json:"error"`
}

type Span struct {
	Data Data `json:"data"`
}

type Root struct {
	Spans []Span `json:"spans"`
}
