// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instana_test

import (
	"os"
	"testing"

	instana "github.com/instana/go-sensor"
)

const TestServiceName = "test_service"

func TestMain(m *testing.M) {
	instana.InitSensor(&instana.Options{
		Service: TestServiceName,
		Tracer: instana.TracerOptions{
			CollectableHTTPHeaders: []string{"x-custom-header-1", "x-custom-header-2"},
		},
	})

	os.Exit(m.Run())
}
