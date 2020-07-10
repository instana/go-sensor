package instana_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
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

func (drv sqlDriver) Open(name string) (driver.Conn, error) { return sqlConn{drv.Error}, nil } //nolint:gosimple

type sqlConn struct{ Error error }

func (conn sqlConn) Prepare(query string) (driver.Stmt, error) { return sqlStmt{conn.Error}, nil } //nolint:gosimple
func (sqlConn) Close() error                                   { return driver.ErrSkip }
func (sqlConn) Begin() (driver.Tx, error)                      { return nil, driver.ErrSkip }

func (conn sqlConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	return sqlResult{}, conn.Error
}

type sqlStmt struct{ Error error }

func (sqlStmt) Close() error                                         { return nil }
func (sqlStmt) NumInput() int                                        { return -1 }
func (stmt sqlStmt) Exec(args []driver.Value) (driver.Result, error) { return sqlResult{}, stmt.Error }
func (stmt sqlStmt) Query(args []driver.Value) (driver.Rows, error)  { return sqlRows{}, stmt.Error }

type sqlResult struct{}

func (sqlResult) LastInsertId() (int64, error) { return 42, nil }
func (sqlResult) RowsAffected() (int64, error) { return 100, nil }

type sqlRows struct{}

func (sqlRows) Columns() []string              { return []string{"col1", "col2"} }
func (sqlRows) Close() error                   { return nil }
func (sqlRows) Next(dest []driver.Value) error { return io.EOF }
