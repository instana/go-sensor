// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2021

package instana

import (
	"encoding/json"
	"fmt"
	"strconv"

	otlog "github.com/opentracing/opentracing-go/log"
)

type stringWriter interface {
	WriteString(string) (int, error)
}

type otLogEncoder struct {
	buf stringWriter
}

func newOpenTracingLogEncoder(buf stringWriter) *otLogEncoder {
	return &otLogEncoder{
		buf: buf,
	}
}

// EmitString writes an opentracing/log.LogField containing string value to the buffer
func (enc *otLogEncoder) EmitString(key, value string) {
	enc.writeField(key, strconv.Quote(value))
}

// EmitBool writes an opentracing/log.LogField containing bool value to the buffer
func (enc *otLogEncoder) EmitBool(key string, value bool) {
	enc.writeField(key, strconv.FormatBool(value))
}

// EmitInt writes an opentracing/log.LogField containing int value to the buffer
func (enc *otLogEncoder) EmitInt(key string, value int) {
	enc.writeField(key, strconv.FormatInt(int64(value), 10))
}

// EmitInt32 writes an opentracing/log.LogField containing int32 value to the buffer
func (enc *otLogEncoder) EmitInt32(key string, value int32) {
	enc.writeField(key, strconv.FormatInt(int64(value), 10))
}

// EmitInt64 writes an opentracing/log.LogField containing int64 value to the buffer
func (enc *otLogEncoder) EmitInt64(key string, value int64) {
	enc.writeField(key, strconv.FormatInt(value, 10))
}

// EmitUint32 writes an opentracing/log.LogField containing uint32 value to the buffer
func (enc *otLogEncoder) EmitUint32(key string, value uint32) {
	enc.writeField(key, strconv.FormatUint(uint64(value), 10))
}

// EmitUint64 writes an opentracing/log.LogField containing uint64 value to the buffer
func (enc *otLogEncoder) EmitUint64(key string, value uint64) {
	enc.writeField(key, strconv.FormatUint(value, 10))
}

// EmitFloat32 writes an opentracing/log.LogField containing float32 value to the buffer
func (enc *otLogEncoder) EmitFloat32(key string, value float32) {
	enc.writeField(key, strconv.FormatFloat(float64(value), 'g', -1, 32))
}

// EmitFloat64 writes an opentracing/log.LogField containing float64 value to the buffer
func (enc *otLogEncoder) EmitFloat64(key string, value float64) {
	enc.writeField(key, strconv.FormatFloat(float64(value), 'g', -1, 64))
}

// EmitObject writes the JSON representation of an object value of opentracing/log.LogField to the buffer.
// In case json.Marshal() returns an error the object representation is replaced by the error message.
func (enc *otLogEncoder) EmitObject(key string, value interface{}) {
	data, err := json.Marshal(value)
	if err != nil {
		enc.writeField(key, fmt.Sprintf("<JSON marshaling failed: %s>", err))
		return
	}

	enc.writeField(key, string(data))
}

// EmitLazyLogger delegates value writing to the LazyLogger
func (enc *otLogEncoder) EmitLazyLogger(value otlog.LazyLogger) {
	value(enc)
}

func (enc *otLogEncoder) writeField(key, value string) {
	enc.buf.WriteString(key)
	enc.buf.WriteString(": ")
	enc.buf.WriteString(value)
}
