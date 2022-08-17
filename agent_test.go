// (c) Copyright IBM Corp. 2022

package instana

import (
	"bytes"
	"github.com/instana/testify/assert"

	"io/ioutil"
	"net/http"
	"strings"
	"testing"
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
