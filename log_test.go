// (c) Copyright IBM Corp. 2026

package instana

import (
	"sync"
	"testing"

	"github.com/instana/go-sensor/logger"
	"github.com/stretchr/testify/assert"
)

// mockLogger is a test implementation of LeveledLogger
type mockLogger struct {
	mu           sync.Mutex
	debugCalls   []string
	infoCalls    []string
	warnCalls    []string
	errorCalls   []string
	debugEnabled bool
	infoEnabled  bool
	warnEnabled  bool
	errorEnabled bool
}

func newMockLogger() *mockLogger {
	return &mockLogger{
		debugEnabled: true,
		infoEnabled:  true,
		warnEnabled:  true,
		errorEnabled: true,
	}
}

func (m *mockLogger) Debug(v ...interface{}) {
	if !m.debugEnabled {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.debugCalls = append(m.debugCalls, formatArgs(v...))
}

func (m *mockLogger) Info(v ...interface{}) {
	if !m.infoEnabled {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.infoCalls = append(m.infoCalls, formatArgs(v...))
}

func (m *mockLogger) Warn(v ...interface{}) {
	if !m.warnEnabled {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.warnCalls = append(m.warnCalls, formatArgs(v...))
}

func (m *mockLogger) Error(v ...interface{}) {
	if !m.errorEnabled {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errorCalls = append(m.errorCalls, formatArgs(v...))
}

func formatArgs(v ...interface{}) string {
	result := ""
	for i, arg := range v {
		if i > 0 {
			result += " "
		}
		result += toString(arg)
	}
	return result
}

func toString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	default:
		return ""
	}
}

func TestLogLevelConstants(t *testing.T) {
	tests := []struct {
		name     string
		level    int
		expected int
	}{
		{"Error level", Error, 0},
		{"Warn level", Warn, 1},
		{"Info level", Info, 2},
		{"Debug level", Debug, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.level)
		})
	}
}

func TestSetLogger(t *testing.T) {
	// Save original logger and restore after test
	originalLogger := defaultLogger
	defer func() {
		defaultLogger = originalLogger
	}()

	t.Run("sets default logger", func(t *testing.T) {
		mock := newMockLogger()
		SetLogger(mock)

		assert.Equal(t, mock, defaultLogger)
	})

	t.Run("updates sensor logger when sensor is initialized", func(t *testing.T) {
		// Save original sensor and restore after test
		muSensor.Lock()
		originalSensor := sensor
		sensor = &sensorS{
			logger: logger.New(nil),
		}
		muSensor.Unlock()

		defer func() {
			muSensor.Lock()
			sensor = originalSensor
			muSensor.Unlock()
		}()

		mock := newMockLogger()
		SetLogger(mock)

		muSensor.RLock()
		assert.Equal(t, mock, sensor.logger)
		muSensor.RUnlock()
	})

	t.Run("does not panic when sensor is nil", func(t *testing.T) {
		// Save original sensor and restore after test
		muSensor.Lock()
		originalSensor := sensor
		sensor = nil
		muSensor.Unlock()

		defer func() {
			muSensor.Lock()
			sensor = originalSensor
			muSensor.Unlock()
		}()

		mock := newMockLogger()
		assert.NotPanics(t, func() {
			SetLogger(mock)
		})
	})
}

func TestSetLogLevel_AllLevels(t *testing.T) {
	// Test that all defined log levels map correctly
	levelMappings := map[int]logger.Level{
		Error: logger.ErrorLevel,
		Warn:  logger.WarnLevel,
		Info:  logger.InfoLevel,
		Debug: logger.DebugLevel,
	}

	for instanaLevel, expectedLoggerLevel := range levelMappings {
		t.Run("level_"+string(rune(instanaLevel+'0')), func(t *testing.T) {
			l := logger.New(nil)
			setLogLevel(l, instanaLevel)
			assert.Equal(t, expectedLoggerLevel, l.GetLevel())
		})
	}
}

func TestLeveledLoggerInterface(t *testing.T) {
	t.Run("mock logger implements LeveledLogger", func(t *testing.T) {
		var _ LeveledLogger = (*mockLogger)(nil)
	})

	t.Run("logger.Logger implements LeveledLogger", func(t *testing.T) {
		var _ LeveledLogger = (*logger.Logger)(nil)
	})
}

func TestLoggerIntegration(t *testing.T) {
	t.Run("SetLogger and setLogLevel work together", func(t *testing.T) {
		// Save original logger
		originalLogger := defaultLogger
		defer func() {
			defaultLogger = originalLogger
		}()

		// Create a new logger and set it as default
		l := logger.New(nil)
		SetLogger(l)

		// Set log level
		setLogLevel(l, Info)

		// Verify the level is set correctly
		assert.Equal(t, logger.InfoLevel, l.GetLevel())
	})
}
