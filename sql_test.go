// // (c) Copyright IBM Corp. 2023

package instana

import (
	"context"
	"database/sql/driver"
	"errors"
	"reflect"
	"testing"
)

type returnNoRows struct{}

type returnError struct{}

type rowsAffected struct{}

type rows struct {
	columns []string
}

var errUnexpected = errors.New("unexpected error")

var connDetails = DbConnDetails{
	RawString:    "connection string",
	Host:         "host",
	Port:         "1234",
	Schema:       "test-schema",
	User:         "user1",
	DatabaseName: "mysql",
}

func (r rows) Columns() []string {
	return r.columns
}

func (r rows) Close() error {
	return nil
}

func (r rows) Next(dest []driver.Value) error {
	return nil
}

func (e returnNoRows) ExecContext(
	ctx context.Context,
	query string,
	args []driver.NamedValue) (driver.Result, error) {
	return driver.ResultNoRows, nil
}

func (e returnError) ExecContext(
	ctx context.Context,
	query string,
	args []driver.NamedValue) (driver.Result, error) {
	return driver.ResultNoRows, errUnexpected
}

func (e rowsAffected) ExecContext(
	ctx context.Context,
	query string,
	args []driver.NamedValue) (driver.Result, error) {
	return driver.RowsAffected(1), nil
}

func (e returnNoRows) QueryContext(
	ctx context.Context,
	query string,
	args []driver.NamedValue) (driver.Rows, error) {
	return rows{[]string{}}, nil
}

func (e returnError) QueryContext(
	ctx context.Context,
	query string,
	args []driver.NamedValue) (driver.Rows, error) {
	return rows{[]string{}}, errUnexpected
}

func (e rowsAffected) QueryContext(
	ctx context.Context,
	query string,
	args []driver.NamedValue) (driver.Rows, error) {
	return rows{[]string{"test"}}, nil
}

func (e returnNoRows) Query(
	query string,
	args []driver.Value) (driver.Rows, error) {
	return rows{[]string{}}, nil
}

func (e returnError) Query(
	query string,
	args []driver.Value) (driver.Rows, error) {
	return rows{[]string{}}, errUnexpected
}

func (e rowsAffected) Query(
	query string,
	args []driver.Value) (driver.Rows, error) {
	return rows{[]string{"test"}}, nil
}

func (e returnNoRows) Exec(
	query string,
	args []driver.Value) (driver.Result, error) {
	return driver.ResultNoRows, nil
}

func (e returnError) Exec(
	query string,
	args []driver.Value) (driver.Result, error) {
	return driver.ResultNoRows, errUnexpected
}

func (e rowsAffected) Exec(
	query string,
	args []driver.Value) (driver.Result, error) {
	return driver.RowsAffected(1), nil
}

