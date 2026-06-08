# Go AMQP Example with Instana Monitoring

This example demonstrates how to instrument a Go application with RabbitMQ (AMQP) messaging using Instana's Go sensor. The application can be deployed on OpenShift or run locally with Docker.

> **Tested on**: Linux with OpenShift & Docker

## Overview

- **Application**: Gin web server with AMQP producer and consumer
- **Message Broker**: RabbitMQ 3.13
- **Monitoring**: Instana Go sensor with AMQP instrumentation
- **Deployment Options**: OpenShift 4.x or Docker/Docker Compose

## Table of Contents

- [OpenShift Deployment](#openshift-deployment)
  - [Prerequisites](#prerequisites)
  - [Installation Steps](#installation-steps)
  - [Testing the Application](#testing-the-application)
  - [Cleanup](#cleanup)
- [Docker Deployment](#docker-deployment)
  - [Prerequisites](#prerequisites-1)
  - [Installation Steps](#installation-steps-1)
  - [Testing the Application](#testing-the-application-1)
  - [Cleanup](#cleanup-1)
- [Monitoring with Instana](#monitoring-with-instana)
- [Resources](#resources)

---

# OpenShift Deployment

## Prerequisites

- OpenShift cluster access (URL and authentication token)
- `oc` CLI installed
- Instana account credentials

## Installation Steps

### 1. Login to OpenShift

Obtain your authentication token from the OpenShift web console:
1. Navigate to your OpenShift web console
2. Click your username → **Copy Login Command**
3. Click **Display Token** and copy the login command

```bash
oc login https://api.your-openshift-cluster.com:6443 --token=sha256~your-token-here
```

Verify your login:
```bash
oc whoami
oc cluster-info
```

### 2. Deploy Instana Agent

Create a dedicated namespace for the Instana agent:

```bash
oc new-project instana-agent
```

Grant required permissions:
```bash
oc adm policy add-scc-to-user privileged -z instana-agent -n instana-agent
```

Install the agent using Helm:
```bash
helm install instana-agent \
   --repo https://agents.instana.io/helm \
   --namespace instana-agent \
   --set openshift=true \
   --set agent.key=YOUR_AGENT_KEY \
   --set agent.downloadKey=YOUR_DOWNLOAD_KEY \
   --set agent.endpointHost=YOUR_ENDPOINT_HOST \
   --set agent.endpointPort=443 \
   --set cluster.name='YOUR_CLUSTER_NAME' \
   instana-agent
```

> **Note**: Replace placeholder values with your actual Instana credentials from **Settings** → **Agents** → **Installing Instana Agents** in the Instana UI.

Verify agent deployment:
```bash
oc get pods -n instana-agent
oc logs -n instana-agent -l app.kubernetes.io/name=instana-agent --max-log-requests=10 --follow
```

### 3. Create Application Project

```bash
oc new-project instana-amqp-demo
oc project
```

### 4. Build Application Image

Navigate to the example directory:
```bash
cd example/openshift-example
```

Create and execute the build:
```bash
oc new-build --name=amqp-service-app --binary --strategy=docker
oc start-build amqp-service-app --from-dir=. --follow
```

The image will be available at:
```
image-registry.openshift-image-registry.svc:5000/instana-amqp-demo/amqp-service-app:latest
```

> **Note**: Ensure `app.yaml` references this image path.

### 5. Create RabbitMQ Secret

Create the secret with a password before deploying RabbitMQ:

**For development/testing** (using default password):
```bash
oc create secret generic rabbitmq-secret \
  --from-literal=rabbitmq-password=guest \
  -n instana-amqp-demo
```

> **Security Note**: Never commit secrets to version control. Always create them separately using secure methods.

### 6. Deploy RabbitMQ

```bash
oc apply -f rabbitmq.yaml
oc wait --for=condition=ready pod -l app=rabbitmq --timeout=120s
```

Verify deployment:
```bash
oc get pods -l app=rabbitmq
oc logs -l app=rabbitmq --tail=20
```

### 7. Deploy Application

```bash
oc apply -f app.yaml
oc wait --for=condition=ready pod -l app=amqp-service-app --timeout=60s
```

Verify deployment:
```bash
oc get pods -l app=amqp-service-app
oc logs -l app=amqp-service-app --tail=20
```

### 8. Access Application

Retrieve the application URL:
```bash
export APP_URL=$(oc get route amqp-service-app -o jsonpath='{.spec.host}')
echo "Application URL: http://$APP_URL"
```

## Testing the Application

### Health Check

```bash
curl http://$APP_URL/health
```

Expected response:
```json
{
  "status": "healthy",
  "rabbitmq": "connected",
  "instana": true
}
```

### Publish Messages

Publish with default message:
```bash
curl http://$APP_URL/publish
```

Publish with custom message:
```bash
curl "http://$APP_URL/publish?message=Hello+World"
```

Expected response:
```json
{
  "status": "success",
  "message": "Message published successfully",
  "content": "Hello World",
  "queue": "instana-test-queue",
  "exchange": "instana-exchange",
  "routing_key": "instana.test"
}
```

### Consume Messages

```bash
curl http://$APP_URL/consume
```

Expected response:
```json
{
  "status": "success",
  "message_count": 1,
  "messages": ["Hello World"]
}
```

### View Available Endpoints

```bash
curl http://$APP_URL/
```

## Cleanup

Remove application resources:
```bash
oc delete -f app.yaml
oc delete -f rabbitmq.yaml
```

Uninstall Instana agent:
```bash
helm uninstall instana-agent --namespace instana-agent
oc delete daemonset instana-agent -n instana-agent
oc delete deployment instana-agent-k8sensor -n instana-agent
oc delete project instana-agent
```

### Troubleshooting: Namespace Stuck in Terminating State

If the namespace remains in "Terminating" state:

```bash
# Check namespace status
oc get namespace instana-agent

# Remove finalizers from InstanaAgent custom resource
oc patch instanaagent instana-agent -n instana-agent -p '{"metadata":{"finalizers":[]}}' --type=merge

# Verify deletion
oc get namespace instana-agent
```

---

# Docker Deployment

## Prerequisites

- Docker Engine or Docker Desktop
- Docker Compose

## Installation Steps

### 1. Deploy Instana Agent

Start the Instana agent container:

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
   --env="INSTANA_AGENT_ENDPOINT=YOUR_ENDPOINT_HOST" \
   --env="INSTANA_AGENT_ENDPOINT_PORT=443" \
   --env="INSTANA_AGENT_KEY=YOUR_AGENT_KEY" \
   --env="INSTANA_DOWNLOAD_KEY=YOUR_DOWNLOAD_KEY" \
   icr.io/instana/agent
```

> **Note**: Replace placeholder values with your actual Instana credentials from the Instana UI.

### 2. Start Application Services

Navigate to the example directory:
```bash
cd example/openshift-amqp
```

> **Note**: The `docker-compose.yaml` uses environment variables for RabbitMQ credentials. For development, the default `guest/guest` credentials are used. For production, set strong passwords via environment variables.

Start RabbitMQ and the application:
```bash
docker compose up -d
```

View logs:
```bash
docker compose logs -f
```

Check service status:
```bash
docker compose ps
```

Services will be available at:
- **Application**: `http://localhost:8085`
- **RabbitMQ AMQP**: `localhost:5672`
- **RabbitMQ Management UI**: `http://localhost:15672`

### 3. Access RabbitMQ Management UI

Navigate to `http://localhost:15672` in your browser.

Login credentials:
- **Username**: `guest`
- **Password**: `guest`

## Testing the Application

### Health Check

```bash
curl http://localhost:8085/health
```

Expected response:
```json
{
  "status": "healthy",
  "rabbitmq": "connected",
  "instana": true
}
```

### Publish Messages

Publish with default message:
```bash
curl http://localhost:8085/publish
```

Publish with custom message:
```bash
curl "http://localhost:8085/publish?message=Hello+from+Docker"
```

Expected response:
```json
{
  "status": "success",
  "message": "Message published successfully",
  "content": "Hello from Docker",
  "queue": "instana-test-queue",
  "exchange": "instana-exchange",
  "routing_key": "instana.test"
}
```

### Consume Messages

```bash
curl http://localhost:8085/consume
```

Expected response:
```json
{
  "status": "success",
  "message_count": 1,
  "messages": ["Hello from Docker"]
}
```

### View Available Endpoints

```bash
curl http://localhost:8085/
```

## Cleanup

Stop and remove containers:
```bash
docker compose down
```

Stop and remove containers with volumes:
```bash
docker compose down -v
```

Remove Instana agent:
```bash
docker stop instana-agent
docker rm instana-agent
```

---

# Monitoring with Instana

## Verify Instrumentation

1. Log in to your Instana dashboard
2. Navigate to **Applications** → `amqp-service`
3. Verify the following data is being collected:
   - HTTP request traces
   - AMQP publish/consume operations
   - RabbitMQ connection metrics
   - Infrastructure metrics
   - OpenShift-specific metadata (pods, deployments, routes) for OpenShift deployments

## Expected Traces

You should observe:
- End-to-end traces from HTTP requests through AMQP message publishing
- Consumer traces showing message processing
- Correlation between producer and consumer spans
- RabbitMQ queue and exchange metrics

---

# Resources

- [Instana Go Sensor Documentation](https://www.ibm.com/docs/en/instana-observability?topic=technologies-monitoring-go)
- [Instana AMQP Instrumentation](https://github.com/instana/go-sensor/tree/main/instrumentation/instaamqp091)
- [Instana OpenShift Monitoring](https://www.ibm.com/docs/en/instana-observability?topic=agents-installing-red-hat-openshift)
- [OpenShift Documentation](https://docs.openshift.com/)
- [RabbitMQ Documentation](https://www.rabbitmq.com/documentation.html)
- [OpenShift Security Context Constraints](https://docs.openshift.com/container-platform/latest/authentication/managing-security-context-constraints.html)