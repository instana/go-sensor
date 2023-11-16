// (c) Copyright IBM Corp. 2023

//go:build integration
// +build integration

package instagocb_test

import (
	"context"
	"testing"

	"github.com/couchbase/gocb/v2"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/autoprofile"
	"github.com/instana/go-sensor/instrumentation/instagocb"
	"github.com/stretchr/testify/assert"
)

var connStr = "couchbase://localhost"
var username = "Administrator"
var password = "password"
var bucketName = "test-bucket"

// helpers

type alwaysReadyClient struct{}

func (alwaysReadyClient) Ready() bool                                       { return true }
func (alwaysReadyClient) SendMetrics(data acceptor.Metrics) error           { return nil }
func (alwaysReadyClient) SendEvent(event *instana.EventData) error          { return nil }
func (alwaysReadyClient) SendSpans(spans []instana.Span) error              { return nil }
func (alwaysReadyClient) SendProfiles(profiles []autoprofile.Profile) error { return nil }
func (alwaysReadyClient) Flush(context.Context) error                       { return nil }

func prepare(t *testing.T) (*instana.Recorder, context.Context, instagocb.Cluster, *assert.Assertions) {
	a := assert.New(t)
	recorder := instana.NewTestRecorder()
	tracer := instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder)
	sensor := instana.NewSensorWithTracer(tracer)

	ctx := context.Background()
	conn, err := instagocb.Connect(sensor, connStr, gocb.ClusterOptions{
		Authenticator: gocb.PasswordAuthenticator{
			Username: username,
			Password: password,
		},
	})

	a.NoError(err)

	// clearing existing bucket
	bucketMgr := conn.Buckets()
	_ = bucketMgr.DropBucket(bucketName, &gocb.DropBucketOptions{})

	return recorder, ctx, conn, a

}

func getLatestSpan(recorder *instana.Recorder) instana.Span {
	spans := recorder.GetQueuedSpans()
	span := spans[len(spans)-1]

	return span
}
