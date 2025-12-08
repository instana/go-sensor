# Oracle Database Example with Instana

Example demonstrating Instana instrumentation for Oracle database using sqlx and godror driver.

## Prerequisites

- Docker and Docker Compose
- Instana agent credentials (INSTANA_AGENT_KEY and INSTANA_DOWNLOAD_KEY)

## Quick Start

1. Set environment variables:
```bash
export INSTANA_AGENT_KEY=your_agent_key
export INSTANA_DOWNLOAD_KEY=your_download_key
```

2. Start services:
```bash
docker-compose up -d
```

3. Test the application:
```bash
curl http://localhost:8080/oracle
curl http://localhost:8080/health
```

## Connection String Formats

The example supports both Oracle connection string formats:

### TNS Format (default)
```bash
scott/tiger@(description=(address=(protocol=tcp)(host=hostdb1)(port=1521))(connect_data=(service_name=FREEPDB1)))
```

### godror Key-Value Format
```bash
user="scott" password="tiger" connectString="hostdb1:1521/FREEPDB1"
```

------

*update the following environment variable in docker-compose file to override the default Oracle connection string.*

```bash
ORACLE_CONNECTION_STRING="your_connection_string"
```
