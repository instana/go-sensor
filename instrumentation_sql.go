// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instana

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"net/url"
	"regexp"
	"strings"
	"sync"

	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"

	_ "unsafe"
)

var (
	sqlDriverRegistrationMu sync.Mutex
)

// InstrumentSQLDriver instruments provided database driver for  use with `sql.Open()`.
// This method will ignore any attempt to register the driver with the same name again.
//
// The instrumented version is registered with `_with_instana` suffix, e.g.
// if `postgres` provided as a name, the instrumented version is registered as
// `postgres_with_instana`.
func InstrumentSQLDriver(sensor *Sensor, name string, driver driver.Driver) {
	sqlDriverRegistrationMu.Lock()
	defer sqlDriverRegistrationMu.Unlock()

	instrumentedName := name + "_with_instana"

	// Check if the instrumented version of a driver has already been registered
	// with database/sql and ignore the second attempt to avoid panicking
	for _, drv := range sql.Drivers() {
		if drv == instrumentedName {
			return
		}
	}

	sql.Register(instrumentedName, &wrappedSQLDriver{
		Driver: driver,
		sensor: sensor,
	})
}

// SQLOpen is a convenience wrapper for `sql.Open()` to use the instrumented version
// of a driver previosly registered using `instana.InstrumentSQLDriver()`
func SQLOpen(driverName, dataSourceName string) (*sql.DB, error) {

	if !strings.HasSuffix(driverName, "_with_instana") {
		driverName += "_with_instana"
	}

	return sql.Open(driverName, dataSourceName)
}

//go:linkname drivers database/sql.drivers
var drivers map[string]driver.Driver

// SQLInstrumentAndOpen returns instrumented `*sql.DB`.
// It takes already registered `driver.Driver` by name, instruments it and additionally registers
// it with different name. After that it returns instrumented `*sql.DB` or error if any.
//
// This function can be used as a convenient shortcut for InstrumentSQLDriver and SQLOpen functions.
// The main difference is that this approach will use the already registered driver and using InstrumentSQLDriver
// requires to explicitly provide an instance of the driver to instrument.
func SQLInstrumentAndOpen(sensor *Sensor, driverName, dataSourceName string) (*sql.DB, error) {
	if d, ok := drivers[driverName]; ok {
		InstrumentSQLDriver(sensor, driverName, d)
	}

	return SQLOpen(driverName, dataSourceName)
}

type wrappedSQLDriver struct {
	driver.Driver

	sensor *Sensor
}

func (drv *wrappedSQLDriver) Open(name string) (driver.Conn, error) {
	conn, err := drv.Driver.Open(name)
	if err != nil {
		return conn, err
	}

	if conn, ok := conn.(*wrappedSQLConn); ok {
		return conn, nil
	}

	return &wrappedSQLConn{
		Conn:    conn,
		details: parseDBConnDetails(name),
		sensor:  drv.sensor,
	}, nil
}

type wrappedSQLConn struct {
	driver.Conn

	details dbConnDetails
	sensor  *Sensor
}

func (conn *wrappedSQLConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	sp := startSQLSpan(ctx, conn.details, query, conn.sensor)
	defer sp.Finish()

	if c, ok := conn.Conn.(driver.QueryerContext); ok {
		res, err := c.QueryContext(ctx, query, args)
		if err != nil && err != driver.ErrSkip {
			sp.LogFields(otlog.Error(err))
		}

		return res, err
	}

	if c, ok := conn.Conn.(driver.Queryer); ok { //nolint:staticcheck
		values, err := sqlNamedValuesToValues(args)
		if err != nil {
			return nil, err
		}

		select {
		default:
		case <-ctx.Done():
			return nil, ctx.Err()
		}

		res, err := c.Query(query, values)
		if err != nil && err != driver.ErrSkip {
			sp.LogFields(otlog.Error(err))
		}

		return res, err
	}

	return nil, driver.ErrSkip
}

func (conn *wrappedSQLConn) Prepare(query string) (driver.Stmt, error) {
	stmt, err := conn.Conn.Prepare(query)
	if err != nil {
		return stmt, err
	}

	if stmt, ok := stmt.(*wrappedSQLStmt); ok {
		return stmt, nil
	}

	return &wrappedSQLStmt{
		Stmt:        stmt,
		connDetails: conn.details,
		query:       query,
		sensor:      conn.sensor,
	}, nil
}

