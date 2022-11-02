// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instasarama_test

import (
	"errors"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/Shopify/sarama"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instasarama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncProducer_SendMessage(t *testing.T) {
	headerFormats := []string{"" /* tests the default behavior */, "binary", "string", "both"}

	for _, headerFormat := range headerFormats {
		os.Setenv(instasarama.KafkaHeaderEnvVarKey, headerFormat)
		recorder := instana.NewTestRecorder()
		sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

		parent := sensor.Tracer().StartSpan("test-span")

		config := sarama.NewConfig()
		config.Version = sarama.V0_11_0_0
		config.Producer.Return.Successes = true

		p := &testSyncProducer{}
		wrapped := instasarama.WrapSyncProducer(p, config, sensor)

		_, _, err := wrapped.SendMessage(
			instasarama.ProducerMessageWithSpan(&sarama.ProducerMessage{Topic: "test-topic"}, parent),
		)
		require.NoError(t, err)

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

		require.Len(t, p.Messages, 1)

		if headerFormat == "both" || headerFormat == "binary" || headerFormat == "" /* -> default, currently both */ {
			assert.Contains(t, p.Messages[0].Headers, sarama.RecordHeader{
				Key:   []byte("X_INSTANA_C"),
				Value: instasarama.PackTraceContextHeader(cSpan.TraceID, cSpan.SpanID),
			})
			assert.Contains(t, p.Messages[0].Headers, sarama.RecordHeader{
				Key:   []byte("X_INSTANA_L"),
				Value: instasarama.PackTraceLevelHeader("1"),
			})
		}

		if headerFormat == "both" || headerFormat == "string" || headerFormat == "" /* -> default, currently both */ {
			assert.Contains(t, p.Messages[0].Headers, sarama.RecordHeader{
				Key:   []byte(instasarama.FieldLS),
				Value: []byte("1"),
			})
			assert.Contains(t, p.Messages[0].Headers, sarama.RecordHeader{
				Key:   []byte(instasarama.FieldT),
				Value: []byte("0000000000000000" + cSpan.TraceID),
			})
			assert.Contains(t, p.Messages[0].Headers, sarama.RecordHeader{
				Key:   []byte(instasarama.FieldS),
				Value: []byte(cSpan.SpanID),
			})
		}

		assert.Equal(t, pSpan.TraceID, cSpan.TraceID)
		assert.Equal(t, pSpan.SpanID, cSpan.ParentID)
		assert.NotEqual(t, pSpan.SpanID, cSpan.SpanID)
		os.Unsetenv(instasarama.KafkaHeaderEnvVarKey)
	}
}

func TestSyncProducer_SendMessage_NoTraceContext(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

	config := sarama.NewConfig()
	config.Version = sarama.V0_11_0_0
	config.Producer.Return.Successes = true

	p := &testSyncProducer{}
	wrapped := instasarama.WrapSyncProducer(p, config, sensor)

	_, _, err := wrapped.SendMessage(&sarama.ProducerMessage{
		Topic: "test-topic",
	})
	require.NoError(t, err)

	spans := recorder.GetQueuedSpans()
	assert.Empty(t, spans)

	require.Len(t, p.Messages, 1)
	assert.Empty(t, p.Messages[0].Headers)
}

func TestSyncProducer_SendMessage_Error(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

	config := sarama.NewConfig()
	config.Version = sarama.V0_11_0_0
	config.Producer.Return.Successes = true

	p := &testSyncProducer{
		Error: errors.New("something went wrong"),
	}
	wrapped := instasarama.WrapSyncProducer(p, config, sensor)

	parent := sensor.Tracer().StartSpan("test-span")
	_, _, err := wrapped.SendMessage(
		instasarama.ProducerMessageWithSpan(&sarama.ProducerMessage{Topic: "test-topic"}, parent),
	)
	parent.Finish()

	assert.Error(t, err)

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 3)

	span, err := extractAgentSpan(spans[0])
	require.NoError(t, err)

	assert.Equal(t, 1, span.Ec)
	assert.EqualValues(t, instana.ExitSpanKind, span.Kind)

	assert.Equal(t, agentKafkaSpanData{
		Service: "test-topic",
		Access:  "send",
	}, span.Data.Kafka)
}

