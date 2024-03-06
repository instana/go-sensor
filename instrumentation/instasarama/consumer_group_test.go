// (c) Copyright IBM Corp. 2024

//go:build go1.17
// +build go1.17

package instasarama_test

import (
	"context"
	"testing"

	"github.com/IBM/sarama"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instasarama"
	"github.com/stretchr/testify/assert"
)

func TestNewConsumerGroup_Consume(t *testing.T) {

	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

	config := NewTestConfig()
	config.ClientID = t.Name()
	config.Version = sarama.V2_0_0_0

	ctx, cancel := context.WithCancel(context.Background())
	h := &handler{t, cancel}

	broker0 := initMockBroker(t)
	defer broker0.Close()

	group, err := instasarama.NewConsumerGroup([]string{broker0.Addr()}, "my-group", config, sensor)
	defer func() { _ = group.Close() }()
	assert.NoError(t, err)

	topics := []string{"my-topic"}
	err = group.Consume(ctx, topics, h)
	assert.Error(t, err)
}

func TestNewConsumerGroupFromClient_Consume(t *testing.T) {

	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(instana.NewTracerWithEverything(&instana.Options{}, recorder))

	config := NewTestConfig()
	config.ClientID = t.Name()
	config.Version = sarama.V2_0_0_0

	ctx, cancel := context.WithCancel(context.Background())
	h := &handler{t, cancel}

	broker0 := initMockBroker(t)
	defer broker0.Close()

	client, err := sarama.NewClient([]string{broker0.Addr()}, config)
	assert.NoError(t, err)

	group, err := instasarama.NewConsumerGroupFromClient("my-group", client, sensor)
	defer func() { _ = group.Close() }()
	assert.NoError(t, err)

	topics := []string{"my-topic"}
	err = group.Consume(ctx, topics, h)
	assert.Error(t, err)
}

type handler struct {
	*testing.T
	cancel context.CancelFunc
}

func (h *handler) Setup(s sarama.ConsumerGroupSession) error   { return nil }
func (h *handler) Cleanup(s sarama.ConsumerGroupSession) error { return nil }
func (h *handler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		sess.MarkMessage(msg, "")
		h.Logf("consumed msg %v", msg)
		h.cancel()
		break
	}
	return nil
}

func NewTestConfig() *sarama.Config {
	config := sarama.NewConfig()
	config.Consumer.Retry.Backoff = 0
	config.Producer.Retry.Backoff = 0
	config.Version = sarama.MinVersion
	return config
}

func initMockBroker(t *testing.T) *sarama.MockBroker {
	topics := []string{"test.topic"}
	mockBroker := sarama.NewMockBroker(t, 0)

	mockBroker.SetHandlerByMap(map[string]sarama.MockResponse{
		"MetadataRequest": sarama.NewMockMetadataResponse(t).
			SetBroker(mockBroker.Addr(), mockBroker.BrokerID()).
			SetLeader(topics[0], 0, mockBroker.BrokerID()).
			SetController(mockBroker.BrokerID()),
	})
	return mockBroker
}
