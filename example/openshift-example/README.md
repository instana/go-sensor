# OpenShift AMQP Example with Instana Monitoring

This example demonstrates running a Go application with RabbitMQ (AMQP) in both local and OpenShift environments, monitored by Instana.

## Overview

- **Application**: Gin web server with AMQP producer and consumer
- **Message Broker**: RabbitMQ 3.13
- **Monitoring**: Instana agent with AMQP instrumentation
- **Platform**: Local (Docker/Docker Compose) and OpenShift 4.x

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [OpenShift Setup](#openshift-setup)
3. [Testing the Application](#testing-the-application)
4. [Architecture](#architecture)
5. [Resources](#resources)

---

## Prerequisites

- Access to an OpenShift cluster (URL, username, password/token)
- `oc` CLI installed on your local machine
- Docker or Podman for building images
- Access to a container registry (OpenShift internal registry, Quay.io, Docker Hub, etc.)

## OpenShift Setup

### Step 1: Login to OpenShift Cluster

```bash
# Get your token from OpenShift web console:
# 1. Login to OpenShift web console
# 2. Click "Display Token"
# 3. Copy the login command

# Login using token
oc login https://api.your-openshift-cluster.com:6443 --token=sha256~your-token-here
```

#### Verify Login

```bash
# Check current user
oc whoami

# Check cluster info
oc cluster-info

# View available projects
oc projects
```

### Step 2: Configure Instana Agent

Before deploying, configure the Instana agent:

1. Log in to your Instana account
2. Navigate to: **Settings** → **Agents** → **Installing Instana Agents**
3. Select **OpenShift (helm chart)** as the platform. You can choose operator also if you like.
4. Follow the steps there to setup the instana agent.

**Run instana agent locally using helm chart**

```bash
oc new-project instana-agent

oc adm policy add-scc-to-user privileged -z instana-agent -n instana-agent

helm install instana-agent \
   --repo https://agents.instana.io/helm \
   --namespace instana-agent \
   --set openshift=true \
   --set agent.key=<redacted> \
   --set agent.downloadKey=<redacted> \
   --set agent.endpointHost=<redacted> \
   --set agent.endpointPort=443 \
   --set cluster.name='gotracer' \
   instana-agent
```

> **Note**: Replace the `<redacted>` placeholders with your actual Instana credentials. For the latest agent installation instructions, visit the **Install Agents** page in your Instana UI.

#### Verify agent is running
oc get pods -n instana-agent
oc logs -n instana-agent -l app.kubernetes.io/name=instana-agent --max-log-requests=10 --follow

### Step 3: Create Application Project

```bash
# Create a new project for the application
oc new-project instana-amqp-demo

# Verify you're in the correct project
oc project
```

### Step 4: Build and Push Application Image

You can build in different ways like using OpenShift Internal Registry with Docker or Using External Registry with Docker (Quay.io, Docker Hub). Here I am using OpenShift Build (Source-to-Image) method.

```bash
# Navigate to the example directory
cd example/openshift-example

# Create a new build configuration
oc new-build --name=amqp-service-app --binary --strategy=docker

# Start the build from current directory
oc start-build amqp-service-app --from-dir=. --follow

# The image will be available at:
# image-registry.openshift-image-registry.svc:5000/instana-amqp-demo/amqp-service-app:latest

# Update app.yaml to use this image
```

> **Note**: OpenShift builds run inside the cluster, so you don't need Podman or Docker for this option.

### Step 5: Deploy RabbitMQ

```bash
# Deploy RabbitMQ with all required resources
oc apply -f rabbitmq.yaml

# Wait for RabbitMQ to be ready (this may take 1-2 minutes)
oc wait --for=condition=ready pod -l app=rabbitmq --timeout=120s

# Verify RabbitMQ is running
oc get pods -l app=rabbitmq
oc logs -l app=rabbitmq --tail=20
```

### Step 7: Deploy the Application

```bash
# Deploy the application
oc apply -f app.yaml

# Wait for application to be ready
oc wait --for=condition=ready pod -l app=amqp-service-app --timeout=60s

# Verify application is running
oc get pods -l app=amqp-service-app
oc logs -l app=amqp-service-app --tail=20
```

### Step 8: Access the Application

#### Get the Route URL

```bash
# Get the application route
oc get route amqp-service-app

# Or get just the URL
export APP_URL=$(oc get route amqp-service-app -o jsonpath='{.spec.host}')
echo "Application URL: http://$APP_URL"
```

#### Get RabbitMQ Management UI URL (Optional)

```bash
# Get the RabbitMQ management route
oc get route rabbitmq-management

# Or get just the URL
export RABBITMQ_URL=$(oc get route rabbitmq-management -o jsonpath='{.spec.host}')
echo "RabbitMQ Management UI: http://$RABBITMQ_URL"
```

---

## Testing the Application

### Basic Health Check

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

```bash
# Publish with default message
curl http://$APP_URL/publish

# Publish with custom message
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
# Consume messages from queue
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

## Verify Monitoring in Instana

1. Log in to your Instana dashboard
2. Navigate to **Applications** → `amqp-service`
3. You should see:
   - Application traces
   - AMQP/RabbitMQ operations (publish/consume)
   - HTTP requests
   - Infrastructure metrics
   - OpenShift-specific data (pods, deployments, routes)

---

## Cleanup (OpenShift)

### Stop All Applications

```bash
# Delete the application deployment and related resources
oc delete -f app.yaml

# Delete RabbitMQ deployment and related resources
oc delete -f rabbitmq.yaml

# Verify all pods are terminated
oc get pods

# Optional: Delete the project (this will remove everything)
# oc delete project instana-amqp-demo

```

### Stop Instana Agent (Helm Installation)

If you installed the Instana agent using Helm, use these commands:

```bash
# List Helm releases to confirm the installation
helm list --namespace instana-agent

# Uninstall the Instana agent
helm uninstall instana-agent --namespace instana-agent

# The helm uninstall may not remove all resources immediately
# Manually delete remaining resources:

# Delete DaemonSet (agent pods on each node)
oc delete daemonset instana-agent -n instana-agent

# Delete Deployment (k8sensor pods)
oc delete deployment instana-agent-k8sensor -n instana-agent

# Verify all pods are terminating/removed
oc get pods -n instana-agent

# If pods still exist, force delete them
oc delete pods --all -n instana-agent --force --grace-period=0

# Optional: Delete the project (this will remove everything)
# oc delete project instana-agent

# Verify namespace is clean
oc get all -n instana-agent
```

---

## Architecture

```
┌────────────────────────────────────────┐
│         OpenShift Cluster              │
│                                        │
│  ┌──────────────────────────────────┐ │
│  │      Instana Agent (DaemonSet)   │ │
│  └────────────┬─────────────────────┘ │
│               │                        │
│  ┌────────────▼─────────────────────┐ │
│  │         Project/Namespace        │ │
│  │                                  │ │
│  │  ┌──────────┐    ┌───────────┐  │ │
│  │  │ RabbitMQ │◄───┤   App     │  │ │
│  │  │  Pod     │    │   Pod     │  │ │
│  │  └──────────┘    └─────┬─────┘  │ │
│  │                        │        │ │
│  │  ┌──────────┐    ┌─────▼─────┐  │ │
│  │  │ Service  │    │  Service  │  │ │
│  │  └──────────┘    └─────┬─────┘  │ │
│  │                        │        │ │
│  │                  ┌─────▼─────┐  │ │
│  │                  │   Route   │  │ │
│  │                  └───────────┘  │ │
│  └──────────────────────────────────┘ │
└────────────────────────────────────────┘
```

## Resources

- [OpenShift Documentation](https://docs.openshift.com/)
- [Instana OpenShift Monitoring](https://www.ibm.com/docs/en/instana-observability/current?topic=agents-installing-openshift)
- [Go Sensor Documentation](https://www.instana.com/docs/ecosystem/go/)
- [RabbitMQ Documentation](https://www.rabbitmq.com/documentation.html)
- [Instana AMQP Instrumentation](https://github.com/instana/go-sensor/tree/main/instrumentation/instaamqp091)
- [OpenShift Security Context Constraints](https://docs.openshift.com/container-platform/latest/authentication/managing-security-context-constraints.html)