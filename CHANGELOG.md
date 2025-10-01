# Changelog

This document provides information related to the releases of the core SDK of
go-sensor project.

## [v1.71.0](https://github.com/instana/go-sensor/releases/tag/v1.71.0)

* feat: add support for disabling log collection at the tracer level

## [v1.70.0](https://github.com/instana/go-sensor/releases/tag/v1.70.0)

* feat: add support for Go version 1.25

## [v1.69.1](https://github.com/instana/go-sensor/releases/tag/v1.69.1)

* chore: update the examples and docs to reflect the correct usage of the `InitCollector` API

## [v1.69.0](https://github.com/instana/go-sensor/releases/tag/v1.69.0)

* feat: cleanup pipeline runs for currency monitoring
* fix: ensure consistent default for MaxLogsPerSpan
* fix: replace secure packages for random number generation
* chore: disable Go toolchain enforcement using `go mod edit -toolchain=none`

## [v1.68.0](https://github.com/instana/go-sensor/releases/tag/v1.68.0)

* refactor: migrate internal/pprof to use google/pprof package
* fix: handle int overflow in Linux CPU stat calculation

## [v1.67.7](https://github.com/instana/go-sensor/releases/tag/v1.67.7)

* fix: add error logging for the APIs

## [v1.67.6](https://github.com/instana/go-sensor/releases/tag/v1.67.6)

* refactor: optimise trace header extraction by adding an early exit when both headers are found
* refactor: replace the custom request and header cloning logic with the standard http.Request.Clone and http.Header.Clone functions

## [v1.67.5](https://github.com/instana/go-sensor/releases/tag/v1.67.5)

* fix: replace deprecated ioutil methods with recommended alternatives
* fix: add a fix for Instagorm unit test failure in Windows

## [v1.67.4](https://github.com/instana/go-sensor/releases/tag/v1.67.4)

* fix: update the default secret matcher to use 'pass' instead of 'password' to align with the tracer specification and other services
* chore: add an example demonstrating HTTP secret matcher configuration

## [v1.67.3](https://github.com/instana/go-sensor/releases/tag/v1.67.3)

* fix: add a fix for tracer config precedence to the correct order

## [v1.67.2](https://github.com/instana/go-sensor/releases/tag/v1.67.2)

* feat: add mongo driver v2 instrumentation
* fix: update the dependencies to the latest version, involving security fixes.

## [v1.67.1](https://github.com/instana/go-sensor/releases/tag/v1.67.1)

* refactor: update go tracer examples with InitCollector API
* refactor: update readme of all instrumentation libraries with InitCollector API
* refactor: update all unit tests in instrumentation libraries with InitCollector API

## [v1.67.0](https://github.com/instana/go-sensor/releases/tag/v1.67.0)

* feat: add support for Go version 1.24
* Chore: container registry for pulling images has been updated

## [v1.66.2](https://github.com/instana/go-sensor/releases/tag/v1.66.2)

* refactor: update Go Tracer unit tests with InitCollector API
* fix: add a fix for the delayed span testcase failures
* fix: add a fix for unit test failures while using instana.InitCollector() API

## [v1.66.1](https://github.com/instana/go-sensor/releases/tag/v1.66.1)

* fix: add a fix for the issue where connection strings for databases such as PostgreSQL were wrongly identified as Redis connections

## [v1.66.0](https://github.com/instana/go-sensor/releases/tag/v1.66.0)

* perf: optimize SQL-based database instrumentation
* refactor: parsing gateway IP with bit shift instead of string on a loop
* fix: add a fix as the test database is not cleaned up from the Cosmos account when integration tests fail
* chore: add a pipeline job to run unit and integration tests against the Golang release candidate
* chore: sonarcloud pipeline

## [v1.65.0](https://github.com/instana/go-sensor/releases/tag/v1.65.0)

* feat: add support for the generic serverless agent.

## [v1.64.0](https://github.com/instana/go-sensor/releases/tag/v1.64.0)

* feat: add support for the latest Golang runtime version 1.23.

## [v1.63.1](https://github.com/instana/go-sensor/releases/tag/v1.63.1)

* fix: add fix for DB2 spans being tagged as MySQL spans instead of generic DB spans.
* chore: update dependencies of example programs.
* chore: add example for the SQL instrumentation using sql.OpenDB API.

## [v1.63.0](https://github.com/instana/go-sensor/releases/tag/v1.63.0)

* feat: add support for Azure Container Apps has been added in Azure agent.
* chore: update the dependencies of example programs.
* chore: improvements in the currency automation script.
* chore: Tekton pipeline for automating Go Tracer Currency Report.
* chore: add support for GitHub action for release summary.
