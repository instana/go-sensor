// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instana

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testMatcher struct{}

func (t testMatcher) Match(s string) bool {
	if s == "testing_matcher" {
		return true
	}
	return false
}

func TestDefaultOptions(t *testing.T) {
	// DefaultOptions now returns minimal/zero-value Options
	opts := DefaultOptions()

	// Verify that DefaultOptions returns zero values
	assert.NotNil(t, opts.Recorder, "Recorder should be initialized")
	assert.NotNil(t, opts.Tracer, "Tracer should be initialized")
}

func TestTracerOptionsPrecedence_InCodeConfigPresent(t *testing.T) {
	secretsRestore := restoreEnvVarFunc("INSTANA_SECRETS")
	headerRestore := restoreEnvVarFunc("INSTANA_EXTRA_HTTP_HEADERS")

	os.Unsetenv("INSTANA_SECRETS")
	os.Unsetenv("INSTANA_EXTRA_HTTP_HEADERS")

	defer secretsRestore()
	defer headerRestore()

	testOpts := &Options{
		AgentHost:                   "localhost",
		AgentPort:                   42699,
		MaxBufferedSpans:            DefaultMaxBufferedSpans,
		ForceTransmissionStartingAt: DefaultForceSpanSendAt,
		Tracer: TracerOptions{
			Secrets:                testMatcher{},
			CollectableHTTPHeaders: []string{"test", "test1"},
		},
	}

	testOpts.applyConfiguration()

	assert.Equal(t, true, testOpts.Tracer.Secrets.Match("testing_matcher"))
	assert.Equal(t, false, testOpts.Tracer.Secrets.Match("foo"))
	assert.Equal(t, false, testOpts.Tracer.tracerDefaultSecrets)

	assert.Equal(t, []string{"test", "test1"}, testOpts.Tracer.CollectableHTTPHeaders)

}

func TestTracerOptionsPrecedence_InCodeConfigAbsent(t *testing.T) {
	secretsRestore := restoreEnvVarFunc("INSTANA_SECRETS")
	headerRestore := restoreEnvVarFunc("INSTANA_EXTRA_HTTP_HEADERS")

	os.Setenv("INSTANA_SECRETS", "contains-ignore-case:key,password1,secret1")
	os.Setenv("INSTANA_EXTRA_HTTP_HEADERS", "abc;def")

	defer secretsRestore()
	defer headerRestore()

	testOpts := &Options{
		AgentHost:                   "localhost",
		AgentPort:                   42699,
		MaxBufferedSpans:            DefaultMaxBufferedSpans,
		ForceTransmissionStartingAt: DefaultForceSpanSendAt,
		Tracer:                      TracerOptions{},
	}

	testOpts.applyConfiguration()

	assert.Equal(t, false, testOpts.Tracer.Secrets.Match("testing_matcher"))
	assert.Equal(t, true, testOpts.Tracer.Secrets.Match("secret1"))
	assert.Equal(t, false, testOpts.Tracer.tracerDefaultSecrets)

	assert.Equal(t, []string{"abc", "def"}, testOpts.Tracer.CollectableHTTPHeaders)

}

func TestTracerOptionsPrecedence_InCodeConfigAndEnvAbsent(t *testing.T) {
	secretsRestore := restoreEnvVarFunc("INSTANA_SECRETS")
	headerRestore := restoreEnvVarFunc("INSTANA_EXTRA_HTTP_HEADERS")

	os.Unsetenv("INSTANA_SECRETS")
	os.Unsetenv("INSTANA_EXTRA_HTTP_HEADERS")

	defer secretsRestore()
	defer headerRestore()

	testOpts := &Options{
		AgentHost:                   "localhost",
		AgentPort:                   42699,
		MaxBufferedSpans:            DefaultMaxBufferedSpans,
		ForceTransmissionStartingAt: DefaultForceSpanSendAt,
		Tracer:                      TracerOptions{},
	}

	testOpts.applyConfiguration()

	assert.Equal(t, false, testOpts.Tracer.Secrets.Match("testing_matcher"))
	assert.Equal(t, true, testOpts.Tracer.Secrets.Match("secret"))
	assert.Equal(t, true, testOpts.Tracer.tracerDefaultSecrets)

	assert.Equal(t, 0, len(testOpts.Tracer.CollectableHTTPHeaders))

}

func restoreEnvVarFunc(key string) func() {
	if oldValue, ok := os.LookupEnv(key); ok {
		return func() { os.Setenv(key, oldValue) }
	}

	return func() { os.Unsetenv(key) }
}