func (conn *wrappedSQLConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	var (
		stmt driver.Stmt
		err  error
	)
	if c, ok := conn.Conn.(driver.ConnPrepareContext); ok {
		stmt, err = c.PrepareContext(ctx, query)
	} else {
		stmt, err = conn.Prepare(query)
	}

	if err != nil {
		return stmt, err
	}

	if stmt, ok := stmt.(*wrappedSQLStmt); ok {
		return stmt, nil
	}

	return &wrappedSQLStmt{
		Stmt:        stmt,
		connDetails: conn.details,
		query:       query,
		sensor:      conn.sensor,
	}, nil
}

func (conn *wrappedSQLConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	sp := startSQLSpan(ctx, conn.details, query, conn.sensor)
	defer sp.Finish()

	if c, ok := conn.Conn.(driver.ExecerContext); ok {
		res, err := c.ExecContext(ctx, query, args)
		if err != nil && err != driver.ErrSkip {
			sp.LogFields(otlog.Error(err))
		}

		return res, err
	}

	if c, ok := conn.Conn.(driver.Execer); ok { //nolint:staticcheck
		values, err := sqlNamedValuesToValues(args)
		if err != nil {
			return nil, err
		}

		select {
		default:
		case <-ctx.Done():
			return nil, ctx.Err()
		}

		res, err := c.Exec(query, values)
		if err != nil && err != driver.ErrSkip {
			sp.LogFields(otlog.Error(err))
		}

		return res, err
	}

	return nil, driver.ErrSkip
}

type wrappedSQLStmt struct {
	driver.Stmt

	connDetails dbConnDetails
	query       string
	sensor      *Sensor
}

