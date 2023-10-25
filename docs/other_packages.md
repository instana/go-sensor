## Tracing Other Go Packages

The IBM Instana Go Tracer allows you to trace the Go standard library for HTTP incoming and outgoing requests, and drivers that implement the database/sql/driver interfaces.

However, there are several popular Go packages commonly used in projects that are desired to be traced as well.

For this purpose, the Go tracer relies on additional packages prefixed with "insta" and followed by the original package name. Eg: gorm -> instagorm and so on.

These additional packages can be imported into your project in order to trace that particular package.
Each package has a separate README.md documentation, as well as its own Go Package website documentation for your convenience.

For instance, if you want to trace HTTP calls for the Gin framework, you can import, along side the tracer, the [instagin](https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instagin) package.

### Available Packages

The easiest way to check which tracing packages are currently supported by the tracer, you can navigate to the [instrumentation](../instrumentation/) directory and check the list of available packages.

If you wish to trace a package that is not part of the list, you have two options: to request the tracer team to instrument a certain package or you can instrument it manually by using our API.

-----
[README](../README.md) |
[Tracer Options](options.md) |
[Tracing HTTP Outgoing Requests](roundtripper.md) |
[Tracing SQL Driver Databases](sql.md) |
[Instrumenting Code Manually](manual_instrumentation.md)
