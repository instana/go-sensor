package instana_test

import (
	"database/sql"
	"database/sql/driver"
	"testing"

	instana "github.com/instana/go-sensor"
	"github.com/stretchr/testify/assert"
)

func TestInstrumentSQLDriver(t *testing.T) {
	if !sqlDriverRegistered("test_driver_with_instana") {
		instana.InstrumentSQLDriver(instana.NewSensor("go-sensor-test"), "test_driver", sqlDriver{})
	}

	assert.Contains(t, sql.Drivers(), "test_driver_with_instana")
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
