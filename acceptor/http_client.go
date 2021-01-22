// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package acceptor

import (
	"errors"
	"net/http"
	"net/url"
	"os"
	"time"
)

// ErrMalformedProxyURL is returned by NewHTTPClient() when the INSTANA_ENDPOINT_PROXY= env var contains
// a malformed URL string
var ErrMalformedProxyURL = errors.New("malformed URL value found in INSTANA_ENDPOINT_PROXY, ignoring")

// NewHTTPClient returns an http.Client configured for sending requests to the Instana serverless acceptor.
// If INSTANA_ENDPOINT_PROXY= env var is populated with a vaild URL, the returned client will use it as an
// HTTP proxy. If the value is malformed, this setting is ignored and ErrMalformedProxyURL error is returned.
// The returned http.Client instance in this case is usable, but does not use a proxy to connect to the acceptor.
func NewHTTPClient(timeout time.Duration) (*http.Client, error) {
	client := &http.Client{}

	if timeout > 0 {
		client.Timeout = timeout
	}

	proxy := os.Getenv("INSTANA_ENDPOINT_PROXY")
	if proxy == "" {
		return client, nil
	}

	proxyURL, err := url.Parse(proxy)
	if err != nil {
		return client, ErrMalformedProxyURL
	}

	client.Transport = &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}

	return client, nil
}
