// (c) Copyright IBM Corp. 2024

package instafasthttp

import (
	"context"
	"time"

	instana "github.com/instana/go-sensor"
	"github.com/valyala/fasthttp"
)

// GetInstrumentedClient returns an instrumented instafasthttp.Client instance derived from a *fasthttp.Client instance
func GetInstrumentedClient(sensor instana.TracerLogger, orgClient *fasthttp.Client) Client {
	return &instaClient{
		Client: orgClient,
		sensor: sensor,
	}
}

// Instrumented fasthttp.Client
//
// Most of the methods are the same as those in fasthttp.Client.
//
// Only Do, DoTimeout, and DoDeadline differ from fasthttp.Client.
// Please use these methods instead of their fasthttp.Client counterparts to enable tracing.
type Client interface {
	// The following methods are from the original *fasthttp.Client; there is no need to implement them.

	// Get returns the status code and body of url.
	//
	// The contents of dst will be replaced by the body and returned, if the dst
	// is too small a new slice will be allocated.
	//
	// The function follows redirects. Use Do* for manually handling redirects.
	Get(dst []byte, url string) (statusCode int, body []byte, err error)

	// GetTimeout returns the status code and body of url.
	//
	// The contents of dst will be replaced by the body and returned, if the dst
	// is too small a new slice will be allocated.
	//
	// The function follows redirects. Use Do* for manually handling redirects.
	//
	// ErrTimeout error is returned if url contents couldn't be fetched
	// during the given timeout.
	GetTimeout(dst []byte, url string, timeout time.Duration) (statusCode int, body []byte, err error)

	// GetDeadline returns the status code and body of url.
	//
	// The contents of dst will be replaced by the body and returned, if the dst
	// is too small a new slice will be allocated.
	//
	// The function follows redirects. Use Do* for manually handling redirects.
	//
	// ErrTimeout error is returned if url contents couldn't be fetched
	// until the given deadline.
	GetDeadline(dst []byte, url string, deadline time.Time) (statusCode int, body []byte, err error)

	// Post sends POST request to the given url with the given POST arguments.
	//
	// The contents of dst will be replaced by the body and returned, if the dst
	// is too small a new slice will be allocated.
	//
	// The function follows redirects. Use Do* for manually handling redirects.
	//
	// Empty POST body is sent if postArgs is nil.
	Post(dst []byte, url string, postArgs *fasthttp.Args) (statusCode int, body []byte, err error)

	// CloseIdleConnections closes any connections which were previously
	// connected from previous requests but are now sitting idle in a
	// "keep-alive" state. It does not interrupt any connections currently
	// in use.
	CloseIdleConnections()

	// DoTimeout performs the given request and waits for response during
	// the given timeout duration.
	//
	// Request must contain at least non-zero RequestURI with full url (including
	// scheme and host) or non-zero Host header + RequestURI.
	//
	// Client determines the server to be requested in the following order:
	//
	//   - from RequestURI if it contains full url with scheme and host;
	//   - from Host header otherwise.
	//
	// The function doesn't follow redirects. Use Get* for following redirects.
	//
	// Response is ignored if resp is nil.
	//
	// ErrTimeout is returned if the response wasn't returned during
	// the given timeout.
	// Immediately returns ErrTimeout if timeout value is negative.
	//
	// ErrNoFreeConns is returned if all Client.MaxConnsPerHost connections
	// to the requested host are busy.
	//
	// It is recommended obtaining req and resp via Acquire
	//
	// Pass a valid context as the first argument for for span correlation
	DoTimeout(ctx context.Context, req *fasthttp.Request, resp *fasthttp.Response, timeout time.Duration) error

	// DoDeadline performs the given request and waits for response until
	// the given deadline.
	//
	// Request must contain at least non-zero RequestURI with full url (including
	// scheme and host) or non-zero Host header + RequestURI.
	//
	// Client determines the server to be requested in the following order:
	//
	//   - from RequestURI if it contains full url with scheme and host;
	//   - from Host header otherwise.
	//
	// The function doesn't follow redirects. Use Get* for following redirects.
	//
	// Response is ignored if resp is nil.
	//
	// ErrTimeout is returned if the response wasn't returned until
	// the given deadline.
	// Immediately returns ErrTimeout if the deadline has already been reached.
	//
	// ErrNoFreeConns is returned if all Client.MaxConnsPerHost connections
	// to the requested host are busy.
	//
	// It is recommended obtaining req and resp via AcquireRequest
	// and AcquireResponse in performance-critical code.
	//
	// Pass a valid context as the first argument for for span correlation
	DoDeadline(ctx context.Context, req *fasthttp.Request, resp *fasthttp.Response, deadline time.Time) error

	// DoRedirects performs the given http request and fills the given http response,
	// following up to maxRedirectsCount redirects. When the redirect count exceeds
	// maxRedirectsCount, ErrTooManyRedirects is returned.
	//
	// Request must contain at least non-zero RequestURI with full url (including
	// scheme and host) or non-zero Host header + RequestURI.
	//
	// Client determines the server to be requested in the following order:
	//
	//   - from RequestURI if it contains full url with scheme and host;
	//   - from Host header otherwise.
	//
	// Response is ignored if resp is nil.
	//
	// ErrNoFreeConns is returned if all DefaultMaxConnsPerHost connections
	// to the requested host are busy.
	//
	// It is recommended obtaining req and resp via AcquireRequest
	// and AcquireResponse in performance-critical code.
	//
	// Pass a valid context as the first argument for for span correlation
	DoRedirects(ctx context.Context, req *fasthttp.Request, resp *fasthttp.Response, maxRedirectsCount int) error

	// Do performs the given http request and fills the given http response.
	//
	// Request must contain at least non-zero RequestURI with full url (including
	// scheme and host) or non-zero Host header + RequestURI.
	//
	// Client determines the server to be requested in the following order:
	//
	//   - from RequestURI if it contains full url with scheme and host;
	//   - from Host header otherwise.
	//
	// Response is ignored if resp is nil.
	//
	// The function doesn't follow redirects. Use Get* for following redirects.
	//
	// ErrNoFreeConns is returned if all Client.MaxConnsPerHost connections
	// to the requested host are busy.
	//
	// It is recommended obtaining req and resp via AcquireRequest
	// and AcquireResponse in performance-critical code.
	//
	// Pass a valid context as the first argument for for span correlation
	Do(ctx context.Context, req *fasthttp.Request, resp *fasthttp.Response) error

	// Unwrap returns the original *fasthttp.Client
	Unwrap() *fasthttp.Client
}

