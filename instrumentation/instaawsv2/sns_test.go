// (c) Copyright IBM Corp. 2023

package instaawsv2_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	instana "github.com/instana/go-sensor"
	instaawsv2 "github.com/instana/go-sensor/instrumentation/instaawsv2"
	"github.com/stretchr/testify/assert"
)

func TestPublishMessage(t *testing.T) {
	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	ctx := context.Background()
	ps := c.StartSpan("aws-parent-span")
	ctx = instana.ContextWithSpan(ctx, ps)

	cfg, err := config.LoadDefaultConfig(ctx)
	assert.NoError(t, err, "Error while configuring aws")

	cfg = applyTestingChanges(cfg)

	instaawsv2.Instrument(c, &cfg)

	snsClient := sns.NewFromConfig(cfg)

	snsMsg := "this is is a test message"
	snsTopicArn := "test-topic-arn"
	_, err = snsClient.Publish(ctx, &sns.PublishInput{
		Message:  &snsMsg,
		TopicArn: &snsTopicArn,
	})
	assert.NoError(t, err, "Error while publishing the sns message")

	ps.Finish()

	recorderSpans := recorder.GetQueuedSpans()
	assert.Equal(t, 2, len(recorderSpans))

	snsSpan := recorderSpans[0]
	assert.IsType(t, instana.AWSSNSSpanData{}, snsSpan.Data)

	data := snsSpan.Data.(instana.AWSSNSSpanData)
	assert.Equal(t, instana.AWSSNSSpanTags{
		TopicARN: snsTopicArn,
		Error:    "",
	}, data.Tags)
}

func TestSNSNoParentSpan(t *testing.T) {
	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx)
	assert.NoError(t, err, "Error while configuring aws")

	cfg = applyTestingChanges(cfg)

	instaawsv2.Instrument(c, &cfg)

	snsClient := sns.NewFromConfig(cfg)

	snsMsg := "this is is a test message"
	snsTopicArn := "test-topic-arn"
	_, err = snsClient.Publish(ctx, &sns.PublishInput{
		Message:  &snsMsg,
		TopicArn: &snsTopicArn,
	})
	assert.NoError(t, err, "Error while publishing the sns message")

	recorderSpans := recorder.GetQueuedSpans()
	assert.Equal(t, 0, len(recorderSpans))
}

func TestSNSUnmonitoredFunction(t *testing.T) {
	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx)
	assert.NoError(t, err, "Error while configuring aws")

	cfg = applyTestingChanges(cfg)

	instaawsv2.Instrument(c, &cfg)

	snsClient := sns.NewFromConfig(cfg)

	_, err = snsClient.CreatePlatformApplication(ctx, &sns.CreatePlatformApplicationInput{
		Attributes: make(map[string]string),
		Name:       testString(5),
		Platform:   testString(5),
	})
	assert.NoError(t, err, "Error while publishing the sns message")

	recorderSpans := recorder.GetQueuedSpans()
	assert.Equal(t, 0, len(recorderSpans))
}
