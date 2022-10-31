// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instana_test

import (
	"runtime"
	"testing"
	"time"

	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/acceptor"
	"github.com/stretchr/testify/assert"
)

func TestSnapshotCollector_Collect(t *testing.T) {
	sc := instana.SnapshotCollector{
		ServiceName:        "test",
		CollectionInterval: 500 * time.Millisecond,
	}

	assert.Equal(t, &acceptor.RuntimeInfo{
		Name:          sc.ServiceName,
		Version:       runtime.Version(),
		Root:          runtime.GOROOT(),
		MaxProcs:      runtime.GOMAXPROCS(0),
		Compiler:      runtime.Compiler,
		NumCPU:        runtime.NumCPU(),
		SensorVersion: instana.Version,
	}, sc.Collect())

	t.Run("second call before collection interval", func(t *testing.T) {
		assert.Nil(t, sc.Collect())
	})

	t.Run("second call after collection interval", func(t *testing.T) {
		oldNumProcs := runtime.GOMAXPROCS(0)
		defer runtime.GOMAXPROCS(oldNumProcs)

		time.Sleep(sc.CollectionInterval)
		runtime.GOMAXPROCS(oldNumProcs + 1)

		assert.Equal(t, &acceptor.RuntimeInfo{
			Name:          sc.ServiceName,
			Version:       runtime.Version(),
			Root:          runtime.GOROOT(),
			MaxProcs:      oldNumProcs + 1,
			Compiler:      runtime.Compiler,
			NumCPU:        runtime.NumCPU(),
			SensorVersion: instana.Version,
		}, sc.Collect())
	})
}
