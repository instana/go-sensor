// (c) Copyright IBM Corp. 2023

package instaawsv2_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instaawsv2"
	"github.com/stretchr/testify/assert"
)

func TestS3GetObjectNoError(t *testing.T) {
	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	ps := c.Tracer().StartSpan("aws-s3-parent-span")

	ctx := instana.ContextWithSpan(context.TODO(), ps)

	cfg, err := config.LoadDefaultConfig(ctx)
	assert.NoError(t, err)

	cfg = applyTestingChanges(cfg)

	instaawsv2.Instrument(c, &cfg)

	s3Client := s3.NewFromConfig(cfg)
	bucket := "s3-test-bucket"
	key := "s3-test-key"

	_, err = s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	assert.NoError(t, err)

	ps.Finish()

	recordedSpans := recorder.GetQueuedSpans()
	assert.Equal(t, 2, len(recordedSpans))

	s3Span := recordedSpans[0]
	assert.IsType(t, instana.AWSS3SpanData{}, s3Span.Data)

	data := s3Span.Data.(instana.AWSS3SpanData)
	assert.Equal(t, instana.AWSS3SpanTags{
		Region:    region,
		Operation: "get",
		Bucket:    bucket,
		Key:       key,
		Error:     "",
	}, data.Tags)
}

func TestMonitoredS3Operations(t *testing.T) {
	s3bucket := "s3-test-bucket"
	s3key := "s3-test-bucket-key"

	testcases := map[string]struct {
		monitoredFunc func(ctx context.Context, client *s3.Client) (interface{}, error)
		expectedOut   instana.AWSS3SpanTags
	}{
		"DeleteBucket": {
			monitoredFunc: func(ctx context.Context, client *s3.Client) (interface{}, error) {
				ip := s3.DeleteBucketInput{
					Bucket: &s3bucket,
				}

				return client.DeleteBucket(ctx, &ip)
			},
			expectedOut: instana.AWSS3SpanTags{
				Region:    region,
				Operation: "deleteBucket",
				Bucket:    s3bucket,
				Key:       "",
				Error:     "",
			},
		},
		"DeleteObject": {
			monitoredFunc: func(ctx context.Context, client *s3.Client) (interface{}, error) {
				ip := s3.DeleteObjectInput{
					Bucket: &s3bucket,
					Key:    &s3key,
				}
				return client.DeleteObject(ctx, &ip)
			},
			expectedOut: instana.AWSS3SpanTags{
				Region:    region,
				Operation: "delete",
				Bucket:    s3bucket,
				Key:       s3key,
				Error:     "",
			},
		},
		"DeleteObjects": {
			monitoredFunc: func(ctx context.Context, client *s3.Client) (interface{}, error) {
				ip := s3.DeleteObjectsInput{
					Bucket: &s3bucket,
					Delete: &s3types.Delete{
						Objects: make([]s3types.ObjectIdentifier, 0),
						Quiet:   nil,
					},
				}
				return client.DeleteObjects(ctx, &ip)
			},
			expectedOut: instana.AWSS3SpanTags{
				Region:    region,
				Operation: "delete",
				Bucket:    s3bucket,
			},
		},
		"CreateBucket": {
			monitoredFunc: func(ctx context.Context, client *s3.Client) (interface{}, error) {
				ip := s3.CreateBucketInput{
					Bucket: &s3bucket,
				}
				return client.CreateBucket(ctx, &ip)
			},
			expectedOut: instana.AWSS3SpanTags{
				Region:    region,
				Operation: "createBucket",
				Bucket:    s3bucket,
				Error:     "",
			},
		},
		"MetaData": {
			monitoredFunc: func(ctx context.Context, client *s3.Client) (interface{}, error) {
				ip := s3.HeadObjectInput{
					Bucket: &s3bucket,
					Key:    &s3key,
				}
				return client.HeadObject(ctx, &ip)
			},
			expectedOut: instana.AWSS3SpanTags{
				Region:    region,
				Operation: "metadata",
				Bucket:    s3bucket,
				Key:       s3key,
				Error:     "",
			},
		},
		"ListObjects": {
			monitoredFunc: func(ctx context.Context, client *s3.Client) (interface{}, error) {
				ip := s3.ListObjectsInput{
					Bucket: &s3bucket,
				}
				return client.ListObjects(ctx, &ip)
			},
			expectedOut: instana.AWSS3SpanTags{
				Region:    region,
				Operation: "list",
				Bucket:    s3bucket,
				Error:     "",
			},
		},
		"ListObjectsV2": {
			monitoredFunc: func(ctx context.Context, client *s3.Client) (interface{}, error) {
				ip := s3.ListObjectsV2Input{
					Bucket: &s3bucket,
				}
				return client.ListObjectsV2(ctx, &ip)
			},
			expectedOut: instana.AWSS3SpanTags{
				Region:    region,
				Operation: "list",
				Bucket:    s3bucket,
				Error:     "",
			},
		},
		"PutObject": {
			monitoredFunc: func(ctx context.Context, client *s3.Client) (interface{}, error) {
				ip := s3.PutObjectInput{
					Bucket: &s3bucket,
					Key:    &s3key,
				}
				return client.PutObject(ctx, &ip)
			},
			expectedOut: instana.AWSS3SpanTags{
				Region:    region,
				Operation: "put",
				Bucket:    s3bucket,
				Key:       s3key,
				Error:     "",
			},
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			recorder := instana.NewTestRecorder()
			c := instana.InitCollector(&instana.Options{
				AgentClient: alwaysReadyClient{},
				Recorder:    recorder,
			})
			defer instana.ShutdownCollector()

			ps := c.Tracer().StartSpan("aws-s3-parent-span")

			ctx := instana.ContextWithSpan(context.TODO(), ps)

			cfg, err := config.LoadDefaultConfig(ctx)
			assert.NoError(t, err)

			cfg = applyTestingChanges(cfg)

			instaawsv2.Instrument(c, &cfg)

			s3Client := s3.NewFromConfig(cfg)

			_, err = testcase.monitoredFunc(ctx, s3Client)
			assert.NoError(t, err)

			ps.Finish()

			recordedSpans := recorder.GetQueuedSpans()
			assert.Equal(t, 2, len(recordedSpans))

			s3Span := recordedSpans[0]
			assert.IsType(t, instana.AWSS3SpanData{}, s3Span.Data)

			data := s3Span.Data.(instana.AWSS3SpanData)
			assert.Equal(t, testcase.expectedOut, data.Tags)
		})
	}
}

func TestErrorNoParentSpanS3(t *testing.T) {
	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx)
	assert.NoError(t, err)

	cfg = applyTestingChanges(cfg)

	instaawsv2.Instrument(c, &cfg)

	s3Client := s3.NewFromConfig(cfg)
	bucket := "s3-test-bucket"
	key := "s3-test-key"

	_, err = s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	assert.NoError(t, err)

	recordedSpans := recorder.GetQueuedSpans()
	assert.Equal(t, 0, len(recordedSpans))
}

func TestErrorUnmonitoredS3Method(t *testing.T) {
	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer instana.ShutdownCollector()

	ctx := context.TODO()

	cfg, err := config.LoadDefaultConfig(ctx)
	assert.NoError(t, err)

	cfg = applyTestingChanges(cfg)

	instaawsv2.Instrument(c, &cfg)

	s3Client := s3.NewFromConfig(cfg)

	s3bucket := "s3-test-bucket"
	s3key := "s3-test-bucket-key"
	s3copy := "s3-copy-source"

	_, err = s3Client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     &s3bucket,
		Key:        &s3key,
		CopySource: &s3copy,
	})

	recordedSpans := recorder.GetQueuedSpans()
	assert.Equal(t, 0, len(recordedSpans))
}
