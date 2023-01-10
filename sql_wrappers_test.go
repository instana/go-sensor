// (c) Copyright IBM Corp. 2023

package instana

import (
	"context"
	"database/sql/driver"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_connAlreadyWrapped(t *testing.T) {
	type args struct {
		conn driver.Conn
	}
	var tests []struct {
		name string
		args args
		want bool
	}

	pCons := []driver.Conn{
		&wConn{}, &w_conn_Execer_ExecerContext_Queryer_QueryerContext_ConnPrepareContext_NamedValueChecker{}, &w_conn_Execer_ExecerContext_Queryer_QueryerContext_ConnPrepareContext{},
	}
	cons := []driver.Conn{
		w_conn_Execer_ExecerContext_Queryer_QueryerContext_ConnPrepareContext_NamedValueChecker{}, w_conn_Execer_ExecerContext_Queryer_QueryerContext_ConnPrepareContext{},
	}

	for _, v := range pCons {
		tests = append(tests, struct {
			name string
			args args
			want bool
		}{
			name: fmt.Sprintf("%T", v),
			args: args{
				v,
			},
			want: true,
		})
	}

	for _, v := range cons {
		tests = append(tests, struct {
			name string
			args args
			want bool
		}{
			name: fmt.Sprintf("%T", v),
			args: args{
				v,
			},
			want: true,
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, connAlreadyWrapped(tt.args.conn), "connAlreadyWrapped(%v)", tt.args.conn)
		})
	}
}

func Test_connAlreadyWrappedFalse(t *testing.T) {
	var a driver.Conn
	assert.False(t, connAlreadyWrapped(a), "connAlreadyWrapped(%v)")
}

func Test_stmtAlreadyWrapped(t *testing.T) {
	type args struct {
		conn driver.Stmt
	}
	var tests []struct {
		name string
		args args
		want bool
	}

	pStmts := []driver.Stmt{
		&wStmt{}, &w_stmt_StmtExecContext_StmtQueryContext_NamedValueChecker_ColumnConverter{}, &w_stmt_StmtExecContext_StmtQueryContext_NamedValueChecker{},
	}
	stmt := []driver.Stmt{w_stmt_StmtExecContext_StmtQueryContext_NamedValueChecker_ColumnConverter{}, w_stmt_StmtExecContext_StmtQueryContext_NamedValueChecker{}}

	for _, v := range pStmts {
		tests = append(tests, struct {
			name string
			args args
			want bool
		}{
			name: fmt.Sprintf("%T", v),
			args: args{
				v,
			},
			want: true,
		})
	}

	for _, v := range stmt {
		tests = append(tests, struct {
			name string
			args args
			want bool
		}{
			name: fmt.Sprintf("%T", v),
			args: args{
				v,
			},
			want: true,
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, stmtAlreadyWrapped(tt.args.conn), "stmtAlreadyWrapped(%v)", tt.args.conn)
		})
	}
}

func Test_stmtAlreadyWrappedFalse(t *testing.T) {
	var a driver.Stmt
	assert.False(t, stmtAlreadyWrapped(a), "stmtAlreadyWrapped(%v)")
}

func Test_stmtWrap(t *testing.T) {
	type args struct {
		stmt driver.Stmt
	}
	var tests []struct {
		name string
		args args
		want interface{}
	}

	var stmts []driver.Stmt
	for _, f := range _stmt_n {
		stmts = append(stmts, f(nil, "", dbConnDetails{}, nil, nil, nil, nil, nil))
	}

	for _, v := range stmts {
		tests = append(tests, struct {
			name string
			args args
			want interface{}
		}{
			name: fmt.Sprintf("%T", v),
			args: args{
				v,
			},
			want: v,
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.IsType(t, tt.want, wrapStmt(tt.args.stmt, "", dbConnDetails{}, nil), tt.name)
		})
	}
}

