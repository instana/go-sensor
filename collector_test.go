// (c) Copyright IBM Corp. 2023

package instana_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"

	"github.com/stretchr/testify/assert"
)

func Test_Collector_Noop(t *testing.T) {
	assert.NotNil(t, instana.GetC(), "instana collector should never be nil and be initialized as noop")

	sc, err := instana.GetC().Extract(nil, nil)
	assert.Nil(t, sc)
	assert.Error(t, err)
	assert.Nil(t, instana.GetC().StartSpan(""))
	assert.Nil(t, instana.GetC().LegacySensor())
}

func Test_Collector_LegacySensor(t *testing.T) {
	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{AgentClient: alwaysReadyClient{}, Recorder: recorder})
	s := c.LegacySensor()
	defer instana.ShutdownCollector()

	assert.NotNil(t, instana.GetC().LegacySensor())

	h := instana.TracingHandlerFunc(s, "/{action}", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(w, "Ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/foo", nil)

	h.ServeHTTP(httptest.NewRecorder(), req)

	assert.Len(t, recorder.GetQueuedSpans(), 1, "Instrumentations should still work fine with instana.GetC().LegacySensor()")
}

func Test_Collector_Singleton(t *testing.T) {
	var ok bool
	var instance instana.TracerLogger

	defer instana.ShutdownCollector()

	_, ok = instana.GetC().(*instana.Collector)
	assert.False(t, ok, "instana collector is noop before InitCollector is called")

	instana.InitCollector(instana.DefaultOptions())

	instance, ok = instana.GetC().(*instana.Collector)
	assert.True(t, ok, "instana collector is of type instana.Collector after InitCollector is called")

	instana.InitCollector(instana.DefaultOptions())

	assert.Equal(t, instana.GetC(), instance, "instana collector is singleton and should not be reassigned if InitCollector is called again")
}

func Test_Collector_EmbeddedTracer(t *testing.T) {
	c := instana.InitCollector(nil)
	defer instana.ShutdownCollector()

	sp := c.StartSpan("my-span")

	carrier := ot.TextMapCarrier(make(map[string]string))

	err := c.Inject(sp.Context(), ot.TextMap, carrier)
	assert.Nil(t, err)

	sctx, err := c.Extract(ot.TextMap, carrier)
	assert.Nil(t, err)

	opt := ext.RPCServerOption(sctx)
	opts := []ot.StartSpanOption{opt}

	cs := c.StartSpan("child-span", opts...)

	parentCtx, ok := sp.Context().(instana.SpanContext)
	assert.True(t, ok)

	childCtx, ok := cs.Context().(instana.SpanContext)
	assert.True(t, ok)

	assert.Equal(t, parentCtx.TraceID, childCtx.TraceID)
	assert.Equal(t, parentCtx.SpanID, childCtx.ParentID)
}

func Test_Collector_Logger(t *testing.T) {
	instana.InitCollector(nil)
	defer instana.ShutdownCollector()

	l := &mylogger{}

	instana.GetC().SetLogger(l)

	instana.GetC().Debug()
	instana.GetC().Info()
	instana.GetC().Warn()
	instana.GetC().Error()
	instana.GetC().Error()

	assert.Equal(t, 1, l.counter["debug"])
	assert.Equal(t, 1, l.counter["info"])
	assert.Equal(t, 1, l.counter["warn"])
	assert.Equal(t, 2, l.counter["error"])
}

var _ instana.LeveledLogger = (*mylogger)(nil)

type mylogger struct {
	counter map[string]int
}

func (l *mylogger) init() {
	if l.counter == nil {
		l.counter = make(map[string]int)
	}
}

func (l *mylogger) Debug(v ...interface{}) {
	l.init()
	l.counter["debug"]++
}

func (l *mylogger) Info(v ...interface{}) {
	l.init()
	l.counter["info"]++
}

func (l *mylogger) Warn(v ...interface{}) {
	l.init()
	l.counter["warn"]++
}

func (l *mylogger) Error(v ...interface{}) {
	l.init()
	l.counter["error"]++
}
