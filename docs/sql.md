## Tracing SQL Driver Databases

The tracer provides means to trace database calls made through the Go standard library, for drivers compliant with the [database/sql/driver package](https://pkg.go.dev/database/sql/driver@go1.21.3).

The tracer is able to auto detect popular databases such as Postgres and MySQL, but will still provide more generic database spans to every driver that fulfills the Go standard library.

Tracing your database calls with the Instana Go Tracer is easier than you think. Simply replace the standard `sql.Open` function by the `instana.SQLInstrumentAndOpen` wrapper function.

In practical terms, your current code that looks something like this:

```go
db, err := sql.Open(s, driverName, connString)
```

Will be changed to this:

```go
c = instana.InitCollector(&instana.Options{
  Service: "Database Call App",
  Tracer:  instana.DefaultTracerOptions(),
})

db, err := instana.SQLInstrumentAndOpen(c, driverName, connString)
```

The `instana.SQLInstrumentAndOpen` will return the expected `(*sql.DB, error)` return, so the rest of your code needs no further changes.

### Complete Example

[MySQL Example](../example/sql-mysql/main.go)
```go
package main

import (
  "io"
  "net/http"

  _ "github.com/go-sql-driver/mysql"
  instana "github.com/instana/go-sensor"
)

func main() {
  s := instana.InitCollector(&instana.Options{
    Service: "MySQL app",
    Tracer:  instana.DefaultTracerOptions(),
  })

  db, err := instana.SQLInstrumentAndOpen(s, "mysql", "go:gopw@tcp(localhost:3306)/godb")
  if err != nil {
    panic(err)
  }

  r, err := db.QueryContext(req.Context(), "SELECT 'Current date is' || CURDATE();")

  if err != nil {
    panic(err)
  }

  var buf, res string

  for r.Next() {
    r.Scan(&buf)
    res += " " + buf
  }
}
```

## Tracing IBM Db2 Driver Databases

In Go Tracer, connection strings are used to distinguish between SQL-like databases based on the database/sql package. This approach has a limitation: MySQL and IBM Db2 connection strings are very similar, which can result in the IBM Db2 driver being incorrectly identified as MySQL.

To resolve this issue, the driver name is now used to differentiate between MySQL and IBM Db2. Currently, only the `go_ibm_db` driver is recognized as the official driver for IBM Db2. If a different driver or driver name is used, the database will still be identified as MySQL. If you need support for a new driver for IBM Db2, please raise a [GitHub issue](https://github.com/instana/go-sensor/issues) in the Go Tracer repository.

Example: Instrumenting Db2
Instrumenting Db2 is the same as in the above example. You only need to pass the driver name as `go_ibm_db` to the `SQLInstrumentAndOpen` function.

```go

  // Example of instrumenting Db2
  db, err := instana.SQLInstrumentAndOpen(s, "go_ibm_db", "connection_string")
  if err != nil {
      log.Fatal(err)
  }

```

Ensure you use `go_ibm_db` as the driver name to correctly identify the IBM Db2 database.

-----
[README](../README.md) |
[Tracer Options](options.md) |
[Tracing HTTP Outgoing Requests](roundtripper.md) |
[Tracing Other Go Packages](other_packages.md) |
[Instrumenting Code Manually](manual_instrumentation.md)