func Test_connWrap(t *testing.T) {
	type args struct {
		conn driver.Conn
	}
	var tests []struct {
		name string
		args args
		want interface{}
	}

	var conns []driver.Conn
	for _, f := range _conn_n {
		conns = append(conns, f(dbConnDetails{}, nil, nil, nil, nil, nil, nil, nil, nil))
	}

	for _, v := range conns {
		tests = append(tests, struct {
			name string
			args args
			want interface{}
		}{
			name: fmt.Sprintf("%T", v),
			args: args{
				v,
			},
			want: v,
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.IsType(t, tt.want, wrapConn(dbConnDetails{}, tt.args.conn, nil), tt.name)
		})
	}
}

type stmtAllInterfacesMock struct {
}

func (s *stmtAllInterfacesMock) CheckNamedValue(value *driver.NamedValue) error {
	return nil
}

func (s *stmtAllInterfacesMock) ColumnConverter(idx int) driver.ValueConverter {
	return nil
}

func (s *stmtAllInterfacesMock) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	return nil, nil
}

func (s *stmtAllInterfacesMock) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	return nil, nil
}

func (s *stmtAllInterfacesMock) Close() error {
	return nil
}

func (s *stmtAllInterfacesMock) NumInput() int {
	return 0
}

func (s *stmtAllInterfacesMock) Exec(args []driver.Value) (driver.Result, error) {
	return nil, nil
}

func (s *stmtAllInterfacesMock) Query(args []driver.Value) (driver.Rows, error) {
	return nil, nil
}

func Test_stmtAllInterfacesCase(t *testing.T) {
	mock := &stmtAllInterfacesMock{}

	d := wrapStmt(mock, "", dbConnDetails{}, nil)

	assert.IsType(t, &w_stmt_StmtExecContext_StmtQueryContext_NamedValueChecker_ColumnConverter{}, d)
}

func Test_stmtWrapOnlyStmt(t *testing.T) {
	var a driver.Stmt

	d := wrapStmt(a, "", dbConnDetails{}, nil)

	assert.IsType(t, &wStmt{}, d)
}

func Test_connWrapOnlyConn(t *testing.T) {
	var a driver.Conn

	d := wrapConn(dbConnDetails{}, a, nil)

	assert.IsType(t, &wConn{}, d)
}

type connAllInterfacesMock struct {
}

func (c *connAllInterfacesMock) CheckNamedValue(value *driver.NamedValue) error {
	return nil
}

func (c *connAllInterfacesMock) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	return nil, nil
}

func (c *connAllInterfacesMock) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	return nil, nil
}

func (c *connAllInterfacesMock) Query(query string, args []driver.Value) (driver.Rows, error) {
	return nil, nil
}

func (c *connAllInterfacesMock) Exec(query string, args []driver.Value) (driver.Result, error) {
	return nil, nil
}

func (c *connAllInterfacesMock) Prepare(query string) (driver.Stmt, error) {
	return nil, nil
}

func (c *connAllInterfacesMock) Close() error {
	return nil
}

func (c *connAllInterfacesMock) Begin() (driver.Tx, error) {
	return nil, nil
}

func Test_connAllInterfacesCase(t *testing.T) {
	mock := &connAllInterfacesMock{}

	d := wrapConn(dbConnDetails{}, mock, nil)

	assert.IsType(t, &w_conn_Execer_ExecerContext_Queryer_QueryerContext_NamedValueChecker{}, d)
}

func Test_convertBooleansToInt(t *testing.T) {
	type args struct {
		args []bool
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "true true true false",
			args: args{args: []bool{true, true, true, false}},
			want: 0b1110,
		},
		{
			name: "true true true true",
			args: args{args: []bool{true, true, true, true}},
			want: 0b1111,
		},
		{
			name: "false true false",
			args: args{args: []bool{false, true, false}},
			want: 0b10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, convertBooleansToInt(tt.args.args...), "convertBooleansToInt(%v)")
		})
	}
}
