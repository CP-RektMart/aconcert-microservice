# gRPC Health Check Implementation Guide

## Overview

This guide documents the gRPC health check implementation for AConcert microservices. We've migrated from HTTP health probes to native gRPC health probes to properly monitor gRPC services.

## What Changed

### Services Updated

The following gRPC services now implement the gRPC Health Checking Protocol:

1. **Event Service** (`services/event`)
2. **Reservation Service** (`services/reservation`)
3. **Location Service** (`services/location`)

### HTTP Services (No Change Required)

These services use Fiber/HTTP and already have proper HTTP health checks:

- Auth Service
- Gateway Service
- Payment Service
- Realtime Service

---

## Implementation Details

### 1. Code Changes

Each gRPC service now includes:

```go
import (
    "google.golang.org/grpc/health"
    healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

// In main():
grpcServer := grpc.NewServer(
    grpc.UnaryInterceptor(grpclogger.LoggingUnaryInterceptor),
)

// Register health check service
healthServer := health.NewServer()
healthpb.RegisterHealthServer(grpcServer, healthServer)

// Set service as serving
healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
healthServer.SetServingStatus("service.ServiceName", healthpb.HealthCheckResponse_SERVING)

// ... register your actual service ...

// On shutdown:
healthServer.SetServingStatus("", healthpb.HealthCheckResponse_NOT_SERVING)
grpcServer.GracefulStop()
```

### 2. Kubernetes Manifest Changes

Updated probe configuration from HTTP to gRPC:

**Before:**
```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8000
  initialDelaySeconds: 30
  periodSeconds: 10
```

**After:**
```yaml
livenessProbe:
  grpc:
    port: 8000
  initialDelaySeconds: 30
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3
```

### 3. Port Naming

Added proper port naming in Kubernetes:

```yaml
ports:
  - containerPort: 8000
    name: grpc  # Named for clarity
```

---

## Deployment Instructions

### Prerequisites

- Kubernetes 1.24+ (gRPC health probes became GA in 1.24)
- Go services already using `google.golang.org/grpc`

### Step 1: Rebuild Docker Images

After code changes, rebuild the affected service images:

```bash
cd aconcert-microservice/.k8s

# Build individual services
docker build -t aconcert/event-service:latest -f ../services/event/Dockerfile ..
docker build -t aconcert/reservation-service:latest -f ../services/reservation/Dockerfile ..
docker build -t aconcert/location-service:latest -f ../services/location/Dockerfile ..

# Or use the build script
./build-images.sh
```

### Step 2: Deploy Updated Manifests

```bash
# Apply updated Kubernetes manifests
kubectl apply -f .k8s/services/event-service.yaml
kubectl apply -f .k8s/services/reservation-service.yaml
kubectl apply -f .k8s/services/location-service.yaml
```

### Step 3: Verify Deployment

```bash
# Check pod status
kubectl get pods -n aconcert -w

# Check for healthy probes (should not see probe failures)
kubectl get events -n aconcert --field-selector reason=Unhealthy

# Describe a pod to see probe status
kubectl describe pod <POD_NAME> -n aconcert | grep -A 5 "Liveness\|Readiness"
```

---

## Testing Health Checks

### Using grpcurl

Install grpcurl if not already available:
```bash
brew install grpcurl  # macOS
# or
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

Test health check locally:
```bash
# Port forward to a pod
kubectl port-forward <POD_NAME> 8000:8000 -n aconcert

# Check health status
grpcurl -plaintext localhost:8000 grpc.health.v1.Health/Check

# Check specific service
grpcurl -plaintext -d '{"service":"event.EventService"}' \
  localhost:8000 grpc.health.v1.Health/Check
```

Expected response:
```json
{
  "status": "SERVING"
}
```

### Using kubectl exec

```bash
# If grpcurl is available in the container
kubectl exec -it <POD_NAME> -n aconcert -- \
  grpcurl -plaintext localhost:8000 grpc.health.v1.Health/Check
```

---

## Troubleshooting

### Issue: Pods Still Restarting

**Symptoms:**
```
waiting for rabbitmq
waiting for postgres
gRPC server starts, then shuts down
```

**Solutions:**

1. **Check Kubernetes Version:**
   ```bash
   kubectl version --short
   ```
   Ensure server version is 1.24+

2. **Verify Image Was Rebuilt:**
   ```bash
   kubectl describe pod <POD_NAME> -n aconcert | grep Image:
   docker images | grep aconcert
   ```

3. **Check if health server is registered:**
   ```bash
   kubectl logs <POD_NAME> -n aconcert
   ```
   Should see: `starting gRPC server port=8000`

4. **Verify probe configuration:**
   ```bash
   kubectl get pod <POD_NAME> -n aconcert -o yaml | grep -A 10 livenessProbe
   ```

### Issue: "Connection Refused" on Health Check

**Cause:** Service hasn't started yet or is listening on wrong port

**Solution:**
- Check `initialDelaySeconds` (should be 30+ seconds)
- Verify PORT environment variable matches probe port
- Check service logs for startup errors

### Issue: Old HTTP Probes Still Running

**Cause:** Manifest not applied or cached

**Solution:**
```bash
# Delete and recreate the deployment
kubectl delete deployment <SERVICE_NAME> -n aconcert
kubectl apply -f .k8s/services/<service>-service.yaml

