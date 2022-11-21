// (c) Copyright IBM Corp. 2021

package instaawssdk_test

import (
	"context"
	"testing"

	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/autoprofile"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instaawssdk"
	"github.com/stretchr/testify/assert"
)

func TestNew1(t *testing.T) {
	defer instana.ShutdownSensor()
	sess := instaawssdk.New(instana.NewSensor("test"), &aws.Config{Region: aws.String("region")}, &aws.Config{Endpoint: aws.String("somestring")})
	assert.IsType(t, &session.Session{}, sess)
}

func TestNew2(t *testing.T) {
	defer instana.ShutdownSensor()
	sess := instaawssdk.New(instana.NewSensor("test"), []*aws.Config{{Region: aws.String("region")}, &aws.Config{Endpoint: aws.String("somestring")}}...)
	assert.IsType(t, &session.Session{}, sess)
}

func TestNewSession1(t *testing.T) {
	defer instana.ShutdownSensor()
	sess, err := instaawssdk.NewSession(instana.NewSensor("test"), &aws.Config{Region: aws.String("region")}, &aws.Config{Endpoint: aws.String("somestring")})
	assert.IsType(t, &session.Session{}, sess)
	assert.NoError(t, err)
}

func TestNewSession2(t *testing.T) {
	defer instana.ShutdownSensor()
	sess, err := instaawssdk.NewSession(instana.NewSensor("test"), []*aws.Config{{Region: aws.String("region")}, &aws.Config{Endpoint: aws.String("somestring")}}...)
	assert.IsType(t, &session.Session{}, sess)
	assert.NoError(t, err)
}

func TestNewSessionWithOptions(t *testing.T) {
	defer instana.ShutdownSensor()
	sess, err := instaawssdk.NewSessionWithOptions(instana.NewSensor("test"), session.Options{})
	assert.IsType(t, &session.Session{}, sess)
	assert.NoError(t, err)
}

type alwaysReadyClient struct{}

func (alwaysReadyClient) Ready() bool                                       { return true }
func (alwaysReadyClient) SendMetrics(data acceptor.Metrics) error           { return nil }
func (alwaysReadyClient) SendEvent(event *instana.EventData) error          { return nil }
func (alwaysReadyClient) SendSpans(spans []instana.Span) error              { return nil }
func (alwaysReadyClient) SendProfiles(profiles []autoprofile.Profile) error { return nil }
func (alwaysReadyClient) Flush(context.Context) error                       { return nil }
