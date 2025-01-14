// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instana_test

import (
	instana "github.com/instana/go-sensor"
)

const TestServiceName = "test_service"

var cleanupFn = func() {
	// reset instana collector
	instana.C = nil
}