func TestSyncProducer_SendMessages_SameTraceContext(t *testing.T) {
	for _, headerFormat := range headerFormats {
		os.Setenv(instasarama.KafkaHeaderEnvVarKey, headerFormat)

		recorder := instana.NewTestRecorder()
		sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

		config := sarama.NewConfig()
		config.Version = sarama.V0_11_0_0
		config.Producer.Return.Successes = true

		p := &testSyncProducer{}
		wrapped := instasarama.WrapSyncProducer(p, config, sensor)

		parent := sensor.Tracer().StartSpan("test-span")
		require.NoError(t, wrapped.SendMessages([]*sarama.ProducerMessage{
			instasarama.ProducerMessageWithSpan(&sarama.ProducerMessage{Topic: "test-topic-1"}, parent),
			instasarama.ProducerMessageWithSpan(&sarama.ProducerMessage{Topic: "test-topic-2"}, parent),
		}))
		parent.Finish()

		spans := recorder.GetQueuedSpans()
		require.Len(t, spans, 2)

		pSpan, err := extractAgentSpan(spans[1])
		require.NoError(t, err)

		cSpan, err := extractAgentSpan(spans[0])
		require.NoError(t, err)

		assert.Equal(t, 0, cSpan.Ec)
		assert.EqualValues(t, instana.ExitSpanKind, cSpan.Kind)
		assert.Equal(t, 2, cSpan.Batch.Size)

		// sort comma-separated list of topics for comparison
		topics := strings.Split(cSpan.Data.Kafka.Service, ",")
		sort.Strings(topics)
		cSpan.Data.Kafka.Service = strings.Join(topics, ",")

		assert.Equal(t, agentKafkaSpanData{
			Service: "test-topic-1,test-topic-2",
			Access:  "send",
		}, cSpan.Data.Kafka)

		require.Len(t, p.Messages, 2)
		for _, msg := range p.Messages {
			if headerFormat == "both" || headerFormat == "binary" || headerFormat == "" /* -> default, currently both */ {
				assert.Contains(t, msg.Headers, sarama.RecordHeader{
					Key:   []byte("X_INSTANA_C"),
					Value: instasarama.PackTraceContextHeader(cSpan.TraceID, cSpan.SpanID),
				})
				assert.Contains(t, msg.Headers, sarama.RecordHeader{
					Key:   []byte("X_INSTANA_L"),
					Value: instasarama.PackTraceLevelHeader("1"),
				})
			}

			if headerFormat == "both" || headerFormat == "string" || headerFormat == "" /* -> default, currently both */ {
				assert.Contains(t, msg.Headers, sarama.RecordHeader{
					Key:   []byte(instasarama.FieldLS),
					Value: []byte("1"),
				})
				assert.Contains(t, msg.Headers, sarama.RecordHeader{
					Key:   []byte(instasarama.FieldT),
					Value: []byte("0000000000000000" + cSpan.TraceID),
				})
				assert.Contains(t, msg.Headers, sarama.RecordHeader{
					Key:   []byte(instasarama.FieldS),
					Value: []byte(cSpan.SpanID),
				})
			}
		}

		assert.Equal(t, pSpan.TraceID, cSpan.TraceID)
		assert.Equal(t, pSpan.SpanID, cSpan.ParentID)
		assert.NotEqual(t, pSpan.SpanID, cSpan.SpanID)
		os.Unsetenv(instasarama.KafkaHeaderEnvVarKey)
	}
}

