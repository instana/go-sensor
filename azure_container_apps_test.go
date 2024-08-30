// (c) Copyright IBM Corp. 2024
//go:build integration
// +build integration

package instana_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	instana "github.com/instana/go-sensor"
	"github.com/stretchr/testify/require"
)

var agent *serverlessAgent

func TestMain(m *testing.M) {
	teardownAzureEnv := setupAzureContainerAppsEnv()
	defer teardownAzureEnv()

	teardownInstanaEnv := setupInstanaEnv()
	defer teardownInstanaEnv()

	var err error
	agent, err = setupServerlessAgent()
	if err != nil {
		log.Fatalf("failed to initialize serverless agent: %s", err)
	}

	os.Exit(m.Run())
}

// func TestAzureContainerApps_SendSpans(t *testing.T) {
// 	defer agent.Reset()

// 	tracer := instana.NewTracer()
// 	sensor := instana.NewSensorWithTracer(tracer)
// 	defer instana.ShutdownSensor()

// 	sp := sensor.Tracer().StartSpan("aca")
// 	sp.Finish()

// 	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
// 	defer cancel()

// 	require.NoError(t, tracer.Flush(ctx))
// 	require.Len(t, agent.Bundles, 1)

// 	var spans []map[string]json.RawMessage
// 	for _, bundle := range agent.Bundles {
// 		var payload struct {
// 			Spans []map[string]json.RawMessage `json:"spans"`
// 		}

// 		require.NoError(t, json.Unmarshal(bundle.Body, &payload), "%s", string(bundle.Body))
// 		spans = append(spans, payload.Spans...)
// 	}

// 	require.Len(t, spans, 1)
// 	assert.JSONEq(t, `{"hl": true, "cp": "azure", "e": "/subscriptions/testgh05-3f0d-4bf9-8f53-209408003632/resourceGroups/testresourcegroup/providers/Microsoft.App/containerapps/azureapp"}`, string(spans[0]["f"]))
// }

func TestAzureAgent_SendSpans_Error(t *testing.T) {
	defer agent.Reset()

	tracer := instana.NewTracer()
	sensor := instana.NewSensorWithTracer(tracer)
	defer instana.ShutdownSensor()

	sp := sensor.Tracer().StartSpan("azf")
	sp.SetTag("returnError", "true")
	sp.Finish()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	require.NoError(t, tracer.Flush(ctx))
	require.Len(t, agent.Bundles, 0)
}

func setupAzureContainerAppsEnv() func() {
	var teardownFuncs []func()

	teardownFuncs = append(teardownFuncs, restoreEnvVarFunc("AZURE_SUBSCRIPTION_ID"))
	os.Setenv("AZURE_SUBSCRIPTION_ID", "testgh05-3f0d-4bf9-8f53-209408003632")

	teardownFuncs = append(teardownFuncs, restoreEnvVarFunc("AZURE_RESOURCE_GROUP"))
	os.Setenv("AZURE_RESOURCE_GROUP", "testresourcegroup")

	teardownFuncs = append(teardownFuncs, restoreEnvVarFunc("CONTAINER_APP_NAME"))
	os.Setenv("CONTAINER_APP_NAME", "azureapp")

	teardownFuncs = append(teardownFuncs, restoreEnvVarFunc("CONTAINER_APP_HOSTNAME"))
	os.Setenv("CONTAINER_APP_HOSTNAME", "azure_app_host")

	return func() {
		for _, f := range teardownFuncs {
			f()
		}
	}
}

func setupInstanaEnv() func() {
	var teardownFuncs []func()

	teardownFuncs = append(teardownFuncs, restoreEnvVarFunc("INSTANA_AGENT_KEY"))
	os.Setenv("INSTANA_AGENT_KEY", "testkey1")

	teardownFuncs = append(teardownFuncs, restoreEnvVarFunc("INSTANA_ZONE"))
	os.Setenv("INSTANA_ZONE", "testzone")

	teardownFuncs = append(teardownFuncs, restoreEnvVarFunc("INSTANA_TAGS"))
	os.Setenv("INSTANA_TAGS", "key1=value1,key2")

	teardownFuncs = append(teardownFuncs, restoreEnvVarFunc("INSTANA_SECRETS"))
	os.Setenv("INSTANA_SECRETS", "contains-ignore-case:key,password,secret,classified")

	teardownFuncs = append(teardownFuncs, restoreEnvVarFunc("CLASSIFIED_DATA"))
	os.Setenv("CLASSIFIED_DATA", "classified")

	return func() {
		for _, f := range teardownFuncs {
			f()
		}
	}
}

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
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Printf("ERROR: failed to read serverless agent spans request body: %s", err)
		body = nil
	}

	var root Root
	var root1 struct {
		Spans []interface{} `json:"spans"`
	}
	fmt.Println(string(body))
	_ = json.Unmarshal(body, &root)
	err = json.Unmarshal(body, &root1)
	fmt.Println("\nHelloo")
	if err != nil {
		fmt.Println("\nHelloo1")
		log.Printf("ERROR: failed to unmarshal serverless agent spans request body: %s", err.Error())
	} else {
		fmt.Println("\nHelloo2")
		fmt.Println(len(root.Spans))

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