# Or force rolling update
kubectl rollout restart deployment/<SERVICE_NAME> -n aconcert
```

---

## Health Check Protocol Details

### Standard gRPC Health Checking Protocol

We implement the standard gRPC health checking protocol defined in:
- [gRPC Health Checking Protocol](https://github.com/grpc/grpc/blob/master/doc/health-checking.md)
- Proto file: `grpc.health.v1.Health`

### Service Status Values

- `SERVING` - Service is healthy and accepting requests
- `NOT_SERVING` - Service is unhealthy or shutting down
- `UNKNOWN` - Health status is unknown
- `SERVICE_UNKNOWN` - Specific service name not found

### Service Names

We register two health check targets per service:

1. **Empty string (`""`)** - Overall server health
2. **Specific service name** - Individual service health (e.g., `event.EventService`)

Kubernetes gRPC probes check the empty string target by default.

---

## Probe Configuration Best Practices

### Liveness Probe
- Checks if the service is alive
- If fails → Kubernetes restarts the container
- Should be lenient to avoid unnecessary restarts

```yaml
livenessProbe:
  grpc:
    port: 8000
  initialDelaySeconds: 30  # Allow time for startup
  periodSeconds: 10        # Check every 10s
  timeoutSeconds: 5        # 5s timeout per check
  failureThreshold: 3      # Restart after 3 failures
```

### Readiness Probe
- Checks if service is ready to accept traffic
- If fails → Removes from service endpoints
- Can be more aggressive than liveness

```yaml
readinessProbe:
  grpc:
    port: 8000
  initialDelaySeconds: 10  # Shorter initial delay
  periodSeconds: 5         # Check more frequently
  timeoutSeconds: 3        # Shorter timeout
  failureThreshold: 2      # Mark unready faster
```

---

## Monitoring and Observability

### Check Probe Success Rate

```bash
# View probe metrics (if metrics-server is installed)
kubectl top pods -n aconcert

# Check pod events for probe failures
kubectl get events -n aconcert --sort-by='.lastTimestamp' | grep -i probe

# Watch pod status in real-time
kubectl get pods -n aconcert -w
```

### Logs to Monitor

Look for these log messages:

**Healthy startup:**
```
starting gRPC server port=8000
```

**Graceful shutdown:**
```
shutting down gRPC server gracefully
gRPC server stopped cleanly
```

**Problems:**
```
failed to serve error=...
failed to connect to postgres error=...
waiting for rabbitmq (repeated many times)
```

---

## Migration Checklist

- [x] Update Event Service code with gRPC health server
- [x] Update Reservation Service code with gRPC health server
- [x] Update Location Service code with gRPC health server
- [x] Update Event Service Kubernetes manifest
- [x] Update Reservation Service Kubernetes manifest
- [x] Update Location Service Kubernetes manifest
- [ ] Rebuild Docker images
- [ ] Push images to registry (if using one)
- [ ] Apply updated Kubernetes manifests
- [ ] Verify pods are healthy
- [ ] Test health checks with grpcurl
- [ ] Monitor for 24 hours to ensure stability

---

## Rollback Plan

If issues occur, rollback by reverting to HTTP health checks:

1. **Revert code changes:**
   ```bash
   git checkout HEAD~1 services/event/cmd/main.go
   git checkout HEAD~1 services/reservation/cmd/main.go
   git checkout HEAD~1 services/location/cmd/main.go
   ```

2. **Revert Kubernetes manifests:**
   ```bash
   git checkout HEAD~1 .k8s/services/event-service.yaml
   git checkout HEAD~1 .k8s/services/reservation-service.yaml
   git checkout HEAD~1 .k8s/services/location-service.yaml
   ```

3. **Rebuild and redeploy:**
   ```bash
   ./build-images.sh
   kubectl apply -f .k8s/services/
   ```

---

## Additional Resources

- [Kubernetes gRPC Probes Documentation](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#define-a-grpc-liveness-probe)
- [gRPC Health Checking Protocol](https://github.com/grpc/grpc/blob/master/doc/health-checking.md)
- [Go gRPC Health Package](https://pkg.go.dev/google.golang.org/grpc/health)

---

## Support

For issues or questions:
1. Check pod logs: `kubectl logs <POD_NAME> -n aconcert`
2. Check pod events: `kubectl describe pod <POD_NAME> -n aconcert`
3. Verify Kubernetes version supports gRPC probes (1.24+)
4. Test health endpoint manually with grpcurl
