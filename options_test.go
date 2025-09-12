package instana

import (
	"os"
	"path/filepath"
	"testing"
)

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
				"http":      true,
				"rpc":       true,
				"messaging": true,
				"logging":   true,
				"graphql":   true,
				"databases": true,
			},
		},
		{
			name:  "Specific categories and types",
			value: "logging, redis, messaging",
			expected: map[string]bool{
				"logging":   true,
				"redis":     true,
				"messaging": true,
			},
		},
		{
			name:  "With extra spaces",
			value: "  http ,  databases  ",
			expected: map[string]bool{
				"http":      true,
				"databases": true,
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
				"http":      true,
				"rpc":       true,
				"messaging": true,
				"logging":   true,
				"databases": true,
			},
		},
		{
			name:     "Disable specific",
			envValue: "logging, redis, messaging",
			expected: map[string]bool{
				"logging":   true,
				"redis":     true,
				"messaging": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			defer os.Unsetenv("INSTANA_TRACING_DISABLE")

			if tt.envValue != "" {
				os.Setenv("INSTANA_TRACING_DISABLE", tt.envValue)
			}

			// Create options with defaults
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
		notDisabled      []string
	}{
		{
			name: "Basic config file parsing",
			yamlContent: `tracing:
  disable:
    - http
    - databases
    - messaging
`,
			useEnvVar:        false,
			expectedError:    false,
			expectedDisabled: []string{"http", "databases", "messaging"},
			notDisabled:      []string{"logging"},
		},
		{
			name: "Config file parsing with environment variable",
			yamlContent: `tracing:
  disable:
    - rpc
    - logging
`,
			useEnvVar:        true,
			expectedError:    false,
			expectedDisabled: []string{"rpc", "logging"},
			notDisabled:      []string{"http"},
		},
		{
			name: "Invalid YAML handling",
			yamlContent: `tracing:
  disable:
    - http
  - invalid indentation
`,
			useEnvVar:        false,
			expectedError:    true,
			expectedDisabled: []string{},
			notDisabled:      []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary YAML file
			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, "config.yaml")

			// Write test YAML content
			err := os.WriteFile(configPath, []byte(tt.yamlContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create test config file: %v", err)
			}

			if tt.useEnvVar {
				t.Setenv("INSTANA_CONFIG_PATH", configPath)

				// Create options and call setDefaults
				opts := &Options{
					Tracer: TracerOptions{},
				}
				opts.setDefaults()

				verifyDisabledCategories(t, opts.Tracer.Disable, tt.expectedDisabled, tt.notDisabled)
			} else {
				// Test parsing the config file directly
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
					verifyDisabledCategories(t, opts.Disable, tt.expectedDisabled, tt.notDisabled)
				}
			}
		})
	}
}

// verifyDisabledCategories checks that expected categories are disabled and others are not
func verifyDisabledCategories(t *testing.T, disableMap map[string]bool, expectedDisabled, notDisabled []string) {
	t.Helper()

	// Verify that the categories were added to the Disable map
	for _, category := range expectedDisabled {
		if !disableMap[category] {
			t.Errorf("Expected category %q to be disabled, but it wasn't", category)
		}
	}

	// Verify that other categories are not disabled
	for _, category := range notDisabled {
		if disableMap[category] {
			t.Errorf("Category %q should not be disabled", category)
		}
	}
}
