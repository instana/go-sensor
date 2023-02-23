// (c) Copyright IBM Corp. 2023

package instagraphql_test

import (
	"sync"
	"testing"
	"time"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instagraphql"
	"github.com/stretchr/testify/assert"
)

func TestExpiringMap(t *testing.T) {
	var wg sync.WaitGroup

	recorder := instana.NewTestRecorder()
	sensor := instana.NewSensorWithTracer(
		instana.NewTracerWithEverything(&instana.Options{AgentClient: alwaysReadyClient{}}, recorder),
	)

	defer instana.ShutdownSensor()

	tracer := sensor.Tracer()

	m := instagraphql.ExpiringMap{}

	m.Set("key1", tracer.StartSpan("span1"), time.Second)
	m.Set("key2", tracer.StartSpan("span2"), time.Second)

	wg.Add(1)

	go func() {
		time.Sleep(time.Millisecond * 900)
		m.Set("key1", tracer.StartSpan("span1 new"), time.Second)
		m.Set("key2", tracer.StartSpan("span2 new"), time.Second)
		time.Sleep(time.Millisecond * 900)
		m.Set("key1", tracer.StartSpan("span1 new 2"), time.Second)
		time.Sleep(time.Millisecond * 900)
		m.Set("key1", tracer.StartSpan("span1 new 3"), time.Second)

		wg.Done()
	}()

	wg.Wait()

	assert.Nil(t, m.Get("key2"), `"key2" should have expired`)
	assert.NotNil(t, m.Get("key1"), `"key1" should still exist`)
}
