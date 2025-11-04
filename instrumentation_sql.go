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

// sqlSpanData contains the data for creating a sql db span
type sqlSpanData struct {
	m           *sync.Mutex
	connDetails DbConnDetails
	query       string
	tags        ot.Tags
}

// sqlSpanOption is a function that applies a configuration to sqlSpanConfig.
type sqlSpanOption func(*sqlSpanData)

// withQuery specifies the query that will be set to sqlSpanData.
func withQuery(query string) sqlSpanOption {
	return func(c *sqlSpanData) {
		c.query = query
	}
}

// getSQLSpanData returns instance of sqlSpanData while creating a connection to DB
func getSQLSpanData(c DbConnDetails, opts ...sqlSpanOption) *sqlSpanData {
	var m sync.Mutex
	tags := make(ot.Tags)

	// Retrieve a tagging function from tagsFuncMap based on the database name.
	// If no specific function is found, default to using withGenericSQLTags.
	// Apply the retrieved or default tagging function to tags.
	tf, ok := tagsFuncMap[db(c.DatabaseName)]
	if !ok {
		tf = withGenericSQLTags
	}

	tf(&c).Apply(tags)

	spanData := &sqlSpanData{
		m:           &m,
		connDetails: c,
		tags:        tags,
	}

	for _, opt := range opts {
		opt(spanData)
	}

	return spanData
}

func (s *sqlSpanData) updateDBNameInSpanData(dbName string) {
	if dbName != "" {
		s.m.Lock()
		defer s.m.Unlock()
		s.connDetails.DatabaseName = dbName
	}
}

func (s *sqlSpanData) addTag(key string, val string) {
	s.m.Lock()
	defer s.m.Unlock()
	s.tags[key] = val
}

func (s *sqlSpanData) updateDBQuery(query string) {
	s.m.Lock()
	defer s.m.Unlock()
	s.query = query
}

// start a new sql span
func (s *sqlSpanData) start(
	ctx context.Context,
	sensor TracerLogger) (sp ot.Span, dbKey string) {

	if s.connDetails.DatabaseName == "" {
		s.parseDatabaseFromQuery()
	}

	switch db(s.connDetails.DatabaseName) {

	case postgres:
		return s.postgresSpan(ctx, sensor), pg_db_key
	case redis:
		return s.redisSpan(ctx, sensor), redis_db_key
	case mysql:
		return s.mySQLSpan(ctx, sensor), mysql_db_key
	case couchbase:
		return s.couchbaseSpan(ctx, sensor), couchbase_db_key
	case cosmos:
		return s.cosmosSpan(ctx, sensor), cosmos_db_key
	}

	return s.genericSQLSpan(ctx, sensor), generic_sql_db_key

}

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
// of a driver previously registered using `instana.InstrumentSQLDriver()`
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

func (s *sqlSpanData) postgresSpan(
	ctx context.Context,
	sensor TracerLogger) (sp ot.Span) {
	s.addTag("pg.stmt", s.query)

	opts := []ot.StartSpanOption{ext.SpanKindRPCClient, s.tags}
	if parentSpan, ok := SpanFromContext(ctx); ok {
		opts = append(opts, ot.ChildOf(parentSpan.Context()))
	}

	return sensor.StartSpan(string(PostgreSQLSpanType), opts...)

}

func (s *sqlSpanData) mySQLSpan(
	ctx context.Context,
	sensor TracerLogger) ot.Span {

	s.addTag("mysql.stmt", s.query)

	opts := []ot.StartSpanOption{ext.SpanKindRPCClient, s.tags}
	if parentSpan, ok := SpanFromContext(ctx); ok {
		opts = append(opts, ot.ChildOf(parentSpan.Context()))
	}

	return sensor.StartSpan(string(MySQLSpanType), opts...)
}

func (s *sqlSpanData) redisSpan(
	ctx context.Context,
	sensor TracerLogger) ot.Span {

	cmd, _ := parseRedisQuery(s.query)
	s.addTag("redis.command", cmd)

	opts := []ot.StartSpanOption{ext.SpanKindRPCClient, s.tags}
	if parentSpan, ok := SpanFromContext(ctx); ok {
		opts = append(opts, ot.ChildOf(parentSpan.Context()))
	}

	return sensor.StartSpan(string(RedisSpanType), opts...)
}

