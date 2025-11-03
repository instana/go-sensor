# Kubernetes Environment with Instana Monitoring

This example demonstrates running a Go application with MySQL database in Kubernetes, monitored by Instana.

## Overview

- **Application**: Gin web server that queries MySQL and returns the current date
- **Database**: MySQL 8.0
- **Monitoring**: Instana agent (DaemonSet + K8s sensor)

## Prerequisites

- Kubernetes cluster (minikube, kind, or cloud provider)
- kubectl CLI
- Docker or Podman
- Go 1.24+ (optional, for local development)

## Quick Start

### 1. Build the Application Image

Choose Docker or Podman based on your setup:

<details>
<summary><b>Using Docker</b></summary>

```bash
cd example/k8s-env

# Build the image
docker build -t mysql-service-app:latest .

# For minikube with Docker driver
eval $(minikube docker-env)
docker build -t mysql-service-app:latest .
```
</details>

<details>
<summary><b>Using Podman</b></summary>

```bash
cd example/k8s-env

# Build the image
podman build -t mysql-service-app:latest .

# Load into minikube
podman save -o mysql-service-app.tar mysql-service-app:latest
minikube image load mysql-service-app.tar
rm mysql-service-app.tar

# Verify image is loaded (should show localhost/mysql-service-app:latest)
minikube image ls | grep mysql-service-app
```

**Note**: Podman images get the `localhost/` prefix in minikube. The app.yaml is already configured for this.
</details>

### 2. Deploy to Kubernetes

```bash
# Deploy MySQL
kubectl apply -f mysql.yaml

# Deploy Instana Agent (update agent key in instana-agent.yaml first)
kubectl apply -f instana-agent.yaml

# Deploy Application
kubectl apply -f app.yaml

# Wait for pods to be ready
kubectl wait --for=condition=ready pod -l app=mysql --timeout=120s
kubectl wait --for=condition=ready pod -l app=mysql-service-app --timeout=60s
```

### 3. Access the Application

**Option 1: Port Forward (Recommended)**
```bash
kubectl port-forward service/mysql-service-app-external 8085:8085
# In another terminal:
curl http://localhost:8085/gin-test
```

**Option 2: From Inside Minikube**
```bash
minikube ssh
curl http://localhost:30085/gin-test
exit
```

**Expected Response:**
```json
{"message":" Current date is2025-10-31 - hello"}
```

> **Note**: Direct access via NodePort (`minikube ip`) or `minikube tunnel` may not work if minikube is running in an isolated network. Use port-forward or SSH into minikube instead.

## Configuration

### Instana Agent

Update the agent key in [`instana-agent.yaml`](instana-agent.yaml:99):
```yaml
data:
  key: <your-base64-encoded-agent-key>
```

### MySQL Credentials

Default credentials (defined in [`mysql.yaml`](mysql.yaml:8)):
- **User**: `go`
- **Password**: `gopw`
- **Database**: `godb`

To change, update the Secret and connection string in [`main.go`](main.go:34).

## Monitoring

The application is instrumented with Instana. View metrics in your Instana dashboard:
1. Log in to Instana
2. Navigate to Applications → `mysql-service`
3. View traces, metrics, and infrastructure data


### Podman-Specific Issues

**"podman-env only compatible with crio runtime"**: Your minikube uses Docker runtime. Use the tarball method instead of `minikube podman-env`.

## Cleanup

```bash
kubectl delete -f app.yaml
kubectl delete -f mysql.yaml
kubectl delete -f instana-agent.yaml
kubectl delete pvc mysql-pvc
```

## Resources

- [Instana Documentation](https://www.instana.com/docs/)
- [Go Sensor Documentation](https://www.instana.com/docs/ecosystem/go/)
- [Kubernetes Monitoring](https://www.instana.com/docs/ecosystem/kubernetes/)