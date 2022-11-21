// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instasarama_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/autoprofile"
	"github.com/instana/go-sensor/instrumentation/instasarama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAsyncProducer_Input(t *testing.T) {
	headerFormats := []string{"" /* tests the default behavior */, "binary", "string", "both"}

	for _, headerFormat := range headerFormats {
		os.Setenv(instasarama.KafkaHeaderEnvVarKey, headerFormat)

		recorder := instana.NewTestRecorder()
		sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder))

		parent := sensor.Tracer().StartSpan("test-span")
		msg := instasarama.ProducerMessageWithSpan(&sarama.ProducerMessage{Topic: "test-topic"}, parent)

		ap := newTestAsyncProducer(nil)
		defer ap.Teardown()

		conf := sarama.NewConfig()
		conf.Version = sarama.V0_11_0_0

		wrapped := instasarama.WrapAsyncProducer(ap, conf, sensor)
		wrapped.Input() <- msg

		var published *sarama.ProducerMessage
	channelSelect:
		select {
		case published = <-ap.input:
			break channelSelect
		case <-time.After(1 * time.Second):
			t.Fatalf("publishing via async producer timed out after 1s")
		}

		parent.Finish()

		spans := recorder.GetQueuedSpans()
		require.Len(t, spans, 2)

		pSpan, err := extractAgentSpan(spans[1])
		require.NoError(t, err)

		cSpan, err := extractAgentSpan(spans[0])
		require.NoError(t, err)

		assert.Equal(t, 0, cSpan.Ec)
		assert.EqualValues(t, instana.ExitSpanKind, cSpan.Kind)

		assert.Equal(t, agentKafkaSpanData{
			Service: "test-topic",
			Access:  "send",
		}, cSpan.Data.Kafka)

		if headerFormat == "both" || headerFormat == "binary" || headerFormat == "" /* -> default, currently both */ {
			assert.Contains(t, published.Headers, sarama.RecordHeader{
				Key:   []byte("X_INSTANA_C"),
				Value: instasarama.PackTraceContextHeader(cSpan.TraceID, cSpan.SpanID),
			})
			assert.Contains(t, published.Headers, sarama.RecordHeader{
				Key:   []byte("X_INSTANA_L"),
				Value: instasarama.PackTraceLevelHeader("1"),
			})
		}

		if headerFormat == "both" || headerFormat == "string" || headerFormat == "" /* -> default, currently both */ {
			assert.Contains(t, published.Headers, sarama.RecordHeader{
				Key:   []byte("X_INSTANA_T"),
				Value: []byte("0000000000000000" + cSpan.TraceID),
			})
			assert.Contains(t, published.Headers, sarama.RecordHeader{
				Key:   []byte("X_INSTANA_S"),
				Value: []byte(cSpan.SpanID),
			})
			assert.Contains(t, published.Headers, sarama.RecordHeader{
				Key:   []byte("X_INSTANA_L_S"),
				Value: []byte("1"),
			})
		}

		assert.Equal(t, pSpan.TraceID, cSpan.TraceID)
		assert.Equal(t, pSpan.SpanID, cSpan.ParentID)
		assert.NotEqual(t, pSpan.SpanID, cSpan.SpanID)

		os.Unsetenv(instasarama.KafkaHeaderEnvVarKey)
	}
}

