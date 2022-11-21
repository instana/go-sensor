// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

package instalogrus_test

import (
	"context"
	"io/ioutil"
	"testing"
	"time"

	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/autoprofile"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instalogrus"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHook_Levels(t *testing.T) {
	sensor := instana.NewSensor("testing")
	defer instana.ShutdownSensor()
	h := instalogrus.NewHook(sensor)

	assert.ElementsMatch(t, []logrus.Level{logrus.ErrorLevel, logrus.WarnLevel}, h.Levels())
}

func TestNewHook_SendLogSpans(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder),
	)
	defer instana.ShutdownSensor()

	logger := logrus.New()
	logger.Level = logrus.DebugLevel
	logger.Out = ioutil.Discard
	logger.Formatter = &logrus.JSONFormatter{
		DisableTimestamp: true, // for easier comparison later
	}

	logger.AddHook(instalogrus.NewHook(sensor))

	examples := map[string]struct {
		Log             func(ctx context.Context)
		ExpectedMessage string
	}{
		"ERROR": {
			Log: func(ctx context.Context) {
				logger.WithContext(ctx).WithFields(logrus.Fields{"value": 42}).Error("log message")
			},
			ExpectedMessage: `{"level":"error", "msg":"log message", "value": 42}`,
		},
		"WARN": {
			Log: func(ctx context.Context) {
				logger.WithContext(ctx).WithFields(logrus.Fields{"value": 42}).Warn("log message")
			},
			ExpectedMessage: `{"level":"warning", "msg":"log message", "value": 42}`,
		},
	}

	for lvl, example := range examples {
		t.Run(lvl, func(t *testing.T) {
			parentSp := sensor.Tracer().StartSpan("testing")
			example.Log(instana.ContextWithSpan(context.Background(), parentSp))
			parentSp.Finish()

			spans := recorder.GetQueuedSpans()
			require.Len(t, spans, 2)

			logSp, sp := spans[0], spans[1]

			assert.Equal(t, sp.TraceID, logSp.TraceID)
			assert.Equal(t, sp.SpanID, logSp.ParentID)
			assert.Equal(t, "log.go", logSp.Name)

			assert.WithinDuration(t,
				time.Unix(int64(sp.Timestamp)/1000, int64(sp.Timestamp)%1000*1e6),
				time.Unix(int64(logSp.Timestamp)/1000, int64(logSp.Timestamp)%1000*1e6),
				// We relax this requirement, because of the rounding we make when calculating the duration and timestamps.
				(time.Duration(sp.Duration)+time.Nanosecond)*time.Millisecond,
			)

			require.IsType(t, instana.LogSpanData{}, logSp.Data)
			data := logSp.Data.(instana.LogSpanData)

			assert.JSONEq(t, example.ExpectedMessage, data.Tags.Message)

			assert.Equal(t, instana.LogSpanTags{
				Message: data.Tags.Message, // tested above
				Level:   lvl,
			}, data.Tags)
		})
	}
}

func TestNewHook_IgnoreLowLevels(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder),
	)
	defer instana.ShutdownSensor()

	logger := logrus.New()
	logger.Level = logrus.DebugLevel
	logger.Out = ioutil.Discard

	logger.AddHook(instalogrus.NewHook(sensor))

	examples := map[string]func(ctx context.Context){
		"INFO": func(ctx context.Context) {
			logger.WithContext(ctx).WithFields(logrus.Fields{"value": 42}).Info("log message")
		},
		"DEBUG": func(ctx context.Context) {
			logger.WithContext(ctx).WithFields(logrus.Fields{"value": 42}).Debug("log message")
		},
		"TRACE": func(ctx context.Context) {
			logger.WithContext(ctx).WithFields(logrus.Fields{"value": 42}).Trace("log message")
		},
	}

	for name, logFn := range examples {
		t.Run(name, func(t *testing.T) {
			parentSp := sensor.Tracer().StartSpan("testing")
			logFn(instana.ContextWithSpan(context.Background(), parentSp))
			parentSp.Finish()

			assert.Len(t, recorder.GetQueuedSpans(), 1)
		})
	}
}

func TestNewHook_NoContext(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(instana.DefaultOptions(), recorder),
	)
	defer instana.ShutdownSensor()

	logger := logrus.New()
	logger.Level = logrus.DebugLevel
	logger.Out = ioutil.Discard

	logger.AddHook(instalogrus.NewHook(sensor))

	logger.Error("log message")

	assert.Empty(t, recorder.GetQueuedSpans())
}

type alwaysReadyClient struct{}

func (alwaysReadyClient) Ready() bool                                       { return true }
func (alwaysReadyClient) SendMetrics(data acceptor.Metrics) error           { return nil }
func (alwaysReadyClient) SendEvent(event *instana.EventData) error          { return nil }
func (alwaysReadyClient) SendSpans(spans []instana.Span) error              { return nil }
func (alwaysReadyClient) SendProfiles(profiles []autoprofile.Profile) error { return nil }
func (alwaysReadyClient) Flush(context.Context) error                       { return nil }
