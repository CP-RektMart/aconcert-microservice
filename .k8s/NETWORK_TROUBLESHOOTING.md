# Kubernetes Network Troubleshooting Guide

## Quick Diagnostic Commands

### 1. Check Overall Cluster Status

```bash
# Check all pods in namespace
kubectl get pods -n aconcert -o wide

# Check all services
kubectl get svc -n aconcert

# Check endpoints (should match pod IPs)
kubectl get endpoints -n aconcert
```

### 2. Check Specific Pod Issues

```bash
# Get pod details
kubectl describe pod <POD_NAME> -n aconcert

# Check logs
kubectl logs <POD_NAME> -n aconcert

# Check previous crash logs
kubectl logs <POD_NAME> -n aconcert --previous

# Check init container logs
kubectl logs <POD_NAME> -c <INIT_CONTAINER_NAME> -n aconcert
```

### 3. Test Network Connectivity

```bash
# Create a test pod for network debugging
kubectl run nettest --image=busybox:1.35 -n aconcert --rm -it -- sh

# Inside the test pod, run:
# Test DNS
nslookup kubernetes.default.svc.cluster.local
nslookup event-service.aconcert.svc.cluster.local

# Test connectivity to databases
nc -zv event-postgres.aconcert.svc.cluster.local 5432
nc -zv rabbitmq.aconcert.svc.cluster.local 5672
nc -zv location-mongo.aconcert.svc.cluster.local 27017

# Test connectivity to services
nc -zv event-service.aconcert.svc.cluster.local 8000
```

---

## Common Network Issues & Solutions

### Issue 1: Init Containers Stuck on "waiting for postgres/rabbitmq"

**Symptoms:**
```
waiting for postgres
waiting for postgres
waiting for postgres
```

**Root Causes:**
1. Database pod not running
2. Service not created
3. Service selector doesn't match pod labels
4. DNS not working

**Solutions:**

```bash
# Check if database pod is running
kubectl get pods -n aconcert | grep postgres

# Check if service exists
kubectl get svc -n aconcert | grep postgres

# Check service endpoints (should show pod IP)
kubectl get endpoints -n aconcert | grep postgres

# If endpoint is empty, check service selector vs pod labels
kubectl get svc event-postgres -n aconcert -o yaml | grep selector
kubectl get pod <POSTGRES_POD> -n aconcert -o yaml | grep labels -A 5

# Check DNS resolution from init container
kubectl logs <POD_NAME> -c wait-for-postgres -n aconcert
```

**Fix:**
```bash
# If database pod is not running, check why
kubectl describe pod <POSTGRES_POD> -n aconcert

# If service selector is wrong, fix the service YAML and reapply
kubectl apply -f .k8s/databases/postgres.yaml

# If DNS is broken, check CoreDNS
kubectl get pods -n kube-system | grep coredns
kubectl logs -n kube-system -l k8s-app=kube-dns
```

---

### Issue 2: Service Endpoints Empty

**Symptoms:**
```bash
$ kubectl get endpoints event-service -n aconcert
NAME            ENDPOINTS   AGE
event-service   <none>      5m
```

**Root Cause:** Service selector doesn't match pod labels

**Solution:**

```bash
# Check service selector
kubectl get svc event-service -n aconcert -o jsonpath='{.spec.selector}'

# Check pod labels
kubectl get pods -n aconcert -l app=event-service --show-labels

# If they don't match, fix the deployment or service YAML
# Ensure deployment has: labels.app = event-service
# Ensure service has: selector.app = event-service

# Fix and reapply
kubectl apply -f .k8s/services/event-service.yaml
```

---

### Issue 3: DNS Resolution Fails

**Symptoms:**
```
nslookup: can't resolve 'event-service.aconcert.svc.cluster.local'
```

**Root Cause:** CoreDNS not working or misconfigured

**Solution:**

```bash
# Check CoreDNS pods
kubectl get pods -n kube-system -l k8s-app=kube-dns

# Check CoreDNS logs for errors
kubectl logs -n kube-system -l k8s-app=kube-dns --tail=50

# Check CoreDNS service
kubectl get svc -n kube-system kube-dns

# Restart CoreDNS if needed
kubectl rollout restart deployment/coredns -n kube-system

# Verify DNS service IP is configured in pods
kubectl exec <POD_NAME> -n aconcert -- cat /etc/resolv.conf
```

