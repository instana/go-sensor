// (c) Copyright IBM Corp. 2022

package instana_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"time"
)

// Driver use case:
// * driver.Conn doesn't implement Exec or ExecContext
// * driver.Conn doesn't implement the driver.NamedValueChecker interface (CheckNamedValue method)
// * Our wrapper ALWAYS implements ExecContext, no matter what

type sqlDriver2 struct{ Error error }

func (drv sqlDriver2) Open(name string) (driver.Conn, error) { return sqlConn2{drv.Error}, nil } //nolint:gosimple
func (drv sqlDriver2) Close() error                          { return nil }                      //nolint:gosimple

type sqlConn2 struct{ Error error }

func (s sqlConn2) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	return &sqlStmt2{s.Error}, nil
}

func (s sqlConn2) Query(query string, args []driver.Value) (driver.Rows, error) {
	return sqlRows2{}, nil
}

func (s sqlConn2) Prepare(query string) (driver.Stmt, error) {
	return &sqlStmt2{s.Error}, nil //nolint:gosimple
}

func (s sqlConn2) Close() error { return driver.ErrSkip }

func (s sqlConn2) Begin() (driver.Tx, error) { return nil, driver.ErrSkip }

type sqlStmt2 struct{ Error error }

func (sqlStmt2) Close() error  { return nil }
func (sqlStmt2) NumInput() int { return -1 }
func (stmt sqlStmt2) Exec(args []driver.Value) (driver.Result, error) {
	return sqlResult2{}, stmt.Error
}

func (stmt sqlStmt2) Query(args []driver.Value) (driver.Rows, error) {
	return sqlRows2{}, stmt.Error
}

////////

func (stmt sqlStmt2) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	return nil, nil
}

func (stmt sqlStmt2) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	return nil, nil
}

//////

func (stmt sqlStmt2) CheckNamedValue(nv *driver.NamedValue) (err error) {
	switch d := nv.Value.(type) {
	case sql.Out:
		err = nil
	case []int:
		temp := make([]int64, len(d))
		for i := 0; i < len(d); i++ {
			temp[i] = int64(d[i])
		}
		nv.Value = temp
		err = nil
	case []int8:
		temp := make([]int64, len(d))
		for i := 0; i < len(d); i++ {
			temp[i] = int64(d[i])
		}
		nv.Value = temp
		err = nil
	case []int16:
		temp := make([]int64, len(d))
		for i := 0; i < len(d); i++ {
			temp[i] = int64(d[i])
		}
		nv.Value = temp
		err = nil
	case []int32:
		temp := make([]int64, len(d))
		for i := 0; i < len(d); i++ {
			temp[i] = int64(d[i])
		}
		nv.Value = temp
		err = nil
	case []int64:
		err = nil
	case []string:
		err = nil
	case []bool:
		err = nil
	case []float64:
		err = nil
	case []float32:
		temp := make([]float64, len(d))
		for i := 0; i < len(d); i++ {
			temp[i] = float64(d[i])
		}
		nv.Value = temp
		err = nil
	case []time.Time:
		err = nil
	default:
		nv.Value, err = driver.DefaultParameterConverter.ConvertValue(nv.Value)
	}
	return err
}

type sqlResult2 struct{}

func (sqlResult2) LastInsertId() (int64, error) { return 42, nil }
func (sqlResult2) RowsAffected() (int64, error) { return 100, nil }

type sqlRows2 struct{}

func (sqlRows2) Columns() []string              { return []string{"col1", "col2"} }
func (sqlRows2) Close() error                   { return nil }
func (sqlRows2) Next(dest []driver.Value) error { return io.EOF }
