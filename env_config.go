// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instana

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/stretchr/testify/assert/yaml"
)

// MaxEnvValueSize is the maximum size of the value of an environment variable.
const MaxEnvValueSize = 32 * 1024

// parseInstanaTags parses the tags string passed via INSTANA_TAGS.
// The tag string is a comma-separated list of keys optionally followed by an '=' character and a string value:
//
//	INSTANA_TAGS := key1[=value1][,key2[=value2],...]
//
// The leading and trailing space is truncated from key names, values are used as-is. If a key does not have
// value associated, it's considered to be nil.
func parseInstanaTags(s string) map[string]interface{} {
	tags := make(map[string]interface{})

	for _, tag := range strings.Split(s, ",") {
		kv := strings.SplitN(tag, "=", 2)

		k := strings.TrimSpace(kv[0])
		if k == "" {
			continue
		}

		var v interface{}
		if len(kv) > 1 {
			v = kv[1]
		}

		tags[k] = v
	}

	if len(tags) == 0 {
		return nil
	}

	return tags
}

// parseInstanaSecrets parses the tags string passed via INSTANA_SECRETS.
// The secrets matcher configuration string is expected to have the following format:
//
//	INSTANA_SECRETS := <matcher>:<secret>[,<secret>]
//
// Where `matcher` is one of:
// * `equals` - matches a string if it's contained in the secrets list
// * `equals-ignore-case` is a case-insensitive version of `equals`
// * `contains` matches a string if it contains any of the secrets list values
// * `contains-ignore-case` is a case-insensitive version of `contains`
// * `regex` matches a string if it fully matches any of the regular expressions provided in the secrets list
//
// This function returns an error if there is no matcher configuration provided.
func parseInstanaSecrets(s string) (Matcher, error) {
	if s == "" {
		return nil, errors.New("empty value for secret matcher configuration")
	}

	ind := strings.Index(s, ":")
	if ind < 0 {
		return nil, fmt.Errorf("malformed secret matcher configuration: %q", s)
	}

	matcher, config := strings.TrimSpace(s[:ind]), strings.Split(s[ind+1:], ",")

	return NamedMatcher(matcher, config)
}

// parseInstanaExtraHTTPHeaders parses the tags string passed via INSTANA_EXTRA_HTTP_HEADERS.
// The header names are expected to come in a semicolon-separated list:
//
//	INSTANA_EXTRA_HTTP_HEADERS := header1[;header2;...]
//
// Any leading and trailing whitespace characters will be trimmed from header names.
func parseInstanaExtraHTTPHeaders(s string) []string {
	var headers []string
	for _, h := range strings.Split(s, ";") {
		h = strings.TrimSpace(h)
		if h == "" {
			continue
		}

		headers = append(headers, h)
	}

	return headers
}

// parseInstanaTimeout parses the Instana backend connection timeout passed via INSTANA_TIMEOUT.
// The value is expected to be an integer number of milliseconds, greate than 0.
// This function returns the default timeout 500ms if provided with an empty string.
func parseInstanaTimeout(s string) (time.Duration, error) {
	if s == "" {
		return defaultServerlessTimeout, nil
	}

	ms, err := strconv.ParseUint(s, 10, 64)
	if err != nil || ms < 1 {
		return 0, fmt.Errorf("invalid timeout value: %q", s)
	}

	return time.Duration(ms) * time.Millisecond, nil
}

// parseInstanaTracingDisable processes the INSTANA_TRACING_DISABLE environment variable value
// and updates the TracerOptions.Disable map accordingly.
//
// When a list of category or type names is specified, those will be disabled.
//
// Example:
// INSTANA_TRACING_DISABLE="logging" - disables logging category
func parseInstanaTracingDisable(value string, opts *TracerOptions) {
	// Initialize the Disable map if it doesn't exist
	if opts.DisableSpans == nil {
		opts.DisableSpans = make(map[string]bool)
	}

	// Trim spaces from the value
	value = strings.TrimSpace(value)

	// if it's not a boolean value, process as a comma-separated list and disable each category.
	items := strings.Split(value, ",")
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item != "" {
			opts.DisableSpans[item] = true
		}
	}
}

// parseConfigFile reads and parses the YAML configuration file at the given path
// and updates the TracerOptions accordingly.
//
// The YAML file must follow this format:
// tracing:
//   disable:
//     - logging: true

func parseConfigFile(path string, opts *TracerOptions) error {
	// Validate the file path and security considerations
	absPath, err := validateFile(path)
	if err != nil {
		return fmt.Errorf("config file validation failed for %s: %w", path, err)
	}

	// Read the file with proper error handling
	data, err := os.ReadFile(absPath)
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

	if opts.DisableSpans == nil {
		opts.DisableSpans = make(map[string]bool)
	}

	// Add the categories configured in the YAML file to the Disable map
	for _, disableMap := range config.Tracing.Disable {
		for category, enabled := range disableMap {
			if enabled {
				opts.DisableSpans[category] = true
			}
		}

	}

	return nil
}

// validateFile ensures the given config file path is safe and usable.
// Security considerations:
// - Resolves symlinks to prevent symlink attacks
// - Ensures the path exists and is a regular file
// - Enforces a reasonable file size limit to avoid DoS
// - Warns if file permissions are too permissive (world-readable)
func validateFile(path string) (absPath string, err error) {
	// Resolve symlinks to avoid symlink attacks
	realPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		return absPath, fmt.Errorf("failed to resolve config file path: %w", err)
	}

	// Get absolute normalized path
	absPath, err = filepath.Abs(realPath)
	if err != nil {
		return absPath, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Check if the path exists and is a regular file
	fileInfo, err := os.Stat(absPath)
	if err != nil {
		return absPath, fmt.Errorf("failed to access config file: %w", err)
	}

	// Ensure it's a regular file, not a directory or special file
	if !fileInfo.Mode().IsRegular() {
		return absPath, fmt.Errorf("config path is not a regular file: %s", absPath)
	}

	// Enforce a maximum file size
	const maxFileSize = 1 * 1024 * 1024 // 1MB
	if fileInfo.Size() > maxFileSize {
		return absPath, fmt.Errorf("config file too large: %d bytes (max allowed: %d bytes)",
			fileInfo.Size(), maxFileSize)
	}

	// Warn if the file is world-readable (optional hardening)
	if fileInfo.Mode().Perm()&0004 != 0 {
		defaultLogger.Warn("config file is world-readable, consider restricting permissions: ", absPath)
	}

	return absPath, nil
}

// LookupValidatedEnv retrieves the value of the environment variable named by key.
// It validates if env value exceeds the configured MaxEnvValueSize limit.
// On success, it returns the variable's value.
func lookupValidatedEnv(key string) (string, bool) {
	envVal, ok := os.LookupEnv(key)
	if !ok {
		return "", false
	}

	if len(envVal) > MaxEnvValueSize {
		defaultLogger.Error(fmt.Errorf("value of %q exceeds safe limit (%d bytes)", key, MaxEnvValueSize))
		return "", false
	}

	return envVal, true
}
