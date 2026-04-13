# GraphQL Example with Instana Instrumentation

This example demonstrates how to use Instana's GraphQL instrumentation with a simple GraphQL API.

## Setup

### Prerequisites
- Go 1.23 or later
- Instana agent running locally (or configure agent endpoint)

### Installation

1. Navigate to the example directory:
```bash
cd example/graphql
```

2. Install dependencies:
```bash
go mod tidy
```

### Running the Application

**Option 1: Custom HTTP Handler (Default)**
```bash
go run .
```

**Option 2: With GraphQL Playground UI**
```bash
go run . -handler
```

The server will start on `http://localhost:9191`

With the `-handler` flag, you can access the GraphQL Playground at `http://localhost:9191/graphql`

## API Testing with cURL

### Queries

#### 1. Get All Characters
```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"query": "{ characters { id name profession crewMember } }"}' \
  http://localhost:9191/graphql
```

#### 2. Get All Ships
```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"query": "{ ships { id name origin } }"}' \
  http://localhost:9191/graphql
```

#### 3. Get Character by ID
```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"query": "{ characters(id: 1) { id name profession crewMember } }"}' \
  http://localhost:9191/graphql
```

#### 4. Get Ship by ID
```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"query": "{ ships(id: 1) { id name origin } }"}' \
  http://localhost:9191/graphql
```

#### 5. Query Multiple Entities (Named Query)
```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"query": "query myQuery { characters { id name profession crewMember } ships { id name origin } }"}' \
  http://localhost:9191/graphql
```

### Mutations

#### 1. Insert New Character
```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"query": "mutation { insertCharacter(name: \"Alex Kamal\", profession: \"Pilot\", crewMember: true) { id name profession crewMember } }"}' \
  http://localhost:9191/graphql
```

#### 2. Insert New Ship
```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"query": "mutation { insertShip(name: \"Razorback\", origin: \"Mars\") { id name origin } }"}' \
  http://localhost:9191/graphql
```

### Error Handling

#### Query with Invalid Field (Error Example)
```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"query": "{ characters { id name profession crewMember invalidField } }"}' \
  http://localhost:9191/graphql
```

## Pretty Print JSON Output

Add `| jq` to any curl command for formatted output:

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"query": "{ characters { id name profession crewMember } }"}' \
  http://localhost:9191/graphql | jq
```

## Data Model

### Character
- `id` (Int): Unique identifier
- `name` (String): Character name
- `profession` (String): Character's profession
- `crewMember` (Boolean): Whether the character is a crew member

### Ship
- `id` (Int): Unique identifier
- `name` (String): Ship name
- `origin` (String): Ship's origin location

## Initial Data

The application loads initial data from `data.json`:
- 4 characters (James Holden, Naomi Nagata, Amos Burton, Josephus Miller)
- 2 ships (Canterbury, Roccinante)

## Instana Tracing

This example uses Instana's GraphQL instrumentation to trace:
- GraphQL queries
- GraphQL mutations
- GraphQL subscriptions
- Query execution time
- Error tracking

Traces will be visible in your Instana dashboard.