type instaClient struct {
	*fasthttp.Client
	sensor instana.TracerLogger
}

func (ic *instaClient) Unwrap() *fasthttp.Client {
	return ic.Client
}

func (ic *instaClient) DoTimeout(ctx context.Context, req *fasthttp.Request, resp *fasthttp.Response, timeout time.Duration) error {
	cfp := &clientFuncParams{
		sensor:         ic.sensor,
		ic:             ic,
		clientFuncType: doWithTimeoutFunc,
		timeout:        timeout,
	}
	_, err := instrumentClient(ctx, req, resp, cfp)
	return err
}

func (ic *instaClient) DoDeadline(ctx context.Context, req *fasthttp.Request, resp *fasthttp.Response, deadline time.Time) error {
	cfp := &clientFuncParams{
		sensor:         ic.sensor,
		ic:             ic,
		clientFuncType: doWithDeadlineFunc,
		deadline:       deadline,
	}
	_, err := instrumentClient(ctx, req, resp, cfp)
	return err
}

func (ic *instaClient) DoRedirects(ctx context.Context, req *fasthttp.Request, resp *fasthttp.Response, maxRedirectsCount int) error {
	cfp := &clientFuncParams{
		sensor:            ic.sensor,
		ic:                ic,
		clientFuncType:    doWithRedirectsFunc,
		maxRedirectsCount: maxRedirectsCount,
	}
	_, err := instrumentClient(ctx, req, resp, cfp)
	return err
}

func (ic *instaClient) Do(ctx context.Context, req *fasthttp.Request, resp *fasthttp.Response) error {
	cfp := &clientFuncParams{
		sensor:         ic.sensor,
		ic:             ic,
		clientFuncType: doFunc,
	}
	_, err := instrumentClient(ctx, req, resp, cfp)
	return err
}
