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
	})

	os.Exit(m.Run())
}
