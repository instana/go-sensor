package instana

import (
	opentracing "github.com/opentracing/opentracing-go"
)

// Tracer extends the opentracing.Tracer interface
type Tracer interface {
	opentracing.Tracer

	// Options gets the Options used in New() or NewWithOptions().
	Options() TracerOptions
}

// Matcher verifies whether a string meets predefined conditions
type Matcher interface {
	Match(s string) bool
}

// TracerOptions allows creating a customized Tracer via NewWithOptions. The object
// must not be updated when there is an active tracer using it.
type TracerOptions struct {
	// DropAllLogs turns log events on all spans into no-ops
	DropAllLogs bool
	// MaxLogsPerSpan limits the number of log records in a span (if set to a non-zero
	// value). If a span has more logs than this value, logs are dropped as
	// necessary
	MaxLogsPerSpan int
	// Secrets is a secrets matcher used to filter out sensitive data from HTTP requests, database
	// connection strings, etc. By default tracer does not filter any values. Package `secrets`
	// provides a set of secret matchers supported by the host agent configuration.
	//
	// See https://www.instana.com/docs/setup_and_manage/host_agent/configuration/#secrets for details
	Secrets Matcher
}