func TestAsyncProducer_Input_WithAwaitResult_Success(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder))

	parent := sensor.Tracer().StartSpan("test-span")
	msg := instasarama.ProducerMessageWithSpan(&sarama.ProducerMessage{Topic: "test-topic"}, parent)

	ap := newTestAsyncProducer(nil)
	defer ap.Teardown()

	conf := sarama.NewConfig()
	conf.Version = sarama.V0_11_0_0
	conf.Producer.Return.Successes = true
	conf.Producer.Return.Errors = true

	wrapped := instasarama.WrapAsyncProducer(ap, conf, sensor)
	wrapped.Input() <- msg

	var published *sarama.ProducerMessage
	select {
	case published = <-ap.input:
		break
	case <-time.After(1 * time.Second):
		t.Fatalf("publishing via async producer timed out after 1s")
	}

	parent.Finish()

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	pSpan, err := extractAgentSpan(spans[0])
	require.NoError(t, err)

	// send error for another message
	ap.errors <- &sarama.ProducerError{
		Msg: &sarama.ProducerMessage{Topic: "another-topic"},
		Err: errors.New("something went wrong"),
	}
	<-wrapped.Errors()
	require.Empty(t, recorder.GetQueuedSpans())

	// send success for another message
	ap.successes <- &sarama.ProducerMessage{Topic: "another-topic"}
	<-wrapped.Successes()
	require.Empty(t, recorder.GetQueuedSpans())

	// send expected success
	ap.successes <- msg
	<-wrapped.Successes()

	spans = recorder.GetQueuedSpans()

	require.Eventually(t, func() bool {
		spans = append(spans, recorder.GetQueuedSpans()...)
		return len(spans) == 1
	}, time.Millisecond*200, time.Millisecond*50)

	cSpan, err := extractAgentSpan(spans[0])
	require.NoError(t, err)

	assert.Equal(t, 0, cSpan.Ec)
	assert.EqualValues(t, instana.ExitSpanKind, cSpan.Kind)

	assert.Equal(t, agentKafkaSpanData{
		Service: "test-topic",
		Access:  "send",
	}, cSpan.Data.Kafka)

	assert.Contains(t, published.Headers, sarama.RecordHeader{
		Key:   []byte("X_INSTANA_C"),
		Value: instasarama.PackTraceContextHeader(cSpan.TraceID, cSpan.SpanID),
	})
	assert.Contains(t, published.Headers, sarama.RecordHeader{
		Key:   []byte("X_INSTANA_L"),
		Value: instasarama.PackTraceLevelHeader("1"),
	})

	assert.Equal(t, pSpan.TraceID, cSpan.TraceID)
	assert.Equal(t, pSpan.SpanID, cSpan.ParentID)
	assert.NotEqual(t, pSpan.SpanID, cSpan.SpanID)
}

func TestAsyncProducer_Input_WithAwaitResult_Error(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder))
	defer instana.ShutdownSensor()

	parent := sensor.Tracer().StartSpan("test-span")
	msg := instasarama.ProducerMessageWithSpan(&sarama.ProducerMessage{Topic: "test-topic"}, parent)

	ap := newTestAsyncProducer(nil)
	defer ap.Teardown()

	conf := sarama.NewConfig()
	conf.Version = sarama.V0_11_0_0
	conf.Producer.Return.Successes = true
	conf.Producer.Return.Errors = true

	wrapped := instasarama.WrapAsyncProducer(ap, conf, sensor)
	wrapped.Input() <- msg

	var published *sarama.ProducerMessage
	select {
	case published = <-ap.input:
		break
	case <-time.After(1 * time.Second):
		t.Fatalf("publishing via async producer timed out after 1s")
	}

	parent.Finish()

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 1)

	pSpan, err := extractAgentSpan(spans[0])
	require.NoError(t, err)

	// send error for another message
	ap.errors <- &sarama.ProducerError{
		Msg: &sarama.ProducerMessage{Topic: "another-topic"},
		Err: errors.New("something went wrong"),
	}
	<-wrapped.Errors()
	require.Empty(t, recorder.GetQueuedSpans())

	// send success for another message
	ap.successes <- &sarama.ProducerMessage{Topic: "another-topic"}
	<-wrapped.Successes()
	require.Empty(t, recorder.GetQueuedSpans())

	// send expected error
	ap.errors <- &sarama.ProducerError{
		Msg: msg,
		Err: errors.New("something went wrong"),
	}
	<-wrapped.Errors()

	spans = recorder.GetQueuedSpans()

	require.Eventually(t, func() bool {
		spans = append(spans, recorder.GetQueuedSpans()...)
		return len(spans) == 2
	}, time.Millisecond*200, time.Millisecond*5)

	cSpan, err := extractAgentSpan(spans[0])
	require.NoError(t, err)

	assert.Equal(t, 1, cSpan.Ec)
	assert.EqualValues(t, instana.ExitSpanKind, cSpan.Kind)

	assert.Equal(t, agentKafkaSpanData{
		Service: "test-topic",
		Access:  "send",
	}, cSpan.Data.Kafka)

	assert.Contains(t, published.Headers, sarama.RecordHeader{
		Key:   []byte("X_INSTANA_C"),
		Value: instasarama.PackTraceContextHeader(cSpan.TraceID, cSpan.SpanID),
	})
	assert.Contains(t, published.Headers, sarama.RecordHeader{
		Key:   []byte("X_INSTANA_L"),
		Value: instasarama.PackTraceLevelHeader("1"),
	})

	assert.Equal(t, pSpan.TraceID, cSpan.TraceID)
	assert.Equal(t, pSpan.SpanID, cSpan.ParentID)
	assert.NotEqual(t, pSpan.SpanID, cSpan.SpanID)
}