func TestParseInstanaTracingDisable(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected map[string]bool
	}{
		{
			name:     "Empty value",
			value:    "",
			expected: map[string]bool{},
		},
		{
			name:  "With extra spaces",
			value: "   logging  ",
			expected: map[string]bool{
				"logging": true,
			},
		},
		{
			name:  "Valid value",
			value: "logging",
			expected: map[string]bool{
				"logging": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := TracerOptions{}
			parseInstanaTracingDisable(tt.value, &opts)

			// Check if the maps have the same size
			if len(opts.DisableSpans) != len(tt.expected) {
				t.Errorf("Expected map size %d, got %d", len(tt.expected), len(opts.DisableSpans))
			}

			// Check if all expected keys are present with correct values
			for k, v := range tt.expected {
				if opts.DisableSpans[k] != v {
					t.Errorf("Expected %s to be %v, got %v", k, v, opts.DisableSpans[k])
				}
			}

			// Check if there are no unexpected keys
			for k := range opts.DisableSpans {
				if _, exists := tt.expected[k]; !exists {
					t.Errorf("Unexpected key in result: %s", k)
				}
			}
		})
	}
}

func TestInstanaTracingDisableEnvVar(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected map[string]bool
		tooLarge bool
	}{
		{
			name:     "No env var",
			envValue: "",
			expected: map[string]bool{},
		},
		{
			name:     "Disable logging",
			envValue: "logging",
			expected: map[string]bool{
				"logging": true,
			},
		},
		{
			name:     "Value exceeds size limit",
			envValue: strings.Repeat("x", MaxEnvValueSize+1),
			expected: map[string]bool{},
			tooLarge: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.envValue != "" {
				t.Setenv("INSTANA_TRACING_DISABLE", tt.envValue)
			}

			opts := DefaultOptions()
			// Apply configuration precedence to read environment variables
			opts.applyConfiguration()

			// For the too large case, we expect the environment variable to be ignored
			// and no categories to be disabled
			if tt.tooLarge {
				assert.Empty(t, opts.Tracer.DisableSpans, "Expected no disabled spans for too large env value")
				return
			}

			// Check if the maps have the expected values
			for k, v := range tt.expected {
				if opts.Tracer.DisableSpans[k] != v {
					t.Errorf("Expected %s to be %v, got %v", k, v, opts.Tracer.DisableSpans[k])
				}
			}
		})
	}
}

func TestConfigFileHandling(t *testing.T) {
	tests := []struct {
		name              string
		yamlContent       string
		useEnvVar         bool
		invalidConfigPath bool
		expectedError     bool
		expectedDisabled  []string
	}{
		{
			name: "Basic config file parsing",
			yamlContent: `tracing:
  disable:
    - logging: true
`,
			useEnvVar:        false,
			expectedError:    false,
			expectedDisabled: []string{"logging"},
		},
		{
			name: "Config file parsing with environment variable",
			yamlContent: `tracing:
  disable:
    - logging: true
`,
			useEnvVar:        true,
			expectedError:    false,
			expectedDisabled: []string{"logging"},
		},
		{
			name: "Invalid YAML handling",
			yamlContent: `tracing:
  disable:
    - logging: true
  - invalid indentation
`,
			useEnvVar:        false,
			expectedError:    true,
			expectedDisabled: []string{},
		},
		{
			name: "Invalid config file path",
			yamlContent: `tracing:
  disable:
    - logging: true
`,
			useEnvVar:         true,
			invalidConfigPath: true,
			expectedError:     true,
			expectedDisabled:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, "config.yaml")

			err := os.WriteFile(configPath, []byte(tt.yamlContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create test config file: %v", err)
			}

			if tt.useEnvVar {
				if tt.invalidConfigPath {
					t.Setenv("INSTANA_CONFIG_PATH", "/invalid/path")
				} else {
					t.Setenv("INSTANA_CONFIG_PATH", configPath)
				}

				opts := &Options{
					Tracer: TracerOptions{},
				}
				opts.applyConfiguration()

				verifyDisabledCategories(t, opts.Tracer.DisableSpans, tt.expectedDisabled)
			} else {
				opts := &TracerOptions{
					DisableSpans: make(map[string]bool),
				}

				err = parseConfigFile(configPath, opts)

				if (err != nil) != tt.expectedError {
					if tt.expectedError {
						t.Error("Expected an error, but didn't get one")
					} else {
						t.Errorf("Got unexpected error: %v", err)
					}
				}

				// Only verify disabled categories if no error was expected
				if !tt.expectedError {
					verifyDisabledCategories(t, opts.DisableSpans, tt.expectedDisabled)
				}
			}
		})
	}
}

