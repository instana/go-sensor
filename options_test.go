// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instana

import (
	"fmt"
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
	}{
		{
			name:     "No env var",
			envValue: "",
			expected: map[string]bool{},
		},
		{
			name:     "DisableSpans all",
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
				opts.setDefaults()

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

func TestValidateFile(t *testing.T) {
	tempDir := t.TempDir()

	validFilePath := filepath.Join(tempDir, "valid.txt")
	err := os.WriteFile(validFilePath, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name          string
		getPathFn     func() (string, error)
		expectedError bool
		errorContains string
	}{
		{
			name: "Valid file",
			getPathFn: func() (string, error) {
				return validFilePath, nil
			},
			expectedError: false,
		},
		{
			name: "Non-existent file",
			getPathFn: func() (string, error) {
				return filepath.Join(tempDir, "nonexistent.txt"), nil
			},
			expectedError: true,
			errorContains: "no such file or directory",
		},
		{
			name: "Symlink to valid file",
			getPathFn: func() (string, error) {
				symlinkPath := filepath.Join(tempDir, "symlink.txt")
				err = os.Symlink(validFilePath, symlinkPath)
				if err != nil {
					return "", fmt.Errorf("Skipping symlink test, could not create symlink: %v", err)
				}
				return symlinkPath, nil
			},
			expectedError: false,
		},
		{
			name: "Directory instead of file",
			getPathFn: func() (string, error) {
				dirPath := filepath.Join(tempDir, "testdir")
				err = os.Mkdir(dirPath, 0755)
				if err != nil {
					return "", fmt.Errorf("Failed to create test directory: %v", err)
				}
				return dirPath, nil
			},
			expectedError: true,
			errorContains: "not a regular file",
		},
		{
			name: "File too large",
			getPathFn: func() (string, error) {
				fpath := filepath.Join(tempDir, "big.conf")
				f, err := os.Create(fpath)
				if err != nil {
					return "", fmt.Errorf("Failed to create test file: %v", err)
				}
				defer f.Close()
				err = f.Truncate(1024*1024 + 1) // >1MB
				if err != nil {
					return "", fmt.Errorf("Failed to truncate test file: %v", err)
				}
				return fpath, nil
			},
			expectedError: true,
			errorContains: "config file too large",
		},
		{
			name: "World-readable file",
			getPathFn: func() (string, error) {
				worldReadablePath := filepath.Join(tempDir, "world-readable.txt")
				err = os.WriteFile(worldReadablePath, []byte("world-readable content"), 0644)
				if err != nil {
					return "", fmt.Errorf("Failed to create world-readable test file: %v", err)
				}
				return worldReadablePath, nil
			},
			expectedError: false, // This should not error, but will log a warning
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := tt.getPathFn()
			if err != nil {
				t.Skip(err)
			}
			absPath, err := validateFile(path)

			if (err != nil) != tt.expectedError {
				if tt.expectedError {
					t.Errorf("Expected error but got none")
				} else {
					t.Errorf("Expected no error but got: %v", err)
				}
			}

			// If am error is expected, check that it contains the expected text
			if tt.expectedError && err != nil && tt.errorContains != "" {
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Error message '%s' does not contain '%s'", err.Error(), tt.errorContains)
				}
			}

			// For successful cases, check that the returned path is absolute
			if !tt.expectedError && err == nil {
				if !filepath.IsAbs(absPath) {
					t.Errorf("Expected absolute path, got: %s", absPath)
				}

				// For the symlink case, verify it was resolved
				if tt.name == "Symlink to valid file" {
					// The resolved path should be an absolute path to the target file
					// Note: On macOS, /var/folders may resolve to /private/var/folders
					// so just check that the base filename matches
					if filepath.Base(absPath) != filepath.Base(validFilePath) {
						t.Errorf("Symlink not properly resolved. Got: %s, Expected file with basename: %s",
							absPath, filepath.Base(validFilePath))
					}
				}
			}
		})
	}
}
