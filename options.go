// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2016

package instana

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/stretchr/testify/assert/yaml"
)

// Options allows the user to configure the to-be-initialized sensor
type Options struct {
	// Service is the global service name that will be used to identify the program in the Instana backend
	Service string
	// AgentHost is the Instana host agent host name
	//
	// Note: This setting has no effect in serverless environments. To specify the serverless acceptor endpoint,
	// INSTANA_ENDPOINT_URL env var. See https://www.instana.com/docs/reference/environment_variables/#serverless-monitoring
	// for more details.
	AgentHost string
	// AgentPort is the Instana host agent port
	//
	// Note: This setting has no effect in serverless environments. To specify the serverless acceptor endpoint,
	// INSTANA_ENDPOINT_URL env var. See https://www.instana.com/docs/reference/environment_variables/#serverless-monitoring
	// for more details.
	AgentPort int
	// MaxBufferedSpans is the maximum number of spans to buffer
	MaxBufferedSpans int
	// ForceTransmissionStartingAt is the number of spans to collect before flushing the buffer to the agent
	ForceTransmissionStartingAt int
	// LogLevel is the initial logging level for the logger used by Instana tracer. The valid log levels are
	// logger.{Error,Warn,Info,Debug}Level provided by the github.com/instana/go-sensor/logger package.
	//
	// Note: this setting is only used to initialize the default logger and has no effect if a custom logger is set via instana.SetLogger()
	LogLevel int
	// EnableAutoProfile enables automatic continuous process profiling when set to true
	EnableAutoProfile bool
	// MaxBufferedProfiles is the maximum number of profiles to buffer
	MaxBufferedProfiles int
	// IncludeProfilerFrames is whether to include profiler calls into the profile or not
	IncludeProfilerFrames bool
	// Tracer contains tracer-specific configuration used by all tracers
	Tracer TracerOptions
	// AgentClient client to communicate with the agent. In most cases, there is no need to provide it.
	// If it is nil the default implementation will be used.
	AgentClient AgentClient

	// Recorder records and manages spans. When this option is not set, instana.NewRecorder() will be used.
	Recorder SpanRecorder

	disableW3CTraceCorrelation bool
}

// DefaultOptions returns the default set of options to configure Instana sensor.
// The service name is set to the name of current executable, the MaxBufferedSpans
// and ForceTransmissionStartingAt are set to instana.DefaultMaxBufferedSpans and
// instana.DefaultForceSpanSendAt correspondingly. The AgentHost and AgentPort are
// taken from the env INSTANA_AGENT_HOST and INSTANA_AGENT_PORT if set, and default
// to localhost and 42699 otherwise.
func DefaultOptions() *Options {
	opts := &Options{
		Tracer: DefaultTracerOptions(),
	}
	opts.setDefaults()

	return opts
}

func (opts *Options) setDefaults() {
	if opts.MaxBufferedSpans == 0 {
		opts.MaxBufferedSpans = DefaultMaxBufferedSpans
	}

	if opts.ForceTransmissionStartingAt == 0 {
		opts.ForceTransmissionStartingAt = DefaultForceSpanSendAt
	}

	if opts.AgentHost == "" {
		opts.AgentHost = agentDefaultHost

		if host := os.Getenv("INSTANA_AGENT_HOST"); host != "" {
			opts.AgentHost = host
		}
	}

	if opts.AgentPort == 0 {
		opts.AgentPort = agentDefaultPort

		if port, err := strconv.Atoi(os.Getenv("INSTANA_AGENT_PORT")); err == nil {
			opts.AgentPort = port
		}
	}

	secretsMatcher, err := parseInstanaSecrets(os.Getenv("INSTANA_SECRETS"))
	if err != nil {
		defaultLogger.Warn("invalid INSTANA_SECRETS= env variable value: ", err, ", ignoring")
		secretsMatcher = opts.Tracer.Secrets
	}

	if secretsMatcher == nil {
		// If secretMatcher is nil, it means no in-code secret matcher has been configured,
		// and the INSTANA_SECRETS environment variable also doesn't have a valid matcher.
		// So, we will set a default secret matcher here and mark tracerDefaultSecrets as true,
		// so that it can be overridden later if a matcher is received from the agent.
		secretsMatcher = DefaultSecretsMatcher()
		opts.Tracer.tracerDefaultSecrets = true
	}

	opts.Tracer.Secrets = secretsMatcher

	if opts.Tracer.MaxLogsPerSpan == 0 {
		opts.Tracer.MaxLogsPerSpan = MaxLogsPerSpan
	}

	if collectableHeaders, ok := os.LookupEnv("INSTANA_EXTRA_HTTP_HEADERS"); ok {
		opts.Tracer.CollectableHTTPHeaders = parseInstanaExtraHTTPHeaders(collectableHeaders)
	}

	// Check if INSTANA_CONFIG_PATH environment variable is set
	if configPath, ok := os.LookupEnv("INSTANA_CONFIG_PATH"); ok {
		if err := parseConfigFile(configPath, &opts.Tracer); err != nil {
			defaultLogger.Warn("invalid INSTANA_CONFIG_PATH= env variable value: ", err, ", ignoring")
		}
		// else check if INSTANA_TRACING_DISABLE environment variable is set
	} else if tracingDisable, ok := os.LookupEnv("INSTANA_TRACING_DISABLE"); ok {
		parseInstanaTracingDisable(tracingDisable, &opts.Tracer)
	}

	opts.disableW3CTraceCorrelation = os.Getenv("INSTANA_DISABLE_W3C_TRACE_CORRELATION") != ""
}

// parseInstanaTracingDisable processes the INSTANA_TRACING_DISABLE environment variable value
// and updates the TracerOptions.Disable map accordingly.
//
// When the value is a boolean (true/false), the whole tracing feature is disabled/enabled.
// When a list of category or type names is specified, only those will be disabled.
//
// Examples:
// INSTANA_TRACING_DISABLE=True - disables all tracing
// INSTANA_TRACING_DISABLE="logging" - disables logging category
func parseInstanaTracingDisable(value string, opts *TracerOptions) {
	// Initialize the Disable map if it doesn't exist
	if opts.Disable == nil {
		opts.Disable = make(map[string]bool)
	}

	// Trim spaces from the value
	value = strings.TrimSpace(value)

	// if it's a boolean value, disable all categories
	if isBooleanTrue(value) {
		opts.DisableAllCategories()
		return
	}

	// if it's not a boolean value, process as a comma-separated list and disable each category.
	items := strings.Split(value, ",")
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item != "" {
			opts.Disable[item] = true
		}
	}
}

// isBooleanTrue checks if a string represents a boolean true value
func isBooleanTrue(value string) bool {
	value = strings.ToLower(value)
	return value == "true"
}

// parseConfigFile reads and parses the YAML configuration file at the given path
// and updates the TracerOptions accordingly.
//
// The YAML file must follow this format:
// tracing:
//   disable:
//     - logging: true

func parseConfigFile(path string, opts *TracerOptions) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	type Config struct {
		Tracing struct {
			Disable []map[string]bool `yaml:"disable"`
		} `yaml:"tracing"`
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	if opts.Disable == nil {
		opts.Disable = make(map[string]bool)
	}

	// Add the categories configured in the YAML file to the Disable map
	for _, disableMap := range config.Tracing.Disable {
		for category, enabled := range disableMap {
			if enabled {
				opts.Disable[category] = true
			}
		}

	}

	return nil
}
