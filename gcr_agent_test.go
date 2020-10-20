// +build gcr_integration

package instana_test

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	instana "github.com/instana/go-sensor"
)

var agent *serverlessAgent

func TestMain(m *testing.M) {
	teardownEnv := setupGCREnv()
	defer teardownEnv()

	defer restoreEnvVarFunc("INSTANA_AGENT_KEY")
	os.Setenv("INSTANA_AGENT_KEY", "testkey1")

	defer restoreEnvVarFunc("INSTANA_ZONE")
	os.Setenv("INSTANA_ZONE", "testzone")

	defer restoreEnvVarFunc("INSTANA_TAGS")
	os.Setenv("INSTANA_TAGS", "key1=value1,key2")

	defer restoreEnvVarFunc("INSTANA_SECRETS")
	os.Setenv("INSTANA_SECRETS", "contains-ignore-case:key,password,secret,classified")

	defer restoreEnvVarFunc("CLASSIFIED_DATA")
	os.Setenv("CLASSIFIED_DATA", "classified")

	var err error
	agent, err = setupServerlessAgent()
	if err != nil {
		log.Fatalf("failed to initialize serverless agent: %s", err)
	}

	instana.InitSensor(instana.DefaultOptions())

	os.Exit(m.Run())
}

func setupGCREnv() func() {
	var teardownFns []func()

	teardownFns = append(teardownFns, restoreEnvVarFunc("K_SERVICE"))
	os.Setenv("K_SERVICE", "test-service")

	teardownFns = append(teardownFns, restoreEnvVarFunc("K_CONFIGURATION"))
	os.Setenv("K_CONFIGURATION", "test-configuration")

	teardownFns = append(teardownFns, restoreEnvVarFunc("K_REVISION"))
	os.Setenv("K_REVISION", "test-revision")

	return func() {
		for _, fn := range teardownFns {
			fn()
		}
	}
}

func setupMetadataServer() func() {
	mux := http.NewServeMux()
	mux.HandleFunc("/computeMetadata", func(w http.ResponseWriter, req *http.Request) {
		http.ServeFile(w, req, "gcloud/computeMetadata.json")
	})

	srv := httptest.NewServer(mux)

	teardown := restoreEnvVarFunc("GOOGLE_CLOUD_RUN_METADATA_ENDPOINT")
	os.Setenv("GOOGLE_CLOUD_RUN_METADATA_ENDPOINT", "http://"+srv.URL)

	return func() {
		teardown()
		srv.Close()
	}
}
