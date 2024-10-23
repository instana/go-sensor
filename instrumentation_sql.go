// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package instana

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"net/url"
	"regexp"
	"slices"
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
func InstrumentSQLDriver(sensor TracerLogger, name string, driver driver.Driver) {
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
		Driver:     driver,
		driverName: name,
		sensor:     sensor,
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
func SQLInstrumentAndOpen(sensor TracerLogger, driverName, dataSourceName string) (*sql.DB, error) {
	if d, ok := drivers[driverName]; ok {
		InstrumentSQLDriver(sensor, driverName, d)
	}

	return SQLOpen(driverName, dataSourceName)
}

type wrappedSQLDriver struct {
	driver.Driver

	driverName string
	sensor     TracerLogger
}

func (drv *wrappedSQLDriver) Open(name string) (driver.Conn, error) {
	conn, err := drv.Driver.Open(name)
	if err != nil {
		return conn, err
	}

	if connAlreadyWrapped(conn) {
		return conn, nil
	}

	w := wrapConn(getDBConnDetails(name, drv.driverName), conn, drv.sensor)

	return w, nil
}

// getDBConnDetails returns db connection details parsing connection URI and checking driver name
func getDBConnDetails(connStr, driverName string) DbConnDetails {

	if isDB2driver(driverName) {
		if details, ok := parseDB2ConnDetailsKV(connStr); ok {
			return details
		}

		return DbConnDetails{RawString: connStr}
	}

	return ParseDBConnDetails(connStr)
}

func postgresSpan(ctx context.Context, conn DbConnDetails, query string, sensor TracerLogger) ot.Span {
	tags := ot.Tags{
		"pg.stmt": query,
		"pg.user": conn.User,
		"pg.host": conn.Host,
	}

	if conn.Schema != "" {
		tags["pg.db"] = conn.Schema
	} else {
		tags["pg.db"] = conn.RawString
	}

	if conn.Port != "" {
		tags["pg.port"] = conn.Port
	}

	opts := []ot.StartSpanOption{ext.SpanKindRPCClient, tags}
	if parentSpan, ok := SpanFromContext(ctx); ok {
		opts = append(opts, ot.ChildOf(parentSpan.Context()))
	}

	return sensor.StartSpan(string(PostgreSQLSpanType), opts...)
}

func mySQLSpan(ctx context.Context, conn DbConnDetails, query string, sensor TracerLogger) ot.Span {
	tags := ot.Tags{
		"mysql.stmt": query,
		"mysql.user": conn.User,
		"mysql.host": conn.Host,
	}

	if conn.Schema != "" {
		tags["mysql.db"] = conn.Schema
	} else {
		tags["mysql.db"] = conn.RawString
	}

	if conn.Port != "" {
		tags["mysql.port"] = conn.Port
	}

	opts := []ot.StartSpanOption{ext.SpanKindRPCClient, tags}
	if parentSpan, ok := SpanFromContext(ctx); ok {
		opts = append(opts, ot.ChildOf(parentSpan.Context()))
	}

	return sensor.StartSpan(string(MySQLSpanType), opts...)
}

func redisSpan(ctx context.Context, conn DbConnDetails, query string, cmd string, sensor TracerLogger) ot.Span {

	// The command string will be empty if the database name is determined
	// from the connection URL rather than the query.
	// Therefore, the Redis command should be parsed from the query.
	if cmd == "" {
		cmd, _ = parseRedisQuery(query)
	}

	tags := ot.Tags{
		"redis.command": cmd,
	}

	if conn.Error != nil {
		tags["redis.error"] = conn.Error.Error()
	}

	connection := conn.Host + ":" + conn.Port

	if conn.Host == "" || conn.Port == "" {
		i := strings.LastIndex(conn.RawString, "@")
		connection = conn.RawString[i+1:]
	}

	tags["redis.connection"] = connection

	opts := []ot.StartSpanOption{ext.SpanKindRPCClient, tags}
	if parentSpan, ok := SpanFromContext(ctx); ok {
		opts = append(opts, ot.ChildOf(parentSpan.Context()))
	}

	return sensor.StartSpan(string(RedisSpanType), opts...)
}

func couchbaseSpan(ctx context.Context, conn DbConnDetails, query string, sensor TracerLogger) ot.Span {
	tags := ot.Tags{
		"couchbase.hostname": conn.RawString,
		"couchbase.sql":      query,
	}

	opts := []ot.StartSpanOption{ext.SpanKindRPCClient, tags}
	if parentSpan, ok := SpanFromContext(ctx); ok {
		opts = append(opts, ot.ChildOf(parentSpan.Context()))
	}

	return sensor.StartSpan(string(CouchbaseSpanType), opts...)
}

func cosmosSpan(ctx context.Context, conn DbConnDetails, query string, sensor TracerLogger) ot.Span {
	tags := ot.Tags{
		"cosmos.cmd": query,
	}

	opts := []ot.StartSpanOption{ext.SpanKindRPCClient, tags}
	if parentSpan, ok := SpanFromContext(ctx); ok {
		opts = append(opts, ot.ChildOf(parentSpan.Context()))
	}

	return sensor.StartSpan(string(CosmosSpanType), opts...)
}

func genericSQLSpan(ctx context.Context, conn DbConnDetails, query string, sensor TracerLogger) ot.Span {
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

	return sensor.StartSpan("sdk.database", opts...)
}

// retrieveDBNameAndCmd attempts to guess what is the database based on the query.
// It accepts a string that may be a database query
// And returns the database name and query command
func retrieveDBNameAndCmd(q string) (cmd string, dbName string) {

	if cmd, ok := parseRedisQuery(q); ok {
		return cmd, "redis"
	}

	return cmd, dbName
}

// parseRedisQuery attempts to guess if the input string is a valid Redis query.
// parameters:
//   - query (string): a string that may be a redis query
//
// returns:
//   - command (string): The Redis command if the input is identified as a Redis query.
//     This would typically be the first word of the Redis command, such as "SET", "CONFIG GET" etc.
//     If the input is not a Redis query, this value will be an empty string.
//   - isRedis (bool): A boolean value, `true` if the input is recognized as a Redis query,
//     otherwise `false`.
func parseRedisQuery(query string) (command string, isRedis bool) {
	query = strings.TrimSpace(query)
	if len(query) == 0 {
		return "", false
	}

	// getting first two words of the query
	parts := strings.SplitN(query, " ", 3)
	command = strings.ToUpper(parts[0])

	_, isRedis = redisCommands[command]
	if !isRedis && len(parts) > 1 {
		command = strings.ToUpper(parts[0] + " " + parts[1])
		_, isRedis = redisCommands[command]
	}

	return
}

// StartSQLSpan creates a span based on DbConnDetails and a query, and attempts to detect which kind of database it belongs.
// If a database is detected and it is already part of the registered spans, the span details will be specific to that
// database.
// Otherwise, the span will have generic database fields.
func StartSQLSpan(ctx context.Context, conn DbConnDetails, query string, sensor TracerLogger) (sp ot.Span, dbKey string) {
	return startSQLSpan(ctx, conn, query, sensor)
}

func startSQLSpan(ctx context.Context, conn DbConnDetails, query string, sensor TracerLogger) (sp ot.Span, dbKey string) {

	var dbCmd string
	if conn.DatabaseName == "" {
		dbCmd, conn.DatabaseName = retrieveDBNameAndCmd(query)
	}

	switch conn.DatabaseName {
	case "postgres":
		return postgresSpan(ctx, conn, query, sensor), "pg"
	case "redis":
		return redisSpan(ctx, conn, query, dbCmd, sensor), "redis"
	case "mysql":
		return mySQLSpan(ctx, conn, query, sensor), "mysql"
	case "couchbase":
		return couchbaseSpan(ctx, conn, query, sensor), "couchbase"
	case "cosmos":
		return cosmosSpan(ctx, conn, query, sensor), "cosmos"
	}

	return genericSQLSpan(ctx, conn, query, sensor), "db"
}

type DbConnDetails struct {
	RawString    string
	Host, Port   string
	Schema       string
	User         string
	DatabaseName string
	Error        error
}

func ParseDBConnDetails(connStr string) DbConnDetails {
	strategies := [...]func(string) (DbConnDetails, bool){
		parseMySQLGoSQLDriver,
		parsePostgresConnDetailsKV,
		parseMySQLConnDetailsKV,
		parseRedisConnString,
		parseDBConnDetailsURI,
	}
	for _, parseFn := range strategies {
		if details, ok := parseFn(connStr); ok {
			return details
		}
	}

	return DbConnDetails{RawString: connStr}
}

// isDB2driver checks whether the driver belongs to IBM Db2 database.
// The driver name is checked against known Db2 drivers.
func isDB2driver(name string) bool {

	// Add the driver name in the knownDB2drivers array
	// if a new Db2 driver needs to be supported.
	knownDB2drivers := []string{
		"go_ibm_db",
	}

	return slices.Contains(knownDB2drivers, name)
}

// parseDBConnDetailsURI attempts to parse a connection string as an URI, assuming that it has
// following format: [scheme://][user[:[password]]@]host[:port][/schema][?attribute1=value1&attribute2=value2...]
func parseDBConnDetailsURI(connStr string) (DbConnDetails, bool) {
	u, err := url.Parse(connStr)
	if err != nil {
		return DbConnDetails{}, false
	}

	if u.Scheme == "" {
		return DbConnDetails{}, false
	}

	path := ""
	if len(u.Path) > 1 {
		path = u.Path[1:]
	}

	details := DbConnDetails{
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

	if u.Scheme == "postgres" {
		details.DatabaseName = u.Scheme
	}

	return details, true
}

var postgresKVPasswordRegex = regexp.MustCompile(`(^|\s)password=[^\s]+(\s|$)`)

// parsePostgresConnDetailsKV parses a space-separated PostgreSQL-style connection string
func parsePostgresConnDetailsKV(connStr string) (DbConnDetails, bool) {
	var details DbConnDetails

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
		return DbConnDetails{}, false
	}

	details.RawString = postgresKVPasswordRegex.ReplaceAllString(connStr, " ")
	details.DatabaseName = "postgres"

	return details, true
}

var mysqlKVPasswordRegex = regexp.MustCompile(`(?i)(^|;)Pwd=[^;]+(;|$)`)

// parseMySQLConnDetailsKV parses a semicolon-separated MySQL-style connection string
func parseMySQLConnDetailsKV(connStr string) (DbConnDetails, bool) {
	details := DbConnDetails{RawString: connStr, DatabaseName: "mysql"}

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
		return DbConnDetails{}, false
	}

	details.RawString = mysqlKVPasswordRegex.ReplaceAllString(connStr, ";")

	return details, true
}

