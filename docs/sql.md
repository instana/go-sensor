## Tracing SQL Driver Databases

The tracer provides means to trace database calls made trough the Go standard library, that is for drivers compliant with the [database/sql/driver package](https://pkg.go.dev/database/sql/driver@go1.21.3).

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

-----
[README](../README.md) |
[Tracer Options](options.md) |
[Tracing HTTP Outgoing Requests](roundtripper.md) |
[Tracing Other Go Packages](other_packages.md) |
[Instrumenting Code Manually](manual_instrumentation.md)
