// (c) Copyright IBM Corp. 2023

package instana

import (
	"context"
	"errors"

	ot "github.com/opentracing/opentracing-go"
)

var (
	_                   TracerLogger = (*noopCollector)(nil)
	noopCollectorErrMsg string       = "collector not initialized. make sure to initialize the Collector. eg: instana.InitCollector"
	noopCollectorErr    error        = errors.New(noopCollectorErrMsg)
)

type noopCollector struct {
	l LeveledLogger
}

func newNoopCollector() TracerLogger {
	c := &noopCollector{
		l: defaultLogger,
	}
	return c
}

func (c *noopCollector) Extract(format interface{}, carrier interface{}) (ot.SpanContext, error) {
	c.l.Error(noopCollectorErrMsg)
	return nil, noopCollectorErr
}

func (c *noopCollector) Inject(sm ot.SpanContext, format interface{}, carrier interface{}) error {
	c.l.Error(noopCollectorErrMsg)
	return noopCollectorErr
}

func (c *noopCollector) StartSpan(operationName string, opts ...ot.StartSpanOption) ot.Span {
	c.l.Error(noopCollectorErrMsg)
	return nil
}

func (c *noopCollector) StartSpanWithOptions(operationName string, opts ot.StartSpanOptions) ot.Span {
	c.l.Error(noopCollectorErrMsg)
	return nil
}

func (c *noopCollector) Options() TracerOptions {
	return TracerOptions{}
}

func (c *noopCollector) Flush(ctx context.Context) error {
	return noopCollectorErr
}

func (c *noopCollector) Debug(v ...interface{}) {
	c.l.Error(noopCollectorErrMsg)
}

func (c *noopCollector) Info(v ...interface{}) {
	c.l.Error(noopCollectorErrMsg)
}

func (c *noopCollector) Warn(v ...interface{}) {
	c.l.Error(noopCollectorErrMsg)
}

func (c *noopCollector) Error(v ...interface{}) {
	c.l.Error(noopCollectorErrMsg)
}

func (c *noopCollector) LegacySensor() *Sensor {
	c.l.Error(noopCollectorErrMsg)
	return nil
}

func (c *noopCollector) Tracer() ot.Tracer {
	c.l.Error(noopCollectorErrMsg)
	return nil
}

func (c *noopCollector) Logger() LeveledLogger {
	c.l.Error(noopCollectorErrMsg)
	return nil
}

// SetLogger sets the logger
func (c *noopCollector) SetLogger(l LeveledLogger) {}
