// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instana

import (
	"os"
	"path/filepath"
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
	assert.Equal(t, &Options{
		AgentHost:                   "localhost",
		AgentPort:                   42699,
		MaxBufferedSpans:            DefaultMaxBufferedSpans,
		ForceTransmissionStartingAt: DefaultForceSpanSendAt,
		Tracer:                      DefaultTracerOptions(),
	}, DefaultOptions())
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

	testOpts.setDefaults()

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

	testOpts.setDefaults()

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

	testOpts.setDefaults()

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
			name:  "Boolean true",
			value: "True",
			expected: map[string]bool{
				"logging": true,
			},
		},
		{
			name:  "With extra spaces",
			value: "   logging  ",
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
			if len(opts.Disable) != len(tt.expected) {
				t.Errorf("Expected map size %d, got %d", len(tt.expected), len(opts.Disable))
			}

			// Check if all expected keys are present with correct values
			for k, v := range tt.expected {
				if opts.Disable[k] != v {
					t.Errorf("Expected %s to be %v, got %v", k, v, opts.Disable[k])
				}
			}

			// Check if there are no unexpected keys
			for k := range opts.Disable {
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
	}{
		{
			name:     "No env var",
			envValue: "",
			expected: map[string]bool{},
		},
		{
			name:     "Disable all",
			envValue: "True",
			expected: map[string]bool{
				"logging": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.envValue != "" {
				t.Setenv("INSTANA_TRACING_DISABLE", tt.envValue)
			}

			opts := DefaultOptions()

			// Check if the maps have the expected values
			for k, v := range tt.expected {
				if opts.Tracer.Disable[k] != v {
					t.Errorf("Expected %s to be %v, got %v", k, v, opts.Tracer.Disable[k])
				}
			}
		})
	}
}

func TestConfigFileHandling(t *testing.T) {
	tests := []struct {
		name             string
		yamlContent      string
		useEnvVar        bool
		expectedError    bool
		expectedDisabled []string
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
				t.Setenv("INSTANA_CONFIG_PATH", configPath)

				opts := &Options{
					Tracer: TracerOptions{},
				}
				opts.setDefaults()

				verifyDisabledCategories(t, opts.Tracer.Disable, tt.expectedDisabled)
			} else {
				opts := &TracerOptions{
					Disable: make(map[string]bool),
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
					verifyDisabledCategories(t, opts.Disable, tt.expectedDisabled)
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
