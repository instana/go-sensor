package instana

import (
	"sync"
	"time"

	ot "github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"
)

type spanS struct {
	tracer *tracerS
	sync.Mutex
	raw RawSpan
}

func (r *spanS) BaggageItem(key string) string {
	r.Lock()
	defer r.Unlock()

	return r.raw.Context.Baggage[key]
}

func (r *spanS) SetBaggageItem(key, val string) ot.Span {
	if r.trim() {
		return r
	}

	r.Lock()
	defer r.Unlock()
	r.raw.Context = r.raw.Context.WithBaggageItem(key, val)

	return r
}

func (r *spanS) Context() ot.SpanContext {
	return r.raw.Context
}

func (r *spanS) Finish() {
	r.FinishWithOptions(ot.FinishOptions{})
}

func (r *spanS) FinishWithOptions(opts ot.FinishOptions) {
	finishTime := opts.FinishTime
	if finishTime.IsZero() {
		finishTime = time.Now()
	}

	duration := finishTime.Sub(r.raw.Start)
	r.Lock()
	defer r.Unlock()
	for _, lr := range opts.LogRecords {
		r.appendLog(lr)
	}

	for _, ld := range opts.BulkLogData {
		r.appendLog(ld.ToLogRecord())
	}

	r.raw.Duration = duration
	// FIXME
	//r.tracer.options.Recorder.RecordSpan(r.raw)
}

func (r *spanS) appendLog(lr ot.LogRecord) {
	maxLogs := r.tracer.options.MaxLogsPerSpan
	if maxLogs == 0 || len(r.raw.Logs) < maxLogs {
		r.raw.Logs = append(r.raw.Logs, lr)
	}
}

func (r *spanS) Log(ld ot.LogData) {
	r.Lock()
	defer r.Unlock()
	if r.trim() || r.tracer.options.DropAllLogs {
		return
	}

	if ld.Timestamp.IsZero() {
		ld.Timestamp = time.Now()
	}

	r.appendLog(ld.ToLogRecord())
}

func (r *spanS) trim() bool {
	return !r.raw.Context.Sampled && r.tracer.options.TrimUnsampledSpans
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
	lr := ot.LogRecord{
		Fields: fields,
	}

	r.Lock()
	defer r.Unlock()
	if r.trim() || r.tracer.options.DropAllLogs {
		return
	}

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
	r.Lock()
	defer r.Unlock()
	r.raw.Operation = operationName

	return r
}

func (r *spanS) SetTag(key string, value interface{}) ot.Span {
	r.Lock()
	defer r.Unlock()
	if r.trim() {
		return r
	}

	if r.raw.Tags == nil {
		r.raw.Tags = ot.Tags{}
	}

	r.raw.Tags[key] = value

	return r
}

func (r *spanS) Tracer() ot.Tracer {
	return r.tracer
}
