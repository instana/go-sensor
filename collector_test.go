// (c) Copyright IBM Corp. 2023

package instana_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	instana "github.com/instana/go-sensor"

	"github.com/stretchr/testify/assert"
)

func Test_Collector_Noop(t *testing.T) {
	assert.NotNil(t, instana.C, "instana.C should never be nil and be initialized as noop")

	sc, err := instana.C.Extract(nil, nil)
	assert.Nil(t, sc)
	assert.Error(t, err)
	assert.Nil(t, instana.C.StartSpan(""))
	assert.Nil(t, instana.C.LegacySensor())
}

func Test_Collector_LegacySensor(t *testing.T) {
	recorder := instana.NewTestRecorder()
	c := instana.InitCollector(&instana.Options{AgentClient: alwaysReadyClient{}, Recorder: recorder})
	s := c.LegacySensor()
	defer instana.ShutdownSensor()

	assert.NotNil(t, instana.C.LegacySensor())

	h := instana.TracingHandlerFunc(s, "/{action}", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(w, "Ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/foo", nil)

	h.ServeHTTP(httptest.NewRecorder(), req)

	assert.Len(t, recorder.GetQueuedSpans(), 1, "Instrumentations should still work fine with instana.C.LegacySensor()")
}

func Test_Collector_Singleton(t *testing.T) {
	instana.C = nil
	var ok bool
	var instance instana.TracerLogger

	_, ok = instana.C.(*instana.Collector)
	assert.False(t, ok, "instana.C is noop before InitCollector is called")

	instana.InitCollector(instana.DefaultOptions())

	instance, ok = instana.C.(*instana.Collector)
	assert.True(t, ok, "instana.C is of type instana.Collector after InitCollector is called")

	instana.InitCollector(instana.DefaultOptions())

	assert.Equal(t, instana.C, instance, "instana.C is singleton and should not be reassigned if InitCollector is called again")
}