func TestSyncProducer_SendMessages_DifferentTraceContext(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

	parentOne := sensor.Tracer().StartSpan("test-span")
	defer parentOne.Finish()

	parentTwo := sensor.Tracer().StartSpan("test-span")
	defer parentTwo.Finish()

	examples := map[string][]*sarama.ProducerMessage{
		"different parent spans": {
			instasarama.ProducerMessageWithSpan(&sarama.ProducerMessage{Topic: "test-topic-1"}, parentOne),
			instasarama.ProducerMessageWithSpan(&sarama.ProducerMessage{Topic: "test-topic-2"}, parentTwo),
		},
		"with message without trace context": {
			instasarama.ProducerMessageWithSpan(&sarama.ProducerMessage{Topic: "test-topic-1"}, parentOne),
			&sarama.ProducerMessage{Topic: "test-topic-3"},
		},
	}

	for name, messages := range examples {
		t.Run(name, func(t *testing.T) {
			config := sarama.NewConfig()
			config.Version = sarama.V0_11_0_0
			config.Producer.Return.Successes = true

			p := &testSyncProducer{}
			wrapped := instasarama.WrapSyncProducer(p, config, sensor)

			require.NoError(t, wrapped.SendMessages(messages))

			assert.Empty(t, recorder.GetQueuedSpans())
			assert.ElementsMatch(t, messages, p.Messages)
		})
	}
}

func TestSyncProducer_SendMessages_NoTraceContext(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

	config := sarama.NewConfig()
	config.Version = sarama.V0_11_0_0
	config.Producer.Return.Successes = true

	p := &testSyncProducer{}
	wrapped := instasarama.WrapSyncProducer(p, config, sensor)

	require.NoError(t, wrapped.SendMessages([]*sarama.ProducerMessage{
		{Topic: "test-topic-1"},
		{Topic: "test-topic-2"},
	}))

	spans := recorder.GetQueuedSpans()
	assert.Empty(t, spans)

	require.Len(t, p.Messages, 2)
	assert.Empty(t, p.Messages[0].Headers)
	assert.Empty(t, p.Messages[1].Headers)
}

func TestSyncProducer_SendMessages_Error(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

	config := sarama.NewConfig()
	config.Version = sarama.V0_11_0_0
	config.Producer.Return.Successes = true

	p := &testSyncProducer{
		Error: errors.New("something went wrong"),
	}
	wrapped := instasarama.WrapSyncProducer(p, config, sensor)

	parent := sensor.Tracer().StartSpan("test-span")
	assert.Error(t, wrapped.SendMessages([]*sarama.ProducerMessage{
		instasarama.ProducerMessageWithSpan(&sarama.ProducerMessage{Topic: "test-topic-1"}, parent),
		instasarama.ProducerMessageWithSpan(&sarama.ProducerMessage{Topic: "test-topic-2"}, parent),
	}))
	parent.Finish()

	spans := recorder.GetQueuedSpans()
	require.Len(t, spans, 3)

	span, err := extractAgentSpan(spans[0])
	require.NoError(t, err)

	assert.Equal(t, 1, span.Ec)
	assert.EqualValues(t, instana.ExitSpanKind, span.Kind)
	assert.Equal(t, 2, span.Batch.Size)

	// sort comma-separated list of topics for comparison
	topics := strings.Split(span.Data.Kafka.Service, ",")
	sort.Strings(topics)
	span.Data.Kafka.Service = strings.Join(topics, ",")

	assert.Equal(t, agentKafkaSpanData{
		Service: "test-topic-1,test-topic-2",
		Access:  "send",
	}, span.Data.Kafka)
}

func TestSyncProducer_Close(t *testing.T) {
	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

	config := sarama.NewConfig()
	config.Version = sarama.V0_11_0_0
	config.Producer.Return.Successes = true

	p := &testSyncProducer{}
	wrapped := instasarama.WrapSyncProducer(p, config, sensor)
	wrapped.Close()

	assert.True(t, p.Closed)
}

type testSyncProducer struct {
	Error    error
	Messages []*sarama.ProducerMessage
	Closed   bool
}

func (p *testSyncProducer) SendMessage(msg *sarama.ProducerMessage) (partition int32, offset int64, err error) {
	p.Messages = append(p.Messages, msg)

	return 0, 0, p.Error
}

func (p *testSyncProducer) SendMessages(msgs []*sarama.ProducerMessage) error {
	p.Messages = append(p.Messages, msgs...)

	return p.Error
}

func (p *testSyncProducer) Close() error {
	p.Closed = true
	return p.Error
}
