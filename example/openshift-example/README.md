# OpenShift AMQP Example with Instana Monitoring

This example demonstrates running a Go application with RabbitMQ (AMQP) in both local and OpenShift environments, monitored by Instana.

## Overview

- **Application**: Gin web server with AMQP producer and consumer
- **Message Broker**: RabbitMQ 3.13 (Bitnami image from Docker Hub)
- **Monitoring**: Instana agent with AMQP instrumentation
- **Platform**: Local (Docker/Docker Compose) and OpenShift 4.x

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Local Setup (macOS)](#local-setup-macos)
3. [OpenShift Setup](#openshift-setup)
4. [Testing the Application](#testing-the-application)
5. [Troubleshooting](#troubleshooting)

---

## Prerequisites

### For Local Setup
- **Docker Engine** or Docker Desktop
- **Docker Compose** (`docker compose` plugin or `docker-compose`)
- Go 1.24+ (optional, for local development)
- Instana agent running locally (optional, for full tracing)

### For OpenShift Setup
- Access to an OpenShift cluster (URL, username, password/token)
- `oc` CLI installed on your local machine
- Docker or Podman for building images
- Access to a container registry (OpenShift internal registry, Quay.io, Docker Hub, etc.)

### Installing Docker and Tools

#### Ubuntu VM

```bash
# Install Docker Engine and Docker Compose plugin
sudo apt-get update
sudo apt-get install -y docker.io docker-compose-v2

# Enable and start Docker
sudo systemctl enable --now docker

# Optional: run Docker without sudo after re-login
sudo usermod -aG docker $USER

# Install OpenShift CLI if needed later
# See: https://docs.openshift.com/container-platform/latest/cli_reference/openshift_cli/getting-started-cli.html

# Verify installations
docker --version
docker compose version
```

#### macOS

```bash
# Install Docker Desktop, then verify
docker --version
docker compose version

# Install OpenShift CLI if needed
brew install openshift-cli
oc version
```

---

## Local Setup (Docker)

### Step 1: Start Instana Agent (Optional)

If you want full tracing locally, ensure the Instana agent is running on your Mac. If not installed, you can run the application without it (traces won't be sent).

### Step 2: Navigate to Example Directory

```bash
cd example/openshift-example
```

### Step 3: Start Services

#### Option A: Using Docker Compose (Recommended)

```bash
# Start RabbitMQ and the application
docker compose up -d

# View logs
docker compose logs -f

# Check service status
docker compose ps

# Stop services
docker compose down
```

#### Option B: Using Docker Directly (Without Compose)

```bash
# Create a network
docker network create amqp-network

# Start RabbitMQ
docker run -d \
  --name rabbitmq \
  --network amqp-network \
  -p 5672:5672 \
  -p 15672:15672 \
  -e RABBITMQ_USERNAME=guest \
  -e RABBITMQ_PASSWORD=guest \
  bitnami/rabbitmq:3.13

# Wait for RabbitMQ to be ready (about 30 seconds)
sleep 30

# Build the application image
docker build -t amqp-service-app:latest -f example/openshift-example/Dockerfile ../..

# Run the application
docker run -d \
  --name amqp-app \
  --network amqp-network \
  -p 8085:8085 \
  -e RABBITMQ_URL=amqp://guest:guest@rabbitmq:5672/ \
  -e INSTANA_AGENT_HOST=host.docker.internal \
  -e INSTANA_AGENT_PORT=42699 \
  -e INSTANA_SERVICE_NAME=amqp-service \
  -e INSTANA_DEBUG=true \
  amqp-service-app:latest

# View logs
docker logs -f amqp-app
```

#### Option C: Using docker-compose (Legacy CLI)

```bash
docker-compose up -d
docker-compose logs -f
docker-compose ps
```

This will start:
- **RabbitMQ** on `localhost:5672` (AMQP) and `localhost:15672` (Management UI)
- **AMQP Application** on `localhost:8085`

### Step 4: Access RabbitMQ Management UI

Open your browser and navigate to:
```
http://localhost:15672
```

Login credentials:
- Username: `guest`
- Password: `guest`

### Step 5: Test the Application Locally

```bash
# Check health
curl http://localhost:8085/health

# Publish a message
curl "http://localhost:8085/publish?message=Hello+from+local"

# Consume messages
curl http://localhost:8085/consume

# View available endpoints
curl http://localhost:8085/
```

### Step 6: Stop Services

#### Using Docker Compose

```bash
# Stop and remove containers
docker compose down

# Stop and remove containers with volumes
docker compose down -v
```

#### Using Docker Directly

```bash
# Stop and remove containers
docker stop amqp-app rabbitmq
docker rm amqp-app rabbitmq

# Remove network
docker network rm amqp-network

# Remove images (optional)
docker rmi amqp-service-app:latest
```

#### Using docker-compose (Legacy CLI)

```bash
docker-compose down
docker-compose down -v
```

---

## OpenShift Setup

### Step 1: Login to OpenShift Cluster

#### Option A: Login with Username and Password

```bash
# Login to your OpenShift cluster
oc login https://api.your-openshift-cluster.com:6443

# Enter your username and password when prompted
```

#### Option B: Login with Token

```bash
# Get your token from OpenShift web console:
# 1. Login to OpenShift web console
# 2. Click your username in top-right corner
# 3. Click "Copy login command"
# 4. Click "Display Token"
# 5. Copy the login command

# Login using token
oc login --token=sha256~your-token-here --server=https://api.your-openshift-cluster.com:6443
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

### Step 2: Create a New Project

```bash
# Create a new project for this application
oc new-project instana-amqp-demo

# Verify you're in the correct project
oc project
```

### Step 3: Configure Instana Agent

Before deploying, configure the Instana agent:

1. Log in to your Instana account
2. Navigate to: **Settings** → **Agents** → **Installing Instana Agents**
3. Select **OpenShift** or **Kubernetes** as the platform
4. Copy the generated YAML configuration
5. Save it as `instana-agent.yaml` in this directory

> **Important**: The `instana-agent.yaml` file is not included in this repository. You must create it with your actual Instana configuration.

### Step 4: Build and Push Application Image

#### Option A: Using OpenShift Internal Registry with Docker (Recommended)

```bash
# Navigate to the example directory
cd example/openshift-example

# Expose the internal registry (if not already exposed)
oc patch configs.imageregistry.operator.openshift.io/cluster --patch '{"spec":{"defaultRoute":true}}' --type=merge

# Get the registry route
export REGISTRY=$(oc get route default-route -n openshift-image-registry -o jsonpath='{.spec.host}')
echo "Registry: $REGISTRY"

# Login to the registry using Docker
docker login -u $(oc whoami) -p $(oc whoami -t) $REGISTRY

# Build and push using Docker
docker build -t $REGISTRY/instana-amqp-demo/amqp-service-app:latest -f example/openshift-example/Dockerfile ../..
docker push $REGISTRY/instana-amqp-demo/amqp-service-app:latest

# Update app.yaml to use the internal registry image
# image: image-registry.openshift.svc:5000/instana-amqp-demo/amqp-service-app:latest
```

#### Option B: Using External Registry with Docker (Quay.io, Docker Hub)

```bash
# Navigate to the example directory
cd example/openshift-example

# Build and push to external registry
# Replace <your-registry> with quay.io/<username> or docker.io/<username>
export IMAGE_REPO=quay.io/<your-username>/amqp-service-app

# Build the image with Docker
docker build -t $IMAGE_REPO:latest -f example/openshift-example/Dockerfile ../..

# Login to your registry
docker login quay.io  # or docker.io

# Push the image
docker push $IMAGE_REPO:latest

# Update app.yaml with your image
# image: quay.io/<your-username>/amqp-service-app:latest
```

#### Option C: Using OpenShift Build (Source-to-Image)

```bash
# Navigate to the example directory
cd example/openshift-example

# Create a new build configuration
oc new-build --name=amqp-service-app --binary --strategy=docker

# Start the build from current directory
oc start-build amqp-service-app --from-dir=. --follow

# The image will be available at:
# image-registry.openshift.svc:5000/instana-amqp-demo/amqp-service-app:latest

# Update app.yaml to use this image
```

```bash
# Navigate to the example directory
cd example/openshift-example

# Create a new build configuration
oc new-build --name=amqp-service-app --binary --strategy=docker

# Start the build from current directory
oc start-build amqp-service-app --from-dir=. --follow

# The image will be available at:
# image-registry.openshift.svc:5000/instana-amqp-demo/amqp-service-app:latest

# Update app.yaml to use this image
```

> **Note**: OpenShift builds run inside the cluster, so you don't need Podman or Docker for this option.

### Step 5: Update Deployment Configuration

Edit `app.yaml` to use your image:

```bash
# Open app.yaml and update the image field
# Change this line:
#   image: image-registry.openshift.svc:5000/instana-agent/amqp-service-app:latest
# To your actual image location, for example:
#   image: image-registry.openshift.svc:5000/instana-amqp-demo/amqp-service-app:latest
# Or:
#   image: quay.io/<your-username>/amqp-service-app:latest
```

### Step 6: Deploy RabbitMQ

```bash
# Deploy RabbitMQ with all required resources
oc apply -f rabbitmq.yaml

# Wait for RabbitMQ to be ready (this may take 1-2 minutes)
oc wait --for=condition=ready pod -l app=rabbitmq --timeout=120s

# Verify RabbitMQ is running
oc get pods -l app=rabbitmq
oc logs -l app=rabbitmq --tail=20
```

### Step 7: Deploy Instana Agent

```bash
# Deploy the Instana agent (ensure you've configured it in Step 3)
oc apply -f instana-agent.yaml

# Verify agent is running
oc get pods -n instana-agent

# Check agent logs
oc logs -n instana-agent -l app.kubernetes.io/name=instana-agent --tail=50
```

### Step 8: Deploy the Application

```bash
# Deploy the application
oc apply -f app.yaml

# Wait for application to be ready
oc wait --for=condition=ready pod -l app=amqp-service-app --timeout=60s

# Verify application is running
oc get pods -l app=amqp-service-app
oc logs -l app=amqp-service-app --tail=20
```

### Step 9: Access the Application

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
# Local
curl http://localhost:8085/health

# OpenShift
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
curl http://localhost:8085/publish
# or
curl http://$APP_URL/publish

# Publish with custom message
curl "http://localhost:8085/publish?message=Hello+World"
# or
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
curl http://localhost:8085/consume
# or
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
curl http://localhost:8085/
# or
curl http://$APP_URL/
```

### Test Complete Flow

```bash
# 1. Publish multiple messages
for i in {1..5}; do
  curl "http://localhost:8085/publish?message=Message+$i"
  echo ""
done

# 2. Consume messages
curl http://localhost:8085/consume

# 3. Check health
curl http://localhost:8085/health
```

---

## Troubleshooting

### Local Setup Issues

#### RabbitMQ Not Starting

**Using Docker Compose:**
```bash
# Check RabbitMQ logs
docker compose logs rabbitmq

# Restart RabbitMQ
docker compose restart rabbitmq

# Check if port is already in use
lsof -i :5672
lsof -i :15672
```

**Using Docker Directly:**
```bash
# Check RabbitMQ logs
docker logs rabbitmq

# Restart RabbitMQ
docker restart rabbitmq

# Check if port is already in use
lsof -i :5672
lsof -i :15672
```

#### Application Can't Connect to RabbitMQ

**Using Docker Compose:**
```bash
# Check if RabbitMQ is healthy
docker compose ps

# Check application logs
docker compose logs amqp-app

# Restart the application
docker compose restart amqp-app
```

**Using Docker Directly:**
```bash
# Check if RabbitMQ is healthy
docker ps

# Check application logs
docker logs amqp-app

# Restart the application
docker restart amqp-app
```

#### Build Failures

**Using Docker Compose:**
```bash
# Clean and rebuild
docker compose down
docker compose build --no-cache
docker compose up -d
```

**Using Docker Directly:**
```bash
# Remove containers
docker stop amqp-app rabbitmq
docker rm amqp-app rabbitmq

# Rebuild image
docker build --no-cache -t amqp-service-app:latest -f example/openshift-example/Dockerfile ../..

# Restart services (use commands from Step 3, Option B)
```

#### Podman Machine Issues

```bash
# Check Podman machine status
podman machine list

# Restart Podman machine
podman machine stop
podman machine start

# If issues persist, recreate the machine
podman machine rm podman-machine-default
podman machine init
podman machine start
```

### OpenShift Setup Issues

#### Check Pod Status

```bash
# View all pods in current project
oc get pods

# Describe a specific pod
oc describe pod <pod-name>

# View pod logs
oc logs <pod-name>

# Follow logs in real-time
oc logs -f <pod-name>

# View previous container logs (if pod restarted)
oc logs <pod-name> --previous
```

#### Check Application Logs

```bash
# Application logs
oc logs -l app=amqp-service-app --tail=50

# RabbitMQ logs
oc logs -l app=rabbitmq --tail=50

# Instana agent logs
oc logs -n instana-agent -l app.kubernetes.io/name=instana-agent --tail=50
```

#### Check Events

```bash
# View recent events in the project
oc get events --sort-by='.lastTimestamp'

# Watch events in real-time
oc get events --watch
```

#### Debug Pod Issues

```bash
# Get detailed pod information
oc describe pod -l app=amqp-service-app

# Check security context
oc get pod <pod-name> -o yaml | grep -A 10 securityContext

# Start a debug session
oc debug deployment/amqp-service-app

# Execute commands in running pod
oc exec -it deployment/amqp-service-app -- sh
```

#### Test RabbitMQ Connectivity

```bash
# From application pod
oc exec -it deployment/amqp-service-app -- sh
# Inside the pod:
nc -zv rabbitmq 5672
exit

# From RabbitMQ pod
oc exec -it deployment/rabbitmq -- rabbitmq-diagnostics ping
```

#### Image Pull Issues

```bash
# Check image stream
oc get imagestream

# Check build status (if using OpenShift builds)
oc get builds
oc logs -f bc/amqp-service-app

# Describe deployment to see image pull errors
oc describe deployment amqp-service-app
```

#### Route Issues

```bash
# Check route configuration
oc get route amqp-service-app -o yaml

# Check service endpoints
oc get endpoints amqp-service-app

# Test service internally
oc run test-pod --image=curlimages/curl --rm -it --restart=Never -- curl http://amqp-service-app:8085/health
```

### Viewing Resources

```bash
# View all resources in the project
oc get all

# View specific resource types
oc get deployments
oc get services
oc get routes
oc get pods

# View with more details
oc get pods -o wide
```

### Scaling the Application

```bash
# Scale application to 3 replicas
oc scale deployment/amqp-service-app --replicas=3

# Verify scaling
oc get pods -l app=amqp-service-app

# Scale back to 1 replica
oc scale deployment/amqp-service-app --replicas=1
```

### Updating the Application

```bash
# After making code changes, rebuild and push the image
docker build -t $IMAGE_REPO:latest -f example/openshift-example/Dockerfile ../..
docker push $IMAGE_REPO:latest

# Trigger a rollout
oc rollout restart deployment/amqp-service-app

# Watch the rollout status
oc rollout status deployment/amqp-service-app

# View rollout history
oc rollout history deployment/amqp-service-app
```

---

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

## Cleanup

### Local Cleanup

**Using Docker Compose:**
```bash
# Stop and remove containers
docker compose down

# Stop and remove containers with volumes
docker compose down -v
```

**Using Docker Directly:**
```bash
# Stop and remove containers
docker stop amqp-app rabbitmq
docker rm amqp-app rabbitmq

# Remove network
docker network rm amqp-network

# Remove volumes (if any)
docker volume prune -f
```

**Using docker-compose (Legacy CLI):**
```bash
docker-compose down
docker-compose down -v
```

### OpenShift Cleanup

```bash
# Delete all application resources
oc delete -f app.yaml
oc delete -f rabbitmq.yaml
oc delete -f instana-agent.yaml

# Delete the entire project (optional)
oc delete project instana-amqp-demo
```

---

## Architecture

### Local Architecture

```
┌─────────────────┐
│  Your Mac       │
│                 │
│  ┌───────────┐  │
│  │  Instana  │  │
│  │  Agent    │  │
│  │ (Optional)│  │
│  └─────┬─────┘  │
│        │        │
│  ┌─────▼─────┐  │
│  │   Docker  │  │
│  │  Compose  │  │
│  │           │  │
│  │ ┌───────┐ │  │
│  │ │RabbitMQ│ │  │
│  │ └───┬───┘ │  │
│  │     │     │  │
│  │ ┌───▼───┐ │  │
│  │ │  App  │ │  │
│  │ └───────┘ │  │
│  └───────────┘  │
└─────────────────┘
```

### OpenShift Architecture

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

---

## Security Notes

This example includes OpenShift-specific security configurations:

- **Security Context Constraints (SCC)**: Pods run with restricted SCC
- **Non-root containers**: Both application and RabbitMQ run as non-root users
- **Capability dropping**: Unnecessary Linux capabilities are dropped
- **Seccomp profiles**: Runtime default seccomp profiles applied
- **Read-only root filesystem**: Where applicable

### Default RabbitMQ Credentials

For production use, change the default credentials:

- **User**: `guest`
- **Password**: `guest`

Update the Secret in `rabbitmq.yaml` and connection string in `main.go` and `docker-compose.yaml`.

---

## Resources

- [OpenShift Documentation](https://docs.openshift.com/)
- [Instana OpenShift Monitoring](https://www.ibm.com/docs/en/instana-observability/current?topic=agents-installing-openshift)
- [Go Sensor Documentation](https://www.instana.com/docs/ecosystem/go/)
- [RabbitMQ Documentation](https://www.rabbitmq.com/documentation.html)
- [Instana AMQP Instrumentation](https://github.com/instana/go-sensor/tree/main/instrumentation/instaamqp091)
- [OpenShift Security Context Constraints](https://docs.openshift.com/container-platform/latest/authentication/managing-security-context-constraints.html)