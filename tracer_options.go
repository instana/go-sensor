// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2017

package instana

// HTTPExitSettings holds opt-in configuration for HTTP exit span error classification.
type HTTPExitSettings struct {
	// ClassifyAll4xxAsErrors marks all 4xx responses as errors on exit spans when true.
	// Ignored when ClassifyAsErrors is non-empty.
	ClassifyAll4xxAsErrors bool
	// ClassifyAsErrors lists specific 4xx status codes (400–499) to mark as errors on exit spans.
	// When non-empty, takes full precedence over ClassifyAll4xxAsErrors.
	ClassifyAsErrors []int
}

// TracerOptions carry the tracer configuration
type TracerOptions struct {
	// DropAllLogs turns log events on all spans into no-ops
	DropAllLogs bool
	// MaxLogsPerSpan defines the maximum number of log records that can be attached to a span.
	// If a span contains more logs than this limit, the excess logs will be dropped.
	// If set to `0`, default value will be used.
	MaxLogsPerSpan int
	// Secrets is a secrets matcher used to filter out sensitive data from HTTP requests, database
	// connection strings, etc. By default tracer does not filter any values. Package `secrets`
	// provides a set of secret matchers supported by the host agent configuration.
	//
	// See https://www.instana.com/docs/setup_and_manage/host_agent/configuration/#secrets for details
	Secrets Matcher
	// CollectableHTTPHeaders is a case-insensitive list of HTTP headers to be collected from HTTP requests and sent to the agent
	//
	// See https://www.instana.com/docs/setup_and_manage/host_agent/configuration/#capture-custom-http-headers for details
	CollectableHTTPHeaders []string

	// Disable allows the exclusion of specific traces or calls from tracing based on the category (technology)
	// or type (frameworks, libraries, instrumentations) supported by the traces.
	// The main benefit of disabling is reducing the overall amount of data being collected and processed.
	DisableSpans map[string]bool

	// HTTP holds HTTP-specific tracer configuration.
	HTTP struct {
		Exit HTTPExitSettings
	}

	// tracerDefaultSecrets flag is used to identify whether tracerOptions.Secrets
	// contains the default secret matcher configured by the Instana SDK.
	//
	// This flag will be set during the collector initialization process, and the following logic will be followed:
	//
	// - If the INSTANA_SECRETS environment variable is set, it will be given the highest priority.
	// - If not, the "Secrets" configured within the code will be preferred.
	// - If neither of the above is available, a default configuration will be applied to the Secrets,
	//   and this flag will be set to true.
	//
	// Later, this flag will also be used to override the secret matcher configuration received from the agent during the handshake.
	tracerDefaultSecrets bool

	// http4xxExitDefaultClassifyAll is true when ClassifyAll4xxAsErrors has not been explicitly
	// set via env var or in-code config, allowing agent configuration to override it.
	http4xxExitDefaultClassifyAll bool

	// http4xxExitDefaultClassifyList is true when ClassifyAsErrors has not been explicitly
	// set via env var or in-code config, allowing agent configuration to override it.
	http4xxExitDefaultClassifyList bool
}

// DefaultTracerOptions returns the default set of options to configure a tracer
func DefaultTracerOptions() TracerOptions {
	return TracerOptions{
		MaxLogsPerSpan:                 MaxLogsPerSpan,
		Secrets:                        DefaultSecretsMatcher(),
		http4xxExitDefaultClassifyAll:  true,
		http4xxExitDefaultClassifyList: true,
	}
}
