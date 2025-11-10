# gRPC Health Check - Quick Start Guide

## What Was Fixed

Your pods were shutting down because:
- ❌ gRPC services had **HTTP health probes** configured
- ❌ HTTP probes can't check gRPC servers (different protocols)
- ✅ Now using **native gRPC health probes**

## What Changed

### Code Changes
Added gRPC health server to 3 services:
- `services/event/cmd/main.go`
- `services/reservation/cmd/main.go`
- `services/location/cmd/main.go`

### Kubernetes Changes
Updated health probes from HTTP to gRPC:
- `.k8s/services/event-service.yaml`
- `.k8s/services/reservation-service.yaml`
- `.k8s/services/location-service.yaml`

---

## Quick Deploy (3 Steps)

### Step 1: Rebuild Images

```bash
cd aconcert-microservice

# Build all three services
docker build -t aconcert/event-service:latest -f services/event/Dockerfile .
docker build -t aconcert/reservation-service:latest -f services/reservation/Dockerfile .
docker build -t aconcert/location-service:latest -f services/location/Dockerfile .
```

### Step 2: Redeploy to Kubernetes

```bash
cd .k8s

# Option A: Use the automated script
./redeploy-grpc-services.sh

# Option B: Manual deployment
kubectl delete deployment event-service reservation-service location-service -n aconcert
kubectl apply -f services/event-service.yaml
kubectl apply -f services/reservation-service.yaml
kubectl apply -f services/location-service.yaml
```

### Step 3: Verify

```bash
# Watch pods come up (should stay running now!)
kubectl get pods -n aconcert -w

# Check for NO probe failures
kubectl get events -n aconcert --field-selector reason=Unhealthy

# View logs to confirm gRPC server started
kubectl logs -l app=event-service -n aconcert --tail=20
```

---

## Expected Results

### ✅ Success Indicators

**Pods stay running:**
```bash
$ kubectl get pods -n aconcert
NAME                               READY   STATUS    RESTARTS   AGE
event-service-xxx                  1/1     Running   0          2m
reservation-service-xxx            1/1     Running   0          2m
location-service-xxx               1/1     Running   0          2m
```

**Logs show clean startup:**
```
waiting for rabbitmq (or postgres)
time=2025-XX-XXTXX:XX:XX.XXXZ level=INFO msg="starting gRPC server" port=8000
(no shutdown messages)
```

**No unhealthy events:**
```bash
$ kubectl get events -n aconcert --field-selector reason=Unhealthy
No resources found in aconcert namespace.
```

### ❌ If Still Failing

**Check Kubernetes version:**
```bash
kubectl version --short
# Server version must be v1.24.0 or higher
```

**Check if images were rebuilt:**
```bash
docker images | grep aconcert
# Should show recent timestamps
```

**View detailed pod status:**
```bash
kubectl describe pod <POD_NAME> -n aconcert
# Look for "Liveness probe failed" or "Readiness probe failed"
```

---

## Test Health Checks

### Using grpcurl

```bash
# Install grpcurl
brew install grpcurl  # macOS
# or
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# Port forward to test
kubectl port-forward svc/event-service 8000:8000 -n aconcert

# Test health check
grpcurl -plaintext localhost:8000 grpc.health.v1.Health/Check
```

**Expected response:**
```json
{
  "status": "SERVING"
}
```

---

## Troubleshooting

### Problem: Pods keep restarting

**Solution 1:** Ensure Kubernetes version supports gRPC probes
```bash
kubectl version --short
# Need v1.24.0+
```

**Solution 2:** Verify images were rebuilt
```bash
# Rebuild with --no-cache
docker build --no-cache -t aconcert/event-service:latest -f services/event/Dockerfile .

# Force pod recreation
kubectl rollout restart deployment/event-service -n aconcert
```

### Problem: "Connection refused" errors

**Check if port is correct:**
```bash
kubectl get pod <POD_NAME> -n aconcert -o yaml | grep -A 5 "livenessProbe"
# Port should be 8000
```

**Check if service is listening:**
```bash
kubectl exec -it <POD_NAME> -n aconcert -- netstat -tlnp | grep 8000
```

### Problem: Old HTTP probes still running

**Force update:**
```bash
kubectl delete deployment event-service -n aconcert
kubectl apply -f .k8s/services/event-service.yaml
```

---

## How It Works

### Before (HTTP Probes - BROKEN)
```yaml
livenessProbe:
  httpGet:              # ❌ HTTP protocol
    path: /health
    port: 8000
```
→ Fails because gRPC server doesn't understand HTTP GET requests

### After (gRPC Probes - WORKING)
```yaml
livenessProbe:
  grpc:                 # ✅ gRPC protocol
    port: 8000
```
→ Works because it uses gRPC Health Checking Protocol

### Code Implementation
```go
import (
    "google.golang.org/grpc/health"
    healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

// Register health server
healthServer := health.NewServer()
healthpb.RegisterHealthServer(grpcServer, healthServer)

// Mark as healthy
healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
```

---

## Files Modified

### Go Code (3 files)
- ✅ `services/event/cmd/main.go`
- ✅ `services/reservation/cmd/main.go`
- ✅ `services/location/cmd/main.go`

### Kubernetes Manifests (3 files)
- ✅ `.k8s/services/event-service.yaml`
- ✅ `.k8s/services/reservation-service.yaml`
- ✅ `.k8s/services/location-service.yaml`

### New Documentation
- 📄 `.k8s/GRPC_HEALTH_CHECK_GUIDE.md` (detailed guide)
- 📄 `GRPC_HEALTH_QUICKSTART.md` (this file)

### Scripts
- 🔧 `.k8s/redeploy-grpc-services.sh` (automated deployment)

---

## Next Steps

1. **Deploy the changes:**
   ```bash
   cd aconcert-microservice/.k8s
   ./redeploy-grpc-services.sh
   ```

2. **Monitor for 5-10 minutes:**
   ```bash
   kubectl get pods -n aconcert -w
   ```

3. **Check logs if needed:**
   ```bash
   kubectl logs -l app=event-service -n aconcert --tail=50 -f
   ```

4. **Verify no restarts:**
   ```bash
   kubectl get pods -n aconcert
   # RESTARTS column should be 0 or very low
   ```

---

## More Information

- **Detailed Guide:** `.k8s/GRPC_HEALTH_CHECK_GUIDE.md`
- **Kubernetes Docs:** https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#define-a-grpc-liveness-probe
- **gRPC Health Protocol:** https://github.com/grpc/grpc/blob/master/doc/health-checking.md

---

## Summary

✅ **Added:** gRPC health check servers to event, reservation, location services
✅ **Updated:** Kubernetes probes from HTTP to gRPC
✅ **Result:** Pods should now stay running without unexpected shutdowns

**Deploy command:**
```bash
cd aconcert-microservice/.k8s && ./redeploy-grpc-services.sh
```
