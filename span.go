// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2016

package instana

import (
	"bytes"
	"strings"
	"sync"
	"time"

	"github.com/instana/go-sensor/logger"
	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
)

const minSpanLogLevel = logger.WarnLevel

var _ ot.Span = (*spanS)(nil)

type spanCategory string

const (
	logging spanCategory = "logging"
	unknown spanCategory = "unknown"
)

type spanS struct {
	Service     string
	Operation   string
	Start       time.Time
	Duration    time.Duration
	Correlation EUMCorrelationData
	Tags        ot.Tags
	Logs        []ot.LogRecord
	ErrorCount  int

	tracer *tracerS
	mu     sync.Mutex

	context SpanContext
}

func (r *spanS) BaggageItem(key string) string {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.context.Baggage[key]
}

func (r *spanS) SetBaggageItem(key, val string) ot.Span {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.context = r.context.WithBaggageItem(key, val)

	return r
}

func (r *spanS) Context() ot.SpanContext {
	return r.context
}

func (r *spanS) Finish() {
	r.FinishWithOptions(ot.FinishOptions{})
}

func (r *spanS) FinishWithOptions(opts ot.FinishOptions) {
	finishTime := opts.FinishTime
	if finishTime.IsZero() {
		finishTime = time.Now()
	}

	duration := finishTime.Sub(r.Start)

	r.mu.Lock()
	defer r.mu.Unlock()

	for _, lr := range opts.LogRecords {
		r.appendLog(lr)
	}

	for _, ld := range opts.BulkLogData {
		r.appendLog(ld.ToLogRecord())
	}

	r.Duration = duration
	if r.sendSpanToAgent() {
		if sensor.Agent().Ready() {
			r.tracer.recorder.RecordSpan(r)
		} else {
			delayed.append(r)
		}
		r.sendOpenTracingLogRecords()
	}
}

func (r *spanS) sendSpanToAgent() bool {
	// Span shouldn't be forwarded if the span category is configured as disabled
	if r.getSpanCategory().disabled() {
		return false
	}

	//if suppress tag is present, span shouldn't be forwarded
	if r.context.Suppressed {
		return false
	}

	if !isRootExitSpan(r.Tags[string(ext.SpanKind)], r.context.ParentID == 0) {
		// if the span is an entry span, intermediate span, exit span with a parent
		// it should be forwarded to the agent
		return true
	}

	// if the span is an exit span without a parent span, then it should be forwarded
	// only if ALLOW_ROOT_EXIT_SPAN is configured by the user
	return allowRootExitSpan()
}

func (r *spanS) appendLog(lr ot.LogRecord) {
	maxLogs := r.tracer.Options().MaxLogsPerSpan
	if len(r.Logs) < maxLogs {
		r.Logs = append(r.Logs, lr)
	}
}

func (r *spanS) Log(ld ot.LogData) {
	if r.tracer.Options().DropAllLogs {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if ld.Timestamp.IsZero() {
		ld.Timestamp = time.Now()
	}

	r.appendLog(ld.ToLogRecord())
}

func (r *spanS) LogEvent(event string) {
	r.Log(ot.LogData{
		Event: event})
}

func (r *spanS) LogEventWithPayload(event string, payload interface{}) {
	r.Log(ot.LogData{
		Event:   event,
		Payload: payload})
}

func (r *spanS) LogFields(fields ...otlog.Field) {

	for _, v := range fields {
		// If this tag indicates an error, increase the error count
		if openTracingLogFieldLevel(v) == logger.ErrorLevel {
			r.ErrorCount++
		}
	}

	lr := ot.LogRecord{
		Fields: fields,
	}

	if r.tracer.Options().DropAllLogs {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if lr.Timestamp.IsZero() {
		lr.Timestamp = time.Now()
	}

	r.appendLog(lr)
}

func (r *spanS) LogKV(keyValues ...interface{}) {
	fields, err := otlog.InterleavedKVToFields(keyValues...)
	if err != nil {
		r.LogFields(otlog.Error(err), otlog.String("function", "LogKV"))

		return
	}

	r.LogFields(fields...)
}

func (r *spanS) SetOperationName(operationName string) ot.Span {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.Operation = operationName

	return r
}

func (r *spanS) SetTag(key string, value interface{}) ot.Span {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Tags == nil {
		r.Tags = ot.Tags{}
	}

	// If this tag indicates an error, increase the error count
	if key == "error" {
		r.ErrorCount++
	}

	if key == suppressTracingTag {
		r.context.Suppressed = true
		return r
	}

	r.Tags[key] = value

	return r
}

func (r *spanS) Tracer() ot.Tracer {
	return r.tracer
}

// sendOpenTracingLogRecords converts OpenTracing log records that contain errors
// to Instana log spans and sends them to the agent
func (r *spanS) sendOpenTracingLogRecords() {
	if logging.disabled() {
		return
	}

	for _, lr := range r.Logs {
		r.sendOpenTracingLogRecord(lr)
	}
}

func (r *spanS) sendOpenTracingLogRecord(lr ot.LogRecord) {
	lvl := openTracingHighestLogRecordLevel(lr)

	if lvl.Less(minSpanLogLevel) {
		return
	}

	buf := bytes.NewBuffer(nil)

	enc := newOpenTracingLogEncoder(buf)
	for _, lf := range lr.Fields {
		lf.Marshal(enc)
		buf.WriteByte(' ')
	}

	r.tracer.StartSpan(
		"log.go",
		ot.ChildOf(r.context),
		ot.StartTime(lr.Timestamp),
		ot.Tags{
			"log.level":   lvl.String(),
			"log.message": strings.TrimSpace(buf.String()),
		},
	).FinishWithOptions(
		ot.FinishOptions{
			FinishTime: lr.Timestamp,
		},
	)
}

// openTracingHighestLogRecordLevel determines the level of this record by inspecting its fields.
// If there are multiple fields suggesting the log level, i.e. both "error" and "warn" are present,
// the highest one takes precedence.
func openTracingHighestLogRecordLevel(lr ot.LogRecord) logger.Level {
	highestLvl := logger.DebugLevel

	for _, lf := range lr.Fields {
		if lvl := openTracingLogFieldLevel(lf); highestLvl.Less(lvl) {
			highestLvl = lvl
		}
	}

	return highestLvl
}

func openTracingLogFieldLevel(lf otlog.Field) logger.Level {
	switch lf.Key() {
	case "error", "error.object":
		return logger.ErrorLevel
	case "warn":
		return logger.WarnLevel
	default:
		return logger.DebugLevel
	}
}

func (c spanCategory) string() string {
	return string(c)
}

func (r *spanS) getSpanCategory() spanCategory {
	// return span category if it is a registered span type
	switch RegisteredSpanType(r.Operation) {
	case LogSpanType:
		return logging
	default:
		return unknown
	}
}

func (c spanCategory) disabled() bool {
	// unrecognized categories are always enabled
	if c == unknown {
		return false
	}

	// Check if sensor or options are nil
	muSensor.RLock()
	defer muSensor.RUnlock()

	if sensor == nil || sensor.options.Tracer.DisableSpans == nil {
		return false
	}

	return sensor.options.Tracer.DisableSpans[c.string()]
}
