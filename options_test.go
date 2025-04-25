// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instana

import (
	"os"
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
	assert.Equal(t, false, testOpts.Tracer.agentOverrideSecrets)

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
	assert.Equal(t, false, testOpts.Tracer.agentOverrideSecrets)

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
	assert.Equal(t, true, testOpts.Tracer.agentOverrideSecrets)

	assert.Equal(t, 0, len(testOpts.Tracer.CollectableHTTPHeaders))

}

func restoreEnvVarFunc(key string) func() {
	if oldValue, ok := os.LookupEnv(key); ok {
		return func() { os.Setenv(key, oldValue) }
	}

	return func() { os.Unsetenv(key) }
}
