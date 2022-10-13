// (c) Copyright IBM Corp. 2022

package instana

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/instana/testify/assert"
)

func Test_agentS_SendSpans(t *testing.T) {
	tests := []struct {
		name  string
		spans []Span
	}{
		{
			name: "big span",
			spans: []Span{
				{
					Data: HTTPSpanData{
						Tags: HTTPSpanTags{
							URL: strings.Repeat("1", maxContentLength),
						},
					},
				},
			},
		},
		{
			name: "multiple big span",
			spans: []Span{
				{Data: HTTPSpanData{Tags: HTTPSpanTags{URL: strings.Repeat("1", maxContentLength)}}},
				{Data: HTTPSpanData{Tags: HTTPSpanTags{URL: strings.Repeat("1", maxContentLength)}}},
				{Data: HTTPSpanData{Tags: HTTPSpanTags{URL: strings.Repeat("1", maxContentLength)}}},
				{Data: HTTPSpanData{Tags: HTTPSpanTags{URL: strings.Repeat("1", maxContentLength)}}},
				{Data: HTTPSpanData{Tags: HTTPSpanTags{URL: strings.Repeat("1", maxContentLength)}}},
				{Data: HTTPSpanData{Tags: HTTPSpanTags{URL: strings.Repeat("1", maxContentLength)}}},
			},
		},
		{
			name: "not really a big span",
			spans: []Span{
				{
					Data: HTTPSpanData{
						Tags: HTTPSpanTags{
							URL: strings.Repeat("1", maxContentLength/2),
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &agentS{host: "", from: &fromS{}, logger: defaultLogger, client: &httpClientMock{
				resp: &http.Response{
					StatusCode: 200,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte("{}"))),
				},
			}}
			err := agent.SendSpans(tt.spans)

			assert.NoError(t, err)
		})
	}
}

type httpClientMock struct {
	resp *http.Response
	err  error
}

func (h httpClientMock) Do(req *http.Request) (*http.Response, error) {
	return h.resp, h.err
}

func Test_agentResponse_getExtraHTTPHeaders(t *testing.T) {

	tests := []struct {
		name         string
		originalJSON string
		want         []string
	}{
		{
			name:         "old config",
			originalJSON: `{"pid":37808,"agentUuid":"88:66:5a:ff:fe:05:a5:f0","extraHeaders":["expected-value"],"secrets":{"matcher":"contains-ignore-case","list":["key","pass","secret"]}}`,
			want:         []string{"expected-value"},
		},
		{
			name:         "new config",
			originalJSON: `{"pid":38381,"agentUuid":"88:66:5a:ff:fe:05:a5:f0","tracing":{"extra-http-headers":["expected-value"]},"extraHeaders":["non-expected-value"],"secrets":{"matcher":"contains-ignore-case","list":["key","pass","secret"]}}`,
			want:         []string{"expected-value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &agentResponse{}
			json.Unmarshal([]byte(tt.originalJSON), r)
			assert.Equalf(t, tt.want, r.getExtraHTTPHeaders(), "getExtraHTTPHeaders()")
		})
	}
}

func Test_agentApplyHostSettings(t *testing.T) {
	agent := &agentS{}
	response := agentResponse{
		Pid:    37892,
		HostID: "myhost",
		Tracing: struct {
			ExtraHTTPHeaders []string `json:"extra-http-headers"`
		}{
			ExtraHTTPHeaders: []string{"my-unwanted-custom-headers"},
		},
	}

	opts := &Options{
		Service: "test_service",
		Tracer: TracerOptions{
			CollectableHTTPHeaders: []string{"x-custom-header-1", "x-custom-header-2"},
		},
	}

	sensor = newSensor(opts)
	defer func() {
		sensor = nil
	}()

	agent.applyHostAgentSettings(response)

	assert.NotContains(t, sensor.options.Tracer.CollectableHTTPHeaders, "my-unwanted-custom-headers")
}
