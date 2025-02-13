// (c) Copyright IBM Corp. 2023

package instaawsv2_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instaawsv2"
	"github.com/stretchr/testify/assert"
)

// Added testcase for testing unsupported aws services
func TestUnSupportedService(t *testing.T) {
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

	//RDS is currently unsupported
	rdsClient := rds.NewFromConfig(cfg)

	_, err = rdsClient.CreateDBCluster(ctx, &rds.CreateDBClusterInput{
		DBClusterIdentifier: testString(10),
		Engine:              testString(6),
	})

	recorderSpans := recorder.GetQueuedSpans()
	assert.Equal(t, 0, len(recorderSpans))
}