func TestAsyncProducer_Input_NoTraceContext(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder))

	msg := &sarama.ProducerMessage{
		Topic: "topic-1",
	}

	ap := newTestAsyncProducer(nil)
	defer ap.Teardown()

	wrapped := instasarama.WrapAsyncProducer(ap, sarama.NewConfig(), sensor)
	wrapped.Input() <- msg

	select {
	case published := <-ap.input:
		assert.Equal(t, msg, published)
	case <-time.After(1 * time.Second):
		t.Fatalf("publishing via async producer timed out after 1s")
	}

	assert.Empty(t, recorder.GetQueuedSpans())
}

func TestAsyncProducer_Successes(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder))

	msg := &sarama.ProducerMessage{
		Topic: "topic-1",
	}

	ap := newTestAsyncProducer(nil)
	defer ap.Teardown()

	ap.successes <- msg

	wrapped := instasarama.WrapAsyncProducer(ap, sarama.NewConfig(), sensor)

	select {
	case received := <-wrapped.Successes():
		assert.Equal(t, msg, received)
	case <-time.After(1 * time.Second):
		t.Fatalf("reading a success message from async producer timed out after 1s")
	}
}

func TestAsyncProducer_Errors(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder))

	msg := &sarama.ProducerError{
		Err: errors.New("something went wrong"),
		Msg: &sarama.ProducerMessage{Topic: "topic-1"},
	}

	ap := newTestAsyncProducer(nil)
	defer ap.Teardown()

	ap.errors <- msg

	wrapped := instasarama.WrapAsyncProducer(ap, sarama.NewConfig(), sensor)

	select {
	case received := <-wrapped.Errors():
		assert.Equal(t, msg, received)
	case <-time.After(1 * time.Second):
		t.Fatalf("reading an error message from async producer timed out after 1s")
	}
}

type testAsyncProducer struct {
	Error  error
	Closed bool
	Async  bool

	input     chan *sarama.ProducerMessage
	successes chan *sarama.ProducerMessage
	errors    chan *sarama.ProducerError
}

func newTestAsyncProducer(returnedErr error) *testAsyncProducer {
	return &testAsyncProducer{
		Error:     returnedErr,
		input:     make(chan *sarama.ProducerMessage, 1),
		successes: make(chan *sarama.ProducerMessage, 1),
		errors:    make(chan *sarama.ProducerError, 1),
	}
}

func (p *testAsyncProducer) AsyncClose() {
	p.Closed = true
	p.Async = true
}

func (p *testAsyncProducer) Close() error {
	p.Closed = true
	return p.Error
}

func (p *testAsyncProducer) Input() chan<- *sarama.ProducerMessage     { return p.input }
func (p *testAsyncProducer) Successes() <-chan *sarama.ProducerMessage { return p.successes }
func (p *testAsyncProducer) Errors() <-chan *sarama.ProducerError      { return p.errors }

func (p *testAsyncProducer) Teardown() {
	close(p.input)
	close(p.successes)
	close(p.errors)
}

type alwaysReadyClient struct{}

func (alwaysReadyClient) Ready() bool                                       { return true }
func (alwaysReadyClient) SendMetrics(data acceptor.Metrics) error           { return nil }
func (alwaysReadyClient) SendEvent(event *instana.EventData) error          { return nil }
func (alwaysReadyClient) SendSpans(spans []instana.Span) error              { return nil }
func (alwaysReadyClient) SendProfiles(profiles []autoprofile.Profile) error { return nil }
func (alwaysReadyClient) Flush(context.Context) error                       { return nil }
