// (c) Copyright IBM Corp. 2023

package instana_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	instana "github.com/instana/go-sensor"
	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/stretchr/testify/assert"
)

func Test_Collector_Noop(t *testing.T) {
	c, err := instana.GetCollector()
	assert.Error(t, err, "should return error as collector has not been initialized yet.")

	assert.NotNil(t, c, "instana collector should never be nil and be initialized as noop")

	sc, err := c.Extract(nil, nil)
	assert.Nil(t, sc)
	assert.Error(t, err)
	assert.Nil(t, c.StartSpan(""))
	assert.Nil(t, c.LegacySensor())
}

func Test_Collector_LegacySensor(t *testing.T) {
	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{AgentClient: alwaysReadyClient{}, Recorder: recorder})
	s := c.LegacySensor()
	defer instana.ShutdownCollector()

	assert.NotNil(t, c.LegacySensor())

	h := instana.TracingHandlerFunc(s, "/{action}", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(w, "Ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/foo", nil)

	h.ServeHTTP(httptest.NewRecorder(), req)

	assert.Len(t, recorder.GetQueuedSpans(), 1, "Instrumentations should still work fine with LegacySensor() method")
}

func Test_Collector_Singleton(t *testing.T) {
	var ok bool
	var instance instana.TracerLogger

	c, err := instana.GetCollector()
	assert.Error(t, err, "should return error as collector has not been initialized yet.")

	defer instana.ShutdownCollector()

	_, ok = c.(*instana.Collector)
	assert.False(t, ok, "instana collector is noop before InitCollector is called")

	c = instana.InitCollector(instana.DefaultOptions())

	instance, ok = c.(*instana.Collector)
	assert.True(t, ok, "instana collector is of type instana.Collector after InitCollector is called")

	c = instana.InitCollector(instana.DefaultOptions())

	assert.Equal(t, c, instance, "instana collector is singleton and should not be reassigned if InitCollector is called again")
}

func Test_InitCollector_With_Goroutines(t *testing.T) {

	defer instana.ShutdownCollector()

	var wg sync.WaitGroup
	wg.Add(3)

	for i := 0; i < 3; i++ {
		go func(id int) {
			defer wg.Done()
			c := instana.InitCollector(instana.DefaultOptions())

			_, ok := c.(*instana.Collector)
			assert.True(t, ok, "instana collector is of type instana.Collector after InitCollector is called")

			assert.NotNil(t, c)
		}(i)
	}
	wg.Wait()
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
	c := instana.InitCollector(nil)
	defer instana.ShutdownCollector()

	l := &mylogger{}

	c.SetLogger(l)

	c.Debug()
	c.Info()
	c.Warn()
	c.Error()
	c.Error()

	assert.Equal(t, 1, l.counter["debug"])
	assert.Equal(t, 1, l.counter["info"])
	assert.Equal(t, 1, l.counter["warn"])
	assert.Equal(t, 2, l.counter["error"])
}

func Test_Collector_MaxLogsPerSpan(t *testing.T) {
	type args struct {
		opts *instana.Options
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "tracer options is empty",
			args: args{
				opts: &instana.Options{},
			},
			want: instana.MaxLogsPerSpan,
		},
		{
			name: "default tracer options",
			args: args{
				opts: &instana.Options{
					Tracer: instana.DefaultTracerOptions(),
				},
			},
			want: instana.MaxLogsPerSpan,
		},
		{
			name: "tracer options are set by user",
			args: args{
				opts: &instana.Options{
					Tracer: instana.TracerOptions{
						MaxLogsPerSpan: 1000,
					},
				},
			},
			want: 1000,
		},
		{
			name: "tracer options are set but not MaxLogsPerSpan",
			args: args{
				opts: &instana.Options{
					Tracer: instana.TracerOptions{
						Secrets: instana.DefaultSecretsMatcher(),
					},
				},
			},
			want: instana.MaxLogsPerSpan,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer instana.ShutdownCollector()
			if got := instana.InitCollector(tt.args.opts); got.Options().MaxLogsPerSpan != tt.want {
				t.Errorf("MaxLogsPerSpan = %v, want %v", got.Options().MaxLogsPerSpan, tt.want)
			}
		})
	}
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