// verifyDisabledCategories checks that expected categories are disabled
func verifyDisabledCategories(t *testing.T, disableMap map[string]bool, expectedDisabled []string) {
	t.Helper()

	for _, category := range expectedDisabled {
		if !disableMap[category] {
			t.Errorf("Expected category %q to be disabled, but it wasn't", category)
		}
	}
}

// TestApplyBasicDefaults tests the default values for basic options
func TestApplyBasicDefaults(t *testing.T) {
	tests := []struct {
		name                        string
		maxBufferedSpans            int
		forceTransmissionStartingAt int
		expectedMaxBufferedSpans    int
		expectedForceTransmission   int
	}{
		{
			name:                        "Both zero - apply defaults",
			maxBufferedSpans:            0,
			forceTransmissionStartingAt: 0,
			expectedMaxBufferedSpans:    DefaultMaxBufferedSpans,
			expectedForceTransmission:   DefaultForceSpanSendAt,
		},
		{
			name:                        "Custom values - keep them",
			maxBufferedSpans:            500,
			forceTransmissionStartingAt: 250,
			expectedMaxBufferedSpans:    500,
			expectedForceTransmission:   250,
		},
		{
			name:                        "Only MaxBufferedSpans set",
			maxBufferedSpans:            800,
			forceTransmissionStartingAt: 0,
			expectedMaxBufferedSpans:    800,
			expectedForceTransmission:   DefaultForceSpanSendAt,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &Options{
				MaxBufferedSpans:            tt.maxBufferedSpans,
				ForceTransmissionStartingAt: tt.forceTransmissionStartingAt,
			}

			opts.applyBasicDefaults()

			assert.Equal(t, tt.expectedMaxBufferedSpans, opts.MaxBufferedSpans)
			assert.Equal(t, tt.expectedForceTransmission, opts.ForceTransmissionStartingAt)
		})
	}
}

// TestApplyAgentConfiguration tests agent host and port configuration precedence
func TestApplyAgentConfiguration(t *testing.T) {
	tests := []struct {
		name         string
		inCodeHost   string
		inCodePort   int
		envHost      string
		envPort      string
		expectedHost string
		expectedPort int
	}{
		{
			name:         "No config - use defaults",
			inCodeHost:   "",
			inCodePort:   0,
			envHost:      "",
			envPort:      "",
			expectedHost: agentDefaultHost,
			expectedPort: agentDefaultPort,
		},
		{
			name:         "In-code only",
			inCodeHost:   "custom-host",
			inCodePort:   12345,
			envHost:      "",
			envPort:      "",
			expectedHost: "custom-host",
			expectedPort: 12345,
		},
		{
			name:         "ENV overrides in-code",
			inCodeHost:   "custom-host",
			inCodePort:   12345,
			envHost:      "env-host",
			envPort:      "54321",
			expectedHost: "env-host",
			expectedPort: 54321,
		},
		{
			name:         "ENV overrides default",
			inCodeHost:   "",
			inCodePort:   0,
			envHost:      "env-host",
			envPort:      "54321",
			expectedHost: "env-host",
			expectedPort: 54321,
		},
		{
			name:         "Invalid ENV port - keep in-code",
			inCodeHost:   "custom-host",
			inCodePort:   12345,
			envHost:      "env-host",
			envPort:      "invalid",
			expectedHost: "env-host",
			expectedPort: 12345,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment
			hostRestore := restoreEnvVarFunc("INSTANA_AGENT_HOST")
			portRestore := restoreEnvVarFunc("INSTANA_AGENT_PORT")
			defer hostRestore()
			defer portRestore()

			if tt.envHost != "" {
				os.Setenv("INSTANA_AGENT_HOST", tt.envHost)
			} else {
				os.Unsetenv("INSTANA_AGENT_HOST")
			}

			if tt.envPort != "" {
				os.Setenv("INSTANA_AGENT_PORT", tt.envPort)
			} else {
				os.Unsetenv("INSTANA_AGENT_PORT")
			}

			opts := &Options{
				AgentHost: tt.inCodeHost,
				AgentPort: tt.inCodePort,
			}

			opts.applyAgentConfiguration()

			assert.Equal(t, tt.expectedHost, opts.AgentHost)
			assert.Equal(t, tt.expectedPort, opts.AgentPort)
		})
	}
}