func Test_wExecerContext_ExecContext(t *testing.T) {

	recorder := NewTestRecorder()

	c := InitCollector(&Options{
		Service:     "go-sensor-test",
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer ShutdownCollector()

	type fields struct {
		ExecerContext driver.ExecerContext
		sensor        TracerLogger
		sqlSpan       *sqlSpanData
	}
	type args struct {
		ctx   context.Context
		query string
		args  []driver.NamedValue
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    driver.Result
		wantErr bool
	}{
		{
			name: "ResultNoRows",
			fields: fields{
				ExecerContext: returnNoRows{},
				sensor:        c,
				sqlSpan:       getSQLSpanData(connDetails),
			},
			args: args{
				ctx:   context.Background(),
				query: "TEST QUERY",
				args:  []driver.NamedValue{},
			},
			want:    driver.ResultNoRows,
			wantErr: false,
		},
		{
			name: "ReturnError",
			fields: fields{
				ExecerContext: returnError{},
				sensor:        c,
				sqlSpan:       getSQLSpanData(connDetails),
			},
			args: args{
				ctx:   context.Background(),
				query: "TEST QUERY",
				args:  []driver.NamedValue{},
			},
			want:    driver.ResultNoRows,
			wantErr: true,
		},
		{
			name: "RowsAffected",
			fields: fields{
				ExecerContext: rowsAffected{},
				sensor:        c,
				sqlSpan:       getSQLSpanData(connDetails),
			},
			args: args{
				ctx:   context.Background(),
				query: "TEST QUERY",
				args:  []driver.NamedValue{},
			},
			want:    driver.RowsAffected(1),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := &wExecerContext{
				ExecerContext: tt.fields.ExecerContext,
				sensor:        tt.fields.sensor,
				sqlSpan:       tt.fields.sqlSpan,
			}
			got, err := conn.ExecContext(tt.args.ctx, tt.args.query, tt.args.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("wExecerContext.ExecContext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("wExecerContext.ExecContext() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_wExecer_Exec(t *testing.T) {

	recorder := NewTestRecorder()
	c := InitCollector(&Options{
		Service:     "go-sensor-test",
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer ShutdownCollector()

	type fields struct {
		Execer  driver.Execer
		sensor  TracerLogger
		sqlSpan *sqlSpanData
	}
	type args struct {
		query string
		args  []driver.Value
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    driver.Result
		wantErr bool
	}{
		{
			name: "ResultNoRows",
			fields: fields{
				Execer:  returnNoRows{},
				sensor:  c,
				sqlSpan: getSQLSpanData(connDetails),
			},
			args: args{
				query: "TEST QUERY",
				args:  []driver.Value{},
			},
			want:    driver.ResultNoRows,
			wantErr: false,
		},
		{
			name: "ReturnError",
			fields: fields{
				Execer:  returnError{},
				sensor:  c,
				sqlSpan: getSQLSpanData(connDetails),
			},
			args: args{
				query: "TEST QUERY",
				args:  []driver.Value{},
			},
			want:    driver.ResultNoRows,
			wantErr: true,
		},
		{
			name: "RowsAffected",
			fields: fields{
				Execer:  rowsAffected{},
				sensor:  c,
				sqlSpan: getSQLSpanData(connDetails),
			},
			args: args{
				query: "TEST QUERY",
				args:  []driver.Value{},
			},
			want:    driver.RowsAffected(1),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := &wExecer{
				Execer:  tt.fields.Execer,
				sensor:  tt.fields.sensor,
				sqlSpan: tt.fields.sqlSpan,
			}
			got, err := conn.Exec(tt.args.query, tt.args.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("wExecer.Exec() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("wExecer.Exec() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_wQueryerContext_QueryContext(t *testing.T) {

	recorder := NewTestRecorder()
	c := InitCollector(&Options{
		Service:     "go-sensor-test",
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer ShutdownCollector()

	type fields struct {
		QueryerContext driver.QueryerContext
		sensor         TracerLogger
		sqlSpan        *sqlSpanData
	}
	type args struct {
		ctx   context.Context
		query string
		args  []driver.NamedValue
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    driver.Rows
		wantErr bool
	}{
		{
			name: "ResultNoRows",
			fields: fields{
				QueryerContext: returnNoRows{},
				sensor:         c,
				sqlSpan:        getSQLSpanData(connDetails),
			},
			args: args{
				ctx:   context.Background(),
				query: "TEST QUERY",
				args:  []driver.NamedValue{},
			},
			want:    rows{[]string{}},
			wantErr: false,
		},
		{
			name: "ReturnError",
			fields: fields{
				QueryerContext: returnError{},
				sensor:         c,
				sqlSpan:        getSQLSpanData(connDetails),
			},
			args: args{
				ctx:   context.Background(),
				query: "TEST QUERY",
				args:  []driver.NamedValue{},
			},
			want:    rows{[]string{}},
			wantErr: true,
		},
		{
			name: "RowsAffected",
			fields: fields{
				QueryerContext: rowsAffected{},
				sensor:         c,
				sqlSpan:        getSQLSpanData(connDetails),
			},
			args: args{
				ctx:   context.Background(),
				query: "TEST QUERY",
				args:  []driver.NamedValue{},
			},
			want:    rows{[]string{"test"}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := &wQueryerContext{
				QueryerContext: tt.fields.QueryerContext,
				sensor:         tt.fields.sensor,
				sqlSpan:        tt.fields.sqlSpan,
			}
			got, err := conn.QueryContext(tt.args.ctx, tt.args.query, tt.args.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("wQueryerContext.QueryContext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("wQueryerContext.QueryContext() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_wQueryer_Query(t *testing.T) {

	recorder := NewTestRecorder()
	c := InitCollector(&Options{
		Service:     "go-sensor-test",
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer ShutdownCollector()

	type fields struct {
		Queryer driver.Queryer
		sensor  TracerLogger
		sqlSpan *sqlSpanData
	}
	type args struct {
		query string
		args  []driver.Value
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    driver.Rows
		wantErr bool
	}{
		{
			name: "ResultNoRows",
			fields: fields{
				Queryer: returnNoRows{},
				sensor:  c,
				sqlSpan: getSQLSpanData(connDetails),
			},
			args: args{
				query: "TEST QUERY",
				args:  []driver.Value{},
			},
			want:    rows{[]string{}},
			wantErr: false,
		},
		{
			name: "ReturnError",
			fields: fields{
				Queryer: returnError{},
				sensor:  c,
				sqlSpan: getSQLSpanData(connDetails),
			},
			args: args{
				query: "TEST QUERY",
				args:  []driver.Value{},
			},
			want:    rows{[]string{}},
			wantErr: true,
		},
		{
			name: "RowsAffected",
			fields: fields{
				Queryer: rowsAffected{},
				sensor:  c,
				sqlSpan: getSQLSpanData(connDetails),
			},
			args: args{
				query: "TEST QUERY",
				args:  []driver.Value{},
			},
			want:    rows{[]string{"test"}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := &wQueryer{
				Queryer: tt.fields.Queryer,
				sensor:  tt.fields.sensor,
				sqlSpan: tt.fields.sqlSpan,
			}
			got, err := conn.Query(tt.args.query, tt.args.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("wQueryer.Query() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("wQueryer.Query() = %v, want %v", got, tt.want)
			}
		})
	}
}

type stmt struct{}

func (s stmt) Query(args []driver.Value) (driver.Rows, error) {
	return rows{[]string{}}, nil
}

func (s stmt) Exec(args []driver.Value) (driver.Result, error) {
	return driver.ResultNoRows, nil
}

func (s stmt) Close() error { return nil }

func (s stmt) NumInput() int { return -1 }

type stmtErr struct{}

func (s stmtErr) Query(args []driver.Value) (driver.Rows, error) {
	return rows{[]string{}}, errUnexpected
}

func (s stmtErr) Exec(args []driver.Value) (driver.Result, error) {
	return driver.ResultNoRows, errUnexpected
}

func (s stmtErr) Close() error { return nil }

func (s stmtErr) NumInput() int { return -1 }

func Test_wStmt_Query(t *testing.T) {

	recorder := NewTestRecorder()
	c := InitCollector(&Options{
		Service:     "go-sensor-test",
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer ShutdownCollector()

	type fields struct {
		Stmt    driver.Stmt
		sqlSpan *sqlSpanData
		sensor  TracerLogger
	}
	type args struct {
		args []driver.Value
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    driver.Rows
		wantErr bool
	}{
		{
			name: "sql stmt Query",
			fields: fields{
				Stmt:    stmt{},
				sensor:  c,
				sqlSpan: getSQLSpanData(connDetails),
			},
			args: args{
				args: []driver.Value{},
			},
			want:    rows{[]string{}},
			wantErr: false,
		},
		{
			name: "sql stmt Query with error",
			fields: fields{
				Stmt:    stmtErr{},
				sensor:  c,
				sqlSpan: getSQLSpanData(connDetails),
			},
			args: args{
				args: []driver.Value{},
			},
			want:    rows{[]string{}},
			wantErr: true,
		},
		{
			name: "sql stmt Exec with error",
			fields: fields{
				Stmt:    stmtErr{},
				sensor:  c,
				sqlSpan: getSQLSpanData(connDetails),
			},
			args: args{
				args: []driver.Value{},
			},
			want:    rows{[]string{}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stmt := &wStmt{
				Stmt:    tt.fields.Stmt,
				sqlSpan: tt.fields.sqlSpan,
				sensor:  tt.fields.sensor,
			}
			got, err := stmt.Query(tt.args.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("wStmt.Query() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("wStmt.Query() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_wStmt_Exec(t *testing.T) {

	recorder := NewTestRecorder()
	c := InitCollector(&Options{
		Service:     "go-sensor-test",
		AgentClient: alwaysReadyClient{},
		Recorder:    recorder,
	})
	defer ShutdownCollector()

	type fields struct {
		Stmt    driver.Stmt
		sqlSpan *sqlSpanData
		sensor  TracerLogger
	}
	type args struct {
		args []driver.Value
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    driver.Result
		wantErr bool
	}{
		{
			name: "sql stmt Exec with error",
			fields: fields{
				Stmt:    stmtErr{},
				sensor:  c,
				sqlSpan: getSQLSpanData(connDetails),
			},
			args: args{
				args: []driver.Value{},
			},
			want:    driver.ResultNoRows,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stmt := &wStmt{
				Stmt:    tt.fields.Stmt,
				sqlSpan: tt.fields.sqlSpan,
				sensor:  tt.fields.sensor,
			}
			got, err := stmt.Exec(tt.args.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("wStmt.Query() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("wStmt.Query() = %v, want %v", got, tt.want)
			}
		})
	}
}
