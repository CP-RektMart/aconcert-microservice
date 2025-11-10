# AConcert K8s - Command Reference Card

## Quick Commands

### Verification
```bash
./verify-setup.sh              # Verify all setup requirements
```

### Build
```bash
make build                     # Build all Docker images
make build SERVICE=auth        # Build specific service (not implemented)
```

### Deploy
```bash
make deploy                    # Full automated deployment
make deploy-databases          # Deploy only databases
make deploy-services           # Deploy only services
./deploy.sh                    # Same as make deploy
```

### Monitor
```bash
make status                    # Show all resources
kubectl get pods -n aconcert   # List all pods
kubectl get svc -n aconcert    # List all services
kubectl get pvc -n aconcert    # List persistent volumes
```

### Logs
```bash
make logs SERVICE=gateway      # Follow gateway logs
kubectl logs -f deployment/auth-service -n aconcert
kubectl logs <pod-name> -n aconcert --previous  # Previous instance
```

### Management
```bash
make restart SERVICE=gateway   # Restart a service
make scale SERVICE=gateway REPLICAS=2  # Scale service
kubectl rollout status deployment/gateway -n aconcert
```

### Cleanup
```bash
make clean                     # Interactive cleanup
kubectl delete namespace aconcert  # Force delete everything
```

## Detailed Commands

### Pod Management
```bash
# Get pod details
kubectl describe pod <pod-name> -n aconcert

# Execute command in pod
kubectl exec -it <pod-name> -n aconcert -- sh

# Port forward
kubectl port-forward -n aconcert svc/gateway 8000:8000
```

### Debugging
```bash
# Check events
kubectl get events -n aconcert --sort-by='.lastTimestamp'

# Check resource usage
kubectl top pods -n aconcert
kubectl top nodes

# View init container logs
kubectl logs <pod-name> -c wait-for-postgres -n aconcert
```

### ConfigMap & Secrets
```bash
# Edit ConfigMap
kubectl edit configmap gateway-config -n aconcert

# View ConfigMap
kubectl get configmap gateway-config -n aconcert -o yaml

# Restart after config change
kubectl rollout restart deployment/gateway -n aconcert
```

### Docker Commands
```bash
# List images
docker images | grep aconcert

# Remove all aconcert images
docker rmi $(docker images | grep aconcert | awk '{print $3}')

# Build single service
docker build -t aconcert/auth-service:latest \
  -f services/auth/Dockerfile .

# Tag image
docker tag aconcert/gateway:latest aconcert/gateway:v1.0.0
```

### Network Debugging
```bash
# Test connectivity from debug pod
kubectl run -it --rm debug --image=busybox \
  --restart=Never -n aconcert -- sh

# Inside pod:
nc -zv auth-postgres.aconcert.svc.cluster.local 5432
nc -zv auth-service.aconcert.svc.cluster.local 8000
```

### Rollout Management
```bash
# View rollout history
kubectl rollout history deployment/gateway -n aconcert

# Rollback to previous version
kubectl rollout undo deployment/gateway -n aconcert

# Rollback to specific revision
kubectl rollout undo deployment/gateway --to-revision=2 -n aconcert
```

## Service URLs

### External (NodePort)
```bash
# Gateway API
curl http://localhost:30000/health

# Realtime WebSocket
wscat -c ws://localhost:30001
# or
curl http://localhost:30001/health
```

### Internal (from within cluster)
```
auth-service.aconcert.svc.cluster.local:8000
event-service.aconcert.svc.cluster.local:8000
location-service.aconcert.svc.cluster.local:8000
reservation-service.aconcert.svc.cluster.local:8000
payment-service.aconcert.svc.cluster.local:8000
realtime-service.aconcert.svc.cluster.local:8000
```

## Troubleshooting Commands

### Pod Stuck in Pending
```bash
kubectl describe pod <pod-name> -n aconcert
kubectl get events -n aconcert | grep <pod-name>
```

### Pod in CrashLoopBackOff
```bash
kubectl logs <pod-name> -n aconcert
kubectl logs <pod-name> -n aconcert --previous
kubectl describe pod <pod-name> -n aconcert
```

### Image Pull Errors
```bash
# For Docker Desktop - images should be available
docker images | grep aconcert

# For Minikube
eval $(minikube docker-env)
make build

# For Kind
kind load docker-image aconcert/gateway:latest --name <cluster-name>
```

### Service Not Accessible
```bash
# Check service
kubectl get svc gateway -n aconcert

# Check endpoints
kubectl get endpoints gateway -n aconcert

# Port forward to test
kubectl port-forward svc/gateway 8000:8000 -n aconcert
curl http://localhost:8000/health
```

## Make Targets

```bash
make help              # Show all available commands
make build             # Build all Docker images
make deploy            # Full deployment
make deploy-databases  # Deploy databases only
make deploy-services   # Deploy services only
make deploy-gateway    # Deploy gateway only
make status            # Show deployment status
make logs SERVICE=name # View service logs
make restart SERVICE=name  # Restart service
make scale SERVICE=name REPLICAS=n  # Scale service
make clean             # Cleanup everything
```

## One-Liners

```bash
# Quick full deploy
cd .k8s && ./verify-setup.sh && make build && make deploy

# Watch all pods
watch kubectl get pods -n aconcert

# Get all pod IPs
kubectl get pods -n aconcert -o wide

# Delete and redeploy gateway
kubectl delete -f services/gateway.yaml && kubectl apply -f services/gateway.yaml

# Restart all services
for dep in $(kubectl get deployments -n aconcert -o name); do
  kubectl rollout restart $dep -n aconcert
done

# Check all pod logs
for pod in $(kubectl get pods -n aconcert -o name); do
  echo "=== $pod ===" && kubectl logs $pod -n aconcert --tail=5
done
```

## Useful Aliases

Add to your `~/.bashrc` or `~/.zshrc`:

```bash
alias k='kubectl'
alias ka='kubectl -n aconcert'
alias kgp='kubectl get pods -n aconcert'
alias kgs='kubectl get svc -n aconcert'
alias kl='kubectl logs -f -n aconcert'
alias kd='kubectl describe -n aconcert'
alias ke='kubectl exec -it -n aconcert'
```

## Emergency Procedures

### Complete Reset
```bash
cd .k8s
make clean  # or
kubectl delete namespace aconcert
kubectl delete pvc --all -n aconcert
```

### Force Delete Stuck Resources
```bash
kubectl delete pod <pod-name> -n aconcert --force --grace-period=0
kubectl patch pvc <pvc-name> -n aconcert -p '{"metadata":{"finalizers":null}}'
```

### Check Cluster Health
```bash
kubectl cluster-info
kubectl get nodes
kubectl get componentstatuses
```

---

💡 **Tip**: Keep this reference handy! Bookmark it or print it out.