// TestApplyServiceConfiguration tests service name configuration precedence
func TestApplyServiceConfiguration(t *testing.T) {
	tests := []struct {
		name            string
		inCodeService   string
		envService      string
		expectedService string
	}{
		{
			name:            "No config - empty service",
			inCodeService:   "",
			envService:      "",
			expectedService: "",
		},
		{
			name:            "In-code only",
			inCodeService:   "my-service",
			envService:      "",
			expectedService: "my-service",
		},
		{
			name:            "ENV overrides in-code",
			inCodeService:   "my-service",
			envService:      "env-service",
			expectedService: "env-service",
		},
		{
			name:            "ENV set but empty - overrides in-code",
			inCodeService:   "my-service",
			envService:      "",
			expectedService: "my-service",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			restore := restoreEnvVarFunc("INSTANA_SERVICE_NAME")
			defer restore()

			if tt.envService != "" {
				os.Setenv("INSTANA_SERVICE_NAME", tt.envService)
			} else {
				os.Unsetenv("INSTANA_SERVICE_NAME")
			}

			opts := &Options{
				Service: tt.inCodeService,
			}

			opts.applyServiceConfiguration()

			assert.Equal(t, tt.expectedService, opts.Service)
		})
	}
}

// TestApplyProfilingConfiguration tests profiling configuration precedence
func TestApplyProfilingConfiguration(t *testing.T) {
	tests := []struct {
		name                  string
		inCodeEnableProfile   bool
		envAutoProfile        string
		expectedEnableProfile bool
	}{
		{
			name:                  "No config - disabled",
			inCodeEnableProfile:   false,
			envAutoProfile:        "",
			expectedEnableProfile: false,
		},
		{
			name:                  "In-code enabled",
			inCodeEnableProfile:   true,
			envAutoProfile:        "",
			expectedEnableProfile: true,
		},
		{
			name:                  "ENV set - enables profiling",
			inCodeEnableProfile:   false,
			envAutoProfile:        "1",
			expectedEnableProfile: true,
		},
		{
			name:                  "ENV set with any value - enables profiling",
			inCodeEnableProfile:   false,
			envAutoProfile:        "true",
			expectedEnableProfile: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			restore := restoreEnvVarFunc("INSTANA_AUTO_PROFILE")
			defer restore()

			if tt.envAutoProfile != "" {
				os.Setenv("INSTANA_AUTO_PROFILE", tt.envAutoProfile)
			} else {
				os.Unsetenv("INSTANA_AUTO_PROFILE")
			}

			opts := &Options{
				EnableAutoProfile: tt.inCodeEnableProfile,
			}

			opts.applyProfilingConfiguration()

			assert.Equal(t, tt.expectedEnableProfile, opts.EnableAutoProfile)
		})
	}
}

// TestApplySecretsConfiguration tests secrets matcher configuration precedence
func TestApplySecretsConfiguration(t *testing.T) {
	tests := []struct {
		name                   string
		inCodeSecrets          Matcher
		envSecrets             string
		expectedMatchSecret    bool
		expectedMatchPassword  bool
		expectedDefaultSecrets bool
	}{
		{
			name:                   "No config - use default",
			inCodeSecrets:          nil,
			envSecrets:             "",
			expectedMatchSecret:    true,
			expectedMatchPassword:  true,
			expectedDefaultSecrets: true,
		},
		{
			name:                   "In-code custom matcher",
			inCodeSecrets:          testMatcher{},
			envSecrets:             "",
			expectedMatchSecret:    false,
			expectedMatchPassword:  false,
			expectedDefaultSecrets: false,
		},
		{
			name:                   "ENV overrides in-code",
			inCodeSecrets:          testMatcher{},
			envSecrets:             "contains-ignore-case:secret,password",
			expectedMatchSecret:    true,
			expectedMatchPassword:  true,
			expectedDefaultSecrets: false,
		},
		{
			name:                   "Invalid ENV - falls back to in-code",
			inCodeSecrets:          testMatcher{},
			envSecrets:             "invalid-format",
			expectedMatchSecret:    false,
			expectedMatchPassword:  false,
			expectedDefaultSecrets: false,
		},
		{
			name:                   "Invalid ENV - falls back to default when no in-code",
			inCodeSecrets:          nil,
			envSecrets:             "invalid-format",
			expectedMatchSecret:    true,
			expectedMatchPassword:  true,
			expectedDefaultSecrets: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			restore := restoreEnvVarFunc("INSTANA_SECRETS")
			defer restore()

			if tt.envSecrets != "" {
				os.Setenv("INSTANA_SECRETS", tt.envSecrets)
			} else {
				os.Unsetenv("INSTANA_SECRETS")
			}

			opts := &Options{
				Tracer: TracerOptions{
					Secrets: tt.inCodeSecrets,
				},
			}

			opts.applySecretsConfiguration()

			assert.Equal(t, tt.expectedMatchSecret, opts.Tracer.Secrets.Match("secret"))
			assert.Equal(t, tt.expectedMatchPassword, opts.Tracer.Secrets.Match("password"))
			assert.Equal(t, tt.expectedDefaultSecrets, opts.Tracer.tracerDefaultSecrets)
		})
	}
}

