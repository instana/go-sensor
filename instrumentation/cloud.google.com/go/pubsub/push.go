// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package pubsub

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/instana/go-sensor/instrumentation/cloud.google.com/go/internal/tags"

	instana "github.com/instana/go-sensor"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// TracingHandlerFunc wraps an HTTP handler that handles Google Cloud Pub/Sub push deliveries sent
// via POST requests. If a request uses any other method, the wrapper uses instana.TracingHandlerFunc()
// to trace it as a regular HTTP request.
//
// Please note, that this wrapper consumes the request body in order to to extract the trace context
// from the message, thus the (net/http.Request).Body is a copy of received data.
func TracingHandlerFunc(sensor *instana.Sensor, pathTemplate string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			instana.TracingHandlerFunc(sensor, pathTemplate, handler)(w, req)
			return
		}

		data, err := ioutil.ReadAll(req.Body)
		if err != nil {
			sensor.Logger().Error("failed to read google cloud pub/sub request body:", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		req.Body = ioutil.NopCloser(bytes.NewReader(data))

		sp, err := startConsumePushSpan(data, sensor)
		if err != nil {
			sensor.Logger().Warn("failed to start google cloud pub/sub consumer trace:", err)
			instana.TracingHandlerFunc(sensor, pathTemplate, handler)(w, req)
			return
		}

		defer sp.Finish()

		handler(w, req.WithContext(instana.ContextWithSpan(req.Context(), sp)))
	}
}

func startConsumePushSpan(body []byte, sensor *instana.Sensor) (opentracing.Span, error) {
	var delivery struct {
		Message struct {
			Attributes map[string]string `json:"attributes"`
			ID         string            `json:"messageId"`
		} `json:"message"`
		Subscription string `json:"subscription"`
	}

	if err := json.Unmarshal(body, &delivery); err != nil {
		return nil, fmt.Errorf("failed to unmarshal delivery: %s", err)
	}

	projectID, subscription, ok := parseFullyQualifiedSubscriptionName(delivery.Subscription)
	if !ok {
		return nil, fmt.Errorf("unexpected subscription name format: %s", delivery.Subscription)
	}

	opts := []opentracing.StartSpanOption{
		ext.SpanKindConsumer,
		opentracing.Tags{
			tags.GcpsOp:     "CONSUME",
			tags.GcpsProjid: projectID,
			tags.GcpsSub:    subscription,
			tags.GcpsMsgid:  delivery.Message.ID,
		},
	}

	spCtx, err := sensor.Tracer().Extract(opentracing.TextMap, opentracing.TextMapCarrier(delivery.Message.Attributes))
	switch err {
	case nil:
		opts = append(opts, opentracing.ChildOf(spCtx))
	case opentracing.ErrInvalidCarrier, opentracing.ErrInvalidSpanContext, opentracing.ErrUnsupportedFormat:
		sensor.Logger().Debug(
			"unexpected google cloud pub/sub trace context propagation format (", err, "): %#v",
			delivery.Message.Attributes,
		)
	case opentracing.ErrSpanContextNotFound:
		// do nothing
	default:
		sensor.Logger().Warn("failed to extract google cloud pub/sub trace context:", err)
	}

	return sensor.Tracer().StartSpan("gcps", opts...), nil

}

func parseFullyQualifiedSubscriptionName(fqsn string) (projectID string, subscription string, ok bool) {
	projectID, fqsn, ok = parsePathResourceID(fqsn, "projects")
	if !ok {
		return "", "", false
	}

	subscription, _, ok = parsePathResourceID(fqsn, "subscriptions")

	return projectID, subscription, ok
}

func parsePathResourceID(path, resource string) (id, rest string, ok bool) {
	if !strings.HasPrefix(path, resource+"/") {
		fmt.Println("no ", resource, " prefix in ", path)
		return "", "", false
	}

	path = strings.TrimPrefix(path, resource+"/")

	ind := strings.IndexByte(path, '/')
	if ind < 0 {
		return path, "", true
	}

	return path[:ind], path[ind+1:], true
}
