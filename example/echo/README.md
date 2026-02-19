# Echo v5 with Instana Instrumentation Example

This example demonstrates how to instrument an Echo v5 web application with Instana for distributed tracing and application performance monitoring (APM).

## Overview

This application showcases:
- **Echo v5 Framework**: Modern, high-performance HTTP web framework for Go
- **Instana Instrumentation Integration**: Automatic distributed tracing for HTTP requests and database queries
- **SQLite Database**: In-memory database with instrumented queries
- **Context Propagation**: Demonstrates how trace context flows from HTTP requests through to database operations

## Features

- ✅ Automatic HTTP request tracing with Echo v5
- ✅ Database query instrumentation with SQLite
- ✅ Context propagation across service boundaries
- ✅ RESTful API endpoints
- ✅ Error handling and HTTP status codes

## Prerequisites

- Go 1.25.0 or higher
- Instana agent running (for trace collection)/ Serverless Endpoint

## Installation

1. Clone the repository and navigate to this example:
```bash
cd example/echo
```

2. Install dependencies:
```bash
go mod download
```

## Running the Application

Start the server:
```bash
go run server.go
```

The server will start on `http://localhost:8080`

## API Endpoints

### 1. Health Check Endpoint
```
GET /myendpoint
```

**Response:**
```json
{
  "message": "pong"
}
```

### 2. Get User by ID
```
GET /users/:id
```

**Parameters:**
- `id` (path parameter): User ID (1, 2, or 3)

**Response (Success - 200):**
```json
{
  "id": 1,
  "name": "Alice",
  "email": "alice@example.com"
}
```

**Response (Not Found - 404):**
```json
{
  "error": "User not found"
}
```

**Response (Server Error - 500):**
```json
{
  "error": "Database error"
}
```

**Example:**
```bash
curl http://localhost:8080/users/1
```

## Sample Data

The application initializes with three sample users:
- **User 1**: Alice (alice@example.com)
- **User 2**: Bob (bob@example.com)
- **User 3**: Charlie (charlie@example.com)

## Code Structure

### Instana Initialization
```go
collector := instana.InitCollector(&instana.Options{
    Service: "echo-v5-example",
})
```
Initializes the Instana collector with a service name for identification in the Instana dashboard.

### Database Instrumentation
```go
db, err := instana.SQLInstrumentAndOpen(collector, "sqlite", ":memory:")
```
`SQLInstrumentAndOpen` is a drop-in replacement for `sql.Open()` that automatically instruments the database driver for distributed tracing.

### Echo Instrumentation
```go
e := instaechov2.New(collector)
```
Creates an instrumented Echo v5 instance that automatically traces all HTTP requests.

### Context Propagation
```go
ctx := c.Request().Context()
err := db.QueryRowContext(ctx, "SELECT id, name, email FROM users WHERE id = ?", c.Param("id"))
```
The trace context from the HTTP request is propagated to the database query, creating a complete trace from the entry point (HTTP) to the exit point (database).

## Key Concepts

### Distributed Tracing
Every HTTP request automatically creates an **entry span** in Instana. When the application makes a database query using the instrumented context, an **exit span** is created and linked to the parent HTTP span. This creates a complete trace showing:
1. HTTP request received
2. Database query executed
3. Response returned

### Automatic Instrumentation
The Instana Go sensor provides automatic instrumentation for:
- HTTP handlers (via `instaechov2.New()`)
- Database queries (via `instana.SQLInstrumentAndOpen()`)
- Context propagation (via standard Go `context.Context`)

No manual span creation is required - the instrumentation handles everything automatically.

## Viewing Traces

Once the application is running and receiving requests:
1. Open your Instana dashboard
2. Navigate to the "Applications" view
3. Find the "echo-v5-example" service
4. View individual traces showing the complete request flow

## License

MIT License - See LICENSE file for details

## Copyright

SPDX-FileCopyrightText: 2026 IBM Corp.
