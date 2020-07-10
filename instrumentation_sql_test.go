package instana_test

import (
	"database/sql"
	"database/sql/driver"
	"testing"

	instana "github.com/instana/go-sensor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstrumentSQLDriver(t *testing.T) {
	if !sqlDriverRegistered("test_driver_with_instana") {
		instana.InstrumentSQLDriver(instana.NewSensor("go-sensor-test"), "test_driver", sqlDriver{})
	}

	assert.Contains(t, sql.Drivers(), "test_driver_with_instana")
}

func TestOpenSQLDB(t *testing.T) {
	if !sqlDriverRegistered("test_driver_with_instana") {
		instana.InstrumentSQLDriver(instana.NewSensor("go-sensor-test"), "test_driver", sqlDriver{})
	}

	_, err := instana.OpenSQLDB("test_driver", "connection string")
	require.NoError(t, err)
}

func sqlDriverRegistered(name string) bool {
	for _, drv := range sql.Drivers() {
		if drv == name {
			return true
		}
	}

	return false
}

type sqlDriver struct{ Error error }

func (drv sqlDriver) Open(name string) (driver.Conn, error) {
	return nil, nil
}