---

### Issue 4: Connection Refused

**Symptoms:**
```
dial tcp 10.x.x.x:5432: connect: connection refused
```

**Root Causes:**
1. Service is not listening on expected port
2. Pod not ready yet
3. Firewall/network policy blocking

**Solution:**

```bash
# Check if pod is ready
kubectl get pods -n aconcert

# Check if service is listening on correct port
kubectl exec <POD_NAME> -n aconcert -- netstat -tlnp

# For databases, verify port in deployment
kubectl get deployment event-postgres -n aconcert -o yaml | grep -A 3 containerPort

# Check if network policies exist
kubectl get networkpolicies -n aconcert

# Test direct pod-to-pod connectivity (bypass service)
POD_IP=$(kubectl get pod <TARGET_POD> -n aconcert -o jsonpath='{.status.podIP}')
kubectl exec <SOURCE_POD> -n aconcert -- nc -zv $POD_IP 5432
```

---

### Issue 5: Pods Can't Reach External Services

**Symptoms:**
```
dial tcp: lookup api.external.com: no such host
```

**Root Cause:** DNS can't resolve external names

**Solution:**

```bash
# Test external DNS from pod
kubectl exec <POD_NAME> -n aconcert -- nslookup google.com

# Check if CoreDNS can forward to upstream DNS
kubectl get configmap coredns -n kube-system -o yaml

# Should have a forward section like:
# forward . /etc/resolv.conf

# If broken, edit CoreDNS config
kubectl edit configmap coredns -n kube-system
kubectl rollout restart deployment/coredns -n kube-system
```

---

## Automated Diagnostics Script

Run the comprehensive network debugging script:

```bash
cd aconcert-microservice/.k8s
./debug-network.sh
```

This script will:
- ✅ Check namespace and pod status
- ✅ Verify service and endpoint configuration
- ✅ Test DNS resolution
- ✅ Test connectivity to all databases and services
- ✅ Check network policies
- ✅ Verify CoreDNS health
- ✅ Analyze pod logs for network errors

---

## Service-Specific Network Requirements

### Event Service
- **Needs:** event-postgres:5432, event-redis:6379, rabbitmq:5672
- **Listens on:** 8000 (gRPC)

### Reservation Service
- **Needs:** reservation-postgres:5432, reservation-redis:6379
- **Listens on:** 8000 (gRPC)

### Location Service
- **Needs:** location-mongo:27017, location-redis:6379
- **Listens on:** 8000 (gRPC)

### Auth Service
- **Needs:** auth-postgres:5432, auth-redis:6379
- **Listens on:** 8000 (HTTP)

### Gateway Service
- **Needs:** auth-service:8000, event-service:8000, location-service:8000, reservation-service:8000
- **Listens on:** 8000 (HTTP)

---

## DNS Naming Convention

Services in Kubernetes can be accessed using:

1. **Short name** (same namespace): `service-name`
2. **Namespace qualified**: `service-name.namespace`
3. **FQDN**: `service-name.namespace.svc.cluster.local`

**Example:**
```bash
# All of these should work from within aconcert namespace:
nc -zv event-postgres 5432
nc -zv event-postgres.aconcert 5432
nc -zv event-postgres.aconcert.svc.cluster.local 5432
```

---

## Checking Service Connectivity Matrix

Create a connectivity test between all services:

```bash
#!/bin/bash
# Save as test-connectivity.sh

NAMESPACE="aconcert"
SERVICES=("event-postgres:5432" "reservation-postgres:5432" "auth-postgres:5432"
          "rabbitmq:5672" "location-mongo:27017"
          "event-redis:6379" "reservation-redis:6379" "auth-redis:6379")

echo "Testing connectivity from a test pod..."
kubectl run nettest --image=busybox:1.35 -n $NAMESPACE --rm -it -- sh -c "
for svc in ${SERVICES[@]}; do
  IFS=':' read -r host port <<< \$svc
  echo -n \"Testing \$host:\$port... \"
  if nc -zw2 \$host.\$NAMESPACE.svc.cluster.local \$port 2>/dev/null; then
    echo 'OK'
  else
    echo 'FAILED'
  fi
done
"
```