// TestApplyHTTPHeadersConfiguration tests HTTP headers configuration precedence
func TestApplyHTTPHeadersConfiguration(t *testing.T) {
	tests := []struct {
		name            string
		inCodeHeaders   []string
		envHeaders      string
		expectedHeaders []string
	}{
		{
			name:            "No config - empty",
			inCodeHeaders:   nil,
			envHeaders:      "",
			expectedHeaders: nil,
		},
		{
			name:            "In-code only",
			inCodeHeaders:   []string{"X-Custom-1", "X-Custom-2"},
			envHeaders:      "",
			expectedHeaders: []string{"X-Custom-1", "X-Custom-2"},
		},
		{
			name:            "ENV overrides in-code",
			inCodeHeaders:   []string{"X-Custom-1", "X-Custom-2"},
			envHeaders:      "X-Env-1;X-Env-2",
			expectedHeaders: []string{"X-Env-1", "X-Env-2"},
		},
		{
			name:            "ENV with spaces",
			inCodeHeaders:   nil,
			envHeaders:      " X-Header-1 ; X-Header-2 ",
			expectedHeaders: []string{"X-Header-1", "X-Header-2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			restore := restoreEnvVarFunc("INSTANA_EXTRA_HTTP_HEADERS")
			defer restore()

			if tt.envHeaders != "" {
				os.Setenv("INSTANA_EXTRA_HTTP_HEADERS", tt.envHeaders)
			} else {
				os.Unsetenv("INSTANA_EXTRA_HTTP_HEADERS")
			}

			opts := &Options{
				Tracer: TracerOptions{
					CollectableHTTPHeaders: tt.inCodeHeaders,
				},
			}

			opts.applyHTTPHeadersConfiguration()

			assert.Equal(t, tt.expectedHeaders, opts.Tracer.CollectableHTTPHeaders)
		})
	}
}

// TestApplyW3CConfiguration tests W3C trace correlation configuration
func TestApplyW3CConfiguration(t *testing.T) {
	tests := []struct {
		name               string
		envDisableW3C      string
		expectedDisableW3C bool
	}{
		{
			name:               "Not set - W3C enabled",
			envDisableW3C:      "",
			expectedDisableW3C: false,
		},
		{
			name:               "Set to any value - W3C disabled",
			envDisableW3C:      "1",
			expectedDisableW3C: true,
		},
		{
			name:               "Set to true - W3C disabled",
			envDisableW3C:      "true",
			expectedDisableW3C: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			restore := restoreEnvVarFunc("INSTANA_DISABLE_W3C_TRACE_CORRELATION")
			defer restore()

			if tt.envDisableW3C != "" {
				os.Setenv("INSTANA_DISABLE_W3C_TRACE_CORRELATION", tt.envDisableW3C)
			} else {
				os.Unsetenv("INSTANA_DISABLE_W3C_TRACE_CORRELATION")
			}

			opts := &Options{}

			opts.applyW3CConfiguration()

			assert.Equal(t, tt.expectedDisableW3C, opts.disableW3CTraceCorrelation)
		})
	}
}

