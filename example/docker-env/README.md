# Docker Environment with Instana Monitoring

A simple Go application with MySQL database running in Docker Compose, monitored by Instana.

> **Tested on**: Linux with Docker

## Overview

- **Application**: Gin web server that queries MySQL and returns the current date
- **Database**: MySQL 8.0
- **Monitoring**: Instana agent (requires local agent running)

## Prerequisites

- Docker and Docker Compose
- Instana agent running locally (see setup below)

## Setup Instana Agent

Before running the application, start the Instana agent:

```bash
docker run \
   --detach \
   --name instana-agent \
   --volume /var/run:/var/run \
   --volume /run:/run \
   --volume /dev:/dev:ro \
   --volume /sys:/sys:ro \
   --volume /var/log:/var/log:ro \
   --privileged \
   -p 42699:42699 \
   --pid=host \
   --env="INSTANA_AGENT_ENDPOINT=****" \
   --env="INSTANA_AGENT_ENDPOINT_PORT=443" \
   --env="INSTANA_AGENT_KEY=****" \
   --env="INSTANA_DOWNLOAD_KEY=****" \
   icr.io/instana/agent
```

> **Note**: Replace `****` with your actual Instana credentials. For the latest agent installation instructions, visit the **Install Agents** page in your Instana UI.

## Quick Start

```bash
# Navigate to docker-env directory
cd example/docker-env

# Start the application and MySQL
docker compose up --build

# Wait for services to be ready (MySQL health check takes ~30s)
```

## Access the Application

```bash
# Test the endpoint
curl http://localhost:8085/gin-test
```

## Configuration

### MySQL Credentials

Default credentials (defined in [`docker-compose.yaml`](docker-compose.yaml:14-17)):
- **User**: `go`
- **Password**: `gopw`
- **Database**: `godb`

### Instana Agent

The application connects to Instana agent at `host.docker.internal`. Ensure your local Instana agent is running and accessible.

Environment variables in [`docker-compose.yaml`](docker-compose.yaml:30-32):
```yaml
INSTANA_AGENT_HOST: host.docker.internal
INSTANA_DEBUG: true
INSTANA_LOG_LEVEL: debug
```

## Monitoring

View metrics in your Instana dashboard:
1. Log in to Instana
2. Navigate to Applications → `mysql-service`
3. View traces and metrics

## Troubleshooting

### MySQL Connection Errors

Wait for MySQL health check to pass (~30 seconds on first start):
```bash
docker compose logs mysql
```

### Instana Agent Connection

Verify agent is accessible:
```bash
# Check if agent is running
curl http://localhost:42699/status
```

## Cleanup

```bash
# Stop and remove containers
docker compose down

# Remove volumes (resets database)
docker compose down -v
```

