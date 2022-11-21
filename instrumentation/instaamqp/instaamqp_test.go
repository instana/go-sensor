// (c) Copyright IBM Corp. 2022

package instaamqp_test

import (
	"context"
	"errors"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/autoprofile"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instaamqp"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type amqpChannelMock struct {
	ch           chan amqp.Delivery
	publishError bool
	consumeError bool
}

func (am amqpChannelMock) Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	d := amqp.Delivery{}
	d.Body = msg.Body
	d.Headers = msg.Headers

	if am.publishError {
		return errors.New("error publishing message")
	}
	am.ch <- d
	return nil
}

func (am amqpChannelMock) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {

	if am.consumeError {
		return nil, errors.New("error consuming message")
	}
	return am.ch, nil
}

func newAmqpChannelMock() amqpChannelMock {
	chMock := &amqpChannelMock{}
	chMock.ch = make(chan amqp.Delivery)
	return *chMock
}

func TestClient(t *testing.T) {
	wg := sync.WaitGroup{}

	testCases := []struct {
		name         string
		hasHeaders   bool
		publishError bool
		consumeError bool
	}{
		{
			name: "Using empty headers",
		},
		{
			name:       "Using user headers",
			hasHeaders: true,
		},
		{
			name:         "With error on publish",
			publishError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			wg.Add(1)

			url := "amqp://user:password@some_host:9999"
			chMock := newAmqpChannelMock()
			chMock.publishError = tc.publishError
			chMock.consumeError = tc.consumeError
			recorder := instana.NewTestRecorder()
			sensor := instana.NewSensorWithTracer(
				instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder),
			)

			instaCh := instaamqp.WrapChannel(sensor, chMock, url)

			// Start waiting for messages to consume
			go func(s string) {
				defer instana.ShutdownSensor()
				ch, _ := instaCh.Consume("queue", "consumer", true, false, false, false, nil)

				for range ch {
					// Give it some time to assure that all spans are added
					time.Sleep(time.Second)

					spans := recorder.GetQueuedSpans()
					require.Len(t, spans, 3, "Expects sdk entry span, publish exit span and consume entry span")

					spanMap := getSpanMap(spans)

					sdkEntrySp := spanMap[string(instana.SDKSpanType)+"_entry"]
					publishSp := spanMap[string(instana.RabbitMQSpanType)+"_exit"]
					consumeSp := spanMap[string(instana.RabbitMQSpanType)+"_entry"]

					checkCorrelation(t, sdkEntrySp, publishSp, consumeSp)

					if publisherData, ok := publishSp.Data.(instana.RabbitMQSpanData); ok {
						checkPublisher(t, publisherData, false)
					}

					if consumerData, ok := consumeSp.Data.(instana.RabbitMQSpanData); ok {
						assert.Equal(t, "amqp://some_host:9999", consumerData.Tags.Address)
						assert.Equal(t, "consume", consumerData.Tags.Sort)
					}
					wg.Done()
				}
			}(tc.name)

			// Publish the message

			entrySpan := sensor.Tracer().StartSpan("testing")
			ext.SpanKind.Set(entrySpan, ext.SpanKindRPCServerEnum)
			defer entrySpan.Finish()

			msg := amqp.Publishing{
				Body: []byte("Test message"),
			}

			if tc.hasHeaders {
				msg.Headers = amqp.Table{
					"some-key": "some-value",
				}
			}

			go func(entrySpan opentracing.Span) {
				err := instaCh.Publish(entrySpan, "my-exchange", "key", false, false, msg)

				if err != nil {
					spans := recorder.GetQueuedSpans()

					require.Len(t, spans, 3, "Expects SDK entry span, failed exit publish span and log error span")

					spanMap := getSpanMap(spans)

					sdkEntrySp := spanMap[string(instana.SDKSpanType)+"_entry"]
					publishSp := spanMap[string(instana.RabbitMQSpanType)+"_exit"]
					errorSp := spanMap[string(instana.LogSpanType)+"_entry"]

					checkCorrelation(t, sdkEntrySp, publishSp, errorSp)

					if publisherData, ok := publishSp.Data.(instana.RabbitMQSpanData); ok {
						checkPublisher(t, publisherData, tc.publishError)
					}

					wg.Done()
				}
			}(entrySpan)
		})
	}
	wg.Wait()
}

func checkCorrelation(t *testing.T, entrySp1, exitSp1, entrySp2 instana.Span) {
	assert.Equal(t, entrySp1.TraceID, exitSp1.TraceID)
	assert.Equal(t, exitSp1.TraceID, entrySp2.TraceID)
	assert.Equal(t, entrySp1.SpanID, exitSp1.ParentID)
	assert.Equal(t, exitSp1.SpanID, entrySp2.ParentID)
}

func checkPublisher(t *testing.T, publisherData instana.RabbitMQSpanData, hasError bool) {
	assert.Equal(t, "amqp://some_host:9999", publisherData.Tags.Address)
	assert.Equal(t, "publish", publisherData.Tags.Sort)
	assert.Equal(t, "my-exchange", publisherData.Tags.Exchange)
	if hasError {
		assert.Equal(t, "Error publishing message", publisherData.Tags.Error)
	}
}

// Returns a map of spans out of the span list.
// Sometimes, the spans provided by the recorder may not be in the expected order, so we remap them into a map, using
// the span type + span kind as a key map, allowing us to use spans in the desired order.
// The key is the span type + underscore + span kind. eg: "sdk_entry", "rabbitmq_exit" and so on.
func getSpanMap(spans []instana.Span) map[string]instana.Span {
	spanMap := make(map[string]instana.Span)

	for _, sp := range spans {
		suffix := "_" + instana.SpanKind(sp.Kind).String()

		switch sp.Data.(type) {
		case instana.LogSpanData:
			spanMap[string(instana.LogSpanType)+suffix] = sp
		case instana.RabbitMQSpanData:
			spanMap[string(instana.RabbitMQSpanType)+"_"+suffix] = sp
		case instana.SDKSpanData:
			spanMap[string(instana.SDKSpanType)+"_"+suffix] = sp
		default:
			log.Printf("Unexpected span data: %v", sp)
		}
	}

	return spanMap
}

type alwaysReadyClient struct{}

func (alwaysReadyClient) Ready() bool                                       { return true }
func (alwaysReadyClient) SendMetrics(data acceptor.Metrics) error           { return nil }
func (alwaysReadyClient) SendEvent(event *instana.EventData) error          { return nil }
func (alwaysReadyClient) SendSpans(spans []instana.Span) error              { return nil }
func (alwaysReadyClient) SendProfiles(profiles []autoprofile.Profile) error { return nil }
func (alwaysReadyClient) Flush(context.Context) error                       { return nil }