// TestApplyConfiguration_FullIntegration tests the complete configuration flow
func TestApplyConfiguration_FullIntegration(t *testing.T) {
	// Setup environment
	hostRestore := restoreEnvVarFunc("INSTANA_AGENT_HOST")
	portRestore := restoreEnvVarFunc("INSTANA_AGENT_PORT")
	serviceRestore := restoreEnvVarFunc("INSTANA_SERVICE_NAME")
	secretsRestore := restoreEnvVarFunc("INSTANA_SECRETS")
	headersRestore := restoreEnvVarFunc("INSTANA_EXTRA_HTTP_HEADERS")
	profileRestore := restoreEnvVarFunc("INSTANA_AUTO_PROFILE")
	disableRestore := restoreEnvVarFunc("INSTANA_TRACING_DISABLE")
	w3cRestore := restoreEnvVarFunc("INSTANA_DISABLE_W3C_TRACE_CORRELATION")

	defer hostRestore()
	defer portRestore()
	defer serviceRestore()
	defer secretsRestore()
	defer headersRestore()
	defer profileRestore()
	defer disableRestore()
	defer w3cRestore()

	// Set environment variables
	os.Setenv("INSTANA_AGENT_HOST", "env-host")
	os.Setenv("INSTANA_AGENT_PORT", "9999")
	os.Setenv("INSTANA_SERVICE_NAME", "env-service")
	os.Setenv("INSTANA_SECRETS", "contains:secret,password")
	os.Setenv("INSTANA_EXTRA_HTTP_HEADERS", "X-Env-Header")
	os.Setenv("INSTANA_AUTO_PROFILE", "1")
	os.Setenv("INSTANA_TRACING_DISABLE", "logging")
	os.Setenv("INSTANA_DISABLE_W3C_TRACE_CORRELATION", "1")

	// Create options with in-code configuration
	opts := &Options{
		AgentHost:         "code-host",
		AgentPort:         8888,
		Service:           "code-service",
		EnableAutoProfile: false,
		Tracer: TracerOptions{
			Secrets:                testMatcher{},
			CollectableHTTPHeaders: []string{"X-Code-Header"},
			DisableSpans: map[string]bool{
				"database": true,
			},
		},
	}

	// Apply configuration
	opts.applyConfiguration()

	// Verify ENV overrides in-code
	assert.Equal(t, "env-host", opts.AgentHost, "ENV should override in-code for AgentHost")
	assert.Equal(t, 9999, opts.AgentPort, "ENV should override in-code for AgentPort")
	assert.Equal(t, "env-service", opts.Service, "ENV should override in-code for Service")
	assert.True(t, opts.EnableAutoProfile, "ENV should enable auto profile")
	assert.True(t, opts.disableW3CTraceCorrelation, "ENV should disable W3C")

	// Verify secrets - ENV overrides in-code
	assert.True(t, opts.Tracer.Secrets.Match("secret"), "ENV secrets should override in-code")
	assert.False(t, opts.Tracer.Secrets.Match("testing_matcher"), "In-code matcher should be overridden")

	// Verify HTTP headers - ENV overrides in-code
	assert.Equal(t, []string{"X-Env-Header"}, opts.Tracer.CollectableHTTPHeaders)

	// Verify tracing disable - ENV adds to in-code (both should be present)
	assert.True(t, opts.Tracer.DisableSpans["logging"], "ENV should set logging disabled")
	assert.True(t, opts.Tracer.DisableSpans["database"], "In-code disable should still be present")

	// Verify defaults are applied
	assert.Equal(t, DefaultMaxBufferedSpans, opts.MaxBufferedSpans)
	assert.Equal(t, DefaultForceSpanSendAt, opts.ForceTransmissionStartingAt)
}

// TestApplyConfiguration_DefaultsOnly tests configuration with only defaults
func TestApplyConfiguration_DefaultsOnly(t *testing.T) {
	// Clear all environment variables
	hostRestore := restoreEnvVarFunc("INSTANA_AGENT_HOST")
	portRestore := restoreEnvVarFunc("INSTANA_AGENT_PORT")
	serviceRestore := restoreEnvVarFunc("INSTANA_SERVICE_NAME")
	secretsRestore := restoreEnvVarFunc("INSTANA_SECRETS")
	headersRestore := restoreEnvVarFunc("INSTANA_EXTRA_HTTP_HEADERS")
	profileRestore := restoreEnvVarFunc("INSTANA_AUTO_PROFILE")
	disableRestore := restoreEnvVarFunc("INSTANA_TRACING_DISABLE")

	defer hostRestore()
	defer portRestore()
	defer serviceRestore()
	defer secretsRestore()
	defer headersRestore()
	defer profileRestore()
	defer disableRestore()

	os.Unsetenv("INSTANA_AGENT_HOST")
	os.Unsetenv("INSTANA_AGENT_PORT")
	os.Unsetenv("INSTANA_SERVICE_NAME")
	os.Unsetenv("INSTANA_SECRETS")
	os.Unsetenv("INSTANA_EXTRA_HTTP_HEADERS")
	os.Unsetenv("INSTANA_AUTO_PROFILE")
	os.Unsetenv("INSTANA_TRACING_DISABLE")

	opts := &Options{}

	opts.applyConfiguration()

	// Verify defaults
	assert.Equal(t, agentDefaultHost, opts.AgentHost)
	assert.Equal(t, agentDefaultPort, opts.AgentPort)
	assert.Equal(t, "", opts.Service)
	assert.False(t, opts.EnableAutoProfile)
	assert.Equal(t, DefaultMaxBufferedSpans, opts.MaxBufferedSpans)
	assert.Equal(t, DefaultForceSpanSendAt, opts.ForceTransmissionStartingAt)

	// Verify default secrets matcher
	assert.True(t, opts.Tracer.tracerDefaultSecrets)
	assert.True(t, opts.Tracer.Secrets.Match("secret"))
	assert.True(t, opts.Tracer.Secrets.Match("password"))

	// Verify empty collections
	assert.Nil(t, opts.Tracer.CollectableHTTPHeaders)
	assert.Nil(t, opts.Tracer.DisableSpans)
}

