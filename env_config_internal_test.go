// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instana

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/instana/go-sensor/secrets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseInstanaTags(t *testing.T) {
	examples := map[string]struct {
		Value    string
		Expected map[string]interface{}
	}{
		"empty":                   {"", nil},
		"single tag, empty key":   {"=value", nil},
		"single tag, no value":    {"key", map[string]interface{}{"key": nil}},
		"single tag, empty value": {"key=", map[string]interface{}{"key": ""}},
		"single tag, with value":  {"key=value", map[string]interface{}{"key": "value"}},
		"multiple tags, mixed": {
			`key1,  key2=  , key3   ="",key4=42`,
			map[string]interface{}{
				"key1": nil,
				"key2": "  ",
				"key3": `""`,
				"key4": "42",
			},
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, example.Expected, parseInstanaTags(example.Value))
		})
	}
}
func TestParseInstanaEmptySecrets(t *testing.T) {
	examples := map[string]struct {
		Value    string
		Expected error
	}{
		"empty": {"", errors.New("empty value for secret matcher configuration")},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			_, err := parseInstanaSecrets(example.Value)
			require.Error(t, err)
			assert.Equal(t, example.Expected, err)
		})
	}
}

func TestParseInstanaSecrets(t *testing.T) {
	regexMatcher, err := secrets.NewRegexpMatcher(regexp.MustCompile("a|b|c"), regexp.MustCompile("d"))
	require.NoError(t, err)

	examples := map[string]struct {
		Value    string
		Expected Matcher
	}{
		"equals":               {"equals:a,b,c", secrets.NewEqualsMatcher("a", "b", "c")},
		"equals-ignore-case":   {"equals-ignore-case:a,b,c", secrets.NewEqualsIgnoreCaseMatcher("a", "b", "c")},
		"contains":             {"contains:a,b,c", secrets.NewContainsMatcher("a", "b", "c")},
		"contains-ignore-case": {"contains-ignore-case:a,b,c", secrets.NewContainsIgnoreCaseMatcher("a", "b", "c")},
		"regexp":               {"regex:a|b|c,d", regexMatcher},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			m, err := parseInstanaSecrets(example.Value)
			require.NoError(t, err)
			assert.Equal(t, example.Expected, m)
		})
	}
}

func TestParseInstanaSecrets_Error(t *testing.T) {
	examples := map[string]string{
		"unknown matcher": "magic:pew,pew",
		"malformed":       "equals;a,b,c",
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			_, err := parseInstanaSecrets(example)
			assert.Error(t, err)
		})
	}
}

func TestParseInstanaExtraHTTPHeaders(t *testing.T) {
	examples := map[string]struct {
		Value    string
		Expected []string
	}{
		"empty":    {"", nil},
		"one":      {"a", []string{"a"}},
		"multiple": {"a; ;  b  ;c;", []string{"a", "b", "c"}},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, example.Expected, parseInstanaExtraHTTPHeaders(example.Value))
		})
	}
}

func TestParseInstanaTimeout(t *testing.T) {
	examples := map[string]struct {
		Value    string
		Expected time.Duration
	}{
		"empty":            {"", defaultServerlessTimeout},
		"positive integer": {"123", 123 * time.Millisecond},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			d, err := parseInstanaTimeout(example.Value)
			require.NoError(t, err)
			assert.Equal(t, example.Expected, d)
		})
	}
}

func TestParseInstanaTimeout_Error(t *testing.T) {
	examples := map[string]string{
		"non-number":       "twenty",
		"non-integer":      "12.5:",
		"zero":             "0",
		"negative integer": "-100",
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			_, err := parseInstanaTimeout(example)
			assert.Error(t, err)
		})
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