---

## Network Policy Debugging

If network policies are blocking traffic:

```bash
# List all network policies
kubectl get networkpolicies -n aconcert

# Describe a specific policy
kubectl describe networkpolicy <POLICY_NAME> -n aconcert

# Temporarily delete all policies (TESTING ONLY!)
kubectl delete networkpolicies --all -n aconcert

# Test connectivity without policies
# If it works, the issue was network policies
```

---

## CNI Plugin Issues

Check which CNI plugin is being used:

```bash
# Check for common CNI plugins
kubectl get pods -n kube-system | grep -iE "calico|flannel|weave|cilium|canal"

# Check CNI plugin logs
kubectl logs -n kube-system <CNI_POD_NAME>

# For Calico
kubectl get pods -n kube-system -l k8s-app=calico-node

# For Flannel
kubectl get pods -n kube-system -l app=flannel
```

---

## Port Forwarding for Local Testing

Test services locally without entering the cluster:

```bash
# Forward database port
kubectl port-forward svc/event-postgres 5432:5432 -n aconcert

# Forward application service
kubectl port-forward svc/event-service 8000:8000 -n aconcert

# Now test from your local machine
nc -zv localhost 5432
grpcurl -plaintext localhost:8000 grpc.health.v1.Health/Check
```

---

## Emergency Network Debug Pod

Deploy a feature-rich debug pod:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: network-debug
  namespace: aconcert
spec:
  containers:
  - name: debug
    image: nicolaka/netshoot
    command: ["sleep", "3600"]
```

Deploy it:
```bash
kubectl apply -f - <<EOF
apiVersion: v1
kind: Pod
metadata:
  name: network-debug
  namespace: aconcert
spec:
  containers:
  - name: debug
    image: nicolaka/netshoot
    command: ["sleep", "3600"]
EOF

# Shell into it
kubectl exec -it network-debug -n aconcert -- bash

# Now you have access to: ping, traceroute, nslookup, dig, curl, etc.
```

---

## Common Fixes Summary

| Issue | Quick Fix |
|-------|-----------|
| Init container stuck | Check if target pod/service exists |
| DNS not working | Restart CoreDNS: `kubectl rollout restart deployment/coredns -n kube-system` |
| Service has no endpoints | Check service selector matches pod labels |
| Connection refused | Verify pod is running and listening on correct port |
| Can't reach service | Check if service exists: `kubectl get svc -n aconcert` |
| Network policy blocking | List policies: `kubectl get networkpolicies -n aconcert` |

---

## Step-by-Step Debugging Process

1. **Check pod status**
   ```bash
   kubectl get pods -n aconcert
   ```

2. **Check pod logs**
   ```bash
   kubectl logs <POD_NAME> -n aconcert
   ```

3. **Check services exist**
   ```bash
   kubectl get svc -n aconcert
   ```

4. **Check service endpoints**
   ```bash
   kubectl get endpoints -n aconcert
   ```

5. **Test DNS from pod**
   ```bash
   kubectl exec <POD_NAME> -n aconcert -- nslookup <SERVICE_NAME>
   ```

6. **Test connectivity from pod**
   ```bash
   kubectl exec <POD_NAME> -n aconcert -- nc -zv <SERVICE_NAME> <PORT>
   ```

7. **Check network policies**
   ```bash
   kubectl get networkpolicies -n aconcert
   ```

8. **Check CoreDNS**
   ```bash
   kubectl get pods -n kube-system -l k8s-app=kube-dns
   ```

---

## Additional Resources

- [Kubernetes Network Troubleshooting](https://kubernetes.io/docs/tasks/debug/debug-application/debug-service/)
- [DNS Debugging](https://kubernetes.io/docs/tasks/administer-cluster/dns-debugging-resolution/)
- [Debug Services](https://kubernetes.io/docs/tasks/debug/debug-application/debug-service/)

---

## Get Help

If network issues persist:

1. Run automated diagnostics: `./debug-network.sh`
2. Collect all relevant logs
3. Check Kubernetes cluster version compatibility
4. Verify CNI plugin is working correctly
5. Check if this is a local (minikube/kind) or cloud cluster issue