// TestApplyMetricsConfiguration_Default tests default metrics transmission delay
func TestApplyMetricsConfiguration_Default(t *testing.T) {
	restore := restoreEnvVarFunc("INSTANA_METRICS_TRANSMISSION_DELAY")
	defer restore()
	os.Unsetenv("INSTANA_METRICS_TRANSMISSION_DELAY")

	opts := &Options{}
	opts.applyMetricsConfiguration()

	assert.Equal(t, 1000, opts.Metrics.TransmissionDelay, "Default should be 1000ms")
}

// TestApplyMetricsConfiguration_ValidENV tests valid ENV override
func TestApplyMetricsConfiguration_ENVValidation(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected int
	}{
		// Valid values
		{
			name:     "Valid 1000ms (min boundary)",
			envValue: "1000",
			expected: 1000,
		},
		{
			name:     "Valid 2000ms (mid-range)",
			envValue: "2000",
			expected: 2000,
		},
		{
			name:     "Valid 5000ms (max boundary)",
			envValue: "5000",
			expected: 5000,
		},
		// Invalid values - fall back to default
		{
			name:     "Non-numeric value",
			envValue: "invalid",
			expected: 1000,
		},
		{
			name:     "Empty string",
			envValue: "",
			expected: 1000,
		},
		{
			name:     "Float value",
			envValue: "1500.5",
			expected: 1000,
		},
		// Below minimum - enforced to minimum
		{
			name:     "Zero value",
			envValue: "0",
			expected: 1000,
		},
		{
			name:     "Negative value",
			envValue: "-500",
			expected: 1000,
		},
		{
			name:     "Below minimum (999ms)",
			envValue: "999",
			expected: 1000,
		},
		// Above maximum - capped at maximum
		{
			name:     "Above maximum (6000ms)",
			envValue: "6000",
			expected: 5000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			restore := restoreEnvVarFunc("INSTANA_METRICS_TRANSMISSION_DELAY")
			defer restore()

			if tt.envValue != "" {
				os.Setenv("INSTANA_METRICS_TRANSMISSION_DELAY", tt.envValue)
			} else {
				os.Unsetenv("INSTANA_METRICS_TRANSMISSION_DELAY")
			}

			opts := &Options{}
			opts.applyMetricsConfiguration()

			assert.Equal(t, tt.expected, opts.Metrics.TransmissionDelay)
		})
	}
}

// TestApplyMetricsConfiguration_ENVPrecedence tests ENV overrides code configuration
func TestApplyMetricsConfiguration_ENVPrecedence(t *testing.T) {
	tests := []struct {
		name         string
		inCodeValue  int
		envValue     string
		expectedCode int
		expectedENV  int
	}{
		{
			name:         "ENV overrides in-code",
			inCodeValue:  2000,
			envValue:     "3000",
			expectedCode: 2000, // Without ENV
			expectedENV:  3000, // With ENV
		},
		{
			name:         "ENV overrides default",
			inCodeValue:  0, // Will use default 1000
			envValue:     "2500",
			expectedCode: 1000, // Default
			expectedENV:  2500, // ENV override
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			restore := restoreEnvVarFunc("INSTANA_METRICS_TRANSMISSION_DELAY")
			defer restore()

			// Test without ENV
			os.Unsetenv("INSTANA_METRICS_TRANSMISSION_DELAY")
			opts := &Options{
				Metrics: MetricsOptions{
					TransmissionDelay: tt.inCodeValue,
				},
			}
			opts.applyMetricsConfiguration()
			assert.Equal(t, tt.expectedCode, opts.Metrics.TransmissionDelay, "In-code value should be used without ENV")

			// Test with ENV
			os.Setenv("INSTANA_METRICS_TRANSMISSION_DELAY", tt.envValue)
			opts = &Options{
				Metrics: MetricsOptions{
					TransmissionDelay: tt.inCodeValue,
				},
			}
			opts.applyMetricsConfiguration()
			assert.Equal(t, tt.expectedENV, opts.Metrics.TransmissionDelay, "ENV should override in-code value")
		})
	}
}