// parseDB2ConnDetailsKV parses a semicolon-separated MySQL-style connection string for DB2
func parseDB2ConnDetailsKV(connStr string) (DbConnDetails, bool) {
	details := DbConnDetails{RawString: connStr, DatabaseName: "db2"}

	for _, field := range strings.Split(connStr, ";") {
		fieldNorm := strings.ToLower(field)

		var (
			prefix   string
			fieldPtr *string
		)
		switch {
		case strings.HasPrefix(fieldNorm, "server="):
			address := field[len("server="):]
			addrFields := strings.Split(address, ":")
			details.Host = addrFields[0]
			if len(addrFields) > 1 {
				details.Port = addrFields[1]
			}
			continue
		case strings.HasPrefix(fieldNorm, "hostname="):
			prefix, fieldPtr = "hostname=", &details.Host
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

	details.RawString = mysqlKVPasswordRegex.ReplaceAllString(connStr, ";")

	return details, true
}

var mySQLGoDriverRe = regexp.MustCompile(`^(.*):(.*)@((.*)\((.*):([0-9]+)\))?\/(.*)$`)

// parseMySQLGoSQLDriver parses the connection string from https://github.com/go-sql-driver/mysql
// Format: user:password@protocol(host:port)/databasename
// When protocol(host:port) is omitted, assume "tcp(localhost:3306)"
func parseMySQLGoSQLDriver(connStr string) (DbConnDetails, bool) {
	// Expected matches
	// 0 - Entire match. eg: go:gopw@tcp(localhost:3306)/godb
	// 1 - User
	// 2 - password
	// 3 - protocol+host+port. Eg: tcp(localhost:3306)
	// 4 - protocol (if "" use tcp)
	// 5 - host (if "" use localhost)
	// 6 - port (if "" use 3306)
	// 7 - database name
	matches := mySQLGoDriverRe.FindAllStringSubmatch(connStr, -1)

	if len(matches) == 0 {
		return DbConnDetails{}, false
	}

	values := matches[0]

	host := values[5]
	port := values[6]

	if host == "" {
		host = "localhost"
	}

	if port == "" {
		port = "3306"
	}

	d := DbConnDetails{
		RawString:    connStr,
		User:         values[1],
		Host:         host,
		Port:         port,
		Schema:       values[7],
		DatabaseName: "mysql",
	}

	return d, true
}

var redisOptionalUser = regexp.MustCompile(`^(.*:\/\/)?(.+)?:.+@(.+):(\d+)`)

// parseRedisConnString attempts to parse: user:password@host:port
// Based on conn string from github.com/bonede/go-redis-driver
func parseRedisConnString(connStr string) (DbConnDetails, bool) {
	// Expected matches
	// 0 - mysql://user:password@localhost:9898 or db://user:password@localhost:9898 and so on
	// 1 - mysql:// or db:// and so on
	// 2 - user
	// 3 - localhost
	// 4 - 1234
	matches := redisOptionalUser.FindAllStringSubmatch(connStr, -1)

	var d = DbConnDetails{}

	if len(matches) == 0 {
		return d, false
	}

	// We want to ignore the first match. for instance db:// or mysql:// will be ignored if matched
	if matches[0][1] == "" {
		return DbConnDetails{
				Host:         matches[0][3],
				Port:         matches[0][4],
				DatabaseName: "redis",
				RawString:    connStr,
			},
			true
	}

	return d, false
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
