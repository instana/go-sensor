// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instana

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"net/url"
	"regexp"
	"strings"
	"sync"

	_ "unsafe"

	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
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

	if connAlreadyWrapped(conn) {
		return conn, nil
	}

	w := wrapConn(parseDBConnDetails(name), conn, drv.sensor)

	return w, nil
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
