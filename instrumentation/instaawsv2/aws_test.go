// (c) Copyright IBM Corp. 2023

package instaawsv2_test

import (
	"bytes"
	"context"
	"io"
	"math/rand"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/acceptor"
	"github.com/instana/go-sensor/autoprofile"
)

type mockTransport struct{}

func (t mockTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader([]byte(`{"message": "response-body-bytes"}`))),
		Header: http.Header{
			"x-amzn-requestid": []string{requestID},
		},
	}, nil
}

type mockCredentials struct{}

func (cred mockCredentials) Retrieve(context.Context) (aws.Credentials, error) {
	return aws.Credentials{}, nil
}

func applyTestingChanges(cfg aws.Config) aws.Config {
	cfg.Credentials = mockCredentials{}
	cfg.Region = region
	cfg.HTTPClient = &http.Client{
		Transport: &mockTransport{},
	}

	return cfg
}

func testString(l int) *string {
	const alphabets = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	buf := make([]byte, l)
	for i := range buf {
		buf[i] = alphabets[rand.Intn(len(alphabets))]
	}
	str := string(buf)

	return &str
}

type alwaysReadyClient struct{}

func (alwaysReadyClient) Ready() bool                              { return true }
func (alwaysReadyClient) SendMetrics(acceptor.Metrics) error       { return nil }
func (alwaysReadyClient) SendEvent(*instana.EventData) error       { return nil }
func (alwaysReadyClient) SendSpans([]instana.Span) error           { return nil }
func (alwaysReadyClient) SendProfiles([]autoprofile.Profile) error { return nil }
func (alwaysReadyClient) Flush(context.Context) error              { return nil }
