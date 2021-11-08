// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

package instalogrus_test

import (
	"context"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instalogrus"
	"github.com/sirupsen/logrus"
)

// This example demostrates how to use instalogrus.NewHook() to instrument the global logrus logger
// with Instana. The instrumented logger instance will then send any ERROR and WARN log messages
// to Instana, associating them with the current operation span.
func Example_globalLogger() {
	sensor := instana.NewSensor("my-service")
	ctx := context.Background()

	// Add instalogrus hook to instrument the logger instance
	logrus.AddHook(instalogrus.NewHook(sensor))

	// Start and inject a span into context. Normally our instrumentation code does it for you.
	sp := sensor.Tracer().StartSpan("entry")
	defer sp.Finish()

	ctx = instana.ContextWithSpan(ctx, sp)

	logrus.
		// Make sure to add context to the log entry, so that the hook could corellate
		// this log record to current operation.
		WithContext(ctx).
		// Use your instrumented logger as usual
		WithFields(logrus.Fields{"data": "..."}).
		Error("something went wrong")
}

// This example demostrates how to use instalogrus.NewHook() to instrument a logrus.Logger instance
// with Instana. The instrumented logger instance will then send any ERROR and WARN log messages
// to Instana, associating them with the current operation span.
func Example_loggerInstance() {
	sensor := instana.NewSensor("my-service")
	ctx := context.Background()

	log := logrus.New()
	// Configure logger
	// ...

	// Add instalogrus hook to instrument the logger instance
	log.AddHook(instalogrus.NewHook(sensor))

	// Start and inject a span into context. Normally our instrumentation code does it for you.
	sp := sensor.Tracer().StartSpan("entry")
	defer sp.Finish()

	ctx = instana.ContextWithSpan(ctx, sp)

	log.
		// Make sure to add context to the log entry, so that the hook could corellate
		// this log record to current operation.
		WithContext(ctx).
		// Use your instrumented logger as usual
		WithFields(logrus.Fields{"data": "..."}).
		Error("something went wrong")
}
