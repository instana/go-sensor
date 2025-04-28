// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2017

package instana

// TracerOptions carry the tracer configuration
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
	// CollectableHTTPHeaders is a case-insensitive list of HTTP headers to be collected from HTTP requests and sent to the agent
	//
	// See https://www.instana.com/docs/setup_and_manage/host_agent/configuration/#capture-custom-http-headers for details
	CollectableHTTPHeaders []string

	// tracerDefaultSecrets flag is used to override the secret matcher configuration received from the agent.
	// This flag will be set during the init collector process, and the following logic will apply:
	//
	// - If the INSTANA_SECRETS environment variable is set, it will be given the highest priority.
	// - If not, the "Secrets" configured within the code will be preferred.
	// - If neither of the above is available, the configuration received from the agent will be used.
	tracerDefaultSecrets bool
}

// DefaultTracerOptions returns the default set of options to configure a tracer
func DefaultTracerOptions() TracerOptions {
	return TracerOptions{
		MaxLogsPerSpan: MaxLogsPerSpan,
		Secrets:        DefaultSecretsMatcher(),
	}
}