// TestApplyMetricsConfiguration_CodeConfiguration tests code-based configuration
func TestApplyMetricsConfiguration_CodeConfiguration(t *testing.T) {
	tests := []struct {
		name        string
		inCodeValue int
		expected    int
		description string
	}{
		{
			name:        "Valid code value 2000ms",
			inCodeValue: 2000,
			expected:    2000,
			description: "Should use in-code value",
		},
		{
			name:        "Code value below minimum (500ms)",
			inCodeValue: 500,
			expected:    1000,
			description: "Should enforce minimum 1000ms",
		},
		{
			name:        "Code value at max 5000ms",
			inCodeValue: 5000,
			expected:    5000,
			description: "Should use max value",
		},
		{
			name:        "Code value above max",
			inCodeValue: 6000,
			expected:    5000,
			description: "Should cap at 5000ms",
		},
		{
			name:        "Zero value uses default",
			inCodeValue: 0,
			expected:    1000,
			description: "Should use default 1000ms",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			restore := restoreEnvVarFunc("INSTANA_METRICS_TRANSMISSION_DELAY")
			defer restore()
			os.Unsetenv("INSTANA_METRICS_TRANSMISSION_DELAY")

			opts := &Options{
				Metrics: MetricsOptions{
					TransmissionDelay: tt.inCodeValue,
				},
			}
			opts.applyMetricsConfiguration()

			assert.Equal(t, tt.expected, opts.Metrics.TransmissionDelay, tt.description)
		})
	}
}

// TestApplyMetricsConfiguration_BackwardCompatibility tests backward compatibility
func TestApplyMetricsConfiguration_BackwardCompatibility(t *testing.T) {
	tests := []struct {
		name        string
		setupOpts   func() *Options
		expected    int
		description string
	}{
		{
			name: "Empty Options struct",
			setupOpts: func() *Options {
				return &Options{}
			},
			expected:    1000,
			description: "Should use default 1000ms",
		},
		{
			name: "Options with other fields set",
			setupOpts: func() *Options {
				return &Options{
					AgentHost: "localhost",
					AgentPort: 42699,
					Service:   "test-service",
				}
			},
			expected:    1000,
			description: "Should use default 1000ms when Metrics not set",
		},
		{
			name: "DefaultOptions()",
			setupOpts: func() *Options {
				return DefaultOptions()
			},
			expected:    1000,
			description: "DefaultOptions should result in 1000ms",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			restore := restoreEnvVarFunc("INSTANA_METRICS_TRANSMISSION_DELAY")
			defer restore()
			os.Unsetenv("INSTANA_METRICS_TRANSMISSION_DELAY")

			opts := tt.setupOpts()
			opts.applyMetricsConfiguration()

			assert.Equal(t, tt.expected, opts.Metrics.TransmissionDelay, tt.description)
		})
	}
}

// BenchmarkApplyMetricsConfiguration benchmarks the configuration overhead
func BenchmarkApplyMetricsConfiguration(b *testing.B) {
	restore := restoreEnvVarFunc("INSTANA_METRICS_TRANSMISSION_DELAY")
	defer restore()

	benchmarks := []struct {
		name     string
		envValue string
	}{
		{
			name:     "Default (no ENV)",
			envValue: "",
		},
		{
			name:     "Valid ENV",
			envValue: "2000",
		},
		{
			name:     "Invalid ENV",
			envValue: "invalid",
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			if bm.envValue != "" {
				os.Setenv("INSTANA_METRICS_TRANSMISSION_DELAY", bm.envValue)
			} else {
				os.Unsetenv("INSTANA_METRICS_TRANSMISSION_DELAY")
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				opts := &Options{}
				opts.applyMetricsConfiguration()
			}
		})
	}
}

// TestApplyConfiguration_WithMetrics tests full integration with applyConfiguration
func TestApplyConfiguration_WithMetrics(t *testing.T) {
	restore := restoreEnvVarFunc("INSTANA_METRICS_TRANSMISSION_DELAY")
	defer restore()

	tests := []struct {
		name        string
		envValue    string
		inCodeValue int
		expected    int
	}{
		{
			name:        "Full integration - ENV override",
			envValue:    "3000",
			inCodeValue: 2000,
			expected:    3000,
		},
		{
			name:        "Full integration - code only",
			envValue:    "",
			inCodeValue: 2500,
			expected:    2500,
		},
		{
			name:        "Full integration - default",
			envValue:    "",
			inCodeValue: 0,
			expected:    1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv("INSTANA_METRICS_TRANSMISSION_DELAY", tt.envValue)
			} else {
				os.Unsetenv("INSTANA_METRICS_TRANSMISSION_DELAY")
			}

			opts := &Options{
				Metrics: MetricsOptions{
					TransmissionDelay: tt.inCodeValue,
				},
			}

			// Call the full applyConfiguration which should call applyMetricsConfiguration
			opts.applyConfiguration()

			assert.Equal(t, tt.expected, opts.Metrics.TransmissionDelay)
		})
	}
}
