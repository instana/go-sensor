# Changelog

This document provides information related to the releases of the core SDK of
go-sensor project.

## [v1.71.0](https://github.com/instana/go-sensor/releases/tag/v1.71.0)

* feat: Ability to disable log collection at the tracer level

## [v1.70.0](https://github.com/instana/go-sensor/releases/tag/v1.70.0)

* feat: Support added for go version 1.25

## [v1.69.1](https://github.com/instana/go-sensor/releases/tag/v1.69.1)

* chore: updated the examples and docs to reflect the correct usage of the `InitCollector` API.

## [v1.69.0](https://github.com/instana/go-sensor/releases/tag/v1.69.0)

* feat: cleanup pipeline runs for currency monitoring
* fix: Ensure consistent default for MaxLogsPerSpan
* fix: Replace secure packages for random number generation
* chore: Disable Go toolchain enforcement using `go mod edit -toolchain=none`

## [v1.68.0](https://github.com/instana/go-sensor/releases/tag/v1.68.0)

* refactor: Migrated internal/pprof to use google/pprof package
* fix: Handled int overflow in Linux CPU stat calculation

## [v1.67.7](https://github.com/instana/go-sensor/releases/tag/v1.67.7)

* fix: added error logging for the APIs

## [v1.67.6](https://github.com/instana/go-sensor/releases/tag/v1.67.6)

* refactor: Optimised trace header extraction by adding an early exit when both headers are found
* refactor: Replaced the custom request and header cloning logic with the standard http.Request.Clone and http.Header.Clone functions

## [v1.67.5](https://github.com/instana/go-sensor/releases/tag/v1.67.5)

* fix: Replaced deprecated ioutil methods with recommended alternatives
* fix: Fixed Instagorm unit test failure in Windows.

## [v1.67.4](https://github.com/instana/go-sensor/releases/tag/v1.67.4)

* fix: Updated the default secret matcher to use 'pass' instead of 'password' to align with the tracer specification and other services
* chore: Added an example demonstrating HTTP secret matcher configuration

## [v1.67.3](https://github.com/instana/go-sensor/releases/tag/v1.67.3)

* fix: Fixed tracer config precedence to the correct order

## [v1.67.2](https://github.com/instana/go-sensor/releases/tag/v1.67.2)

* feat: Added mongo driver v2 instrumentation
* fix: Updated the dependencies to the latest version, involving security fixes.

## [v1.67.1](https://github.com/instana/go-sensor/releases/tag/v1.67.1)

* refactor: Updated go tracer examples with InitCollector API
* refactor: Updated readme of all instrumentation libraries with InitCollector API
* refactor: Updated all unit tests in instrumentation libraries with InitCollector API

## [v1.67.0](https://github.com/instana/go-sensor/releases/tag/v1.67.0)

* feat: Support added for go version 1.24
* Chore: Container registry for pulling images has been updated.

## [v1.66.2](https://github.com/instana/go-sensor/releases/tag/v1.66.2)

* refactor: Update Go Tracer unit tests with InitCollector API
* fix: Added fix for the delayed span testcase failures
* fix: Fixed unit test failures while using instana.InitCollector() API

## [v1.66.1](https://github.com/instana/go-sensor/releases/tag/v1.66.1)

* fix: Addressed an issue where connection strings for databases such as PostgreSQL were wrongly identified as Redis connections

## [v1.66.0](https://github.com/instana/go-sensor/releases/tag/v1.66.0)

* perf: optimize SQL-based database instrumentation.
* refactor: parsing gateway IP with bit shift instead of string on a loop.
* fix: the test database is not cleaned up from the Cosmos account when integration tests fail.
* chore: added a pipeline job to run unit and integration tests against the Golang release candidate.
* chore: sonarcloud pipeline.

## [v1.65.0](https://github.com/instana/go-sensor/releases/tag/v1.65.0)

* feat:Added support for generic the serverless agent.

## [v1.64.0](https://github.com/instana/go-sensor/releases/tag/v1.64.0)

* feat:Added support for the latest Golang runtime version 1.23.

## [v1.63.1](https://github.com/instana/go-sensor/releases/tag/v1.63.1)

* fix:Fix for DB2 spans being tagged as MySQL spans instead of generic DB spans.
* chore:Updated dependencies of example programs.
* chore:Added example for the SQL instrumentation using sql.OpenDB API.

## [v1.63.0](https://github.com/instana/go-sensor/releases/tag/v1.63.0)

* feat:Support for Azure Container Apps has been added in Azure agent.
* chore:Updated the dependencies of example programs.
* chore:Improvements in the currency automation script.
* chore:Tekton pipeline for automating Go Tracer Currency Report.
* chore:Support for GitHub action for release summary.