func (s *sqlSpanData) couchbaseSpan(
	ctx context.Context,
	sensor TracerLogger) ot.Span {

	s.addTag("couchbase.sql", s.query)

	opts := []ot.StartSpanOption{ext.SpanKindRPCClient, s.tags}
	if parentSpan, ok := SpanFromContext(ctx); ok {
		opts = append(opts, ot.ChildOf(parentSpan.Context()))
	}

	return sensor.StartSpan(string(CouchbaseSpanType), opts...)
}

func (s *sqlSpanData) cosmosSpan(
	ctx context.Context,
	sensor TracerLogger) ot.Span {

	s.addTag("cosmos.cmd", s.query)

	opts := []ot.StartSpanOption{ext.SpanKindRPCClient, s.tags}
	if parentSpan, ok := SpanFromContext(ctx); ok {
		opts = append(opts, ot.ChildOf(parentSpan.Context()))
	}

	return sensor.StartSpan(string(CosmosSpanType), opts...)
}

func (s *sqlSpanData) genericSQLSpan(
	ctx context.Context,
	sensor TracerLogger) ot.Span {

	s.addTag(string(ext.DBStatement), s.query)

	opts := []ot.StartSpanOption{ext.SpanKindRPCClient, s.tags}
	if parentSpan, ok := SpanFromContext(ctx); ok {
		opts = append(opts, ot.ChildOf(parentSpan.Context()))
	}

	return sensor.StartSpan("sdk.database", opts...)
}

// parseDatabaseFromQuery attempts to guess what is the database based on the query
func (s *sqlSpanData) parseDatabaseFromQuery() {
	s.updateSpanDataIfRedis()
}

// updateSpanDataIfRedis validates the db query, checks if the database is redis.
// if it is redis, update database name and tags with redis command.
func (s *sqlSpanData) updateSpanDataIfRedis() {

	if _, ok := parseRedisQuery(s.query); ok {

		// The tags will be containing generic SQL tags as it was treated generic SQL span,
		// as go sensor was not able to determine the database from connection details.
		// Hence removing the existing tags and adding default redis connection tags.
		s.tags = make(ot.Tags)
		withRedisTags(&s.connDetails).Apply(s.tags)

		s.updateDBNameInSpanData(Redis)
	}
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

	s := getSQLSpanData(conn, withQuery(query))
	return s.start(ctx, sensor)
}

// DbConnDetails holds the details of a database connection parsed from a connection string.
type DbConnDetails struct {
	RawString    string
	Host, Port   string
	Schema       string
	User         string
	DatabaseName string
	Error        error
}

// ParseDBConnDetails parses a database connection string (connStr) and returns a DbConnDetails struct.
// This struct contains the details necessary to establish a connection to the database.
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

	//
	// From the official postgresql doc - https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING
	//
	// "The URI scheme designator can be either postgresql:// or postgres://"
	//
	if u.Scheme == "postgres" || u.Scheme == "postgresql" {
		details.DatabaseName = "postgres"
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

var mySQLGoDriverRe = regexp.MustCompile(`^(.*):(.*)@((.*)\((.*?)(?::([0-9]+))?\))?\/(.*)$`)

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

	schema := values[7]

	// To remove params if any are present
	if i := strings.Index(schema, "?"); i != -1 {
		schema = schema[:i]
	}

	if port == "" {
		port = "3306"
	}

	d := DbConnDetails{
		RawString:    connStr,
		User:         values[1],
		Host:         host,
		Port:         port,
		Schema:       schema,
		DatabaseName: "mysql",
	}

	return d, true
}

// parseRedisConnString attempts to parse: user:password@host:port
// Based on conn string from github.com/bonede/go-redis-driver
func parseRedisConnString(connStr string) (DbConnDetails, bool) {

	//
	// Regex : ^(.*:\/\/)?([^:]+)?(?::(?:[^@]+))?@([^:]+):(\d+)
	//
	// This updated regular expression addresses an issue where connection strings for
	// databases like PostgreSQL were incorrectly identified as Redis connections.
	// The regex is divided into five parts:
	// 1. The URI scheme (optional)
	// 2. The username (optional part of the URI)
	// 3. The optional password, which is separated from the username by a colon
	// 4. The hostname
	// 5. The port number
	//
	redisOptionalUser := regexp.MustCompile(`^(.*:\/\/)?([^:]+)?(?::(?:[^@]+))?@([^:]+):(\d+)`)

	// Expected matches and index
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

	// We want to ignore the first match. for instance db://, mysql:// or postgresql:// will be ignored if matched
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
