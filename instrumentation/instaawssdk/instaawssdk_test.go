// (c) Copyright IBM Corp. 2021

package instaawssdk_test

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instaawssdk"
	"github.com/instana/testify/assert"
	"testing"
)

func TestNew1(t *testing.T) {
	sess := instaawssdk.New(instana.NewSensor("test"), &aws.Config{Region: aws.String("region")}, &aws.Config{Endpoint: aws.String("somestring")})
	assert.IsType(t, &session.Session{}, sess)
}

func TestNew2(t *testing.T) {
	sess := instaawssdk.New(instana.NewSensor("test"), []*aws.Config{{Region: aws.String("region")}, &aws.Config{Endpoint: aws.String("somestring")}}...)
	assert.IsType(t, &session.Session{}, sess)
}

func TestNewSession1(t *testing.T) {
	sess, err := instaawssdk.NewSession(instana.NewSensor("test"), &aws.Config{Region: aws.String("region")}, &aws.Config{Endpoint: aws.String("somestring")})
	assert.IsType(t, &session.Session{}, sess)
	assert.NoError(t, err)
}

func TestNewSession2(t *testing.T) {
	sess, err := instaawssdk.NewSession(instana.NewSensor("test"), []*aws.Config{{Region: aws.String("region")}, &aws.Config{Endpoint: aws.String("somestring")}}...)
	assert.IsType(t, &session.Session{}, sess)
	assert.NoError(t, err)
}

func TestNewSessionWithOptions(t *testing.T) {
	sess, err := instaawssdk.NewSessionWithOptions(instana.NewSensor("test"), session.Options{})
	assert.IsType(t, &session.Session{}, sess)
	assert.NoError(t, err)
}