func (stmt *wrappedSQLStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	sp := startSQLSpan(ctx, stmt.connDetails, stmt.query, stmt.sensor)
	defer sp.Finish()

	if s, ok := stmt.Stmt.(driver.StmtExecContext); ok {
		res, err := s.ExecContext(ctx, args)
		if err != nil && err != driver.ErrSkip {
			sp.LogFields(otlog.Error(err))
		}

		return res, err
	}

	values, err := sqlNamedValuesToValues(args)
	if err != nil {
		return nil, err
	}

	select {
	default:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	res, err := stmt.Exec(values) //nolint:staticcheck
	if err != nil && err != driver.ErrSkip {
		sp.LogFields(otlog.Error(err))
	}

	return res, err
}

func (stmt *wrappedSQLStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	sp := startSQLSpan(ctx, stmt.connDetails, stmt.query, stmt.sensor)
	defer sp.Finish()

	if s, ok := stmt.Stmt.(driver.StmtQueryContext); ok {
		res, err := s.QueryContext(ctx, args)
		if err != nil && err != driver.ErrSkip {
			sp.LogFields(otlog.Error(err))
		}

		return res, err
	}

	values, err := sqlNamedValuesToValues(args)
	if err != nil {
		return nil, err
	}

	select {
	default:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	res, err := stmt.Stmt.Query(values) //nolint:staticcheck
	if err != nil && err != driver.ErrSkip {
		sp.LogFields(otlog.Error(err))
	}

	return res, err
}

func startSQLSpan(ctx context.Context, conn dbConnDetails, query string, sensor *Sensor) ot.Span {
	tags := ot.Tags{
		string(ext.DBType):      "sql",
		string(ext.DBStatement): query,
		string(ext.PeerAddress): conn.RawString,
	}

	if conn.Schema != "" {
		tags[string(ext.DBInstance)] = conn.Schema
	} else {
		tags[string(ext.DBInstance)] = conn.RawString
	}

	if conn.Host != "" {
		tags[string(ext.PeerHostname)] = conn.Host
	}

	if conn.Port != "" {
		tags[string(ext.PeerPort)] = conn.Port
	}

	opts := []ot.StartSpanOption{ext.SpanKindRPCClient, tags}
	if parentSpan, ok := SpanFromContext(ctx); ok {
		opts = append(opts, ot.ChildOf(parentSpan.Context()))
	}

	return sensor.Tracer().StartSpan("sdk.database", opts...)
}

type dbConnDetails struct {
	RawString  string
	Host, Port string
	Schema     string
	User       string
}

func parseDBConnDetails(connStr string) dbConnDetails {
	strategies := [...]func(string) (dbConnDetails, bool){
		parseDBConnDetailsURI,
		parsePostgresConnDetailsKV,
		parseMySQLConnDetailsKV,
	}
	for _, parseFn := range strategies {
		if details, ok := parseFn(connStr); ok {
			return details
		}
	}

	return dbConnDetails{RawString: connStr}
}

// parseDBConnDetailsURI attempts to parse a connection string as an URI, assuming that it has
// following format: [scheme://][user[:[password]]@]host[:port][/schema][?attribute1=value1&attribute2=value2...]
func parseDBConnDetailsURI(connStr string) (dbConnDetails, bool) {
	u, err := url.Parse(connStr)
	if err != nil {
		return dbConnDetails{}, false
	}

	if u.Scheme == "" {
		return dbConnDetails{}, false
	}

	path := ""
	if len(u.Path) > 1 {
		path = u.Path[1:]
	}

	details := dbConnDetails{
		RawString: connStr,
		Host:      u.Hostname(),
		Port:      u.Port(),
		Schema:    path,
	}

	if u.User != nil {
		details.User = u.User.Username()

		// create a copy without user password
		u := cloneURL(u)
		u.User = url.User(details.User)
		details.RawString = u.String()
	}

	return details, true
}

var postgresKVPasswordRegex = regexp.MustCompile(`(^|\s)password=[^\s]+(\s|$)`)

// parsePostgresConnDetailsKV parses a space-separated PostgreSQL-style connection string
func parsePostgresConnDetailsKV(connStr string) (dbConnDetails, bool) {
	var details dbConnDetails

	for _, field := range strings.Split(connStr, " ") {
		fieldNorm := strings.ToLower(field)

		var (
			prefix   string
			fieldPtr *string
		)
		switch {
		case strings.HasPrefix(fieldNorm, "host="):
			if details.Host != "" {
				// hostaddr= takes precedence
				continue
			}

			prefix, fieldPtr = "host=", &details.Host
		case strings.HasPrefix(fieldNorm, "hostaddr="):
			prefix, fieldPtr = "hostaddr=", &details.Host
		case strings.HasPrefix(fieldNorm, "port="):
			prefix, fieldPtr = "port=", &details.Port
		case strings.HasPrefix(fieldNorm, "user="):
			prefix, fieldPtr = "user=", &details.User
		case strings.HasPrefix(fieldNorm, "dbname="):
			prefix, fieldPtr = "dbname=", &details.Schema
		default:
			continue
		}

		*fieldPtr = field[len(prefix):]
	}

	if details.Schema == "" {
		return dbConnDetails{}, false
	}

	details.RawString = postgresKVPasswordRegex.ReplaceAllString(connStr, " ")

	return details, true
}

var mysqlKVPasswordRegex = regexp.MustCompile(`(?i)(^|;)Pwd=[^;]+(;|$)`)

// parseMySQLConnDetailsKV parses a semicolon-separated MySQL-style connection string
func parseMySQLConnDetailsKV(connStr string) (dbConnDetails, bool) {
	details := dbConnDetails{RawString: connStr}

	for _, field := range strings.Split(connStr, ";") {
		fieldNorm := strings.ToLower(field)

		var (
			prefix   string
			fieldPtr *string
		)
		switch {
		case strings.HasPrefix(fieldNorm, "server="):
			prefix, fieldPtr = "server=", &details.Host
		case strings.HasPrefix(fieldNorm, "port="):
			prefix, fieldPtr = "port=", &details.Port
		case strings.HasPrefix(fieldNorm, "uid="):
			prefix, fieldPtr = "uid=", &details.User
		case strings.HasPrefix(fieldNorm, "database="):
			prefix, fieldPtr = "database=", &details.Schema
		default:
			continue
		}

		*fieldPtr = field[len(prefix):]
	}

	if details.Schema == "" {
		return dbConnDetails{}, false
	}

	details.RawString = mysqlKVPasswordRegex.ReplaceAllString(connStr, ";")

	return details, true
}

// The following code is ported from $GOROOT/src/database/sql/ctxutil.go
//
// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
func sqlNamedValuesToValues(named []driver.NamedValue) ([]driver.Value, error) {
	dargs := make([]driver.Value, len(named))
	for n, param := range named {
		if len(param.Name) > 0 {
			return nil, errors.New("sql: driver does not support the use of Named Parameters")
		}
		dargs[n] = param.Value
	}
	return dargs, nil
}

type dsnConnector struct {
	dsn    string
	driver driver.Driver
}

func (t dsnConnector) Connect(_ context.Context) (driver.Conn, error) {
	return t.driver.Open(t.dsn)
}

func (t dsnConnector) Driver() driver.Driver {
	return t.driver
}
